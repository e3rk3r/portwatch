package alert

import (
	"context"
	"time"
)

// ObserveDispatcher wraps a Dispatcher and records latency and error metrics
// via an Observer. It is transparent: the original error is always returned.
type ObserveDispatcher struct {
	next     Dispatcher
	observer *Observer
}

// NewObserveDispatcher creates an ObserveDispatcher.
// Panics if next or observer is nil.
func NewObserveDispatcher(next Dispatcher, obs *Observer) *ObserveDispatcher {
	if next == nil {
		panic("observe: next dispatcher must not be nil")
	}
	if obs == nil {
		panic("observe: observer must not be nil")
	}
	return &ObserveDispatcher{next: next, observer: obs}
}

// Send records timing and outcome then forwards to the next dispatcher.
func (d *ObserveDispatcher) Send(ctx context.Context, n Notification) error {
	start := time.Now()
	err := d.next.Send(ctx, n)
	d.observer.Record(err, time.Since(start), false)
	return err
}
