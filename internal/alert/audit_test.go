package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubDispatcher struct {
	err error
}

func (s *stubDispatcher) Send(_ context.Context, _ Notification) error { return s.err }

func baseNotif(port int, state string) Notification {
	return Notification{Port: port, State: state, Title: "test", Body: "body"}
}

func TestAuditLog_DefaultCapacity(t *testing.T) {
	al := NewAuditLog(0)
	if al.cap != DefaultAuditCapacity {
		t.Fatalf("expected cap %d, got %d", DefaultAuditCapacity, al.cap)
	}
}

func TestAuditLog_RecordAndSnapshot(t *testing.T) {
	al := NewAuditLog(4)
	al.Record(AuditEntry{Port: 80, State: "open", Success: true})
	al.Record(AuditEntry{Port: 443, State: "closed", Success: false, Err: "timeout"})

	snap := al.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	if snap[0].Port != 80 || !snap[0].Success {
		t.Errorf("unexpected first entry: %+v", snap[0])
	}
	if snap[1].Err != "timeout" {
		t.Errorf("expected err 'timeout', got %q", snap[1].Err)
	}
}

func TestAuditLog_Overflow(t *testing.T) {
	al := NewAuditLog(3)
	for i := 0; i < 5; i++ {
		al.Record(AuditEntry{Port: i})
	}
	if al.Len() != 3 {
		t.Fatalf("expected 3 entries after overflow, got %d", al.Len())
	}
	snap := al.Snapshot()
	if snap[0].Port != 2 {
		t.Errorf("expected oldest retained port=2, got %d", snap[0].Port)
	}
}

func TestAuditDispatcher_RecordsSuccess(t *testing.T) {
	log := NewAuditLog(10)
	stub := &stubDispatcher{}
	ad := NewAuditDispatcher(stub, log, "webhook")
	fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ad.clock = func() time.Time { return fixedTime }

	if err := ad.Send(context.Background(), baseNotif(8080, "open")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := log.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap))
	}
	e := snap[0]
	if e.Port != 8080 || e.State != "open" || !e.Success || e.Channel != "webhook" {
		t.Errorf("unexpected entry: %+v", e)
	}
	if !e.Timestamp.Equal(fixedTime) {
		t.Errorf("unexpected timestamp: %v", e.Timestamp)
	}
}

func TestAuditDispatcher_RecordsFailure(t *testing.T) {
	log := NewAuditLog(10)
	stub := &stubDispatcher{err: errors.New("connection refused")}
	ad := NewAuditDispatcher(stub, log, "script")

	err := ad.Send(context.Background(), baseNotif(9090, "closed"))
	if err == nil {
		t.Fatal("expected error to propagate")
	}
	snap := log.Snapshot()
	if snap[0].Success {
		t.Error("expected Success=false")
	}
	if snap[0].Err != "connection refused" {
		t.Errorf("unexpected Err: %q", snap[0].Err)
	}
}

func TestAuditDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil next")
		}
	}()
	NewAuditDispatcher(nil, NewAuditLog(10), "x")
}

func TestAuditKey_Format(t *testing.T) {
	if got := auditKey(80, "open", "webhook"); got != "80:open:webhook" {
		t.Errorf("unexpected key: %q", got)
	}
}
