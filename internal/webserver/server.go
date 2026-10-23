// Package webserver exposes the WARPSCOUT core (the same JSON protocol the
// Android app uses through mobileapi) over HTTP, together with a web UI.
package webserver

import (
	"crypto/subtle"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vernette/warpscout/core"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

//go:embed static
var staticFiles embed.FS

// Listener receives JSON events from the core.
type Listener interface{ OnEvent(eventJSON string) }

// Runner is the core entry point. mobileapi.Start/Cancel satisfy it.
type Runner interface {
	Start(requestJSON string, listener Listener) error
	Cancel()
}

type Config struct {
	Runner  Runner
	Store   *Store
	Token   string // empty disables authentication
	Version string
	// TrustProxy makes the rate limiter use X-Forwarded-For. Enable it only
	// behind a reverse proxy that sets the header.
	TrustProxy bool
}

type Server struct {
	cfg Config

	mu     sync.Mutex
	cur    *job
	nextID int64

	auth *authState
}

func New(cfg Config) *Server {
	return &Server{cfg: cfg, auth: newAuthState()}
}

const cookieName = "warpscout_session"

var runOperations = map[string]bool{
	"register": true, "scan": true, "find-junk": true, "find-sni": true, "socks": true,
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.handleLogout)
	mux.HandleFunc("GET /api/state", s.requireAuth(s.handleState))
	mux.HandleFunc("POST /api/run", s.requireAuth(s.handleRun))
	mux.HandleFunc("POST /api/cancel", s.requireAuth(s.handleCancel))
	mux.HandleFunc("GET /api/events", s.requireAuth(s.handleEvents))
	mux.HandleFunc("POST /api/render", s.requireAuth(s.handleRender))
	mux.HandleFunc("GET /api/account", s.requireAuth(s.handleAccountGet))
	mux.HandleFunc("PUT /api/account", s.requireAuth(s.handleAccountPut))
	mux.HandleFunc("DELETE /api/account", s.requireAuth(s.handleAccountDelete))
	mux.HandleFunc("POST /api/account/export", s.requireAuth(s.handleAccountExport))
	mux.HandleFunc("GET /api/tools", s.requireAuth(s.handleTools))
	mux.HandleFunc("GET /api/history", s.requireAuth(s.handleHistoryList))
	mux.HandleFunc("DELETE /api/history", s.requireAuth(s.handleHistoryClear))
	mux.HandleFunc("GET /api/history/{id}", s.requireAuth(s.handleHistoryGet))
	mux.HandleFunc("DELETE /api/history/{id}", s.requireAuth(s.handleHistoryDelete))

	sub, _ := fs.Sub(staticFiles, "static")
	mux.Handle("GET /", http.FileServerFS(sub))
	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; img-src 'self' data:; frame-ancestors 'none'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

// ---- auth -----------------------------------------------------------------

func (s *Server) ip(r *http.Request) string { return clientIP(r, s.cfg.TrustProxy) }

// checkToken validates a presented token and feeds the per-IP lockout.
// The bool reports whether the IP is currently locked out.
func (s *Server) checkToken(r *http.Request, presented string) (ok, locked bool) {
	ip := s.ip(r)
	if _, l := s.auth.locked(ip); l {
		return false, true
	}
	if subtle.ConstantTimeCompare([]byte(presented), []byte(s.cfg.Token)) == 1 {
		s.auth.succeed(ip)
		return true, false
	}
	s.auth.fail(ip)
	return false, false
}

func (s *Server) authorized(w http.ResponseWriter, r *http.Request) bool {
	if s.cfg.Token == "" {
		return true
	}
	if c, err := r.Cookie(cookieName); err == nil && s.auth.validSession(c.Value) {
		return true
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		ok, locked := s.checkToken(r, strings.TrimPrefix(h, "Bearer "))
		if locked {
			s.writeLocked(w, r)
			return false
		}
		return ok
	}
	return false
}

func (s *Server) writeLocked(w http.ResponseWriter, r *http.Request) {
	left, _ := s.auth.locked(s.ip(r))
	w.Header().Set("Retry-After", fmt.Sprint(int(left.Seconds())+1))
	writeError(w, http.StatusTooManyRequests, "too_many_attempts", "too many failed attempts, try again later")
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.authorized(w, r) {
			if w.Header().Get("Retry-After") == "" {
				writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
			}
			return
		}
		// State-changing requests must carry a custom header: a cross-site form
		// cannot set it, which closes the CSRF hole even without SameSite.
		if r.Method != http.MethodGet && r.Header.Get("X-WarpScout") != "1" {
			writeError(w, http.StatusForbidden, "csrf", "missing X-WarpScout header")
			return
		}
		next(w, r)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-WarpScout") != "1" {
		writeError(w, http.StatusForbidden, "csrf", "missing X-WarpScout header")
		return
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := decodeBody(r, &body, 4096); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if s.cfg.Token != "" {
		ok, locked := s.checkToken(r, body.Token)
		if locked {
			s.writeLocked(w, r)
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, "bad_token", "invalid token")
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieName, Value: s.auth.newSession(), Path: "/", HttpOnly: true,
		SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		MaxAge: int(sessionTTL.Seconds()),
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-WarpScout") != "1" {
		writeError(w, http.StatusForbidden, "csrf", "missing X-WarpScout header")
		return
	}
	if c, err := r.Cookie(cookieName); err == nil {
		s.auth.dropSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---- state / run / cancel -------------------------------------------------

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	_, hasAccount := s.cfg.Store.Account()
	state := map[string]any{"version": s.cfg.Version, "hasAccount": hasAccount, "running": false, "authRequired": s.cfg.Token != ""}
	s.mu.Lock()
	cur := s.cur
	s.mu.Unlock()
	if cur != nil {
		state["job"] = map[string]any{
			"id": cur.id, "operation": cur.operation, "startedAt": cur.started,
			"running": !cur.isDone(), "seq": cur.seq(),
		}
		state["running"] = !cur.isDone()
	}
	writeJSON(w, http.StatusOK, state)
}

type socksRequest struct {
	Endpoint string        `json:"endpoint"`
	Port     int           `json:"port"`
	Protocol core.Protocol `json:"protocol"`
}

// runRequest is what the browser sends. Scan options come as a ScanSpec
// (preset + expert form) and are resolved here, so the preset rules live in
// one place (core.ScanSpec.Resolve).
type runRequest struct {
	Operation string          `json:"operation"`
	Spec      *core.ScanSpec  `json:"spec,omitempty"`
	Socks     *socksRequest   `json:"socks,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"` // register / render only
}

func (in runRequest) spec() core.ScanSpec {
	if in.Spec != nil {
		return *in.Spec
	}
	return core.ScanSpec{}
}

func buildPayload(in runRequest) (json.RawMessage, *core.ScanOptions, error) {
	switch in.Operation {
	case "register":
		if len(in.Payload) == 0 {
			return json.RawMessage("{}"), nil, nil
		}
		return in.Payload, nil, nil
	case "scan":
		opts := in.spec().Resolve()
		data, err := json.Marshal(opts)
		return data, &opts, err
	case "find-junk":
		opts := in.spec().Resolve()
		opts.Protocol = core.ProtocolAWG
		data, err := json.Marshal(opts)
		return data, &opts, err
	case "find-sni":
		opts := in.spec().Resolve()
		if opts.Protocol != core.ProtocolMASQUEH3 && opts.Protocol != core.ProtocolMASQUEH2 {
			opts.Protocol = core.ProtocolMASQUEH3
		}
		data, err := json.Marshal(opts)
		return data, &opts, err
	case "socks":
		if in.Socks == nil || strings.TrimSpace(in.Socks.Endpoint) == "" {
			return nil, nil, errors.New("socks.endpoint is required")
		}
		port := in.Socks.Port
		if port == 0 {
			port = 1080
		}
		if port < 1 || port > 65535 {
			return nil, nil, errors.New("socks.port must be between 1 and 65535")
		}
		opts := in.spec().Resolve()
		protocol := in.Socks.Protocol
		if protocol == "" {
			protocol = opts.Protocol
		}
		opts.Protocol = protocol
		data, err := json.Marshal(core.SocksOptions{Endpoint: strings.TrimSpace(in.Socks.Endpoint), Port: port, Protocol: protocol, Scan: opts})
		return data, &opts, err
	}
	return nil, nil, errors.New("unknown operation")
}

type coreRequest struct {
	SchemaVersion int             `json:"schemaVersion"`
	Operation     string          `json:"operation"`
	AccountJSON   string          `json:"accountJson,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

type funcListener func(string)

func (f funcListener) OnEvent(e string) { f(e) }

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	var in runRequest
	if err := decodeBody(r, &in, 1<<20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if !runOperations[in.Operation] {
		writeError(w, http.StatusBadRequest, "unknown_operation", "unknown operation")
		return
	}
	payload, resolved, err := buildPayload(in)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	account, hasAccount := s.cfg.Store.Account()
	if in.Operation != "register" && !hasAccount {
		writeError(w, http.StatusConflict, "account_required", "register or import a WARP account first")
		return
	}
	req := coreRequest{SchemaVersion: 1, Operation: in.Operation, Payload: payload}
	if in.Operation != "register" {
		req.AccountJSON = account
	}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	s.mu.Lock()
	if s.cur != nil && !s.cur.isDone() {
		s.mu.Unlock()
		writeError(w, http.StatusConflict, "operation_active", "another operation is already running")
		return
	}
	s.nextID++
	j := newJob(s.nextID, in.Operation)
	s.cur = j
	s.mu.Unlock()

	go func() {
		defer j.finish()
		defer func() {
			if rec := recover(); rec != nil {
				j.add(errorEvent("core_panic", fmt.Sprint(rec)))
			}
		}()
		err := s.cfg.Runner.Start(string(reqJSON), funcListener(func(e string) {
			j.add(e)
			s.afterEvent(in, resolved, e)
		}))
		if err != nil && !j.hasTerminal() {
			j.add(errorEvent("operation_failed", err.Error()))
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]any{"jobId": j.id})
}

// afterEvent persists results that the UI should not lose on reload.
func (s *Server) afterEvent(in runRequest, resolved *core.ScanOptions, event string) {
	var ev struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if json.Unmarshal([]byte(event), &ev) != nil || ev.Type != "completed" {
		return
	}
	switch in.Operation {
	case "register":
		var acc struct {
			RawJSON string `json:"rawJson"`
		}
		if json.Unmarshal(ev.Payload, &acc) == nil && strings.TrimSpace(acc.RawJSON) != "" {
			_ = s.cfg.Store.SaveAccount(acc.RawJSON)
		}
	case "scan":
		spec := in.spec()
		specJSON, _ := json.Marshal(spec)
		var optsJSON json.RawMessage
		if resolved != nil {
			optsJSON, _ = json.Marshal(resolved)
		}
		_, _ = s.cfg.Store.AddHistory(string(spec.Preset), specJSON, optsJSON, ev.Payload)
	case "find-junk", "find-sni":
		_ = s.cfg.Store.SaveTool(in.Operation, ev.Payload)
	}
}

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.cfg.Store.Tools())
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	s.cfg.Runner.Cancel()
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	j := s.cur
	s.mu.Unlock()
	if j == nil {
		writeError(w, http.StatusNotFound, "no_job", "no operation has been started")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "no_stream", "streaming unsupported")
		return
	}
	after := 0
	raw := r.URL.Query().Get("after")
	if raw == "" {
		raw = r.Header.Get("Last-Event-ID")
	}
	fmt.Sscanf(raw, "%d", &after)

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "retry: 2000\nevent: job\ndata: {\"jobId\":%d,\"operation\":%q}\n\n", j.id, j.operation)
	flusher.Flush()

	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()
	for {
		events, first, done, wait := j.since(after)
		for i, e := range events {
			seq := first + i
			fmt.Fprintf(w, "id: %d\ndata: %s\n\n", seq, e)
			after = seq
		}
		if len(events) > 0 {
			flusher.Flush()
		}
		if done && len(events) == 0 {
			fmt.Fprint(w, "event: end\ndata: {}\n\n")
			flusher.Flush()
			return
		}
		if len(events) > 0 {
			continue
		}
		select {
		case <-wait:
		case <-keepalive.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// ---- render ---------------------------------------------------------------

type collector struct {
	mu     sync.Mutex
	events []string
}

func (c *collector) OnEvent(e string) {
	c.mu.Lock()
	c.events = append(c.events, e)
	c.mu.Unlock()
}

func (s *Server) handleRender(w http.ResponseWriter, r *http.Request) {
	var in runRequest
	if err := decodeBody(r, &in, 4<<20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if in.Operation != "render-report" && in.Operation != "render-config" {
		writeError(w, http.StatusBadRequest, "unknown_operation", "unknown operation")
		return
	}
	// Rendering shares the core's backend lock; a running scan would block it.
	s.mu.Lock()
	busy := s.cur != nil && !s.cur.isDone()
	s.mu.Unlock()
	if busy {
		writeError(w, http.StatusConflict, "operation_active", "wait for the running operation to finish")
		return
	}
	payload := in.Payload
	if in.Operation == "render-config" {
		merged := map[string]json.RawMessage{}
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &merged); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_request", "payload must be an object")
				return
			}
		}
		opts, _ := json.Marshal(in.spec().Resolve())
		merged["options"] = opts
		payload, _ = json.Marshal(merged)
	}
	req := coreRequest{SchemaVersion: 1, Operation: in.Operation, Payload: payload}
	if in.Operation == "render-config" {
		account, ok := s.cfg.Store.Account()
		if !ok {
			writeError(w, http.StatusConflict, "account_required", "register or import a WARP account first")
			return
		}
		req.AccountJSON = account
	}
	reqJSON, _ := json.Marshal(req)
	var c collector
	_ = s.cfg.Runner.Start(string(reqJSON), &c)
	for _, e := range c.events {
		var ev struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
			Error   *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal([]byte(e), &ev) != nil {
			continue
		}
		switch ev.Type {
		case "completed":
			writeJSON(w, http.StatusOK, map[string]json.RawMessage{"result": ev.Payload})
			return
		case "error":
			if ev.Error != nil {
				writeError(w, http.StatusBadRequest, ev.Error.Code, ev.Error.Message)
				return
			}
		}
	}
	writeError(w, http.StatusInternalServerError, "no_result", "the core returned no result")
}

// ---- account --------------------------------------------------------------

func (s *Server) handleAccountGet(w http.ResponseWriter, r *http.Request) {
	_, ok := s.cfg.Store.Account()
	writeJSON(w, http.StatusOK, map[string]bool{"registered": ok})
}

func (s *Server) handleAccountPut(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RawJSON string `json:"rawJson"`
	}
	if err := decodeBody(r, &body, 1<<20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	var probe struct {
		PrivateKey    string `json:"private_key"`
		PeerPublicKey string `json:"peer_public_key"`
	}
	if !json.Valid([]byte(body.RawJSON)) || json.Unmarshal([]byte(body.RawJSON), &probe) != nil {
		writeError(w, http.StatusBadRequest, "invalid_account", "account JSON is invalid")
		return
	}
	if err := s.cfg.Store.SaveAccount(body.RawJSON); err != nil {
		writeError(w, http.StatusInternalServerError, "save_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"registered": true})
}

func (s *Server) handleAccountDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.cfg.Store.DeleteAccount(); err != nil {
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"registered": false})
}

// handleAccountExport returns the private account file. It asks for the
// access token again, so a stolen session cookie alone cannot take the key.
func (s *Server) handleAccountExport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := decodeBody(r, &body, 4096); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if s.cfg.Token != "" {
		ok, locked := s.checkToken(r, body.Token)
		if locked {
			s.writeLocked(w, r)
			return
		}
		if !ok {
			log.Printf("account export denied for %s: wrong token", s.ip(r))
			writeError(w, http.StatusForbidden, "bad_token", "invalid token")
			return
		}
	}
	raw, ok := s.cfg.Store.Account()
	if !ok {
		writeError(w, http.StatusNotFound, "no_account", "no account")
		return
	}
	log.Printf("account exported by %s", s.ip(r))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="warpscout-account.json"`)
	io.WriteString(w, raw)
}

// ---- history --------------------------------------------------------------

func (s *Server) handleHistoryList(w http.ResponseWriter, r *http.Request) {
	list, err := s.cfg.Store.History()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "history_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (s *Server) handleHistoryGet(w http.ResponseWriter, r *http.Request) {
	data, err := s.cfg.Store.HistoryEntry(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "history entry not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleHistoryDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.cfg.Store.DeleteHistory(r.PathValue("id")); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "not_found", "history entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleHistoryClear(w http.ResponseWriter, r *http.Request) {
	if err := s.cfg.Store.ClearHistory(); err != nil {
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---- helpers --------------------------------------------------------------

func decodeBody(r *http.Request, v any, limit int64) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, limit))
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func errorEvent(code, message string) string {
	data, _ := json.Marshal(map[string]any{
		"schemaVersion": 1, "type": "error",
		"error": map[string]any{"code": code, "message": message, "retryable": false},
	})
	return string(data)
}
