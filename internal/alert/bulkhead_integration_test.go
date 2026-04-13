package alert_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/alert"
)

func TestBulkhead_PipelineIntegration(t *testing.T) {
	// Simulate a pipeline: Bulkhead -> Retry -> slow next
	var calls int64
	block := make(chan struct{})

	base := alert.DispatcherFunc(func(_ context.Context, _ alert.Notification) error {
		atomic.AddInt64(&calls, 1)
		<-block
		return nil
	})

	retryPolicy := alert.DefaultRetryPolicy()
	retryPolicy.MaxAttempts = 1
	withRetry := alert.NewRetryer(retryPolicy, base)

	bhPolicy := alert.BulkheadPolicy{MaxConcurrent: 2}
	d := alert.NewBulkheadDispatcher(bhPolicy, withRetry)

	notif := alert.Notification{Port: 8080, State: "closed"}

	var wg sync.WaitGroup
	rejected := int64(0)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := d.Send(context.Background(), notif); errors.Is(err, alert.ErrBulkheadFull) {
				atomic.AddInt64(&rejected, 1)
			}
		}()
	}

	time.Sleep(40 * time.Millisecond)
	close(block)
	wg.Wait()

	if atomic.LoadInt64(&rejected) == 0 {
		t.Fatal("expected at least one rejection from bulkhead")
	}
	if atomic.LoadInt64(&calls) > 2 {
		t.Fatalf("expected at most 2 concurrent calls, got %d", atomic.LoadInt64(&calls))
	}
}

func TestBulkhead_RecoverAfterDrain(t *testing.T) {
	next := alert.DispatcherFunc(func(_ context.Context, _ alert.Notification) error {
		return nil
	})
	d := alert.NewBulkheadDispatcher(alert.BulkheadPolicy{MaxConcurrent: 1}, next)
	notif := alert.Notification{Port: 443, State: "open"}

	for i := 0; i < 5; i++ {
		if err := d.Send(context.Background(), notif); err != nil {
			t.Fatalf("iteration %d: unexpected error %v", i, err)
		}
	}
}
