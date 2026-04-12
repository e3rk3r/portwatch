package alert

import (
	"fmt"
	"sync"
	"time"
)

// ShadowPolicy controls how the shadow dispatcher compares primary vs secondary.
type ShadowPolicy struct {
	// LogDivergence, when true, records mismatches between primary and shadow.
	LogDivergence bool
}

// DefaultShadowPolicy returns a sensible default.
func DefaultShadowPolicy() ShadowPolicy {
	return ShadowPolicy{LogDivergence: true}
}

// Divergence records a single mismatch between primary and shadow dispatchers.
type Divergence struct {
	Key       string
	Primary   error
	Shadow    error
	RecordedAt time.Time
}

// ShadowDispatcher sends every notification to both a primary and a shadow
// dispatcher. The shadow runs asynchronously and its errors never affect the
// primary result. Divergences (where exactly one side errors) are optionally
// recorded for later inspection.
type ShadowDispatcher struct {
	primary   Dispatcher
	shadow    Dispatcher
	policy    ShadowPolicy
	clock     func() time.Time

	mu          sync.Mutex
	divergences []Divergence
}

// NewShadowDispatcher creates a ShadowDispatcher wrapping primary and shadow.
func NewShadowDispatcher(primary, shadow Dispatcher, policy ShadowPolicy) *ShadowDispatcher {
	if primary == nil {
		panic("shadow: primary dispatcher must not be nil")
	}
	if shadow == nil {
		panic("shadow: shadow dispatcher must not be nil")
	}
	return &ShadowDispatcher{
		primary: primary,
		shadow:  shadow,
		policy:  policy,
		clock:   time.Now,
	}
}

// Send dispatches to the primary synchronously and to the shadow asynchronously.
func (s *ShadowDispatcher) Send(n Notification) error {
	primaryErr := s.primary.Send(n)

	go func() {
		shadowErr := s.shadow.Send(n)
		if s.policy.LogDivergence {
			if (primaryErr == nil) != (shadowErr == nil) {
				s.mu.Lock()
				s.divergences = append(s.divergences, Divergence{
					Key:        fmt.Sprintf("%d:%s", n.Port, n.State),
					Primary:    primaryErr,
					Shadow:     shadowErr,
					RecordedAt: s.clock(),
				})
				s.mu.Unlock()
			}
		}
	}()

	return primaryErr
}

// Divergences returns a snapshot of all recorded divergences.
func (s *ShadowDispatcher) Divergences() []Divergence {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Divergence, len(s.divergences))
	copy(out, s.divergences)
	return out
}

// Reset clears recorded divergences.
func (s *ShadowDispatcher) Reset() {
	s.mu.Lock()
	s.divergences = s.divergences[:0]
	s.mu.Unlock()
}
