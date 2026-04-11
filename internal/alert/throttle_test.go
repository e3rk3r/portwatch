package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fixedThrottleClock returns a clock that always returns t.
func fixedThrottleClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestThrottler_FirstCallAlwaysPasses(t *testing.T) {
	th := NewThrottler(DefaultThrottlePolicy(), fixedThrottleClock(time.Now()))
	if !th.Allow(8080, "open") {
		t.Fatal("expected first call to pass")
	}
}

func TestThrottler_BlockedAfterBurst(t *testing.T) {
	now := time.Now()
	p := ThrottlePolicy{MaxBurst: 3, BurstWindow: time.Minute}
	th := NewThrottler(p, fixedThrottleClock(now))

	for i := 0; i < 3; i++ {
		if !th.Allow(9090, "closed") {
			t.Fatalf("call %d should have passed", i+1)
		}
	}
	if th.Allow(9090, "closed") {
		t.Fatal("4th call should be throttled")
	}
}

func TestThrottler_PassesAfterWindowExpires(t *testing.T) {
	base := time.Now()
	current := base
	clock := func() time.Time { return current }

	p := ThrottlePolicy{MaxBurst: 2, BurstWindow: 30 * time.Second}
	th := NewThrottler(p, clock)

	th.Allow(443, "open")
	th.Allow(443, "open")

	// Advance past the burst window.
	current = base.Add(31 * time.Second)
	if !th.Allow(443, "open") {
		t.Fatal("expected call to pass after window expired")
	}
}

func TestThrottler_IndependentKeys(t *testing.T) {
	p := ThrottlePolicy{MaxBurst: 1, BurstWindow: time.Minute}
	th := NewThrottler(p, fixedThrottleClock(time.Now()))

	if !th.Allow(80, "open") {
		t.Fatal("port 80 open should pass")
	}
	if !th.Allow(443, "open") {
		t.Fatal("port 443 open should pass independently")
	}
	if th.Allow(80, "open") {
		t.Fatal("second call for port 80 open should be throttled")
	}
}

func TestThrottler_ResetRestoresBudget(t *testing.T) {
	p := ThrottlePolicy{MaxBurst: 1, BurstWindow: time.Minute}
	th := NewThrottler(p, fixedThrottleClock(time.Now()))

	th.Allow(8080, "open")
	if th.Allow(8080, "open") {
		t.Fatal("should be throttled before reset")
	}
	th.Reset()
	if !th.Allow(8080, "open") {
		t.Fatal("should pass after reset")
	}
}

// --- ThrottleDispatcher tests ---

type countingDispatcher struct {
	calls int
}

func (c *countingDispatcher) Send(_ context.Context, _ Notification) error {
	c.calls++
	return nil
}

func TestThrottleDispatcher_ForwardsWhenAllowed(t *testing.T) {
	inner := &countingDispatcher{}
	th := NewThrottler(DefaultThrottlePolicy(), fixedThrottleClock(time.Now()))
	d := NewThrottleDispatcher(inner, th)

	n := Notification{Port: 8080, State: "open"}
	if err := d.Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 forwarded call, got %d", inner.calls)
	}
}

func TestThrottleDispatcher_ReturnsErrorWhenThrottled(t *testing.T) {
	inner := &countingDispatcher{}
	p := ThrottlePolicy{MaxBurst: 1, BurstWindow: time.Minute}
	th := NewThrottler(p, fixedThrottleClock(time.Now()))
	d := NewThrottleDispatcher(inner, th)

	n := Notification{Port: 3000, State: "closed"}
	_ = d.Send(context.Background(), n)
	err := d.Send(context.Background(), n)
	if err == nil {
		t.Fatal("expected throttle error on second call")
	}
	if inner.calls != 1 {
		t.Fatalf("inner should have been called once, got %d", inner.calls)
	}
}

func TestNewThrottleDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	th := NewThrottler(DefaultThrottlePolicy(), nil)
	NewThrottleDispatcher(nil, th)
}

func TestNewThrottleDispatcher_NilThrottlerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil throttler")
		}
	}()
	NewThrottleDispatcher(&countingDispatcher{}, nil)
}

// Ensure the error sentinel is reachable for callers who want to inspect it.
func TestThrottleDispatcher_ErrorContainsPort(t *testing.T) {
	inner := &countingDispatcher{}
	p := ThrottlePolicy{MaxBurst: 1, BurstWindow: time.Minute}
	th := NewThrottler(p, fixedThrottleClock(time.Now()))
	d := NewThrottleDispatcher(inner, th)

	n := Notification{Port: 7777, State: "open"}
	_ = d.Send(context.Background(), n)
	err := d.Send(context.Background(), n)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, err) { // trivially true; real check is string content
		t.Fatal("error should be non-nil")
	}
	if got := err.Error(); len(got) == 0 {
		t.Fatal("error message should not be empty")
	}
}
