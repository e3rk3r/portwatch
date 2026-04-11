package alert

import (
	"fmt"
	"sync"
	"time"
)

// ThrottlePolicy holds configuration for the token-bucket throttler.
type ThrottlePolicy struct {
	// MaxBurst is the maximum number of notifications allowed in BurstWindow.
	MaxBurst int
	// BurstWindow is the rolling window over which MaxBurst is enforced.
	BurstWindow time.Duration
}

// DefaultThrottlePolicy returns a sensible default: 5 alerts per minute.
func DefaultThrottlePolicy() ThrottlePolicy {
	return ThrottlePolicy{
		MaxBurst:    5,
		BurstWindow: time.Minute,
	}
}

// Throttler tracks per-key sliding-window counters and decides whether a
// notification should be allowed through.
type Throttler struct {
	policy ThrottlePolicy
	clock  func() time.Time
	mu     sync.Mutex
	// timestamps holds the ring of recent event times per key.
	timestamps map[string][]time.Time
}

// NewThrottler creates a Throttler with the given policy.
// Pass nil for clock to use time.Now.
func NewThrottler(p ThrottlePolicy, clock func() time.Time) *Throttler {
	if clock == nil {
		clock = time.Now
	}
	return &Throttler{
		policy:     p,
		clock:      clock,
		timestamps: make(map[string][]time.Time),
	}
}

// throttleKey builds a map key from port and state.
func throttleKey(port int, state string) string {
	return fmt.Sprintf("%d:%s", port, state)
}

// Allow returns true if the notification for (port, state) is within the
// burst budget, recording the attempt in the process.
func (t *Throttler) Allow(port int, state string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.clock()
	cutoff := now.Add(-t.policy.BurstWindow)
	key := throttleKey(port, state)

	// Prune entries outside the window.
	ts := t.timestamps[key]
	valid := ts[:0]
	for _, v := range ts {
		if v.After(cutoff) {
			valid = append(valid, v)
		}
	}

	if len(valid) >= t.policy.MaxBurst {
		t.timestamps[key] = valid
		return false
	}

	t.timestamps[key] = append(valid, now)
	return true
}

// Reset clears the sliding window for all keys, restoring full burst budget.
func (t *Throttler) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timestamps = make(map[string][]time.Time)
}
