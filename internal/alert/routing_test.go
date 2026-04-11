package alert

import (
	"errors"
	"testing"
)

// captureDispatcher records every notification it receives.
type captureDispatcher struct {
	Received []Notification
	Err      error
}

func (c *captureDispatcher) Send(n Notification) error {
	c.Received = append(c.Received, n)
	return c.Err
}

func makeNotif(port int, state string) Notification {
	return Notification{Port: port, State: state, Title: "test"}
}

func TestRouter_Register_NilDispatcher(t *testing.T) {
	r := NewRouter()
	err := r.Register(Route{Name: "bad"}, nil)
	if err == nil {
		t.Fatal("expected error for nil dispatcher")
	}
}

func TestRouter_NoRoutes_DropsNotification(t *testing.T) {
	r := NewRouter()
	// No routes registered — Send should not panic and return nil.
	if err := r.Send(makeNotif(8080, "open")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRouter_MatchAllRoute(t *testing.T) {
	r := NewRouter()
	d := &captureDispatcher{}
	_ = r.Register(Route{Name: "all"}, d)

	r.Send(makeNotif(80, "open"))
	r.Send(makeNotif(443, "closed"))

	if len(d.Received) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(d.Received))
	}
}

func TestRouter_PortFilter(t *testing.T) {
	r := NewRouter()
	d80 := &captureDispatcher{}
	dAll := &captureDispatcher{}

	_ = r.Register(Route{Name: "port80", Ports: []int{80}}, d80)
	_ = r.Register(Route{Name: "all"}, dAll)

	r.Send(makeNotif(80, "open"))
	r.Send(makeNotif(443, "open"))

	if len(d80.Received) != 1 {
		t.Fatalf("port80 dispatcher: expected 1, got %d", len(d80.Received))
	}
	if len(dAll.Received) != 2 {
		t.Fatalf("all dispatcher: expected 2, got %d", len(dAll.Received))
	}
}

func TestRouter_StateFilter(t *testing.T) {
	r := NewRouter()
	dOpen := &captureDispatcher{}
	_ = r.Register(Route{Name: "openOnly", States: []string{"open"}}, dOpen)

	r.Send(makeNotif(8080, "open"))
	r.Send(makeNotif(8080, "closed"))

	if len(dOpen.Received) != 1 {
		t.Fatalf("expected 1 open notification, got %d", len(dOpen.Received))
	}
	if dOpen.Received[0].State != "open" {
		t.Fatalf("expected state 'open', got %q", dOpen.Received[0].State)
	}
}

func TestRouter_PropagatesDispatcherError(t *testing.T) {
	r := NewRouter()
	want := errors.New("send failed")
	d := &captureDispatcher{Err: want}
	_ = r.Register(Route{Name: "err"}, d)

	err := r.Send(makeNotif(9090, "open"))
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestRouter_MulticastToMatchingRoutes(t *testing.T) {
	r := NewRouter()
	d1 := &captureDispatcher{}
	d2 := &captureDispatcher{}
	_ = r.Register(Route{Name: "r1", Ports: []int{80, 443}}, d1)
	_ = r.Register(Route{Name: "r2", States: []string{"closed"}}, d2)

	r.Send(makeNotif(443, "closed")) // matches both
	r.Send(makeNotif(80, "open"))   // matches only d1

	if len(d1.Received) != 2 {
		t.Fatalf("d1: expected 2, got %d", len(d1.Received))
	}
	if len(d2.Received) != 1 {
		t.Fatalf("d2: expected 1, got %d", len(d2.Received))
	}
}
