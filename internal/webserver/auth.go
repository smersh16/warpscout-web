package webserver

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	sessionTTL = 7 * 24 * time.Hour
	maxFails   = 5
	failWindow = 15 * time.Minute
	lockFor    = 15 * time.Minute
)

type failRecord struct {
	count       int
	first       time.Time
	lockedUntil time.Time
}

// authState holds server-side sessions (random ids with an expiry, lost on
// restart) and per-IP failed-attempt counters.
type authState struct {
	mu       sync.Mutex
	now      func() time.Time
	sessions map[string]time.Time
	fails    map[string]*failRecord
}

func newAuthState() *authState {
	return &authState{now: time.Now, sessions: map[string]time.Time{}, fails: map[string]*failRecord{}}
}

func (a *authState) newSession() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	id := hex.EncodeToString(buf)
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	for k, exp := range a.sessions {
		if now.After(exp) {
			delete(a.sessions, k)
		}
	}
	a.sessions[id] = now.Add(sessionTTL)
	return id
}

func (a *authState) validSession(id string) bool {
	if id == "" {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	exp, ok := a.sessions[id]
	if !ok {
		return false
	}
	if a.now().After(exp) {
		delete(a.sessions, id)
		return false
	}
	return true
}

func (a *authState) dropSession(id string) {
	a.mu.Lock()
	delete(a.sessions, id)
	a.mu.Unlock()
}

// locked reports whether ip is currently blocked from token attempts.
func (a *authState) locked(ip string) (time.Duration, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	rec := a.fails[ip]
	if rec == nil {
		return 0, false
	}
	if left := rec.lockedUntil.Sub(a.now()); left > 0 {
		return left, true
	}
	return 0, false
}

func (a *authState) fail(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	rec := a.fails[ip]
	if rec == nil || now.Sub(rec.first) > failWindow {
		rec = &failRecord{first: now}
		a.fails[ip] = rec
	}
	rec.count++
	if rec.count >= maxFails {
		rec.lockedUntil = now.Add(lockFor)
		rec.count = 0
		rec.first = now
	}
	if len(a.fails) > 10000 { // keep memory bounded under a spray of addresses
		for k, r := range a.fails {
			if now.After(r.lockedUntil) && now.Sub(r.first) > failWindow {
				delete(a.fails, k)
			}
		}
	}
}

func (a *authState) succeed(ip string) {
	a.mu.Lock()
	delete(a.fails, ip)
	a.mu.Unlock()
}

// clientIP is the peer address; behind a trusted reverse proxy it is the
// rightmost X-Forwarded-For entry (the one the proxy itself appended).
func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
