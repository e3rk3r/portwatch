package alert

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// countingDispatcher records how many times Send was called.
type countingDispatcher struct {
	count int32
	err   error
}

func (c *countingDispatcher) Send(_ context.Context, _ Notification) error {
	atomic.AddInt32(&c.count, 1)
	return c.err
}

func fanoutNotif() Notification {
	return Notification{Port: 9090, State: "open", Title: "fanout test"}
}

func TestFanoutDispatcher_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty targets")
		}
	}()
	NewFanoutDispatcher()
}

func TestFanoutDispatcher_AllTargetsCalled(t *testing.T) {
	a := &countingDispatcher{}
	b := &countingDispatcher{}
	c := &countingDispatcher{}

	f := NewFanoutDispatcher(a, b, c)
	if err := f.Send(context.Background(), fanoutNotif()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, d := range []*countingDispatcher{a, b, c} {
		if atomic.LoadInt32(&d.count) != 1 {
			t.Errorf("target[%d] called %d times, want 1", i, d.count)
		}
	}
}

func TestFanoutDispatcher_NilTargetSkipped(t *testing.T) {
	a := &countingDispatcher{}
	f := NewFanoutDispatcher(a, nil)
	if err := f.Send(context.Background(), fanoutNotif()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&a.count) != 1 {
		t.Errorf("expected target a to be called once")
	}
}

func TestFanoutDispatcher_CollectsErrors(t *testing.T) {
	ea := errors.New("alpha failed")
	eb := errors.New("beta failed")
	a := &countingDispatcher{err: ea}
	b := &countingDispatcher{err: eb}
	c := &countingDispatcher{}

	f := NewFanoutDispatcher(a, b, c)
	err := f.Send(context.Background(), fanoutNotif())
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	if !strings.Contains(err.Error(), "alpha failed") || !strings.Contains(err.Error(), "beta failed") {
		t.Errorf("error missing expected messages: %v", err)
	}
}

func TestFanoutDispatcher_Concurrent(t *testing.T) {
	const n = 20
	targets := make([]Dispatcher, n)
	counters := make([]*countingDispatcher, n)
	for i := range targets {
		counters[i] = &countingDispatcher{}
		targets[i] = counters[i]
	}
	f := NewFanoutDispatcher(targets...)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := f.Send(ctx, fanoutNotif()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, c := range counters {
		if atomic.LoadInt32(&c.count) != 1 {
			t.Errorf("counter[%d] = %d, want 1", i, c.count)
		}
	}
}
