package alert

import (
	"testing"
	"time"
)

func fixedEscalationClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestEscalator_FirstRecordNoEscalation(t *testing.T) {
	e := NewEscalator(DefaultEscalationPolicy())
	if e.Record("port:8080:closed") {
		t.Fatal("expected no escalation on first record")
	}
}

func TestEscalator_BelowThreshold(t *testing.T) {
	e := NewEscalator(EscalationPolicy{Threshold: 3, MinDuration: 0})
	e.Record("k")
	if e.Record("k") {
		t.Fatal("expected no escalation below threshold")
	}
}

func TestEscalator_ThresholdWithoutDuration(t *testing.T) {
	base := time.Now()
	e := NewEscalator(EscalationPolicy{Threshold: 2, MinDuration: 10 * time.Minute})
	e.clock = fixedEscalationClock(base)
	e.Record("k")
	// Second call still at same time — duration not met
	if e.Record("k") {
		t.Fatal("expected no escalation when duration not met")
	}
}

func TestEscalator_ThresholdAndDurationMet(t *testing.T) {
	base := time.Now()
	e := NewEscalator(EscalationPolicy{Threshold: 2, MinDuration: 5 * time.Minute})
	e.clock = fixedEscalationClock(base)
	e.Record("k")
	// Advance clock past MinDuration
	e.clock = fixedEscalationClock(base.Add(6 * time.Minute))
	if !e.Record("k") {
		t.Fatal("expected escalation when threshold and duration both met")
	}
}

func TestEscalator_ResetClearsState(t *testing.T) {
	e := NewEscalator(EscalationPolicy{Threshold: 1, MinDuration: 0})
	e.Record("k")
	e.Reset("k")
	if e.Count("k") != 0 {
		t.Fatalf("expected count 0 after reset, got %d", e.Count("k"))
	}
}

func TestEscalator_IndependentKeys(t *testing.T) {
	e := NewEscalator(EscalationPolicy{Threshold: 2, MinDuration: 0})
	e.Record("a")
	e.Record("b")
	if e.Count("a") != 1 {
		t.Fatalf("expected count 1 for 'a', got %d", e.Count("a"))
	}
	if e.Count("b") != 1 {
		t.Fatalf("expected count 1 for 'b', got %d", e.Count("b"))
	}
}

func TestEscalator_CountUnknownKey(t *testing.T) {
	e := NewEscalator(DefaultEscalationPolicy())
	if c := e.Count("missing"); c != 0 {
		t.Fatalf("expected 0 for unknown key, got %d", c)
	}
}
