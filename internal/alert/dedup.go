// Package alert provides alerting primitives for portwatch.
package alert

import (
	"sync"
	"time"
)

// DefaultDedupPolicy returns a DedupPolicy with a 5-minute suppression window.
func DefaultDedupPolicy() DedupPolicy {
	return DedupPolicy{Window: 5 * time.Minute}
}

// DedupPolicy controls how long identical events are suppressed.
type DedupPolicy struct {
	Window time.Duration
}

// Deduplicator suppresses repeated notifications for the same key within a
// configurable time window.
type Deduplicator struct {
	policy DedupPolicy
	clock  func() time.Time
	mu     sync.Mutex
	last   map[string]time.Time
}

// NewDeduplicator creates a Deduplicator with the given policy.
// Pass a nil clock to use time.Now.
func NewDeduplicator(p DedupPolicy, clock func() time.Time) *Deduplicator {
	if clock == nil {
		clock = time.Now
	}
	return &Deduplicator{
		policy: p,
		clock:  clock,
		last:   make(map[string]time.Time),
	}
}

// Allow returns true if the notification should be delivered, i.e. the key
// has not been seen within the configured window.
func (d *Deduplicator) Allow(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.clock()
	if t, ok := d.last[key]; ok && now.Sub(t) < d.policy.Window {
		return false
	}
	d.last[key] = now
	return true
}

// Reset clears the dedup state for the given key, allowing the next
// notification through regardless of the window.
func (d *Deduplicator) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.last, key)
}
