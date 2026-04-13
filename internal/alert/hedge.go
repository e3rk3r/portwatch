package alert

import (
	"context"
	"sync"
	"time"
)

// HedgePolicy configures the hedge dispatcher behaviour.
type HedgePolicy struct {
	// Delay is how long to wait before firing the hedge request.
	Delay time.Duration
	// MaxHedges is the maximum number of concurrent hedge attempts (1 = one hedge).
	MaxHedges int
}

// DefaultHedgePolicy returns a sensible default hedge policy.
func DefaultHedgePolicy() HedgePolicy {
	return HedgePolicy{
		Delay:     200 * time.Millisecond,
		MaxHedges: 1,
	}
}

// hedgeDispatcher fires the primary dispatcher and, after Delay, also fires
// a hedge request against the same target. The first success wins; the result
// is returned once at least one attempt completes without error. If all
// attempts fail the last error is returned.
type hedgeDispatcher struct {
	next   Dispatcher
	policy HedgePolicy
}

// NewHedgeDispatcher wraps next with hedge logic defined by policy.
func NewHedgeDispatcher(next Dispatcher, policy HedgePolicy) Dispatcher {
	if next == nil {
		panic("alert: NewHedgeDispatcher: next must not be nil")
	}
	if policy.MaxHedges < 1 {
		policy.MaxHedges = 1
	}
	if policy.Delay <= 0 {
		policy.Delay = DefaultHedgePolicy().Delay
	}
	return &hedgeDispatcher{next: next, policy: policy}
}

func (h *hedgeDispatcher) Send(ctx context.Context, n Notification) error {
	total := h.policy.MaxHedges + 1 // original + hedges

	type result struct{ err error }
	results := make(chan result, total)

	var wg sync.WaitGroup

	launch := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := h.next.Send(ctx, n)
			results <- result{err: err}
		}()
	}

	launch() // primary

	for i := 0; i < h.policy.MaxHedges; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(h.policy.Delay):
			launch()
		case r := <-results:
			if r.err == nil {
				return nil
			}
		}
	}

	// Drain remaining results; first nil wins.
	var lastErr error
	for i := 0; i < total; i++ {
		r := <-results
		if r.err == nil {
			return nil
		}
		lastErr = r.err
	}
	return lastErr
}
