package alert

import (
	"context"
	"errors"
	"testing"
)

type capturingDispatcher struct {
	received []Notification
	err      error
}

func (c *capturingDispatcher) Send(_ context.Context, n Notification) error {
	c.received = append(c.received, n)
	return c.err
}

func TestMuxDispatcher_NilMuxPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil mux")
		}
	}()
	NewMuxDispatcher(nil, map[string]Dispatcher{"k": &capturingDispatcher{}}, nil)
}

func TestMuxDispatcher_EmptyRoutesPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty routes")
		}
	}()
	NewMuxDispatcher(NewMux(nil), map[string]Dispatcher{}, nil)
}

func TestMuxDispatcher_NilRoutePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil route dispatcher")
		}
	}()
	NewMuxDispatcher(NewMux(nil), map[string]Dispatcher{"k": nil}, nil)
}

func TestMuxDispatcher_MatchedRouteReceivesNotification(t *testing.T) {
	ctx := context.Background()
	n := muxNotif(8080, "open")
	key := muxKey(n)

	target := &capturingDispatcher{}
	d := NewMuxDispatcher(NewMux(nil), map[string]Dispatcher{key: target}, nil)

	if err := d.Send(ctx, n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(target.received) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(target.received))
	}
}

func TestMuxDispatcher_FallbackUsedWhenNoMatch(t *testing.T) {
	ctx := context.Background()
	n := muxNotif(9090, "closed")

	fb := &capturingDispatcher{}
	primary := &capturingDispatcher{}
	d := NewMuxDispatcher(NewMux(nil), map[string]Dispatcher{"nomatch": primary}, fb)

	if err := d.Send(ctx, n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fb.received) != 1 {
		t.Fatalf("expected fallback to receive 1 notification, got %d", len(fb.received))
	}
	if len(primary.received) != 0 {
		t.Fatal("primary should not have been called")
	}
}

func TestMuxDispatcher_NoMatchNoFallbackDropsSilently(t *testing.T) {
	ctx := context.Background()
	n := muxNotif(1234, "open")

	target := &capturingDispatcher{}
	d := NewMuxDispatcher(NewMux(nil), map[string]Dispatcher{"other:closed": target}, nil)

	if err := d.Send(ctx, n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(target.received) != 0 {
		t.Fatal("expected no dispatches")
	}
}

func TestMuxDispatcher_PropagatesTargetError(t *testing.T) {
	ctx := context.Background()
	n := muxNotif(8080, "open")
	key := muxKey(n)

	wantErr := errors.New("dispatch failed")
	target := &capturingDispatcher{err: wantErr}
	d := NewMuxDispatcher(NewMux(nil), map[string]Dispatcher{key: target}, nil)

	if err := d.Send(ctx, n); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
