package alert

import (
	"context"
	"fmt"
)

// CircuitDispatcher wraps a Dispatcher with circuit-breaker protection.
// When the circuit is open, Send returns an error immediately without
// forwarding to the downstream dispatcher.
type CircuitDispatcher struct {
	next    Dispatcher
	breaker *CircuitBreaker
}

// NewCircuitDispatcher creates a CircuitDispatcher using the given policy.
// Panics if next or breaker is nil.
func NewCircuitDispatcher(next Dispatcher, policy CircuitPolicy) *CircuitDispatcher {
	if next == nil {
		panic("alert: CircuitDispatcher next dispatcher must not be nil")
	}
	return &CircuitDispatcher{
		next:    next,
		breaker: NewCircuitBreaker(policy),
	}
}

// Send forwards the notification if the circuit allows it, recording
// success or failure based on the downstream result.
func (cd *CircuitDispatcher) Send(ctx context.Context, n Notification) error {
	if !cd.breaker.Allow() {
		return fmt.Errorf("alert: circuit open, dropping notification for port %d", n.Port)
	}
	err := cd.next.Send(ctx, n)
	if err != nil {
		cd.breaker.RecordFailure()
		return err
	}
	cd.breaker.RecordSuccess()
	return nil
}

// State exposes the current circuit state for observability.
func (cd *CircuitDispatcher) State() CircuitState {
	return cd.breaker.State()
}
