package alert

import (
	"errors"
	"testing"
)

// captureDispatcher records every notification it receives.
type captureDispatcher struct {
	got []Notification
	err error
}

func (c *captureDispatcher) Send(n Notification) error {
	c.got = append(c.got, n)
	return c.err
}

func TestFilter_AllowAll(t *testing.T) {
	cap := &captureDispatcher{}
	f := NewFilter(DefaultFilterPolicy(), cap)

	n := Notification{Port: 8080, State: "open"}
	if err := f.Send(n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.got) != 1 {
		t.Fatalf("expected 1 forwarded, got %d", len(cap.got))
	}
}

func TestFilter_PortAllowlist(t *testing.T) {
	cap := &captureDispatcher{}
	f := NewFilter(FilterPolicy{Ports: []int{8080}}, cap)

	f.Send(Notification{Port: 9090, State: "open"})
	f.Send(Notification{Port: 8080, State: "open"})

	if len(cap.got) != 1 {
		t.Fatalf("expected 1 forwarded, got %d", len(cap.got))
	}
	if cap.got[0].Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cap.got[0].Port)
	}
}

func TestFilter_StateAllowlist(t *testing.T) {
	cap := &captureDispatcher{}
	f := NewFilter(FilterPolicy{States: []string{"closed"}}, cap)

	f.Send(Notification{Port: 80, State: "open"})
	f.Send(Notification{Port: 80, State: "closed"})

	if len(cap.got) != 1 {
		t.Fatalf("expected 1 forwarded, got %d", len(cap.got))
	}
	if cap.got[0].State != "closed" {
		t.Fatalf("expected state closed, got %s", cap.got[0].State)
	}
}

func TestFilter_CombinedPortAndState(t *testing.T) {
	cap := &captureDispatcher{}
	f := NewFilter(FilterPolicy{Ports: []int{443}, States: []string{"open"}}, cap)

	f.Send(Notification{Port: 443, State: "closed"}) // wrong state
	f.Send(Notification{Port: 80, State: "open"})    // wrong port
	f.Send(Notification{Port: 443, State: "open"})   // matches

	if len(cap.got) != 1 {
		t.Fatalf("expected 1 forwarded, got %d", len(cap.got))
	}
}

func TestFilter_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream error")
	cap := &captureDispatcher{err: sentinel}
	f := NewFilter(DefaultFilterPolicy(), cap)

	err := f.Send(Notification{Port: 80, State: "open"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestNewFilter_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewFilter(DefaultFilterPolicy(), nil)
}

func TestNewFilterDispatcher_Convenience(t *testing.T) {
	cap := &captureDispatcher{}
	d := NewFilterDispatcher([]int{3000}, []string{"open"}, cap)

	d.Send(Notification{Port: 3000, State: "open"})
	d.Send(Notification{Port: 3000, State: "closed"})

	if len(cap.got) != 1 {
		t.Fatalf("expected 1 forwarded, got %d", len(cap.got))
	}
}
