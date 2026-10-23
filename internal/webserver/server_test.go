package webserver

import (
	"bufio"
	"encoding/json"
	"github.com/vernette/warpscout/core"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func newTestServer(t *testing.T, token string) (*httptest.Server, *Store) {
	t.Helper()
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	srv := New(Config{Runner: NewDemoRunner(), Store: store, Token: token, Version: "test"})
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, store
}

func do(t *testing.T, method, url, token, body string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("X-WarpScout", "1")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var sb strings.Builder
	sc := bufio.NewScanner(res.Body)
	for sc.Scan() {
		sb.WriteString(sc.Text() + "\n")
	}
	return res, sb.String()
}

func TestAuthRequired(t *testing.T) {
	ts, _ := newTestServer(t, "secret")
	if res, _ := do(t, "GET", ts.URL+"/api/state", "", ""); res.StatusCode != 401 {
		t.Fatalf("want 401, got %d", res.StatusCode)
	}
	if res, _ := do(t, "GET", ts.URL+"/api/state", "wrong", ""); res.StatusCode != 401 {
		t.Fatalf("want 401, got %d", res.StatusCode)
	}
	if res, _ := do(t, "GET", ts.URL+"/api/state", "secret", ""); res.StatusCode != 200 {
		t.Fatalf("want 200, got %d", res.StatusCode)
	}
	req, _ := http.NewRequest("POST", ts.URL+"/api/cancel", nil)
	req.Header.Set("Authorization", "Bearer secret")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != 403 {
		t.Fatalf("missing CSRF header must be 403, got %d", res.StatusCode)
	}
}

func TestStaticIndexServed(t *testing.T) {
	ts, _ := newTestServer(t, "")
	res, body := do(t, "GET", ts.URL+"/", "", "")
	if res.StatusCode != 200 || !strings.Contains(body, "WarpScout") {
		t.Fatalf("index not served: %d", res.StatusCode)
	}
}

func TestRegisterScanHistoryFlow(t *testing.T) {
	ts, store := newTestServer(t, "")
	if res, _ := do(t, "POST", ts.URL+"/api/run", "", `{"operation":"scan"}`); res.StatusCode != 409 {
		t.Fatalf("scan without account must be 409, got %d", res.StatusCode)
	}
	if res, _ := do(t, "POST", ts.URL+"/api/run", "", `{"operation":"register","payload":{}}`); res.StatusCode != 202 {
		t.Fatalf("register: %d", res.StatusCode)
	}
	waitEnd(t, ts.URL)
	if _, ok := store.Account(); !ok {
		t.Fatal("account was not persisted after registration")
	}
	if res, _ := do(t, "POST", ts.URL+"/api/run", "", `{"operation":"scan","spec":{"preset":"standard","protocol":"awg"}}`); res.StatusCode != 202 {
		t.Fatalf("scan: %d", res.StatusCode)
	}
	// a second run while active is rejected
	if res, _ := do(t, "POST", ts.URL+"/api/run", "", `{"operation":"scan"}`); res.StatusCode != 409 {
		t.Fatalf("concurrent run must be 409, got %d", res.StatusCode)
	}
	events := waitEnd(t, ts.URL)
	if !strings.Contains(events, `"type":"completed"`) {
		t.Fatalf("no completed event:\n%s", events)
	}
	_, body := do(t, "GET", ts.URL+"/api/history", "", "")
	var list struct{ Items []HistorySummary }
	if err := json.Unmarshal([]byte(body), &list); err != nil || len(list.Items) != 1 || list.Items[0].Working == 0 {
		t.Fatalf("history: %v %s", err, body)
	}
	id := list.Items[0].ID
	if res, _ := do(t, "GET", ts.URL+"/api/history/"+id, "", ""); res.StatusCode != 200 {
		t.Fatal("history entry not readable")
	}
	if res, _ := do(t, "GET", ts.URL+"/api/history/..%2f..%2faccount", "", ""); res.StatusCode != 404 {
		t.Fatalf("path traversal must 404, got %d", res.StatusCode)
	}
	if res, _ := do(t, "DELETE", ts.URL+"/api/history/"+id, "", ""); res.StatusCode != 200 {
		t.Fatal("delete failed")
	}
}

func TestCancel(t *testing.T) {
	ts, store := newTestServer(t, "")
	_ = store.SaveAccount(`{"private_key":"a","peer_public_key":"b"}`)
	do(t, "POST", ts.URL+"/api/run", "", `{"operation":"scan"}`)
	time.Sleep(300 * time.Millisecond)
	do(t, "POST", ts.URL+"/api/cancel", "", `{}`)
	events := waitEnd(t, ts.URL)
	if !strings.Contains(events, `"canceled"`) {
		t.Fatalf("expected cancel error:\n%s", events)
	}
}

func TestRender(t *testing.T) {
	ts, store := newTestServer(t, "")
	if res, _ := do(t, "POST", ts.URL+"/api/render", "", `{"operation":"render-config","spec":{"preset":"standard"},"payload":{}}`); res.StatusCode != 409 {
		t.Fatalf("render-config without account: %d", res.StatusCode)
	}
	_ = store.SaveAccount(`{"private_key":"a","peer_public_key":"b"}`)
	res, body := do(t, "POST", ts.URL+"/api/render", "", `{"operation":"render-config","spec":{"preset":"standard"},"payload":{}}`)
	if res.StatusCode != 200 || !strings.Contains(body, "result") {
		t.Fatalf("render-config: %d %s", res.StatusCode, body)
	}
}

func TestJobResume(t *testing.T) {
	j := newJob(1, "scan")
	for i := 0; i < 3; i++ {
		j.add(`{"type":"progress"}`)
	}
	ev, first, _, _ := j.since(1)
	if len(ev) != 2 || first != 2 {
		t.Fatalf("resume: %d events from %d", len(ev), first)
	}
	for i := 0; i < maxJobEvents+10; i++ {
		j.add(`{"type":"progress"}`)
	}
	if j.seq() != maxJobEvents+13 {
		t.Fatalf("seq after trimming: %d", j.seq())
	}
}

func waitEnd(t *testing.T, base string) string {
	t.Helper()
	res, err := http.Get(base + "/api/events?after=0")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var sb strings.Builder
	sc := bufio.NewScanner(res.Body)
	deadline := time.After(20 * time.Second)
	done := make(chan struct{})
	go func() {
		for sc.Scan() {
			sb.WriteString(sc.Text() + "\n")
			if strings.HasPrefix(sc.Text(), "event: end") {
				close(done)
				return
			}
		}
	}()
	select {
	case <-done:
	case <-deadline:
		t.Fatal("timeout waiting for end event")
	}
	return sb.String()
}

type recordRunner struct {
	*DemoRunner
	mu   sync.Mutex
	reqs []string
}

func (r *recordRunner) Start(req string, l Listener) error {
	r.mu.Lock()
	r.reqs = append(r.reqs, req)
	r.mu.Unlock()
	return r.DemoRunner.Start(req, l)
}

func (r *recordRunner) last() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reqs[len(r.reqs)-1]
}

