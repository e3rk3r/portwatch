package alert

import (
	"testing"
	"time"
)

func TestObserver_DefaultPolicy(t *testing.T) {
	obs := NewObserver(ObservePolicy{})
	if len(obs.policy.LatencyBuckets) == 0 {
		t.Fatal("expected default buckets to be set")
	}
}

func TestObserver_RecordSuccess(t *testing.T) {
	obs := NewObserver(DefaultObservePolicy())
	obs.Record(nil, 10*time.Millisecond, false)
	obs.Record(nil, 20*time.Millisecond, false)

	snap := obs.Snapshot()
	if snap.Total != 2 {
		t.Fatalf("expected total=2, got %d", snap.Total)
	}
	if snap.Errors != 0 {
		t.Fatalf("expected errors=0, got %d", snap.Errors)
	}
	if snap.LatencyP50Ms == 0 {
		t.Fatal("expected non-zero p50")
	}
}

func TestObserver_RecordError(t *testing.T) {
	obs := NewObserver(DefaultObservePolicy())
	obs.Record(errTest, 5*time.Millisecond, false)

	snap := obs.Snapshot()
	if snap.Total != 1 {
		t.Fatalf("expected total=1, got %d", snap.Total)
	}
	if snap.Errors != 1 {
		t.Fatalf("expected errors=1, got %d", snap.Errors)
	}
}

func TestObserver_RecordDropped(t *testing.T) {
	obs := NewObserver(DefaultObservePolicy())
	obs.Record(nil, 0, true)

	snap := obs.Snapshot()
	if snap.Total != 1 {
		t.Fatalf("expected total=1, got %d", snap.Total)
	}
	if snap.Dropped != 1 {
		t.Fatalf("expected dropped=1, got %d", snap.Dropped)
	}
	if len(obs.samples) != 0 {
		t.Fatal("dropped events must not contribute to latency samples")
	}
}

func TestObserver_Reset(t *testing.T) {
	obs := NewObserver(DefaultObservePolicy())
	obs.Record(nil, 15*time.Millisecond, false)
	obs.Reset()

	snap := obs.Snapshot()
	if snap.Total != 0 || snap.Errors != 0 || snap.Dropped != 0 {
		t.Fatal("expected zeroed snapshot after Reset")
	}
}

func TestObserver_P99GreaterThanP50(t *testing.T) {
	obs := NewObserver(DefaultObservePolicy())
	for i := 0; i < 100; i++ {
		obs.Record(nil, time.Duration(i)*time.Millisecond, false)
	}
	snap := obs.Snapshot()
	if snap.LatencyP99Ms <= snap.LatencyP50Ms {
		t.Fatalf("expected p99 (%v) > p50 (%v)", snap.LatencyP99Ms, snap.LatencyP50Ms)
	}
}
