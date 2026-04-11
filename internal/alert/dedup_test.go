package alert

import (
	"testing"
	"time"
)

func fixedDedupClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestDeduplicator_FirstCallAlwaysPasses(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(DefaultDedupPolicy(), fixedDedupClock(now))
	if !d.Allow("8080:open") {
		t.Fatal("expected first call to pass")
	}
}

func TestDeduplicator_BlockedWithinWindow(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(DefaultDedupPolicy(), fixedDedupClock(now))
	d.Allow("8080:open")
	if d.Allow("8080:open") {
		t.Fatal("expected second call within window to be blocked")
	}
}

func TestDeduplicator_PassesAfterWindow(t *testing.T) {
	now := time.Now()
	clock := fixedDedupClock(now)
	d := NewDeduplicator(DedupPolicy{Window: time.Minute}, clock)
	d.Allow("8080:open")
	// advance clock beyond window
	d.clock = fixedDedupClock(now.Add(2 * time.Minute))
	if !d.Allow("8080:open") {
		t.Fatal("expected call after window to pass")
	}
}

func TestDeduplicator_ResetAllowsNext(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(DefaultDedupPolicy(), fixedDedupClock(now))
	d.Allow("9090:closed")
	d.Reset("9090:closed")
	if !d.Allow("9090:closed") {
		t.Fatal("expected call after reset to pass")
	}
}

func TestDeduplicator_IndependentKeys(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(DefaultDedupPolicy(), fixedDedupClock(now))
	d.Allow("8080:open")
	if !d.Allow("9090:open") {
		t.Fatal("expected different key to pass independently")
	}
}

func TestDedupDispatcher_SuppressDuplicate(t *testing.T) {
	now := time.Now()
	ded := NewDeduplicator(DefaultDedupPolicy(), fixedDedupClock(now))
	var sent int
	next := DispatcherFunc(func(n Notification) error { sent++; return nil })
	dd := NewDedupDispatcher(next, ded)
	n := Notification{Port: 8080, State: "open"}
	_ = dd.Send(n)
	_ = dd.Send(n)
	if sent != 1 {
		t.Fatalf("expected 1 delivery, got %d", sent)
	}
}
