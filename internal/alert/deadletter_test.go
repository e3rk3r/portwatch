package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errDelivery = errors.New("delivery failed")

func makeDLNotif(port int) Notification {
	return Notification{Port: port, State: "closed"}
}

func TestDeadLetterQueue_DefaultCapacity(t *testing.T) {
	q := NewDeadLetterQueue(DeadLetterPolicy{})
	if q.cap != 100 {
		t.Fatalf("expected cap 100, got %d", q.cap)
	}
}

func TestDeadLetterQueue_RecordAndLen(t *testing.T) {
	q := NewDeadLetterQueue(DefaultDeadLetterPolicy())
	q.Record(makeDLNotif(8080), errDelivery, time.Now())
	if q.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", q.Len())
	}
}

func TestDeadLetterQueue_Overflow(t *testing.T) {
	q := NewDeadLetterQueue(DeadLetterPolicy{Capacity: 3})
	for i := 0; i < 5; i++ {
		q.Record(makeDLNotif(8000+i), errDelivery, time.Now())
	}
	if q.Len() != 3 {
		t.Fatalf("expected 3 entries after overflow, got %d", q.Len())
	}
	// oldest should have been evicted; first remaining port should be 8002
	snap := q.Snapshot()
	if snap[0].Notification.Port != 8002 {
		t.Fatalf("expected port 8002, got %d", snap[0].Notification.Port)
	}
}

func TestDeadLetterQueue_SnapshotOrder(t *testing.T) {
	q := NewDeadLetterQueue(DeadLetterPolicy{Capacity: 10})
	ports := []int{9001, 9002, 9003}
	for _, p := range ports {
		q.Record(makeDLNotif(p), errDelivery, time.Now())
	}
	snap := q.Snapshot()
	for i, p := range ports {
		if snap[i].Notification.Port != p {
			t.Fatalf("pos %d: expected port %d, got %d", i, p, snap[i].Notification.Port)
		}
	}
}

func TestDeadLetterDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewDeadLetterDispatcher(nil, NewDeadLetterQueue(DefaultDeadLetterPolicy()))
}

func TestDeadLetterDispatcher_NilQueuePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil dlq")
		}
	}()
	NewDeadLetterDispatcher(&logDispatcher{}, nil)
}

func TestDeadLetterDispatcher_RecordsOnError(t *testing.T) {
	q := NewDeadLetterQueue(DefaultDeadLetterPolicy())
	failing := &funcDispatcher{fn: func(_ context.Context, _ Notification) error { return errDelivery }}
	d := NewDeadLetterDispatcher(failing, q)

	err := d.Send(context.Background(), makeDLNotif(8080))
	if !errors.Is(err, errDelivery) {
		t.Fatalf("expected errDelivery, got %v", err)
	}
	if q.Len() != 1 {
		t.Fatalf("expected 1 dead-letter entry, got %d", q.Len())
	}
}

func TestDeadLetterDispatcher_NoRecordOnSuccess(t *testing.T) {
	q := NewDeadLetterQueue(DefaultDeadLetterPolicy())
	ok := &funcDispatcher{fn: func(_ context.Context, _ Notification) error { return nil }}
	d := NewDeadLetterDispatcher(ok, q)

	if err := d.Send(context.Background(), makeDLNotif(9090)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Len() != 0 {
		t.Fatalf("expected 0 dead-letter entries, got %d", q.Len())
	}
}
