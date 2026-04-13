package alert

import (
	"sync"
	"time"
)

// WindowPolicy configures the sliding-window event counter.
type WindowPolicy struct {
	// Size is the duration of the sliding window.
	Size time.Duration
	// MaxEvents is the maximum number of events allowed within the window.
	MaxEvents int
}

// DefaultWindowPolicy returns a sensible default: 10 events per minute.
func DefaultWindowPolicy() WindowPolicy {
	return WindowPolicy{
		Size:      time.Minute,
		MaxEvents: 10,
	}
}

// WindowCounter tracks event timestamps within a sliding window.
type WindowCounter struct {
	policy WindowPolicy
	mu     sync.Mutex
	// timestamps maps a string key to a list of event times.
	timestamps map[string][]time.Time
	clock      func() time.Time
}

// NewWindowCounter creates a WindowCounter with the given policy.
func NewWindowCounter(p WindowPolicy) *WindowCounter {
	if p.Size <= 0 {
		p.Size = DefaultWindowPolicy().Size
	}
	if p.MaxEvents <= 0 {
		p.MaxEvents = DefaultWindowPolicy().MaxEvents
	}
	return &WindowCounter{
		policy:     p,
		timestamps: make(map[string][]time.Time),
		clock:      time.Now,
	}
}

// Allow returns true and records the event if the key is below the window limit.
// It prunes timestamps outside the current window before deciding.
func (w *WindowCounter) Allow(key string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.clock()
	cutoff := now.Add(-w.policy.Size)

	ts := w.timestamps[key]
	pruned := ts[:0]
	for _, t := range ts {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}

	if len(pruned) >= w.policy.MaxEvents {
		w.timestamps[key] = pruned
		return false
	}

	w.timestamps[key] = append(pruned, now)
	return true
}

// Reset clears all recorded timestamps for the given key.
func (w *WindowCounter) Reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.timestamps, key)
}

// Count returns the number of events recorded within the current window for key.
func (w *WindowCounter) Count(key string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.clock()
	cutoff := now.Add(-w.policy.Size)
	count := 0
	for _, t := range w.timestamps[key] {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// Keys returns all keys that currently have at least one event within the window.
func (w *WindowCounter) Keys() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.clock()
	cutoff := now.Add(-w.policy.Size)

	keys := make([]string, 0, len(w.timestamps))
	for key, ts := range w.timestamps {
		for _, t := range ts {
			if t.After(cutoff) {
				keys = append(keys, key)
				break
			}
		}
	}
	return keys
}
