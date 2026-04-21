package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

func fixedCheckpointClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func checkpointNotif(port int, state string) Notification {
	return Notification{Port: port, State: state}
}

func TestCheckpointer_MissOnEmpty(t *testing.T) {
	cp := NewCheckpointer(DefaultCheckpointPolicy(), nil)
	_, ok := cp.Latest(8080, "open")
	if ok {
		t.Fatal("expected miss on empty store")
	}
}

func TestCheckpointer_HitAfterRecord(t *testing.T) {
	cp := NewCheckpointer(DefaultCheckpointPolicy(), nil)
	n := checkpointNotif(8080, "open")
	cp.Record(n)
	got, ok := cp.Latest(8080, "open")
	if !ok {
		t.Fatal("expected hit after record")
	}
	if got.Port != 8080 || got.State != "open" {
		t.Fatalf("unexpected notification: %+v", got)
	}
}

func TestCheckpointer_ExpiredEntry(t *testing.T) {
	now := time.Now()
	cp := NewCheckpointer(CheckpointPolicy{MaxAge: time.Minute}, fixedCheckpointClock(now))
	cp.Record(checkpointNotif(9090, "closed"))

	// advance clock beyond MaxAge
	cp.clock = fixedCheckpointClock(now.Add(2 * time.Minute))
	_, ok := cp.Latest(9090, "closed")
	if ok {
		t.Fatal("expected miss for expired entry")
	}
}

func TestCheckpointer_IndependentKeys(t *testing.T) {
	cp := NewCheckpointer(DefaultCheckpointPolicy(), nil)
	cp.Record(checkpointNotif(80, "open"))
	cp.Record(checkpointNotif(80, "closed"))

	_, ok1 := cp.Latest(80, "open")
	_, ok2 := cp.Latest(80, "closed")
	if !ok1 || !ok2 {
		t.Fatal("both keys should be present independently")
	}
}

func TestCheckpointer_ResetClearsAll(t *testing.T) {
	cp := NewCheckpointer(DefaultCheckpointPolicy(), nil)
	cp.Record(checkpointNotif(443, "open"))
	cp.Reset()
	_, ok := cp.Latest(443, "open")
	if ok {
		t.Fatal("expected miss after reset")
	}
}

func TestCheckpointDispatcher_RecordsOnSuccess(t *testing.T) {
	cp := NewCheckpointer(DefaultCheckpointPolicy(), nil)
	next := &stubDispatcher{}
	d := NewCheckpointDispatcher(next, cp)

	n := checkpointNotif(8080, "open")
	if err := d.Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := cp.Latest(8080, "open")
	if !ok {
		t.Fatal("checkpoint should be recorded after successful send")
	}
}

func TestCheckpointDispatcher_SkipsRecordOnError(t *testing.T) {
	cp := NewCheckpointer(DefaultCheckpointPolicy(), nil)
	next := &stubDispatcher{err: errors.New("send failed")}
	d := NewCheckpointDispatcher(next, cp)

	_ = d.Send(context.Background(), checkpointNotif(8080, "open"))
	_, ok := cp.Latest(8080, "open")
	if ok {
		t.Fatal("checkpoint should NOT be recorded after failed send")
	}
}

func TestCheckpointDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	cp := NewCheckpointer(DefaultCheckpointPolicy(), nil)
	NewCheckpointDispatcher(nil, cp)
}

func TestCheckpointDispatcher_NilCheckpointerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil checkpointer")
		}
	}()
	NewCheckpointDispatcher(&stubDispatcher{}, nil)
}
