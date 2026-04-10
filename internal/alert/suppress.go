// Package alert provides alerting primitives for portwatch.
package alert

import (
	"sync"
	"time"
)

// SuppressPolicy holds configuration for alert suppression windows.
type SuppressPolicy struct {
	// Windows is a list of time ranges (hour:minute) during which alerts are suppressed.
	Windows []TimeWindow
}

// TimeWindow represents a daily recurring suppression window.
type TimeWindow struct {
	Start time.Time // only hour/minute/second are used
	End   time.Time // only hour/minute/second are used
}

// DefaultSuppressPolicy returns a policy with no suppression windows.
func DefaultSuppressPolicy() SuppressPolicy {
	return SuppressPolicy{}
}

// Suppressor decides whether an alert should be suppressed based on the
// current wall-clock time and the configured suppression windows.
type Suppressor struct {
	mu     sync.Mutex
	policy SuppressPolicy
	clock  func() time.Time
}

// NewSuppressor creates a Suppressor with the given policy.
// If clock is nil, time.Now is used.
func NewSuppressor(policy SuppressPolicy, clock func() time.Time) *Suppressor {
	if clock == nil {
		clock = time.Now
	}
	return &Suppressor{policy: policy, clock: clock}
}

// IsSuppressed returns true if the current time falls within any configured
// suppression window.
func (s *Suppressor) IsSuppressed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	nowSecs := secondsOfDay(now)

	for _, w := range s.policy.Windows {
		start := secondsOfDay(w.Start)
		end := secondsOfDay(w.End)

		if start <= end {
			if nowSecs >= start && nowSecs < end {
				return true
			}
		} else {
			// window wraps midnight
			if nowSecs >= start || nowSecs < end {
				return true
			}
		}
	}
	return false
}

// UpdatePolicy replaces the suppression policy at runtime.
func (s *Suppressor) UpdatePolicy(p SuppressPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policy = p
}

func secondsOfDay(t time.Time) int {
	return t.Hour()*3600 + t.Minute()*60 + t.Second()
}
