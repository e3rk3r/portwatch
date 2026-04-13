package alert

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBulkheadDispatcher_ConcurrentCallsWithinLimit(t *testing.T) {
	const limit = 4
	var peak int64
	var mu sync.Mutex
	var current int64

	block := make(chan struct{})
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		v := atomic.AddInt64(&current, 1)
		mu.Lock()
		if v > peak {
			peak = v
		}
		mu.Unlock()
		<-block
		atomic.AddInt64(&current, -1)
		return nil
	})

	d := NewBulkheadDispatcher(BulkheadPolicy{MaxConcurrent: limit}, next)

	var wg sync.WaitGroup
	for i := 0; i < limit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = d.Send(context.Background(), bulkheadNotif())
		}()
	}

	time.Sleep(30 * time.Millisecond)
	close(block)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if peak > limit {
		t.Fatalf("peak concurrency %d exceeded limit %d", peak, limit)
	}
}

func TestBulkheadDispatcher_ReleasesSlotOnError(t *testing.T) {
	sentinel := errors.New("dispatch error")
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		return sentinel
	})
	d := NewBulkheadDispatcher(BulkheadPolicy{MaxConcurrent: 1}, next)

	err := d.Send(context.Background(), bulkheadNotif())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	// slot must be released — second call should succeed (not return ErrBulkheadFull)
	err2 := d.Send(context.Background(), bulkheadNotif())
	if errors.Is(err2, ErrBulkheadFull) {
		t.Fatal("slot was not released after error")
	}
}
