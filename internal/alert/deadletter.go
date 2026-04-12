package alert

import (
	"context"
	"sync"
	"time"
)

// DeadLetterPolicy configures the dead-letter queue.
type DeadLetterPolicy struct {
	// Capacity is the maximum number of failed notifications to retain.
	// Defaults to 100.
	Capacity int
}

// DefaultDeadLetterPolicy returns a DeadLetterPolicy with sensible defaults.
func DefaultDeadLetterPolicy() DeadLetterPolicy {
	return DeadLetterPolicy{Capacity: 100}
}

// DeadLetterEntry records a notification that could not be delivered.
type DeadLetterEntry struct {
	Notification Notification
	Err          error
	FailedAt     time.Time
}

// DeadLetterQueue stores notifications that failed delivery so they can be
// inspected or replayed later.
type DeadLetterQueue struct {
	mu      sync.Mutex
	entries []DeadLetterEntry
	cap     int
}

// NewDeadLetterQueue creates a DeadLetterQueue with the given policy.
func NewDeadLetterQueue(p DeadLetterPolicy) *DeadLetterQueue {
	cap := p.Capacity
	if cap <= 0 {
		cap = DefaultDeadLetterPolicy().Capacity
	}
	return &DeadLetterQueue{cap: cap}
}

// Record stores a failed notification in the queue. If the queue is full the
// oldest entry is evicted.
func (q *DeadLetterQueue) Record(n Notification, err error, now time.Time) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.entries) >= q.cap {
		q.entries = q.entries[1:]
	}
	q.entries = append(q.entries, DeadLetterEntry{Notification: n, Err: err, FailedAt: now})
}

// Snapshot returns a copy of all queued entries in insertion order.
func (q *DeadLetterQueue) Snapshot() []DeadLetterEntry {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]DeadLetterEntry, len(q.entries))
	copy(out, q.entries)
	return out
}

// Len returns the current number of entries.
func (q *DeadLetterQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.entries)
}

// NewDeadLetterDispatcher wraps next so that any delivery error causes the
// notification to be recorded in dlq instead of being silently dropped.
func NewDeadLetterDispatcher(next Dispatcher, dlq *DeadLetterQueue) Dispatcher {
	if next == nil {
		panic("alert: NewDeadLetterDispatcher: next must not be nil")
	}
	if dlq == nil {
		panic("alert: NewDeadLetterDispatcher: dlq must not be nil")
	}
	return &deadLetterDispatcher{next: next, dlq: dlq}
}

type deadLetterDispatcher struct {
	next Dispatcher
	dlq  *DeadLetterQueue
}

func (d *deadLetterDispatcher) Send(ctx context.Context, n Notification) error {
	err := d.next.Send(ctx, n)
	if err != nil {
		d.dlq.Record(n, err, time.Now())
	}
	return err
}
