package alert

import (
	"context"
	"sync"
	"testing"
	"time"
)

func makeBatchNotification(port int, state string) Notification {
	return Notification{Port: port, State: state, Title: "test", Body: "body"}
}

func TestBatcher_AddBelowThreshold(t *testing.T) {
	var mu sync.Mutex
	var flushed [][]Notification
	b := NewBatcher(BatchPolicy{MaxSize: 5, FlushInterval: time.Hour}, func(batch []Notification) {
		mu.Lock()
		defer mu.Unlock()
		flushed = append(flushed, batch)
	})
	b.Add(makeBatchNotification(8080, "open"))
	b.Add(makeBatchNotification(9090, "closed"))
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(flushed) != 0 {
		t.Fatalf("expected no flush yet, got %d", len(flushed))
	}
	if b.Len() != 2 {
		t.Fatalf("expected 2 buffered, got %d", b.Len())
	}
}

func TestBatcher_FlushesAtMaxSize(t *testing.T) {
	var mu sync.Mutex
	var flushed [][]Notification
	b := NewBatcher(BatchPolicy{MaxSize: 3, FlushInterval: time.Hour}, func(batch []Notification) {
		mu.Lock()
		defer mu.Unlock()
		flushed = append(flushed, batch)
	})
	for i := 0; i < 3; i++ {
		b.Add(makeBatchNotification(8080+i, "open"))
	}
	time.Sleep(30 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(flushed) != 1 {
		t.Fatalf("expected 1 flush, got %d", len(flushed))
	}
	if len(flushed[0]) != 3 {
		t.Fatalf("expected batch of 3, got %d", len(flushed[0]))
	}
}

func TestBatcher_FlushesOnInterval(t *testing.T) {
	var mu sync.Mutex
	var flushed [][]Notification
	b := NewBatcher(BatchPolicy{MaxSize: 100, FlushInterval: 40 * time.Millisecond}, func(batch []Notification) {
		mu.Lock()
		defer mu.Unlock()
		flushed = append(flushed, batch)
	})
	b.Add(makeBatchNotification(8080, "open"))
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	go b.Run(ctx)
	<-ctx.Done()
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(flushed) == 0 {
		t.Fatal("expected at least one interval flush")
	}
}

func TestBatcher_FlushOnContextCancel(t *testing.T) {
	var mu sync.Mutex
	var flushed [][]Notification
	b := NewBatcher(BatchPolicy{MaxSize: 100, FlushInterval: time.Hour}, func(batch []Notification) {
		mu.Lock()
		defer mu.Unlock()
		flushed = append(flushed, batch)
	})
	b.Add(makeBatchNotification(3000, "closed"))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { b.Run(ctx); close(done) }()
	cancel()
	<-done
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(flushed) != 1 || len(flushed[0]) != 1 {
		t.Fatalf("expected 1 batch with 1 notification on cancel, got %v", flushed)
	}
}

func TestBatchKey_Format(t *testing.T) {
	n := makeBatchNotification(8080, "open")
	key := batchKey(n)
	if key != "8080:open" {
		t.Fatalf("unexpected batch key: %s", key)
	}
}
