package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultCheckpointPolicy returns a CheckpointPolicy with sensible defaults.
func DefaultCheckpointPolicy() CheckpointPolicy {
	return CheckpointPolicy{
		MaxAge: 5 * time.Minute,
	}
}

// CheckpointPolicy controls how long a checkpoint entry is retained.
type CheckpointPolicy struct {
	MaxAge time.Duration
}

// checkpointEntry records the last successfully dispatched notification for a key.
type checkpointEntry struct {
	Notification Notification
	At           time.Time
}

// Checkpointer records the last successful dispatch per (port, state) key so
// that a downstream component can resume from a known-good position after a
// restart or transient failure.
type Checkpointer struct {
	mu     sync.RWMutex
	policy CheckpointPolicy
	clock  func() time.Time
	store  map[string]checkpointEntry
}

// NewCheckpointer creates a Checkpointer with the given policy.
// Pass nil for clock to use time.Now.
func NewCheckpointer(p CheckpointPolicy, clock func() time.Time) *Checkpointer {
	if clock == nil {
		clock = time.Now
	}
	return &Checkpointer{
		policy: p,
		clock:  clock,
		store:  make(map[string]checkpointEntry),
	}
}

func checkpointKey(n Notification) string {
	return fmt.Sprintf("%d:%s", n.Port, n.State)
}

// Record saves n as the latest checkpoint for its (port, state) key.
func (c *Checkpointer) Record(n Notification) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[checkpointKey(n)] = checkpointEntry{Notification: n, At: c.clock()}
}

// Latest returns the most recent notification for the given port and state,
// along with whether a valid (non-expired) entry was found.
func (c *Checkpointer) Latest(port int, state string) (Notification, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := fmt.Sprintf("%d:%s", port, state)
	e, ok := c.store[key]
	if !ok {
		return Notification{}, false
	}
	if c.policy.MaxAge > 0 && c.clock().Sub(e.At) > c.policy.MaxAge {
		return Notification{}, false
	}
	return e.Notification, true
}

// Reset clears all stored checkpoints.
func (c *Checkpointer) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]checkpointEntry)
}

// NewCheckpointDispatcher wraps next, recording a checkpoint after each
// successful dispatch.
func NewCheckpointDispatcher(next Dispatcher, cp *Checkpointer) Dispatcher {
	if next == nil {
		panic("checkpoint: next dispatcher must not be nil")
	}
	if cp == nil {
		panic("checkpoint: checkpointer must not be nil")
	}
	return &checkpointDispatcher{next: next, cp: cp}
}

type checkpointDispatcher struct {
	next Dispatcher
	cp   *Checkpointer
}

func (d *checkpointDispatcher) Send(ctx context.Context, n Notification) error {
	if err := d.next.Send(ctx, n); err != nil {
		return err
	}
	d.cp.Record(n)
	return nil
}
