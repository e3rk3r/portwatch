package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultReplayPolicy returns a ReplayPolicy with sensible defaults.
func DefaultReplayPolicy() ReplayPolicy {
	return ReplayPolicy{
		MaxEvents: 64,
		MaxAge:    10 * time.Minute,
	}
}

// ReplayPolicy controls which stored events are eligible for replay.
type ReplayPolicy struct {
	// MaxEvents is the maximum number of events to keep in the replay buffer.
	MaxEvents int
	// MaxAge is the maximum age of an event eligible for replay.
	MaxAge time.Duration
}

// ReplayEntry holds a stored notification together with its timestamp.
type ReplayEntry struct {
	Notification Notification
	StoredAt     time.Time
}

// Replayer stores recent notifications and can replay them to a target
// Dispatcher — useful when a downstream channel recovers after an outage.
type Replayer struct {
	mu     sync.Mutex
	policy ReplayPolicy
	buf    []ReplayEntry
	clock  func() time.Time
}

// NewReplayer creates a Replayer with the given policy.
func NewReplayer(p ReplayPolicy) *Replayer {
	if p.MaxEvents <= 0 {
		p.MaxEvents = DefaultReplayPolicy().MaxEvents
	}
	if p.MaxAge <= 0 {
		p.MaxAge = DefaultReplayPolicy().MaxAge
	}
	return &Replayer{policy: p, clock: time.Now}
}

// Record stores a notification in the replay buffer, evicting the oldest
// entry when the buffer is full.
func (r *Replayer) Record(n Notification) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry := ReplayEntry{Notification: n, StoredAt: r.clock()}
	if len(r.buf) >= r.policy.MaxEvents {
		r.buf = r.buf[1:]
	}
	r.buf = append(r.buf, entry)
}

// Replay sends all buffered notifications that are younger than MaxAge to
// dst. Entries that fail to send are retained for a subsequent call.
func (r *Replayer) Replay(ctx context.Context, dst Dispatcher) error {
	r.mu.Lock()
	candidates := make([]ReplayEntry, len(r.buf))
	copy(candidates, r.buf)
	r.mu.Unlock()

	now := r.clock()
	var failed []ReplayEntry
	for _, e := range candidates {
		if now.Sub(e.StoredAt) > r.policy.MaxAge {
			continue // drop expired
		}
		if err := dst.Send(ctx, e.Notification); err != nil {
			failed = append(failed, e)
		}
	}

	r.mu.Lock()
	r.buf = failed
	r.mu.Unlock()

	if len(failed) > 0 {
		return fmt.Errorf("replay: %d notification(s) failed to deliver", len(failed))
	}
	return nil
}

// Len returns the current number of buffered entries.
func (r *Replayer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.buf)
}
