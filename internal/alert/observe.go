package alert

import (
	"sync"
	"time"
)

// ObservePolicy configures the metrics observer.
type ObservePolicy struct {
	// LatencyBuckets are histogram bucket boundaries in milliseconds.
	// Defaults to [5, 10, 25, 50, 100, 250, 500, 1000].
	LatencyBuckets []float64
}

// DefaultObservePolicy returns a sensible default policy.
func DefaultObservePolicy() ObservePolicy {
	return ObservePolicy{
		LatencyBuckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	}
}

// ObserveSnapshot holds a point-in-time view of dispatcher metrics.
type ObserveSnapshot struct {
	Total    int64
	Errors   int64
	Dropped  int64
	LatencyP50Ms float64
	LatencyP99Ms float64
}

// Observer records call counts and latency for a Dispatcher.
type Observer struct {
	mu       sync.Mutex
	policy   ObservePolicy
	total    int64
	errors   int64
	dropped  int64
	samples  []float64
}

// NewObserver creates an Observer with the given policy.
func NewObserver(p ObservePolicy) *Observer {
	if len(p.LatencyBuckets) == 0 {
		p = DefaultObservePolicy()
	}
	return &Observer{policy: p}
}

// Record registers the outcome and latency of a single dispatch attempt.
func (o *Observer) Record(err error, latency time.Duration, dropped bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.total++
	if dropped {
		o.dropped++
		return
	}
	if err != nil {
		o.errors++
	}
	o.samples = append(o.samples, float64(latency.Milliseconds()))
}

// Snapshot returns a copy of the current metrics.
func (o *Observer) Snapshot() ObserveSnapshot {
	o.mu.Lock()
	defer o.mu.Unlock()
	snap := ObserveSnapshot{
		Total:   o.total,
		Errors:  o.errors,
		Dropped: o.dropped,
	}
	if len(o.samples) > 0 {
		sorted := make([]float64, len(o.samples))
		copy(sorted, o.samples)
		sortFloat64s(sorted)
		snap.LatencyP50Ms = percentile(sorted, 50)
		snap.LatencyP99Ms = percentile(sorted, 99)
	}
	return snap
}

// Reset clears all recorded metrics.
func (o *Observer) Reset() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.total, o.errors, o.dropped = 0, 0, 0
	o.samples = o.samples[:0]
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p / 100.0)
	return sorted[idx]
}

func sortFloat64s(s []float64) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
