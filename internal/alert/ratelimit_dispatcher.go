package alert

import (
	"context"
	"fmt"
	"time"
)

// RateLimitDispatcher wraps a Dispatcher and enforces a token-bucket style
// rate limit: at most MaxBurst notifications per Window duration per key.
// Notifications that exceed the limit are dropped and an error is returned.
type RateLimitDispatcher struct {
	next     Dispatcher
	limiter  *Throttler
}

// NewRateLimitDispatcher constructs a RateLimitDispatcher with the given
// throttle policy and downstream dispatcher.
//
// Panics if next is nil.
func NewRateLimitDispatcher(policy ThrottlePolicy, next Dispatcher) *RateLimitDispatcher {
	if next == nil {
		panic("alert: NewRateLimitDispatcher: next dispatcher must not be nil")
	}
	return &RateLimitDispatcher{
		next:    next,
		limiter: NewThrottler(policy, realClock{}),
	}
}

// Send forwards n to the downstream dispatcher only when the per-key rate
// limit has not been exceeded. The key is composed from the notification port
// and state, mirroring the throttle key convention.
func (d *RateLimitDispatcher) Send(ctx context.Context, n Notification) error {
	key := throttleKey(n)
	allow, retryAfter := d.limiter.Allow(key, time.Now())
	if !allow {
		return fmt.Errorf("alert: rate limit exceeded for %s; retry after %s", key, retryAfter.Round(time.Millisecond))
	}
	return d.next.Send(ctx, n)
}

// realClock satisfies the clock interface used by Throttler.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }
