package alert

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func replayNotif(port int) Notification {
	return Notification{Port: port, State: "open"}
}

func fixedReplayClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestReplayer_RecordAndLen(t *testing.T) {
	r := NewReplayer(DefaultReplayPolicy())
	if r.Len() != 0 {
		t.Fatalf("expected 0, got %d", r.Len())
	}
	r.Record(replayNotif(8080))
	r.Record(replayNotif(9090))
	if r.Len() != 2 {
		t.Fatalf("expected 2, got %d", r.Len())
	}
}

func TestReplayer_Overflow(t *testing.T) {
	p := ReplayPolicy{MaxEvents: 3, MaxAge: time.Minute}
	r := NewReplayer(p)
	for i := 0; i < 5; i++ {
		r.Record(replayNotif(8000 + i))
	}
	if r.Len() != 3 {
		t.Fatalf("expected 3, got %d", r.Len())
	}
}

func TestReplayer_ReplaySuccess(t *testing.T) {
	r := NewReplayer(DefaultReplayPolicy())
	r.Record(replayNotif(8080))
	r.Record(replayNotif(9090))

	var calls int32
	dst := dispatcherFunc(func(_ context.Context, _ Notification) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	if err := r.Replay(context.Background(), dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
	if r.Len() != 0 {
		t.Fatalf("buffer should be empty after successful replay")
	}
}

func TestReplayer_ReplayFailureRetained(t *testing.T) {
	r := NewReplayer(DefaultReplayPolicy())
	r.Record(replayNotif(8080))

	dst := dispatcherFunc(func(_ context.Context, _ Notification) error {
		return errors.New("downstream down")
	})

	if err := r.Replay(context.Background(), dst); err == nil {
		t.Fatal("expected error")
	}
	if r.Len() != 1 {
		t.Fatalf("failed entry should be retained, got len=%d", r.Len())
	}
}

func TestReplayer_ExpiredEntriesDropped(t *testing.T) {
	now := time.Now()
	p := ReplayPolicy{MaxEvents: 10, MaxAge: time.Minute}
	r := NewReplayer(p)
	r.clock = fixedReplayClock(now.Add(-2 * time.Minute)) // stored 2 min ago
	r.Record(replayNotif(8080))
	r.clock = fixedReplayClock(now) // replay happens now

	var calls int32
	dst := dispatcherFunc(func(_ context.Context, _ Notification) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	if err := r.Replay(context.Background(), dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 0 {
		t.Fatalf("expired entry should not be replayed")
	}
}
