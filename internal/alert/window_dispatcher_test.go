package alert

import (
	"errors"
	"testing"
	"time"
)

func makeWindowDispatcher(maxEvents int) (*WindowDispatcher, *captureDispatcher) {
	cap := &captureDispatcher{}
	counter := NewWindowCounter(WindowPolicy{Size: time.Minute, MaxEvents: maxEvents})
	return NewWindowDispatcher(cap, counter), cap
}

func windowNotif(port int) Notification {
	return Notification{Port: port, State: "open", Title: "test"}
}

func TestWindowDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewWindowDispatcher(nil, NewWindowCounter(DefaultWindowPolicy()))
}

func TestWindowDispatcher_NilCounterPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil counter")
		}
	}()
	NewWindowDispatcher(&captureDispatcher{}, nil)
}

func TestWindowDispatcher_AllowsWithinLimit(t *testing.T) {
	d, cap := makeWindowDispatcher(3)
	for i := 0; i < 3; i++ {
		if err := d.Send(windowNotif(8080)); err != nil {
			t.Fatalf("unexpected error on send %d: %v", i+1, err)
		}
	}
	if cap.count != 3 {
		t.Fatalf("expected 3 deliveries, got %d", cap.count)
	}
}

func TestWindowDispatcher_BlocksOverLimit(t *testing.T) {
	d, cap := makeWindowDispatcher(2)
	d.Send(windowNotif(8080))
	d.Send(windowNotif(8080))
	err := d.Send(windowNotif(8080))
	if err == nil {
		t.Fatal("expected error when over limit")
	}
	if cap.count != 2 {
		t.Fatalf("expected 2 deliveries, got %d", cap.count)
	}
}

func TestWindowDispatcher_PropagatesNextError(t *testing.T) {
	fail := &errorDispatcher{err: errors.New("downstream failure")}
	counter := NewWindowCounter(WindowPolicy{Size: time.Minute, MaxEvents: 5})
	d := NewWindowDispatcher(fail, counter)
	if err := d.Send(windowNotif(9090)); err == nil {
		t.Fatal("expected error from next dispatcher")
	}
}

func TestWindowDispatcher_IndependentPorts(t *testing.T) {
	d, cap := makeWindowDispatcher(1)
	d.Send(windowNotif(8080))
	// different port → different key, should pass
	if err := d.Send(windowNotif(9090)); err != nil {
		t.Fatalf("expected independent port to pass: %v", err)
	}
	if cap.count != 2 {
		t.Fatalf("expected 2 deliveries, got %d", cap.count)
	}
}
