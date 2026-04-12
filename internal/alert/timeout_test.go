package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

// slowDispatcher blocks for the given duration before returning the given error.
type slowDispatcher struct {
	delay time.Duration
	err   error
}

func (s *slowDispatcher) Send(ctx context.Context, _ Notification) error {
	select {
	case <-time.After(s.delay):
		return s.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func timeoutNotif() Notification {
	return Notification{Port: 8080, State: "open"}
}

func TestTimeoutDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewTimeoutDispatcher(nil, DefaultTimeoutPolicy())
}

func TestTimeoutDispatcher_FastDispatchSucceeds(t *testing.T) {
	next := &slowDispatcher{delay: 1 * time.Millisecond}
	td := NewTimeoutDispatcher(next, TimeoutPolicy{PerDispatch: 100 * time.Millisecond})

	if err := td.Send(context.Background(), timeoutNotif()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestTimeoutDispatcher_SlowDispatchTimesOut(t *testing.T) {
	next := &slowDispatcher{delay: 200 * time.Millisecond}
	td := NewTimeoutDispatcher(next, TimeoutPolicy{PerDispatch: 20 * time.Millisecond})

	err := td.Send(context.Background(), timeoutNotif())
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestTimeoutDispatcher_PropagatesNextError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	next := &slowDispatcher{delay: 1 * time.Millisecond, err: sentinel}
	td := NewTimeoutDispatcher(next, TimeoutPolicy{PerDispatch: 100 * time.Millisecond})

	err := td.Send(context.Background(), timeoutNotif())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestTimeoutDispatcher_ZeroDurationUsesDefault(t *testing.T) {
	next := &slowDispatcher{delay: 1 * time.Millisecond}
	td := NewTimeoutDispatcher(next, TimeoutPolicy{PerDispatch: 0})

	if td.policy.PerDispatch != DefaultTimeoutPolicy().PerDispatch {
		t.Fatalf("expected default duration %v, got %v",
			DefaultTimeoutPolicy().PerDispatch, td.policy.PerDispatch)
	}
}
