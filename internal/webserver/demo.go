package webserver

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DemoRunner is a fake core used with --demo to try the UI without network
// access or a WARP account. It produces plausible events and results.
type DemoRunner struct {
	mu     sync.Mutex
	cancel chan struct{}
}

func NewDemoRunner() *DemoRunner { return &DemoRunner{} }

func (d *DemoRunner) Cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cancel != nil {
		close(d.cancel)
		d.cancel = nil
	}
}

func emitJSON(l Listener, v any) {
	data, _ := json.Marshal(v)
	l.OnEvent(string(data))
}

func (d *DemoRunner) Start(requestJSON string, l Listener) error {
	var req struct {
		Operation string          `json:"operation"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return err
	}
	switch req.Operation {
	case "render-report":
		emitJSON(l, map[string]any{"schemaVersion": 1, "type": "completed", "operation": req.Operation, "payload": "WARPSCOUT demo report\n"})
		return nil
	case "render-config":
		emitJSON(l, map[string]any{"schemaVersion": 1, "type": "completed", "operation": req.Operation, "payload": "[Interface]\n# demo config\nPrivateKey = DEMO\n"})
		return nil
	}
	cancel := make(chan struct{})
	d.mu.Lock()
	d.cancel = cancel
	d.mu.Unlock()

	op := req.Operation
	step := func(phase string, i, n int, msg string, ep map[string]any) bool {
		ev := map[string]any{"schemaVersion": 1, "operation": op, "type": "progress", "phase": phase, "completed": i, "total": n, "message": msg}
		if ep != nil {
			ev["type"] = "endpoint"
			ev["endpoint"] = ep
		}
		emitJSON(l, ev)
		select {
		case <-cancel:
			return false
		case <-time.After(250 * time.Millisecond):
			return true
		}
	}
	fail := func() error {
		emitJSON(l, map[string]any{"schemaVersion": 1, "type": "error", "error": map[string]any{"code": "canceled", "message": "operation canceled"}})
		return nil
	}
	switch op {
	case "register":
		if !step("registration", 0, 1, "Registering WARP account", nil) {
			return fail()
		}
		emitJSON(l, map[string]any{"schemaVersion": 1, "type": "completed", "operation": op,
			"payload": map[string]any{"rawJson": `{"private_key":"DEMO","peer_public_key":"DEMO"}`}})
	case "scan":
		var opts struct {
			Telegram     bool `json:"telegram"`
			TelegramOnly bool `json:"telegramOnly"`
		}
		_ = json.Unmarshal(req.Payload, &opts)
		var results []map[string]any
		for i := 0; i < 12; i++ {
			ep := map[string]any{"endpoint": fmt.Sprintf("162.159.192.%d:2408", i+1), "working": i%4 != 3, "durable": i%4 != 3,
				"endpointPingMs": 20 + float64(i*7%40), "tunnelPingMs": 30 + float64(i*11%50), "country": "NL", "node": "AMS", "region": "EU"}
			if opts.Telegram {
				ok := i%3 != 1
				tg := map[string]any{"ok": ok, "dcsReached": 5, "dcsTotal": 5, "rttMs": 60 + float64(i*13%70)}
				if !ok {
					tg = map[string]any{"ok": false, "dcsReached": i % 3, "dcsTotal": 5}
				}
				ep["telegram"] = tg
				if opts.TelegramOnly && !ok {
					continue
				}
			}
			results = append(results, ep)
			if !step("scan", i+1, 12, "", ep) {
				return fail()
			}
		}
		emitJSON(l, map[string]any{"schemaVersion": 1, "type": "completed", "operation": op, "payload": map[string]any{
			"protocol": "awg", "startedAt": time.Now().UTC().Add(-3 * time.Second), "finishedAt": time.Now().UTC(), "results": results}})
	case "socks":
		if !step("handshake", 0, 1, "demo", nil) || !step("listening", 1, 1, "socks5h://127.0.0.1:1080", nil) {
			return fail()
		}
		<-cancel
		emitJSON(l, map[string]any{"schemaVersion": 1, "type": "completed", "operation": op, "payload": map[string]bool{"stopped": true}})
	default:
		for i := 0; i < 5; i++ {
			if !step("tool", i+1, 5, fmt.Sprintf("candidate %d", i+1), nil) {
				return fail()
			}
		}
		payload := map[string]any{"sni": "example.com", "protocol": "masque", "attempts": 5}
		if op == "find-junk" {
			payload = map[string]any{"junkCount": 4, "junkMin": 40, "junkMax": 70, "i1": "<b 0xdemo>"}
		}
		emitJSON(l, map[string]any{"schemaVersion": 1, "type": "completed", "operation": op, "payload": payload})
	}
	return nil
}
