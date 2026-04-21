package alert

import (
	"context"
	"sync"
)

// TapPolicy controls the behaviour of a tap dispatcher.
type TapPolicy struct {
	// MaxCapacity is the maximum number of notifications held in the tap
	// buffer. Older entries are evicted when the buffer is full.
	MaxCapacity int
}

// DefaultTapPolicy returns a TapPolicy with sensible defaults.
func DefaultTapPolicy() TapPolicy {
	return TapPolicy{MaxCapacity: 256}
}

// Tap is a passive observer that records every notification that flows
// through a dispatcher chain without altering the result. It is useful
// for testing, debugging, and audit purposes.
type Tap struct {
	policy TapPolicy
	mu     sync.Mutex
	buf    []Notification
}

// NewTap creates a Tap with the given policy. If policy is nil the
// default policy is used.
func NewTap(policy *TapPolicy) *Tap {
	p := DefaultTapPolicy()
	if policy != nil {
		p = *policy
	}
	if p.MaxCapacity <= 0 {
		p.MaxCapacity = DefaultTapPolicy().MaxCapacity
	}
	return &Tap{policy: p}
}

// record stores n in the ring buffer, evicting the oldest entry when full.
func (t *Tap) record(n Notification) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.buf) >= t.policy.MaxCapacity {
		t.buf = t.buf[1:]
	}
	t.buf = append(t.buf, n)
}

// Snapshot returns a copy of all captured notifications in arrival order.
func (t *Tap) Snapshot() []Notification {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Notification, len(t.buf))
	copy(out, t.buf)
	return out
}

// Len returns the number of notifications currently held in the tap.
func (t *Tap) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.buf)
}

// Reset clears all captured notifications.
func (t *Tap) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = t.buf[:0]
}

// NewTapDispatcher wraps next so that every notification is recorded in
// tap before being forwarded. The call to next is always made and its
// error is returned unchanged.
func NewTapDispatcher(tap *Tap, next Dispatcher) Dispatcher {
	if tap == nil {
		panic("alert: NewTapDispatcher: tap must not be nil")
	}
	if next == nil {
		panic("alert: NewTapDispatcher: next must not be nil")
	}
	return &tapDispatcher{tap: tap, next: next}
}

type tapDispatcher struct {
	tap  *Tap
	next Dispatcher
}

func (d *tapDispatcher) Send(ctx context.Context, n Notification) error {
	d.tap.record(n)
	return d.next.Send(ctx, n)
}
