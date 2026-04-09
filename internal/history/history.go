// Package history provides a simple in-memory ring buffer for recording
// port state change events that the daemon has observed.
package history

import (
	"sync"
	"time"
)

// Event represents a single state-change record for a monitored port.
type Event struct {
	Port      int       `json:"port"`
	Host      string    `json:"host"`
	State     string    `json:"state"`
	Timestamp time.Time `json:"timestamp"`
}

// Ring is a fixed-capacity, thread-safe circular buffer of Events.
type Ring struct {
	mu       sync.Mutex
	events   []Event
	cap      int
	head     int
	count    int
}

// NewRing creates a Ring that retains at most capacity events.
func NewRing(capacity int) *Ring {
	if capacity <= 0 {
		capacity = 100
	}
	return &Ring{
		events: make([]Event, capacity),
		cap:    capacity,
	}
}

// Record appends an event to the ring, overwriting the oldest entry when full.
func (r *Ring) Record(e Event) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events[r.head] = e
	r.head = (r.head + 1) % r.cap
	if r.count < r.cap {
		r.count++
	}
}

// Snapshot returns a copy of all stored events in chronological order
// (oldest first).
func (r *Ring) Snapshot() []Event {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Event, r.count)
	start := (r.head - r.count + r.cap) % r.cap
	for i := 0; i < r.count; i++ {
		out[i] = r.events[(start+i)%r.cap]
	}
	return out
}

// Len returns the number of events currently stored.
func (r *Ring) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}
