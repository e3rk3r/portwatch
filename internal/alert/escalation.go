package alert

import (
	"sync"
	"time"
)

// EscalationPolicy defines thresholds for escalating alerts to additional channels.
type EscalationPolicy struct {
	// After this many consecutive failures, escalate.
	Threshold int
	// Minimum duration the port must remain in a failed state before escalating.
	MinDuration time.Duration
}

// DefaultEscalationPolicy returns a sensible default escalation policy.
func DefaultEscalationPolicy() EscalationPolicy {
	return EscalationPolicy{
		Threshold:   3,
		MinDuration: 5 * time.Minute,
	}
}

type escalationState struct {
	count     int
	firstSeen time.Time
}

// Escalator tracks consecutive failure counts per key and decides when to escalate.
type Escalator struct {
	mu     sync.Mutex
	policy EscalationPolicy
	clock  func() time.Time
	state  map[string]*escalationState
}

// NewEscalator creates a new Escalator with the given policy.
func NewEscalator(policy EscalationPolicy) *Escalator {
	return &Escalator{
		policy: policy,
		clock:  time.Now,
		state:  make(map[string]*escalationState),
	}
}

// Record registers a failure event for the given key.
// Returns true if the escalation threshold has been reached.
func (e *Escalator) Record(key string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := e.clock()
	s, ok := e.state[key]
	if !ok {
		e.state[key] = &escalationState{count: 1, firstSeen: now}
		return false
	}
	s.count++
	duration := now.Sub(s.firstSeen)
	return s.count >= e.policy.Threshold && duration >= e.policy.MinDuration
}

// Reset clears the escalation state for the given key (e.g., port recovered).
func (e *Escalator) Reset(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.state, key)
}

// Count returns the current failure count for the given key.
func (e *Escalator) Count(key string) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	if s, ok := e.state[key]; ok {
		return s.count
	}
	return 0
}