func newRecordServer(t *testing.T, token string) (*httptest.Server, *Store, *recordRunner) {
	t.Helper()
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rr := &recordRunner{DemoRunner: NewDemoRunner()}
	srv := New(Config{Runner: rr, Store: store, Token: token, Version: "test"})
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, store, rr
}

func TestSpecIsResolvedOnServer(t *testing.T) {
	ts, store, rr := newRecordServer(t, "")
	_ = store.SaveAccount(`{"private_key":"a","peer_public_key":"b"}`)
	do(t, "POST", ts.URL+"/api/run", "", `{"operation":"scan","spec":{"preset":"full","expert":false,"protocol":"wg"}}`)
	waitEnd(t, ts.URL)
	var req struct {
		Payload core.ScanOptions `json:"payload"`
	}
	if err := json.Unmarshal([]byte(rr.last()), &req); err != nil {
		t.Fatal(err)
	}
	p := req.Payload
	if !p.Full || p.SamplePerSubnet != 0 || p.TunnelPingCount != 10 || p.Protocol != core.ProtocolWG {
		t.Fatalf("preset not resolved: %+v", p)
	}
	// history keeps the spec so the form can be restored exactly
	_, body := do(t, "GET", ts.URL+"/api/history", "", "")
	var list struct{ Items []HistorySummary }
	_ = json.Unmarshal([]byte(body), &list)
	_, entry := do(t, "GET", ts.URL+"/api/history/"+list.Items[0].ID, "", "")
	if !strings.Contains(entry, `"spec":{"preset":"full"`) {
		t.Fatalf("spec missing in history: %s", entry)
	}
}

