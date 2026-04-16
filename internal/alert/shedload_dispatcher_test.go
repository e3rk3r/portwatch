package alert

import (
	"context"
	"errors"
	"testing"
)

func TestShedDispatcher_PropagatesNextError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		return sentinel
	})
	d := NewShedDispatcher(ShedPolicy{MaxInFlight: 10}, next)
	err := d.Dispatch(context.Background(), Notification{Port: 443})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestShedDispatcher_ReleasesSlotOnError(t *testing.T) {
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		return errors.New("boom")
	})
	shedder := NewLoadShedder(ShedPolicy{MaxInFlight: 1})
	d := dispatcherFunc(func(ctx context.Context, n Notification) error {
		if err := shedder.Acquire(); err != nil {
			return err
		}
		defer shedder.Release()
		return next.Dispatch(ctx, n)
	})

	_ = d.Dispatch(context.Background(), Notification{Port: 80})
	if shedder.InFlight() != 0 {
		t.Fatalf("slot not released after error; in-flight=%d", shedder.InFlight())
	}
}

func TestShedDispatcher_PipelineIntegration(t *testing.T) {
	var calls int
	base := dispatcherFunc(func(_ context.Context, _ Notification) error {
		calls++
		return nil
	})
	pipe := NewPipeline(
		NewShedDispatcher(ShedPolicy{MaxInFlight: 8}, base),
	)
	for i := 0; i < 5; i++ {
		if err := pipe.Dispatch(context.Background(), Notification{Port: 3000}); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}
	if calls != 5 {
		t.Fatalf("expected 5 calls, got %d", calls)
	}
}
