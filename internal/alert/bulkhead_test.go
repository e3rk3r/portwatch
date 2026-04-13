package alert

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func bulkheadNotif() Notification {
	return Notification{Port: 9000, State: "open"}
}

func TestBulkhead_AcquireRelease(t *testing.T) {
	bh := NewBulkhead(BulkheadPolicy{MaxConcurrent: 2})
	if err := bh.Acquire(); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if bh.Active() != 1 {
		t.Fatalf("expected active=1, got %d", bh.Active())
	}
	bh.Release()
	if bh.Active() != 0 {
		t.Fatalf("expected active=0 after release, got %d", bh.Active())
	}
}

func TestBulkhead_RejectsWhenFull(t *testing.T) {
	bh := NewBulkhead(BulkheadPolicy{MaxConcurrent: 1})
	if err := bh.Acquire(); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	defer bh.Release()
	err := bh.Acquire()
	if !errors.Is(err, ErrBulkheadFull) {
		t.Fatalf("expected ErrBulkheadFull, got %v", err)
	}
}

func TestBulkhead_DefaultPolicy(t *testing.T) {
	bh := NewBulkhead(BulkheadPolicy{})
	if cap(bh.sem) != 8 {
		t.Fatalf("expected default capacity 8, got %d", cap(bh.sem))
	}
}

func TestBulkheadDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewBulkheadDispatcher(DefaultBulkheadPolicy(), nil)
}

func TestBulkheadDispatcher_AllowsUnderLimit(t *testing.T) {
	called := false
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		called = true
		return nil
	})
	d := NewBulkheadDispatcher(BulkheadPolicy{MaxConcurrent: 2}, next)
	if err := d.Send(context.Background(), bulkheadNotif()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next to be called")
	}
}

func TestBulkheadDispatcher_RejectsOverLimit(t *testing.T) {
	block := make(chan struct{})
	var wg sync.WaitGroup
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		<-block
		return nil
	})
	d := NewBulkheadDispatcher(BulkheadPolicy{MaxConcurrent: 1}, next)

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = d.Send(context.Background(), bulkheadNotif())
	}()

	// give goroutine time to acquire
	time.Sleep(20 * time.Millisecond)

	err := d.Send(context.Background(), bulkheadNotif())
	if !errors.Is(err, ErrBulkheadFull) {
		t.Fatalf("expected ErrBulkheadFull, got %v", err)
	}
	close(block)
	wg.Wait()
}
