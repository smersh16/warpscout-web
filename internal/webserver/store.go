package webserver

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vernette/warpscout/core"
)

var idPattern = regexp.MustCompile(`^[0-9A-Za-z-]{1,64}$`)

// Store keeps the WARP account and scan history as plain files in DataDir.
type Store struct {
	dir string
	mu  sync.Mutex
}

type HistoryEntry struct {
	ID        string          `json:"id"`
	CreatedAt time.Time       `json:"createdAt"`
	Preset    string          `json:"preset,omitempty"`
	Spec      json.RawMessage `json:"spec,omitempty"`
	Options   json.RawMessage `json:"options,omitempty"`
	Report    json.RawMessage `json:"report"`
}

type HistorySummary struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Preset    string    `json:"preset,omitempty"`
	Protocol  string    `json:"protocol"`
	Total     int       `json:"total"`
	Working   int       `json:"working"`
	Telegram  int       `json:"telegramOk,omitempty"`
	Best      string    `json:"best,omitempty"`
	BestPing  float64   `json:"bestPingMs,omitempty"`
}

func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(filepath.Join(dir, "history"), 0o700); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) accountPath() string { return filepath.Join(s.dir, "account.json") }

func (s *Store) Account() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.accountPath())
	if err != nil || strings.TrimSpace(string(data)) == "" {
		return "", false
	}
	return string(data), true
}

func (s *Store) SaveAccount(raw string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeAtomic(s.accountPath(), []byte(raw))
}

func (s *Store) DeleteAccount() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.accountPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Store) AddHistory(preset string, spec, options, report json.RawMessage) (string, error) {
	var suffix [3]byte
	_, _ = rand.Read(suffix[:])
	now := time.Now().UTC()
	id := now.Format("20060102-150405") + "-" + hex.EncodeToString(suffix[:])
	entry := HistoryEntry{ID: id, CreatedAt: now, Preset: preset, Spec: spec, Options: options, Report: report}
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return id, writeAtomic(filepath.Join(s.dir, "history", id+".json"), data)
}

func (s *Store) History() ([]HistorySummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	files, err := filepath.Glob(filepath.Join(s.dir, "history", "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	if len(files) > 200 {
		files = files[:200]
	}
	out := make([]HistorySummary, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var entry HistoryEntry
		if json.Unmarshal(data, &entry) != nil {
			continue
		}
		var report core.ScanReport
		_ = json.Unmarshal(entry.Report, &report)
		summary := HistorySummary{ID: entry.ID, CreatedAt: entry.CreatedAt, Preset: entry.Preset, Protocol: string(report.Protocol), Total: len(report.Results)}
		for _, r := range report.Results {
			if !r.Working {
				continue
			}
			summary.Working++
			if r.Telegram != nil && r.Telegram.OK {
				summary.Telegram++
			}
			ping := r.TunnelPingMS
			if ping == 0 {
				ping = r.EndpointPingMS
			}
			if summary.Best == "" || (ping > 0 && ping < summary.BestPing) {
				summary.Best, summary.BestPing = r.Endpoint, ping
			}
		}
		out = append(out, summary)
	}
	return out, nil
}

func (s *Store) HistoryEntry(id string) ([]byte, error) {
	if !idPattern.MatchString(id) {
		return nil, os.ErrNotExist
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.ReadFile(filepath.Join(s.dir, "history", id+".json"))
}

func (s *Store) DeleteHistory(id string) error {
	if !idPattern.MatchString(id) {
		return os.ErrNotExist
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.Remove(filepath.Join(s.dir, "history", id+".json"))
}

func (s *Store) ClearHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	files, _ := filepath.Glob(filepath.Join(s.dir, "history", "*.json"))
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ToolResult is the latest outcome of find-junk / find-sni.
type ToolResult struct {
	At      time.Time       `json:"at"`
	Payload json.RawMessage `json:"payload"`
}

func (s *Store) toolsPath() string { return filepath.Join(s.dir, "tools.json") }

func (s *Store) Tools() map[string]ToolResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readTools()
}

func (s *Store) readTools() map[string]ToolResult {
	out := map[string]ToolResult{}
	if data, err := os.ReadFile(s.toolsPath()); err == nil {
		_ = json.Unmarshal(data, &out)
	}
	return out
}

func (s *Store) SaveTool(name string, payload json.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := s.readTools()
	all[name] = ToolResult{At: time.Now().UTC(), Payload: payload}
	data, err := json.Marshal(all)
	if err != nil {
		return err
	}
	return writeAtomic(s.toolsPath(), data)
}
