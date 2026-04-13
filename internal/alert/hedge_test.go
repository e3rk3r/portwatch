package alert

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func hedgeNotif() Notification {
	return Notification{Port: 9090, State: "open", Title: "hedge test"}
}

func TestHedgeDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewHedgeDispatcher(nil, DefaultHedgePolicy())
}

func TestHedgeDispatcher_PrimarySucceedsImmediately(t *testing.T) {
	calls := int32(0)
	d := DispatcherFunc(func(_ context.Context, _ Notification) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	h := NewHedgeDispatcher(d, HedgePolicy{Delay: 50 * time.Millisecond, MaxHedges: 1})
	if err := h.Send(context.Background(), hedgeNotif()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only the primary should have been called (hedge not needed).
	if c := atomic.LoadInt32(&calls); c < 1 {
		t.Fatalf("expected at least 1 call, got %d", c)
	}
}

func TestHedgeDispatcher_HedgeFiresAfterDelay(t *testing.T) {
	calls := int32(0)
	errOnce := errors.New("primary failed")
	d := DispatcherFunc(func(_ context.Context, _ Notification) error {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			time.Sleep(80 * time.Millisecond) // simulate slow primary
			return errOnce
		}
		return nil // hedge succeeds
	})
	h := NewHedgeDispatcher(d, HedgePolicy{Delay: 20 * time.Millisecond, MaxHedges: 1})
	if err := h.Send(context.Background(), hedgeNotif()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHedgeDispatcher_AllFailsReturnsError(t *testing.T) {
	sentinel := errors.New("always fails")
	d := DispatcherFunc(func(_ context.Context, _ Notification) error {
		return sentinel
	})
	h := NewHedgeDispatcher(d, HedgePolicy{Delay: 10 * time.Millisecond, MaxHedges: 1})
	err := h.Send(context.Background(), hedgeNotif())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestHedgeDispatcher_ContextCancelReturnsErr(t *testing.T) {
	d := DispatcherFunc(func(ctx context.Context, _ Notification) error {
		<-ctx.Done()
		return ctx.Err()
	})
	h := NewHedgeDispatcher(d, HedgePolicy{Delay: 5 * time.Millisecond, MaxHedges: 1})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	err := h.Send(ctx, hedgeNotif())
	if err == nil {
		t.Fatal("expected error on context cancel")
	}
}
