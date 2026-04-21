package alert

import (
	"context"
	"errors"
	"testing"
)

func tapNotif(port int) Notification {
	return Notification{Port: port, State: "open"}
}

func TestTap_DefaultCapacity(t *testing.T) {
	tap := NewTap(nil)
	if tap.policy.MaxCapacity != 256 {
		t.Fatalf("expected 256, got %d", tap.policy.MaxCapacity)
	}
}

func TestTap_RecordAndSnapshot(t *testing.T) {
	tap := NewTap(nil)
	tap.record(tapNotif(8080))
	tap.record(tapNotif(9090))

	snap := tap.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	if snap[0].Port != 8080 || snap[1].Port != 9090 {
		t.Fatalf("unexpected order: %+v", snap)
	}
}

func TestTap_Overflow(t *testing.T) {
	cap := 3
	tap := NewTap(&TapPolicy{MaxCapacity: cap})
	for i := 0; i < 5; i++ {
		tap.record(tapNotif(i))
	}
	if tap.Len() != cap {
		t.Fatalf("expected %d, got %d", cap, tap.Len())
	}
	// oldest entries should have been evicted; first retained port is 2
	snap := tap.Snapshot()
	if snap[0].Port != 2 {
		t.Fatalf("expected port 2, got %d", snap[0].Port)
	}
}

func TestTap_Reset(t *testing.T) {
	tap := NewTap(nil)
	tap.record(tapNotif(1234))
	tap.Reset()
	if tap.Len() != 0 {
		t.Fatal("expected empty tap after reset")
	}
}

func TestTapDispatcher_NilTapPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil tap")
		}
	}()
	NewTapDispatcher(nil, &captureDispatcher{})
}

func TestTapDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewTapDispatcher(NewTap(nil), nil)
}

func TestTapDispatcher_RecordsAndForwards(t *testing.T) {
	tap := NewTap(nil)
	cap := &captureDispatcher{}
	d := NewTapDispatcher(tap, cap)

	n := tapNotif(8080)
	if err := d.Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tap.Len() != 1 {
		t.Fatalf("expected 1 recorded notification, got %d", tap.Len())
	}
	if len(cap.received) != 1 {
		t.Fatalf("expected notification forwarded to next")
	}
}

func TestTapDispatcher_PropagatesError(t *testing.T) {
	tap := NewTap(nil)
	sentinel := errors.New("downstream error")
	next := &errorDispatcher{err: sentinel}
	d := NewTapDispatcher(tap, next)

	err := d.Send(context.Background(), tapNotif(9090))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	// notification should still have been recorded
	if tap.Len() != 1 {
		t.Fatalf("expected tap to record even on downstream error")
	}
}
