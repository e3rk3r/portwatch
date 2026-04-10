package alert

import (
	"fmt"
	"testing"
	"time"
)

func fixedBackoffClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestBackoffLimiter_FirstCallAlwaysPasses(t *testing.T) {
	now := time.Now()
	bl := NewBackoffLimiter(DefaultBackoffPolicy(), fixedBackoffClock(now))
	if !bl.Allow("port:8080:open") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestBackoffLimiter_BlockedBeforeInterval(t *testing.T) {
	now := time.Now()
	bl := NewBackoffLimiter(DefaultBackoffPolicy(), fixedBackoffClock(now))
	bl.Allow("k")
	// Same instant — still within backoff window.
	if bl.Allow("k") {
		t.Fatal("expected call to be blocked within backoff window")
	}
}

func TestBackoffLimiter_PassesAfterInterval(t *testing.T) {
	base := 10 * time.Second
	p := BackoffPolicy{BaseInterval: base, Multiplier: 2.0, MaxInterval: 5 * time.Minute}
	now := time.Now()

	var current time.Time = now
	bl := NewBackoffLimiter(p, func() time.Time { return current })

	bl.Allow("k") // first call at t=0

	current = now.Add(base + time.Millisecond) // advance past base interval
	if !bl.Allow("k") {
		t.Fatal("expected call to pass after base interval")
	}
}

func TestBackoffLimiter_ExponentialGrowth(t *testing.T) {
	base := 10 * time.Second
	p := BackoffPolicy{BaseInterval: base, Multiplier: 2.0, MaxInterval: 10 * time.Minute}
	now := time.Now()
	current := now

	bl := NewBackoffLimiter(p, func() time.Time { return current })

	// Allow first call.
	bl.Allow("k")

	// Each subsequent allow should require a longer wait.
	expected := []time.Duration{base, base * 2, base * 4}
	for i, exp := range expected {
		current = current.Add(exp + time.Millisecond)
		if !bl.Allow("k") {
			t.Fatalf("step %d: expected allow after %v", i, exp)
		}
	}
}

func TestBackoffLimiter_MaxIntervalCapped(t *testing.T) {
	p := BackoffPolicy{BaseInterval: time.Second, Multiplier: 100.0, MaxInterval: 5 * time.Second}
	now := time.Now()
	current := now
	bl := NewBackoffLimiter(p, func() time.Time { return current })

	// Burn through several intervals to hit the cap.
	for i := 0; i < 5; i++ {
		current = current.Add(p.MaxInterval + time.Millisecond)
		bl.Allow("k")
	}

	// Next interval must not exceed MaxInterval.
	current = current.Add(p.MaxInterval + time.Millisecond)
	if !bl.Allow("k") {
		t.Fatal("expected allow after MaxInterval")
	}
}

func TestBackoffLimiter_IndependentKeys(t *testing.T) {
	now := time.Now()
	bl := NewBackoffLimiter(DefaultBackoffPolicy(), fixedBackoffClock(now))

	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("port:%d:open", 8080+i)
		if !bl.Allow(key) {
			t.Errorf("key %s: expected first call to be allowed", key)
		}
	}
}

func TestBackoffLimiter_ResetRestoresKey(t *testing.T) {
	now := time.Now()
	bl := NewBackoffLimiter(DefaultBackoffPolicy(), fixedBackoffClock(now))

	bl.Allow("k")
	if bl.Allow("k") {
		t.Fatal("expected block after first allow")
	}

	bl.Reset("k")
	if !bl.Allow("k") {
		t.Fatal("expected allow after reset")
	}
}
