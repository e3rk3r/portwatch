package alert

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // normal operation
	CircuitOpen                         // blocking calls
	CircuitHalfOpen                     // testing recovery
)

// CircuitPolicy configures the circuit breaker behaviour.
type CircuitPolicy struct {
	FailureThreshold int           // consecutive failures before opening
	SuccessThreshold int           // consecutive successes to close from half-open
	OpenDuration     time.Duration // how long to stay open before half-open
}

// DefaultCircuitPolicy returns a sensible default policy.
func DefaultCircuitPolicy() CircuitPolicy {
	return CircuitPolicy{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		OpenDuration:     30 * time.Second,
	}
}

// circuitClock allows time injection in tests.
type circuitClock func() time.Time

// CircuitBreaker tracks failure/success counts and opens or closes the circuit.
type CircuitBreaker struct {
	mu       sync.Mutex
	policy   CircuitPolicy
	state    CircuitState
	failures int
	successes int
	openedAt time.Time
	now      circuitClock
}

// NewCircuitBreaker creates a CircuitBreaker with the given policy.
func NewCircuitBreaker(p CircuitPolicy) *CircuitBreaker {
	if p.FailureThreshold <= 0 {
		p.FailureThreshold = DefaultCircuitPolicy().FailureThreshold
	}
	if p.OpenDuration <= 0 {
		p.OpenDuration = DefaultCircuitPolicy().OpenDuration
	}
	if p.SuccessThreshold <= 0 {
		p.SuccessThreshold = DefaultCircuitPolicy().SuccessThreshold
	}
	return &CircuitBreaker{policy: p, now: time.Now}
}

// Allow returns true if the call should proceed.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case CircuitOpen:
		if cb.now().Sub(cb.openedAt) >= cb.policy.OpenDuration {
			cb.state = CircuitHalfOpen
			cb.successes = 0
			return true
		}
		return false
	default:
		return true
	}
}

// RecordSuccess records a successful call.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	if cb.state == CircuitHalfOpen {
		cb.successes++
		if cb.successes >= cb.policy.SuccessThreshold {
			cb.state = CircuitClosed
		}
	}
}

// RecordFailure records a failed call.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.successes = 0
	cb.failures++
	if cb.state == CircuitHalfOpen || cb.failures >= cb.policy.FailureThreshold {
		cb.state = CircuitOpen
		cb.openedAt = cb.now()
		cb.failures = 0
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
