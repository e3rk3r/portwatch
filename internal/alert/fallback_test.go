package alert

import (
	"context"
	"errors"
	"log"
	"io"
	"testing"
)

// errDispatcher always returns the configured error from Send.
type errDispatcher struct{ err error }

func (e *errDispatcher) Send(_ context.Context, _ Notification) error { return e.err }

// okDispatcher always succeeds and records the last notification it received.
type okDispatcher struct{ last Notification }

func (o *okDispatcher) Send(_ context.Context, n Notification) error {
	o.last = n
	return nil
}

var silentLogger = log.New(io.Discard, "", 0)

func TestFallbackDispatcher_PanicsOnNilPrimary(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil primary")
		}
	}()
	NewFallbackDispatcher(nil, &okDispatcher{}, silentLogger)
}

func TestFallbackDispatcher_PanicsOnNilFallback(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil fallback")
		}
	}()
	NewFallbackDispatcher(&okDispatcher{}, nil, silentLogger)
}

func TestFallbackDispatcher_PrimarySucceeds(t *testing.T) {
	primary := &okDispatcher{}
	fallback := &okDispatcher{}
	d := NewFallbackDispatcher(primary, fallback, silentLogger)

	n := Notification{Port: 8080, State: "open"}
	if err := d.Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if primary.last.Port != 8080 {
		t.Errorf("primary should have received notification")
	}
	if fallback.last.Port != 0 {
		t.Errorf("fallback should not have been called")
	}
}

func TestFallbackDispatcher_UseFallbackOnPrimaryError(t *testing.T) {
	primErr := errors.New("primary down")
	primary := &errDispatcher{err: primErr}
	fallback := &okDispatcher{}
	d := NewFallbackDispatcher(primary, fallback, silentLogger)

	n := Notification{Port: 9090, State: "closed"}
	if err := d.Send(context.Background(), n); err != nil {
		t.Fatalf("fallback should have succeeded, got: %v", err)
	}
	if fallback.last.Port != 9090 {
		t.Errorf("fallback should have received notification")
	}
}

func TestFallbackDispatcher_BothFail(t *testing.T) {
	primErr := errors.New("primary down")
	fbErr := errors.New("fallback down")
	d := NewFallbackDispatcher(&errDispatcher{err: primErr}, &errDispatcher{err: fbErr}, silentLogger)

	err := d.Send(context.Background(), Notification{Port: 443, State: "open"})
	if !errors.Is(err, fbErr) {
		t.Errorf("expected fallback error, got: %v", err)
	}
}
