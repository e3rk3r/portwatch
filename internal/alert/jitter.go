package alert

import (
	"context"
	"math/rand"
	"time"
)

// JitterPolicy configures the jitter dispatcher.
type JitterPolicy struct {
	// MaxJitter is the upper bound of the random delay added before dispatching.
	MaxJitter time.Duration
}

// DefaultJitterPolicy returns a JitterPolicy with a 500ms max jitter.
func DefaultJitterPolicy() JitterPolicy {
	return JitterPolicy{
		MaxJitter: 500 * time.Millisecond,
	}
}

// JitterDispatcher wraps a Dispatcher and introduces a random delay before
// forwarding each notification. This helps avoid thundering-herd problems when
// many alerts fire simultaneously.
type JitterDispatcher struct {
	next   Dispatcher
	policy JitterPolicy
	rng    *rand.Rand
}

// NewJitterDispatcher creates a JitterDispatcher that delays each Send call by
// a random duration in [0, policy.MaxJitter) before calling next.
// Panics if next is nil.
func NewJitterDispatcher(next Dispatcher, policy JitterPolicy) *JitterDispatcher {
	if next == nil {
		panic("jitter: next dispatcher must not be nil")
	}
	if policy.MaxJitter <= 0 {
		policy.MaxJitter = DefaultJitterPolicy().MaxJitter
	}
	return &JitterDispatcher{
		next:   next,
		policy: policy,
		// #nosec G404 — non-cryptographic jitter is intentional
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Send waits for a random jitter interval then forwards n to the wrapped
// dispatcher. If the context is cancelled during the wait, Send returns the
// context error immediately without forwarding.
func (j *JitterDispatcher) Send(ctx context.Context, n Notification) error {
	delay := time.Duration(j.rng.Int63n(int64(j.policy.MaxJitter)))
	select {
	case <-time.After(delay):
		return j.next.Send(ctx, n)
	case <-ctx.Done():
		return ctx.Err()
	}
}
