package webserver

import (
	"encoding/json"
	"sync"
	"time"
)

const maxJobEvents = 4000

// job is one background operation. Its events are buffered so that a browser
// tab can disconnect and later resume from the last sequence number it saw.
type job struct {
	id        int64
	operation string
	started   time.Time

	mu       sync.Mutex
	events   []string
	base     int // number of events dropped from the front
	done     bool
	terminal bool
	notify   chan struct{}
}

func newJob(id int64, operation string) *job {
	return &job{id: id, operation: operation, started: time.Now().UTC(), notify: make(chan struct{})}
}

func (j *job) add(event string) {
	var head struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal([]byte(event), &head)

	j.mu.Lock()
	defer j.mu.Unlock()
	j.events = append(j.events, event)
	if len(j.events) > maxJobEvents {
		drop := len(j.events) - maxJobEvents/2
		j.events = append([]string(nil), j.events[drop:]...)
		j.base += drop
	}
	if head.Type == "completed" || head.Type == "error" {
		j.terminal = true
	}
	j.wake()
}

func (j *job) finish() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.done = true
	j.wake()
}

func (j *job) wake() {
	close(j.notify)
	j.notify = make(chan struct{})
}

func (j *job) isDone() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.done
}

func (j *job) hasTerminal() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.terminal
}

// seq returns the sequence number of the newest event (0 when empty).
func (j *job) seq() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.base + len(j.events)
}

// since returns events with sequence number > after. Event k (0-based in the
// buffer) has sequence number base+k+1.
func (j *job) since(after int) (events []string, firstSeq int, done bool, wait <-chan struct{}) {
	j.mu.Lock()
	defer j.mu.Unlock()
	start := after - j.base
	if start < 0 {
		start = 0
	}
	if start > len(j.events) {
		start = len(j.events)
	}
	return append([]string(nil), j.events[start:]...), j.base + start + 1, j.done, j.notify
}
