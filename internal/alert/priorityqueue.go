package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Priority levels for notifications.
const (
	PriorityLow    = 0
	PriorityNormal = 1
	PriorityHigh   = 2
	PriorityCritical = 3
)

// DefaultPriorityQueuePolicy returns a sensible default policy.
func DefaultPriorityQueuePolicy() PriorityQueuePolicy {
	return PriorityQueuePolicy{
		Capacity: 256,
		DrainTimeout: 5 * time.Second,
	}
}

// PriorityQueuePolicy controls behaviour of the priority queue dispatcher.
type PriorityQueuePolicy struct {
	Capacity     int
	DrainTimeout time.Duration
}

type priorityItem struct {
	notif    Notification
	priority int
}

// PriorityQueue buffers notifications in priority order before forwarding.
type PriorityQueue struct {
	mu     sync.Mutex
	buckets [4][]Notification // index == priority level
	policy PriorityQueuePolicy
	size   int
}

// NewPriorityQueue creates a PriorityQueue with the given policy.
func NewPriorityQueue(p PriorityQueuePolicy) *PriorityQueue {
	if p.Capacity <= 0 {
		p.Capacity = DefaultPriorityQueuePolicy().Capacity
	}
	if p.DrainTimeout <= 0 {
		p.DrainTimeout = DefaultPriorityQueuePolicy().DrainTimeout
	}
	return &PriorityQueue{policy: p}
}

// Enqueue adds a notification at the given priority level.
// Returns an error if the queue is at capacity.
func (q *PriorityQueue) Enqueue(n Notification, priority int) error {
	if priority < PriorityLow || priority > PriorityCritical {
		return fmt.Errorf("priority %d out of range [0,3]", priority)
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size >= q.policy.Capacity {
		return fmt.Errorf("priority queue at capacity (%d)", q.policy.Capacity)
	}
	q.buckets[priority] = append(q.buckets[priority], n)
	q.size++
	return nil
}

// Dequeue removes and returns the highest-priority notification.
// Returns false if the queue is empty.
func (q *PriorityQueue) Dequeue() (Notification, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for p := PriorityCritical; p >= PriorityLow; p-- {
		if len(q.buckets[p]) > 0 {
			n := q.buckets[p][0]
			q.buckets[p] = q.buckets[p][1:]
			q.size--
			return n, true
		}
	}
	return Notification{}, false
}

// Len returns the total number of queued notifications.
func (q *PriorityQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}

// Drain forwards all queued notifications to next, highest priority first.
func (q *PriorityQueue) Drain(ctx context.Context, next Dispatcher) error {
	for {
		n, ok := q.Dequeue()
		if !ok {
			return nil
		}
		if err := next.Send(ctx, n); err != nil {
			return err
		}
	}
}
