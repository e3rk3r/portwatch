package alert

import (
	"context"
	"errors"
	"testing"
)

func TestPriorityQueue_DefaultPolicy(t *testing.T) {
	pol := DefaultPriorityQueuePolicy()
	if pol.Capacity != 256 {
		t.Fatalf("expected capacity 256, got %d", pol.Capacity)
	}
}

func TestPriorityQueue_EnqueueDequeue_Order(t *testing.T) {
	q := NewPriorityQueue(DefaultPriorityQueuePolicy())

	low := Notification{Title: "low"}
	norm := Notification{Title: "normal"}
	high := Notification{Title: "high"}
	crit := Notification{Title: "critical"}

	_ = q.Enqueue(low, PriorityLow)
	_ = q.Enqueue(norm, PriorityNormal)
	_ = q.Enqueue(high, PriorityHigh)
	_ = q.Enqueue(crit, PriorityCritical)

	expected := []string{"critical", "high", "normal", "low"}
	for _, want := range expected {
		n, ok := q.Dequeue()
		if !ok {
			t.Fatalf("expected item %q, got empty", want)
		}
		if n.Title != want {
			t.Errorf("expected %q, got %q", want, n.Title)
		}
	}
}

func TestPriorityQueue_DequeueEmpty(t *testing.T) {
	q := NewPriorityQueue(DefaultPriorityQueuePolicy())
	_, ok := q.Dequeue()
	if ok {
		t.Fatal("expected false from empty queue")
	}
}

func TestPriorityQueue_CapacityEnforced(t *testing.T) {
	pol := PriorityQueuePolicy{Capacity: 2}
	q := NewPriorityQueue(pol)

	if err := q.Enqueue(Notification{}, PriorityLow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := q.Enqueue(Notification{}, PriorityLow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := q.Enqueue(Notification{}, PriorityLow); err == nil {
		t.Fatal("expected capacity error")
	}
}

func TestPriorityQueue_InvalidPriority(t *testing.T) {
	q := NewPriorityQueue(DefaultPriorityQueuePolicy())
	if err := q.Enqueue(Notification{}, 99); err == nil {
		t.Fatal("expected error for out-of-range priority")
	}
}

func TestPriorityQueue_Len(t *testing.T) {
	q := NewPriorityQueue(DefaultPriorityQueuePolicy())
	_ = q.Enqueue(Notification{Title: "a"}, PriorityHigh)
	_ = q.Enqueue(Notification{Title: "b"}, PriorityLow)
	if q.Len() != 2 {
		t.Fatalf("expected len 2, got %d", q.Len())
	}
	q.Dequeue()
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestPriorityQueue_Drain(t *testing.T) {
	q := NewPriorityQueue(DefaultPriorityQueuePolicy())
	_ = q.Enqueue(Notification{Title: "x"}, PriorityCritical)
	_ = q.Enqueue(Notification{Title: "y"}, PriorityLow)

	var received []string
	next := DispatcherFunc(func(_ context.Context, n Notification) error {
		received = append(received, n.Title)
		return nil
	})

	if err := q.Drain(context.Background(), next); err != nil {
		t.Fatalf("drain error: %v", err)
	}
	if len(received) != 2 || received[0] != "x" {
		t.Errorf("unexpected drain order: %v", received)
	}
}

func TestPriorityQueue_Drain_StopsOnError(t *testing.T) {
	q := NewPriorityQueue(DefaultPriorityQueuePolicy())
	_ = q.Enqueue(Notification{}, PriorityHigh)
	_ = q.Enqueue(Notification{}, PriorityLow)

	sentinel := errors.New("dispatch failed")
	next := DispatcherFunc(func(_ context.Context, _ Notification) error {
		return sentinel
	})

	if err := q.Drain(context.Background(), next); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
	if q.Len() != 1 {
		t.Errorf("expected 1 item remaining after error, got %d", q.Len())
	}
}
