package alert

import (
	"context"
	"fmt"
)

// ThrottleDispatcher wraps a Dispatcher and drops notifications that exceed
// the burst budget defined by a Throttler.
type ThrottleDispatcher struct {
	next      Dispatcher
	throttler *Throttler
}

// NewThrottleDispatcher returns a ThrottleDispatcher that gates calls to next
// using the provided Throttler. Both arguments must be non-nil.
func NewThrottleDispatcher(next Dispatcher, throttler *Throttler) *ThrottleDispatcher {
	if next == nil {
		panic("throttle_dispatcher: next Dispatcher must not be nil")
	}
	if throttler == nil {
		panic("throttle_dispatcher: Throttler must not be nil")
	}
	return &ThrottleDispatcher{next: next, throttler: throttler}
}

// Send forwards the notification to the next dispatcher only when the
// throttler permits it. Throttled notifications return a descriptive error.
func (d *ThrottleDispatcher) Send(ctx context.Context, n Notification) error {
	if !d.throttler.Allow(n.Port, string(n.State)) {
		return fmt.Errorf("throttle_dispatcher: burst limit reached for port %d state %s",
			n.Port, n.State)
	}
	return d.next.Send(ctx, n)
}