func TestFindToolsForceProtocolAndPersist(t *testing.T) {
	ts, store, rr := newRecordServer(t, "")
	_ = store.SaveAccount(`{"private_key":"a","peer_public_key":"b"}`)
	do(t, "POST", ts.URL+"/api/run", "", `{"operation":"find-junk","spec":{"protocol":"wg"}}`)
	waitEnd(t, ts.URL)
	if !strings.Contains(rr.last(), `"protocol":"awg"`) {
		t.Fatalf("find-junk must force awg: %s", rr.last())
	}
	do(t, "POST", ts.URL+"/api/run", "", `{"operation":"find-sni","spec":{"protocol":"awg"}}`)
	waitEnd(t, ts.URL)
	if !strings.Contains(rr.last(), `"protocol":"masque"`) {
		t.Fatalf("find-sni must use masque: %s", rr.last())
	}
	_, body := do(t, "GET", ts.URL+"/api/tools", "", "")
	if !strings.Contains(body, "find-junk") || !strings.Contains(body, "find-sni") {
		t.Fatalf("tool results not saved: %s", body)
	}
}

func TestSocksPayload(t *testing.T) {
	ts, store, rr := newRecordServer(t, "")
	_ = store.SaveAccount(`{"private_key":"a","peer_public_key":"b"}`)
	if res, _ := do(t, "POST", ts.URL+"/api/run", "", `{"operation":"socks"}`); res.StatusCode != 400 {
		t.Fatalf("socks without endpoint must be 400, got %d", res.StatusCode)
	}
	if res, _ := do(t, "POST", ts.URL+"/api/run", "", `{"operation":"socks","socks":{"endpoint":"1.2.3.4:2408","port":70000}}`); res.StatusCode != 400 {
		t.Fatalf("bad port must be 400, got %d", res.StatusCode)
	}
	do(t, "POST", ts.URL+"/api/run", "", `{"operation":"socks","socks":{"endpoint":"1.2.3.4:2408","protocol":"wg"}}`)
	time.Sleep(400 * time.Millisecond)
	if !strings.Contains(rr.last(), `"port":1080`) || !strings.Contains(rr.last(), `"endpoint":"1.2.3.4:2408"`) {
		t.Fatalf("socks payload: %s", rr.last())
	}
	do(t, "POST", ts.URL+"/api/cancel", "", `{}`)
	waitEnd(t, ts.URL)
}

func loginClient(t *testing.T, base, token string) (*http.Client, *http.Response) {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar}
	req, _ := http.NewRequest("POST", base+"/api/login", strings.NewReader(`{"token":"`+token+`"}`))
	req.Header.Set("X-WarpScout", "1")
	res, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	return c, res
}

func TestSessionLifecycle(t *testing.T) {
	ts, _ := newTestServer(t, "secret")
	c, res := loginClient(t, ts.URL, "secret")
	if res.StatusCode != 200 {
		t.Fatalf("login: %d", res.StatusCode)
	}
	cookie := ""
	for _, ck := range res.Cookies() {
		if ck.Name == cookieName {
			cookie = ck.Value
		}
	}
	if len(cookie) != 64 || cookie == "secret" {
		t.Fatalf("session id must be random, got %q", cookie)
	}
	get := func() int {
		r, _ := c.Get(ts.URL + "/api/state")
		r.Body.Close()
		return r.StatusCode
	}
	if get() != 200 {
		t.Fatal("session cookie not accepted")
	}
	req, _ := http.NewRequest("POST", ts.URL+"/api/logout", nil)
	req.Header.Set("X-WarpScout", "1")
	r, _ := c.Do(req)
	r.Body.Close()
	// the old cookie must now be invalid server-side
	old, _ := http.NewRequest("GET", ts.URL+"/api/state", nil)
	old.AddCookie(&http.Cookie{Name: cookieName, Value: cookie})
	or, _ := http.DefaultClient.Do(old)
	or.Body.Close()
	if or.StatusCode != 401 {
		t.Fatalf("logged-out session still valid: %d", or.StatusCode)
	}
}

