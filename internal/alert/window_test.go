package alert

import (
	"testing"
	"time"
)

func fixedWindowClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestWindowCounter_AllowBelowLimit(t *testing.T) {
	w := NewWindowCounter(WindowPolicy{Size: time.Minute, MaxEvents: 3})
	for i := 0; i < 3; i++ {
		if !w.Allow("k") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
}

func TestWindowCounter_BlocksAtLimit(t *testing.T) {
	w := NewWindowCounter(WindowPolicy{Size: time.Minute, MaxEvents: 2})
	w.Allow("k")
	w.Allow("k")
	if w.Allow("k") {
		t.Fatal("expected Allow=false after limit reached")
	}
}

func TestWindowCounter_PrunesExpired(t *testing.T) {
	now := time.Now()
	w := NewWindowCounter(WindowPolicy{Size: time.Second, MaxEvents: 2})
	w.clock = fixedWindowClock(now)
	w.Allow("k")
	w.Allow("k")
	// advance past the window
	w.clock = fixedWindowClock(now.Add(2 * time.Second))
	if !w.Allow("k") {
		t.Fatal("expected Allow=true after window expires")
	}
}

func TestWindowCounter_Reset(t *testing.T) {
	w := NewWindowCounter(WindowPolicy{Size: time.Minute, MaxEvents: 1})
	w.Allow("k")
	w.Reset("k")
	if !w.Allow("k") {
		t.Fatal("expected Allow=true after Reset")
	}
}

func TestWindowCounter_Count(t *testing.T) {
	w := NewWindowCounter(WindowPolicy{Size: time.Minute, MaxEvents: 10})
	w.Allow("k")
	w.Allow("k")
	if got := w.Count("k"); got != 2 {
		t.Fatalf("expected Count=2, got %d", got)
	}
}

func TestWindowCounter_IndependentKeys(t *testing.T) {
	w := NewWindowCounter(WindowPolicy{Size: time.Minute, MaxEvents: 1})
	w.Allow("a")
	if !w.Allow("b") {
		t.Fatal("expected Allow=true for independent key")
	}
}

func TestWindowCounter_DefaultPolicy(t *testing.T) {
	w := NewWindowCounter(WindowPolicy{})
	if w.policy.MaxEvents != 10 {
		t.Fatalf("expected default MaxEvents=10, got %d", w.policy.MaxEvents)
	}
	if w.policy.Size != time.Minute {
		t.Fatalf("expected default Size=1m, got %v", w.policy.Size)
	}
}
