package alert

import (
	"context"
	"fmt"
	"time"
)

// DefaultTimeoutPolicy returns a TimeoutPolicy with a 5-second per-dispatch timeout.
func DefaultTimeoutPolicy() TimeoutPolicy {
	return TimeoutPolicy{PerDispatch: 5 * time.Second}
}

// TimeoutPolicy configures how long a single dispatch attempt may run.
type TimeoutPolicy struct {
	// PerDispatch is the maximum duration allowed for a single Send call.
	PerDispatch time.Duration
}

// TimeoutDispatcher wraps a Dispatcher and enforces a per-call deadline.
type TimeoutDispatcher struct {
	next   Dispatcher
	policy TimeoutPolicy
}

// NewTimeoutDispatcher returns a TimeoutDispatcher that cancels Send calls
// exceeding policy.PerDispatch. Panics if next is nil.
func NewTimeoutDispatcher(next Dispatcher, policy TimeoutPolicy) *TimeoutDispatcher {
	if next == nil {
		panic("alert: TimeoutDispatcher requires a non-nil next dispatcher")
	}
	d := policy.PerDispatch
	if d <= 0 {
		d = DefaultTimeoutPolicy().PerDispatch
	}
	return &TimeoutDispatcher{next: next, policy: TimeoutPolicy{PerDispatch: d}}
}

// Send invokes the next dispatcher with a deadline-bounded context.
// It returns an error if the deadline is exceeded before Send returns.
func (t *TimeoutDispatcher) Send(ctx context.Context, n Notification) error {
	ctx, cancel := context.WithTimeout(ctx, t.policy.PerDispatch)
	defer cancel()

	type result struct{ err error }
	ch := make(chan result, 1)
	go func() {
		ch <- result{err: t.next.Send(ctx, n)}
	}()

	select {
	case r := <-ch:
		return r.err
	case <-ctx.Done():
		return fmt.Errorf("alert: dispatch timed out after %s: %w", t.policy.PerDispatch, ctx.Err())
	}
}
