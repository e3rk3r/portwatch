package alert

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// countingDispatcher records how many times Send was called.
type countingDispatcher struct {
	calls int32
	err   error
}

func (c *countingDispatcher) Send(_ context.Context, _ Notification) error {
	atomic.AddInt32(&c.calls, 1)
	return c.err
}

func jitterNotif() Notification {
	return BuildNotification(8080, "open", "test-host")
}

func TestJitterDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next dispatcher")
		}
	}()
	NewJitterDispatcher(nil, DefaultJitterPolicy())
}

func TestJitterDispatcher_ForwardsNotification(t *testing.T) {
	cd := &countingDispatcher{}
	policy := JitterPolicy{MaxJitter: 10 * time.Millisecond}
	d := NewJitterDispatcher(cd, policy)

	err := d.Send(context.Background(), jitterNotif())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&cd.calls) != 1 {
		t.Fatalf("expected 1 call to next, got %d", cd.calls)
	}
}

func TestJitterDispatcher_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	cd := &countingDispatcher{err: sentinel}
	policy := JitterPolicy{MaxJitter: 5 * time.Millisecond}
	d := NewJitterDispatcher(cd, policy)

	err := d.Send(context.Background(), jitterNotif())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestJitterDispatcher_CancelledContextReturnsEarly(t *testing.T) {
	cd := &countingDispatcher{}
	// Large jitter so the context cancel fires first.
	policy := JitterPolicy{MaxJitter: 10 * time.Second}
	d := NewJitterDispatcher(cd, policy)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := d.Send(ctx, jitterNotif())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if atomic.LoadInt32(&cd.calls) != 0 {
		t.Fatal("next dispatcher should not have been called")
	}
}

func TestJitterDispatcher_DefaultPolicyAppliedOnZeroMaxJitter(t *testing.T) {
	cd := &countingDispatcher{}
	// Zero MaxJitter should fall back to default.
	d := NewJitterDispatcher(cd, JitterPolicy{MaxJitter: 0})
	if d.policy.MaxJitter != DefaultJitterPolicy().MaxJitter {
		t.Fatalf("expected default MaxJitter %v, got %v",
			DefaultJitterPolicy().MaxJitter, d.policy.MaxJitter)
	}
}
