package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultBatchPolicy returns a BatchPolicy that collects up to 10 notifications
// or flushes every 30 seconds, whichever comes first.
func DefaultBatchPolicy() BatchPolicy {
	return BatchPolicy{
		MaxSize:       10,
		FlushInterval: 30 * time.Second,
	}
}

// BatchPolicy controls how a Batcher accumulates and flushes notifications.
type BatchPolicy struct {
	MaxSize       int
	FlushInterval time.Duration
}

// Batcher accumulates notifications and flushes them as a slice to a
// downstream handler once a size or time threshold is reached.
type Batcher struct {
	policy  BatchPolicy
	mu      sync.Mutex
	buf     []Notification
	flushFn func([]Notification)
	now     func() time.Time
}

// NewBatcher creates a Batcher with the given policy and flush callback.
// The flush callback is invoked with a snapshot of accumulated notifications.
func NewBatcher(policy BatchPolicy, flush func([]Notification)) *Batcher {
	if policy.MaxSize <= 0 {
		policy.MaxSize = DefaultBatchPolicy().MaxSize
	}
	if policy.FlushInterval <= 0 {
		policy.FlushInterval = DefaultBatchPolicy().FlushInterval
	}
	return &Batcher{
		policy:  policy,
		flushFn: flush,
		now:     time.Now,
	}
}

// Add appends a notification to the internal buffer. If the buffer reaches
// MaxSize, it is flushed immediately.
func (b *Batcher) Add(n Notification) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, n)
	if len(b.buf) >= b.policy.MaxSize {
		b.flush()
	}
}

// Run starts the interval-based flush loop. It blocks until ctx is cancelled.
func (b *Batcher) Run(ctx context.Context) {
	ticker := time.NewTicker(b.policy.FlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			b.flush()
			b.mu.Unlock()
		case <-ctx.Done():
			b.mu.Lock()
			b.flush()
			b.mu.Unlock()
			return
		}
	}
}

// Len returns the current number of buffered notifications.
func (b *Batcher) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.buf)
}

// batchKey returns a string key for a notification used in batch grouping.
func batchKey(n Notification) string {
	return fmt.Sprintf("%d:%s", n.Port, n.State)
}

// flush drains the buffer and calls flushFn. Must be called with b.mu held.
func (b *Batcher) flush() {
	if len(b.buf) == 0 {
		return
	}
	snap := make([]Notification, len(b.buf))
	copy(snap, b.buf)
	b.buf = b.buf[:0]
	go b.flushFn(snap)
}
