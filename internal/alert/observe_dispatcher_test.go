package alert

import (
	"context"
	"errors"
	"testing"
)

func TestObserveDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewObserveDispatcher(nil, NewObserver(DefaultObservePolicy()))
}

func TestObserveDispatcher_NilObserverPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil observer")
		}
	}()
	NewObserveDispatcher(&captureDispatcher{}, nil)
}

func TestObserveDispatcher_SuccessRecorded(t *testing.T) {
	cap := &captureDispatcher{}
	obs := NewObserver(DefaultObservePolicy())
	d := NewObserveDispatcher(cap, obs)

	n := Notification{Port: 8080, State: "open"}
	if err := d.Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := obs.Snapshot()
	if snap.Total != 1 {
		t.Fatalf("expected total=1, got %d", snap.Total)
	}
	if snap.Errors != 0 {
		t.Fatalf("expected errors=0, got %d", snap.Errors)
	}
	if !cap.called {
		t.Fatal("expected next dispatcher to be called")
	}
}

func TestObserveDispatcher_ErrorRecorded(t *testing.T) {
	sendErr := errors.New("downstream failure")
	next := &errorDispatcher{err: sendErr}
	obs := NewObserver(DefaultObservePolicy())
	d := NewObserveDispatcher(next, obs)

	n := Notification{Port: 9090, State: "closed"}
	err := d.Send(context.Background(), n)
	if !errors.Is(err, sendErr) {
		t.Fatalf("expected original error, got %v", err)
	}

	snap := obs.Snapshot()
	if snap.Total != 1 {
		t.Fatalf("expected total=1, got %d", snap.Total)
	}
	if snap.Errors != 1 {
		t.Fatalf("expected errors=1, got %d", snap.Errors)
	}
}

func TestObserveDispatcher_TransparentPassThrough(t *testing.T) {
	cap := &captureDispatcher{}
	obs := NewObserver(DefaultObservePolicy())
	d := NewObserveDispatcher(cap, obs)

	n := Notification{Port: 443, State: "open"}
	d.Send(context.Background(), n) //nolint:errcheck

	if cap.last.Port != 443 {
		t.Fatalf("expected notification forwarded unchanged, got port %d", cap.last.Port)
	}
}
