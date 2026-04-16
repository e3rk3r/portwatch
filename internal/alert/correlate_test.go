package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

func fixedCorrelateClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func correlateNotif(port int, state string) Notification {
	return Notification{Port: port, State: state, Title: "test"}
}

func TestCorrelator_BelowThreshold(t *testing.T) {
	c := NewCorrelator(DefaultCorrelatePolicy())
	n := correlateNotif(8080, "closed")
	for i := 0; i < 2; i++ {
		if c.Record(n) {
			t.Fatalf("expected false before threshold")
		}
	}
}

func TestCorrelator_AtThreshold(t *testing.T) {
	c := NewCorrelator(DefaultCorrelatePolicy())
	n := correlateNotif(8080, "closed")
	for i := 0; i < 2; i++ {
		c.Record(n)
	}
	if !c.Record(n) {
		t.Fatal("expected true at threshold")
	}
}

func TestCorrelator_AboveThresholdNoRepeat(t *testing.T) {
	c := NewCorrelator(DefaultCorrelatePolicy())
	n := correlateNotif(9090, "closed")
	for i := 0; i < 3; i++ {
		c.Record(n)
	}
	// 4th call should not re-trigger
	if c.Record(n) {
		t.Fatal("expected false after threshold already triggered")
	}
}

func TestCorrelator_WindowReset(t *testing.T) {
	base := time.Now()
	c := NewCorrelator(CorrelatePolicy{WindowDuration: time.Minute, MinEvents: 2})
	c.clock = fixedCorrelateClock(base)
	n := correlateNotif(8080, "open")
	c.Record(n)
	// advance past window
	c.clock = fixedCorrelateClock(base.Add(2 * time.Minute))
	if c.Record(n) {
		t.Fatal("should not trigger after window reset")
	}
}

func TestCorrelator_Reset(t *testing.T) {
	c := NewCorrelator(DefaultCorrelatePolicy())
	n := correlateNotif(8080, "closed")
	c.Record(n)
	c.Record(n)
	c.Reset()
	if c.Record(n) {
		t.Fatal("expected false after reset")
	}
}

func TestCorrelateDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewCorrelateDispatcher(DefaultCorrelatePolicy(), nil)
}

func TestCorrelateDispatcher_ForwardsAtThreshold(t *testing.T) {
	var sent int
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		sent++
		return nil
	})
	d := NewCorrelateDispatcher(CorrelatePolicy{WindowDuration: time.Minute, MinEvents: 2}, next)
	n := correlateNotif(8080, "closed")
	_ = d.Send(context.Background(), n)
	_ = d.Send(context.Background(), n)
	if sent != 1 {
		t.Fatalf("expected 1 send, got %d", sent)
	}
}

func TestCorrelateDispatcher_PropagatesError(t *testing.T) {
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		return errors.New("boom")
	})
	d := NewCorrelateDispatcher(CorrelatePolicy{WindowDuration: time.Minute, MinEvents: 1}, next)
	n := correlateNotif(8080, "closed")
	if err := d.Send(context.Background(), n); err == nil {
		t.Fatal("expected error")
	}
}
