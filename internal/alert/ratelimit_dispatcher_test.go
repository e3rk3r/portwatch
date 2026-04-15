package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

func makeRateLimitDispatcher(maxBurst int, window time.Duration) (*RateLimitDispatcher, *captureDispatcher) {
	cap := &captureDispatcher{}
	policy := ThrottlePolicy{
		MaxBurst: maxBurst,
		Window:   window,
	}
	return NewRateLimitDispatcher(policy, cap), cap
}

func rateLimitNotif(port int) Notification {
	return Notification{Port: port, State: "open"}
}

func TestRateLimitDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewRateLimitDispatcher(DefaultThrottlePolicy(), nil)
}

func TestRateLimitDispatcher_AllowsWithinBurst(t *testing.T) {
	d, cap := makeRateLimitDispatcher(3, time.Minute)
	ctx := context.Background()
	n := rateLimitNotif(9090)

	for i := 0; i < 3; i++ {
		if err := d.Send(ctx, n); err != nil {
			t.Fatalf("call %d: unexpected error: %v", i+1, err)
		}
	}

	if cap.count != 3 {
		t.Fatalf("expected 3 forwarded, got %d", cap.count)
	}
}

func TestRateLimitDispatcher_BlocksAfterBurst(t *testing.T) {
	d, cap := makeRateLimitDispatcher(2, time.Minute)
	ctx := context.Background()
	n := rateLimitNotif(7070)

	// exhaust burst
	_ = d.Send(ctx, n)
	_ = d.Send(ctx, n)

	err := d.Send(ctx, n)
	if err == nil {
		t.Fatal("expected rate limit error on 3rd call")
	}
	if cap.count != 2 {
		t.Fatalf("expected 2 forwarded, got %d", cap.count)
	}
}

func TestRateLimitDispatcher_PropagatesDownstreamError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	fail := &errorDispatcher{err: sentinel}
	policy := DefaultThrottlePolicy()
	d := NewRateLimitDispatcher(policy, fail)

	err := d.Send(context.Background(), rateLimitNotif(1234))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
