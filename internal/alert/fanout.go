package alert

import (
	"context"
	"fmt"
	"strings"
)

// FanoutDispatcher delivers a notification to multiple downstream dispatchers
// concurrently. All errors are collected and returned as a combined error.
// A nil entry in the targets slice is silently skipped.
type FanoutDispatcher struct {
	targets []Dispatcher
}

// NewFanoutDispatcher creates a FanoutDispatcher that broadcasts to each
// provided target. Panics if targets is empty.
func NewFanoutDispatcher(targets ...Dispatcher) *FanoutDispatcher {
	if len(targets) == 0 {
		panic("fanout: at least one target dispatcher is required")
	}
	return &FanoutDispatcher{targets: targets}
}

// Send delivers n to every non-nil target dispatcher concurrently.
// It waits for all goroutines to complete before returning.
// If one or more targets return an error, all errors are combined.
func (f *FanoutDispatcher) Send(ctx context.Context, n Notification) error {
	type result struct {
		idx int
		err error
	}

	ch := make(chan result, len(f.targets))

	for i, d := range f.targets {
		if d == nil {
			ch <- result{i, nil}
			continue
		}
		go func(idx int, disp Dispatcher) {
			ch <- result{idx, disp.Send(ctx, n)}
		}(i, d)
	}

	var errs []string
	for range f.targets {
		if r := <-ch; r.err != nil {
			errs = append(errs, fmt.Sprintf("target[%d]: %v", r.idx, r.err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("fanout errors: %s", strings.Join(errs, "; "))
	}
	return nil
}
