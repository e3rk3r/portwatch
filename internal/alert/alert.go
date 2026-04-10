// Package alert provides rate-limited alerting to prevent webhook/script
// flooding when a port flaps rapidly between states.
package alert

import (
	"sync"
	"time"
)

// Policy controls how often alerts fire for a given key.
type Policy struct {
	// Cooldown is the minimum duration between successive alerts for the same key.
	Cooldown time.Duration
}

// DefaultPolicy returns a Policy with a 30-second cooldown.
func DefaultPolicy() Policy {
	return Policy{Cooldown: 30 * time.Second}
}

// Limiter tracks last-alert timestamps per key and suppresses alerts that
// arrive within the cooldown window.
type Limiter struct {
	mu     sync.Mutex
	last   map[string]time.Time
	policy Policy
	now    func() time.Time
}

// NewLimiter creates a Limiter using the given Policy.
func NewLimiter(p Policy) *Limiter {
	return &Limiter{
		last:   make(map[string]time.Time),
		policy: p,
		now:    time.Now,
	}
}

// Allow returns true if the alert for key should be fired, updating the
// internal timestamp. Returns false when still within the cooldown window.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if t, ok := l.last[key]; ok {
		if now.Sub(t) < l.policy.Cooldown {
			return false
		}
	}
	l.last[key] = now
	return true
}

// Reset clears the recorded timestamp for key, allowing the next alert
// through immediately regardless of cooldown.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, key)
}
