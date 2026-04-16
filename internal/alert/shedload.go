package alert

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrLoadShed is returned when the load shedder drops a notification.
var ErrLoadShed = errors.New("alert: load shed – system overloaded")

// ShedPolicy configures load shedding behaviour.
type ShedPolicy struct {
	// MaxInFlight is the maximum number of concurrent dispatches allowed.
	MaxInFlight int64
}

// DefaultShedPolicy returns a conservative default.
func DefaultShedPolicy() ShedPolicy {
	return ShedPolicy{MaxInFlight: 32}
}

// LoadShedder drops notifications when too many dispatches are in-flight.
type LoadShedder struct {
	policy  ShedPolicy
	inflight atomic.Int64
}

// NewLoadShedder constructs a LoadShedder with the given policy.
func NewLoadShedder(p ShedPolicy) *LoadShedder {
	if p.MaxInFlight <= 0 {
		p.MaxInFlight = DefaultShedPolicy().MaxInFlight
	}
	return &LoadShedder{policy: p}
}

// Acquire attempts to reserve an in-flight slot.
// Returns ErrLoadShed if the limit is already reached.
func (l *LoadShedder) Acquire() error {
	if l.inflight.Add(1) > l.policy.MaxInFlight {
		l.inflight.Add(-1)
		return ErrLoadShed
	}
	return nil
}

// Release frees a previously acquired slot.
func (l *LoadShedder) Release() {
	l.inflight.Add(-1)
}

// InFlight returns the current number of in-flight dispatches.
func (l *LoadShedder) InFlight() int64 {
	return l.inflight.Load()
}

// NewShedDispatcher wraps next with load-shedding logic.
func NewShedDispatcher(policy ShedPolicy, next Dispatcher) Dispatcher {
	if next == nil {
		panic("alert: NewShedDispatcher: next must not be nil")
	}
	shedder := NewLoadShedder(policy)
	return dispatcherFunc(func(ctx context.Context, n Notification) error {
		if err := shedder.Acquire(); err != nil {
			return err
		}
		defer shedder.Release()
		return next.Dispatch(ctx, n)
	})
}
