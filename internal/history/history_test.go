package history_test

import (
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/history"
)

func makeEvent(port int, state string) history.Event {
	return history.Event{
		Port:      port,
		Host:      "localhost",
		State:     state,
		Timestamp: time.Now().UTC(),
	}
}

func TestNewRing_DefaultCapacity(t *testing.T) {
	r := history.NewRing(0)
	if r == nil {
		t.Fatal("expected non-nil Ring")
	}
}

func TestRing_RecordAndLen(t *testing.T) {
	r := history.NewRing(10)
	r.Record(makeEvent(8080, "open"))
	r.Record(makeEvent(9090, "closed"))

	if got := r.Len(); got != 2 {
		t.Fatalf("expected Len 2, got %d", got)
	}
}

func TestRing_SnapshotOrder(t *testing.T) {
	r := history.NewRing(5)
	ports := []int{1, 2, 3}
	for _, p := range ports {
		r.Record(makeEvent(p, "open"))
	}

	snap := r.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 events, got %d", len(snap))
	}
	for i, e := range snap {
		if e.Port != ports[i] {
			t.Errorf("index %d: expected port %d, got %d", i, ports[i], e.Port)
		}
	}
}

func TestRing_Overflow(t *testing.T) {
	cap := 3
	r := history.NewRing(cap)
	for i := 1; i <= 5; i++ {
		r.Record(makeEvent(i, "open"))
	}

	if got := r.Len(); got != cap {
		t.Fatalf("expected Len %d after overflow, got %d", cap, got)
	}

	snap := r.Snapshot()
	// oldest surviving event should be port 3
	if snap[0].Port != 3 {
		t.Errorf("expected oldest port 3, got %d", snap[0].Port)
	}
	if snap[cap-1].Port != 5 {
		t.Errorf("expected newest port 5, got %d", snap[cap-1].Port)
	}
}

func TestRing_TimestampAutoSet(t *testing.T) {
	r := history.NewRing(5)
	e := history.Event{Port: 80, Host: "localhost", State: "open"}
	before := time.Now().UTC()
	r.Record(e)
	after := time.Now().UTC()

	snap := r.Snapshot()
	if snap[0].Timestamp.Before(before) || snap[0].Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]",
			snap[0].Timestamp, before, after)
	}
}