func TestSessionExpires(t *testing.T) {
	a := newAuthState()
	now := time.Now()
	a.now = func() time.Time { return now }
	id := a.newSession()
	if !a.validSession(id) {
		t.Fatal("fresh session invalid")
	}
	now = now.Add(sessionTTL + time.Minute)
	if a.validSession(id) {
		t.Fatal("expired session accepted")
	}
}

func TestLoginLockout(t *testing.T) {
	ts, _ := newTestServer(t, "secret")
	for i := 0; i < maxFails; i++ {
		if _, res := loginClient(t, ts.URL, "bad"); res.StatusCode != 401 {
			t.Fatalf("attempt %d: want 401, got %d", i, res.StatusCode)
		}
	}
	_, res := loginClient(t, ts.URL, "secret")
	if res.StatusCode != 429 || res.Header.Get("Retry-After") == "" {
		t.Fatalf("locked IP must get 429 + Retry-After even with the right token, got %d", res.StatusCode)
	}
}

func TestLockoutExpiresAndResets(t *testing.T) {
	a := newAuthState()
	now := time.Now()
	a.now = func() time.Time { return now }
	for i := 0; i < maxFails; i++ {
		a.fail("1.1.1.1")
	}
	if _, l := a.locked("1.1.1.1"); !l {
		t.Fatal("should be locked")
	}
	if _, l := a.locked("2.2.2.2"); l {
		t.Fatal("other IPs must not be locked")
	}
	now = now.Add(lockFor + time.Second)
	if _, l := a.locked("1.1.1.1"); l {
		t.Fatal("lock must expire")
	}
}

func TestClientIPTrustProxy(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:5555"
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 9.9.9.9")
	if got := clientIP(r, false); got != "10.0.0.1" {
		t.Fatalf("untrusted proxy header must be ignored, got %s", got)
	}
	if got := clientIP(r, true); got != "9.9.9.9" {
		t.Fatalf("want rightmost XFF entry, got %s", got)
	}
}

func TestAccountExportNeedsToken(t *testing.T) {
	ts, store := newTestServer(t, "secret")
	_ = store.SaveAccount(`{"private_key":"KEY","peer_public_key":"b"}`)
	c, _ := loginClient(t, ts.URL, "secret")
	post := func(body string) (int, string) {
		req, _ := http.NewRequest("POST", ts.URL+"/api/account/export", strings.NewReader(body))
		req.Header.Set("X-WarpScout", "1")
		res, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return res.StatusCode, string(b)
	}
	if code, body := post(`{"token":""}`); code != 403 || strings.Contains(body, "KEY") {
		t.Fatalf("export without token must be denied: %d %s", code, body)
	}
	if code, body := post(`{"token":"secret"}`); code != 200 || !strings.Contains(body, "KEY") {
		t.Fatalf("export with token: %d %s", code, body)
	}
	// the old GET route is gone
	res, _ := c.Get(ts.URL + "/api/account/export")
	res.Body.Close()
	if res.StatusCode == 200 {
		t.Fatal("GET export must not exist")
	}
}

func TestTelegramSpecReachesCoreAndHistory(t *testing.T) {
	ts, store, rr := newRecordServer(t, "")
	_ = store.SaveAccount(`{"private_key":"a","peer_public_key":"b"}`)
	do(t, "POST", ts.URL+"/api/run", "", `{"operation":"scan","spec":{"preset":"standard","telegramOnly":true}}`)
	events := waitEnd(t, ts.URL)
	var req struct {
		Payload core.ScanOptions `json:"payload"`
	}
	if err := json.Unmarshal([]byte(rr.last()), &req); err != nil {
		t.Fatal(err)
	}
	if !req.Payload.Telegram || !req.Payload.TelegramOnly {
		t.Fatalf("telegram flags lost: %+v", req.Payload)
	}
	if !strings.Contains(events, `"telegram":{`) || !strings.Contains(events, `"ok":true`) {
		t.Fatalf("telegram status missing from events:\n%s", events)
	}
	_, body := do(t, "GET", ts.URL+"/api/history", "", "")
	var list struct{ Items []HistorySummary }
	_ = json.Unmarshal([]byte(body), &list)
	if len(list.Items) != 1 || list.Items[0].Telegram == 0 || list.Items[0].Telegram != list.Items[0].Working {
		t.Fatalf("history must count Telegram-ok endpoints (only-with-Telegram keeps just those): %+v", list.Items)
	}
}
