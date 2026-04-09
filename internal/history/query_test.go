package history

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
)

func addEvent(r *Ring, port int, state string, ago time.Duration) {
	r.Record(Event{
		Port:      port,
		State:     state,
		Timestamp: time.Now().Add(-ago),
	})
}

func TestQuery_NoFilter(t *testing.T) {
	r := NewRing(10)
	addEvent(r, 8080, "open", 5*time.Minute)
	addEvent(r, 9090, "closed", 3*time.Minute)

	got := r.Query(Filter{})
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
}

func TestQuery_FilterByPort(t *testing.T) {
	r := NewRing(10)
	addEvent(r, 8080, "open", 5*time.Minute)
	addEvent(r, 9090, "closed", 3*time.Minute)
	addEvent(r, 8080, "closed", 1*time.Minute)

	got := r.Query(Filter{Port: 8080})
	if len(got) != 2 {
		t.Fatalf("expected 2 events for port 8080, got %d", len(got))
	}
	for _, e := range got {
		if e.Port != 8080 {
			t.Errorf("unexpected port %d in results", e.Port)
		}
	}
}

func TestQuery_FilterByState(t *testing.T) {
	r := NewRing(10)
	addEvent(r, 8080, "open", 5*time.Minute)
	addEvent(r, 9090, "closed", 3*time.Minute)

	got := r.Query(Filter{State: "open"})
	if len(got) != 1 {
		t.Fatalf("expected 1 open event, got %d", len(got))
	}
}

func TestQuery_FilterBySince(t *testing.T) {
	r := NewRing(10)
	addEvent(r, 8080, "open", 10*time.Minute)
	addEvent(r, 9090, "closed", 2*time.Minute)

	cutoff := time.Now().Add(-5 * time.Minute)
	got := r.Query(Filter{Since: cutoff})
	if len(got) != 1 {
		t.Fatalf("expected 1 recent event, got %d", len(got))
	}
}

func TestQuery_Limit(t *testing.T) {
	r := NewRing(20)
	for i := 0; i < 8; i++ {
		addEvent(r, 8080, "open", time.Duration(i)*time.Minute)
	}

	got := r.Query(Filter{Limit: 3})
	if len(got) != 3 {
		t.Fatalf("expected 3 events with limit, got %d", len(got))
	}
}

func TestSummary_LastState(t *testing.T) {
	r := NewRing(10)
	addEvent(r, 8080, "open", 5*time.Minute)
	addEvent(r, 8080, "closed", 1*time.Minute)

	ports := []config.PortConfig{{Port: 8080}, {Port: 9090}}
	summary := r.Summary(ports)

	if summary[8080] != "closed" {
		t.Errorf("expected 8080 to be closed, got %s", summary[8080])
	}
	if summary[9090] != "unknown" {
		t.Errorf("expected 9090 to be unknown, got %s", summary[9090])
	}
}
