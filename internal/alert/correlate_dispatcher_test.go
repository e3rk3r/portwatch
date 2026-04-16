package alert

import (
	"context"
	"testing"
	"time"
)

func TestCorrelateDispatcher_IndependentPorts(t *testing.T) {
	var sent int
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		sent++
		return nil
	})
	p := CorrelatePolicy{WindowDuration: time.Minute, MinEvents: 2}
	d := NewCorrelateDispatcher(p, next)

	n1 := correlateNotif(8080, "closed")
	n2 := correlateNotif(9090, "closed")

	// Each port needs its own 2 events
	_ = d.Send(context.Background(), n1)
	_ = d.Send(context.Background(), n1) // triggers for 8080
	_ = d.Send(context.Background(), n2)
	_ = d.Send(context.Background(), n2) // triggers for 9090

	if sent != 2 {
		t.Fatalf("expected 2 sends (one per port), got %d", sent)
	}
}

func TestCorrelateDispatcher_DifferentStatesIndependent(t *testing.T) {
	var sent int
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		sent++
		return nil
	})
	p := CorrelatePolicy{WindowDuration: time.Minute, MinEvents: 2}
	d := NewCorrelateDispatcher(p, next)

	open := correlateNotif(8080, "open")
	closed := correlateNotif(8080, "closed")

	_ = d.Send(context.Background(), open)
	_ = d.Send(context.Background(), open)   // triggers open
	_ = d.Send(context.Background(), closed)
	_ = d.Send(context.Background(), closed) // triggers closed

	if sent != 2 {
		t.Fatalf("expected 2 sends, got %d", sent)
	}
}
