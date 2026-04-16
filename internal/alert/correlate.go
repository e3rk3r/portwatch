package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultCorrelatePolicy returns a CorrelatePolicy with sensible defaults.
func DefaultCorrelatePolicy() CorrelatePolicy {
	return CorrelatePolicy{
		WindowDuration: 2 * time.Minute,
		MinEvents:      3,
	}
}

// CorrelatePolicy controls how events are correlated into a single alert.
type CorrelatePolicy struct {
	WindowDuration time.Duration
	MinEvents      int
}

type correlateEntry struct {
	count     int
	firstSeen time.Time
	lastNotif Notification
}

// Correlator groups repeated notifications for the same port+state within a
// time window and only forwards them once the MinEvents threshold is met.
type Correlator struct {
	policy CorrelatePolicy
	clock  func() time.Time
	mu     sync.Mutex
	entries map[string]*correlateEntry
}

// NewCorrelator creates a Correlator with the given policy.
func NewCorrelator(p CorrelatePolicy) *Correlator {
	return &Correlator{
		policy:  p,
		clock:   time.Now,
		entries: make(map[string]*correlateEntry),
	}
}

func correlateKey(n Notification) string {
	return fmt.Sprintf("%d:%s", n.Port, n.State)
}

// Record records a notification and returns true when the correlation
// threshold has been reached for the first time in the current window.
func (c *Correlator) Record(n Notification) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock()
	key := correlateKey(n)
	e, ok := c.entries[key]
	if !ok || now.Sub(e.firstSeen) > c.policy.WindowDuration {
		c.entries[key] = &correlateEntry{count: 1, firstSeen: now, lastNotif: n}
		return false
	}
	e.count++
	e.lastNotif = n
	return e.count == c.policy.MinEvents
}

// Reset clears all correlation state.
func (c *Correlator) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*correlateEntry)
}

// NewCorrelateDispatcher wraps next and only forwards notifications once the
// correlation threshold is met within the configured window.
func NewCorrelateDispatcher(p CorrelatePolicy, next Dispatcher) Dispatcher {
	if next == nil {
		panic("correlate: next dispatcher must not be nil")
	}
	c := NewCorrelator(p)
	return dispatcherFunc(func(ctx context.Context, n Notification) error {
		if c.Record(n) {
			return next.Send(ctx, n)
		}
		return nil
	})
}
