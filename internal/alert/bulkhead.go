package alert

import (
	"context"
	"errors"
	"sync"
)

// ErrBulkheadFull is returned when the bulkhead concurrency limit is reached.
var ErrBulkheadFull = errors.New("bulkhead: concurrency limit reached")

// BulkheadPolicy configures the bulkhead dispatcher.
type BulkheadPolicy struct {
	// MaxConcurrent is the maximum number of in-flight dispatches allowed.
	// Defaults to 8 if zero.
	MaxConcurrent int
}

// DefaultBulkheadPolicy returns a sensible default BulkheadPolicy.
func DefaultBulkheadPolicy() BulkheadPolicy {
	return BulkheadPolicy{MaxConcurrent: 8}
}

// Bulkhead limits the number of concurrent dispatches to protect downstream
// systems from overload. Excess calls are rejected immediately with
// ErrBulkheadFull rather than queued, preserving low latency.
type Bulkhead struct {
	policy BulkheadPolicy
	sem    chan struct{}
	mu     sync.Mutex
	active int
}

// NewBulkhead creates a Bulkhead from the given policy.
func NewBulkhead(p BulkheadPolicy) *Bulkhead {
	if p.MaxConcurrent <= 0 {
		p.MaxConcurrent = DefaultBulkheadPolicy().MaxConcurrent
	}
	return &Bulkhead{
		policy: p,
		sem:    make(chan struct{}, p.MaxConcurrent),
	}
}

// Acquire attempts to acquire a slot. Returns ErrBulkheadFull if none are free.
func (b *Bulkhead) Acquire() error {
	select {
	case b.sem <- struct{}{}:
		b.mu.Lock()
		b.active++
		b.mu.Unlock()
		return nil
	default:
		return ErrBulkheadFull
	}
}

// Release frees a previously acquired slot.
func (b *Bulkhead) Release() {
	<-b.sem
	b.mu.Lock()
	b.active--
	b.mu.Unlock()
}

// Active returns the current number of in-flight dispatches.
func (b *Bulkhead) Active() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active
}

// NewBulkheadDispatcher wraps next with a Bulkhead concurrency limiter.
func NewBulkheadDispatcher(policy BulkheadPolicy, next Dispatcher) Dispatcher {
	if next == nil {
		panic("bulkhead: next dispatcher must not be nil")
	}
	bh := NewBulkhead(policy)
	return dispatcherFunc(func(ctx context.Context, n Notification) error {
		if err := bh.Acquire(); err != nil {
			return err
		}
		defer bh.Release()
		return next.Send(ctx, n)
	})
}
