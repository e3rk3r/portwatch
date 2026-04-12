package alert

import (
	"context"
	"errors"
	"testing"
)

func makeReplayDispatcher(dst Dispatcher) (*ReplayDispatcher, *Replayer) {
	r := NewReplayer(DefaultReplayPolicy())
	return NewReplayDispatcher(dst, r), r
}

func TestReplayDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewReplayDispatcher(nil, NewReplayer(DefaultReplayPolicy()))
}

func TestReplayDispatcher_NilReplayerPanics(t *testing.T) {
	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("expected panic for nil replayer")
		}
	}()
	NewReplayDispatcher(dispatcherFunc(func(_ context.Context, _ Notification) error { return nil }), nil)
}

func TestReplayDispatcher_SuccessRecordsEntry(t *testing.T) {
	dst := dispatcherFunc(func(_ context.Context, _ Notification) error { return nil })
	rd, replayer := makeReplayDispatcher(dst)

	if err := rd.Send(context.Background(), replayNotif(8080)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if replayer.Len() != 1 {
		t.Fatalf("expected 1 buffered entry, got %d", replayer.Len())
	}
}

func TestReplayDispatcher_FailureRecordsEntry(t *testing.T) {
	dst := dispatcherFunc(func(_ context.Context, _ Notification) error {
		return errors.New("fail")
	})
	rd, replayer := makeReplayDispatcher(dst)

	if err := rd.Send(context.Background(), replayNotif(9090)); err == nil {
		t.Fatal("expected error from downstream")
	}
	if replayer.Len() != 1 {
		t.Fatalf("expected 1 buffered entry after failure, got %d", replayer.Len())
	}
}

func TestReplayDispatcher_ReplayAfterRecovery(t *testing.T) {
	var attempt int
	dst := dispatcherFunc(func(_ context.Context, _ Notification) error {
		attempt++
		if attempt == 1 {
			return errors.New("down")
		}
		return nil
	})
	rd, replayer := makeReplayDispatcher(dst)

	// First send fails → buffered.
	_ = rd.Send(context.Background(), replayNotif(8080))
	if replayer.Len() != 1 {
		t.Fatalf("expected 1 entry buffered")
	}

	// Downstream recovers; replay should succeed.
	if err := replayer.Replay(context.Background(), dst); err != nil {
		t.Fatalf("replay error: %v", err)
	}
	if replayer.Len() != 0 {
		t.Fatalf("buffer should be empty after replay")
	}
}
