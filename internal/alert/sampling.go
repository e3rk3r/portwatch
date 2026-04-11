package alert

import (
	"math/rand"
	"sync"
	"time"
)

// SamplePolicy controls probabilistic sampling of notifications.
type SamplePolicy struct {
	// Rate is the fraction of notifications to allow through [0.0, 1.0].
	// A rate of 1.0 means all notifications pass; 0.0 means none.
	Rate float64
}

// DefaultSamplePolicy passes every notification (100% sample rate).
func DefaultSamplePolicy() SamplePolicy {
	return SamplePolicy{Rate: 1.0}
}

// Sampler applies probabilistic sampling to notifications.
type Sampler struct {
	policy SamplePolicy
	rng    *rand.Rand
	mu     sync.Mutex
}

// NewSampler creates a Sampler with the given policy.
// If policy.Rate is outside [0.0, 1.0] it is clamped.
func NewSampler(policy SamplePolicy) *Sampler {
	rate := policy.Rate
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &Sampler{
		policy: SamplePolicy{Rate: rate},
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Allow returns true if the notification should be forwarded based on the
// configured sample rate. Thread-safe.
func (s *Sampler) Allow(_ Notification) bool {
	if s.policy.Rate >= 1.0 {
		return true
	}
	if s.policy.Rate <= 0.0 {
		return false
	}
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v < s.policy.Rate
}

// SampleDispatcher wraps a Dispatcher and drops notifications that fail the
// sampler's Allow check.
type SampleDispatcher struct {
	sampler *Sampler
	next    Dispatcher
}

// NewSampleDispatcher creates a SampleDispatcher.
// Panics if sampler or next is nil.
func NewSampleDispatcher(sampler *Sampler, next Dispatcher) *SampleDispatcher {
	if sampler == nil {
		panic("alert: NewSampleDispatcher: sampler must not be nil")
	}
	if next == nil {
		panic("alert: NewSampleDispatcher: next dispatcher must not be nil")
	}
	return &SampleDispatcher{sampler: sampler, next: next}
}

// Send forwards the notification to the next dispatcher only if the sampler
// allows it; otherwise the notification is silently dropped.
func (d *SampleDispatcher) Send(n Notification) error {
	if !d.sampler.Allow(n) {
		return nil
	}
	return d.next.Send(n)
}
