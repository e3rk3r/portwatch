// Package alert provides rate limiting and notification dispatch for portwatch.
package alert

import (
	"math"
	"sync"
	"time"
)

// BackoffPolicy defines exponential backoff parameters for alert suppression.
type BackoffPolicy struct {
	BaseInterval time.Duration
	Multiplier   float64
	MaxInterval  time.Duration
}

// DefaultBackoffPolicy returns a sensible default exponential backoff policy.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		BaseInterval: 30 * time.Second,
		Multiplier:   2.0,
		MaxInterval:  30 * time.Minute,
	}
}

// backoffEntry tracks the current backoff state for a single key.
type backoffEntry struct {
	attempts int
	nextAt   time.Time
}

// BackoffLimiter applies exponential backoff per key, suppressing repeated
// alerts for the same port/state combination until the backoff interval elapses.
type BackoffLimiter struct {
	mu     sync.Mutex
	policy BackoffPolicy
	clock  func() time.Time
	state  map[string]*backoffEntry
}

// NewBackoffLimiter constructs a BackoffLimiter with the given policy.
// Pass nil for clock to use time.Now.
func NewBackoffLimiter(p BackoffPolicy, clock func() time.Time) *BackoffLimiter {
	if clock == nil {
		clock = time.Now
	}
	return &BackoffLimiter{
		policy: p,
		clock:  clock,
		state:  make(map[string]*backoffEntry),
	}
}

// Allow returns true if the alert for key should be dispatched now.
// On each allowed call the backoff interval for that key is doubled up to MaxInterval.
func (b *BackoffLimiter) Allow(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.clock()
	entry, ok := b.state[key]
	if !ok {
		// First occurrence: allow immediately, schedule next backoff window.
		b.state[key] = &backoffEntry{
			attempts: 1,
			nextAt:   now.Add(b.policy.BaseInterval),
		}
		return true
	}

	if now.Before(entry.nextAt) {
		return false
	}

	// Compute next interval: base * multiplier^attempts, capped at max.
	interval := time.Duration(float64(b.policy.BaseInterval) *
		math.Pow(b.policy.Multiplier, float64(entry.attempts)))
	if interval > b.policy.MaxInterval {
		interval = b.policy.MaxInterval
	}

	entry.attempts++
	entry.nextAt = now.Add(interval)
	return true
}

// Reset clears the backoff state for the given key.
func (b *BackoffLimiter) Reset(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.state, key)
}
