package alert

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeSuppressDispatcher(t *testing.T, srv *httptest.Server) *Dispatcher {
	t.Helper()
	cfg := makeTestDispatcher(t, srv)
	return cfg
}

func TestSuppressDispatcher_AllowsWhenNotSuppressed(t *testing.T) {
	delivered := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delivered = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	inner := makeSuppressDispatcher(t, srv)
	clock := func() time.Time { return time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC) }
	suppressor := NewSuppressor(SuppressPolicy{
		Windows: []TimeWindow{makeWindow(22, 0, 6, 0)},
	}, clock)

	sd := NewSuppressDispatcher(inner, suppressor)
	n := Notification{Port: 8080, State: "open", OccurredAt: time.Now()}
	if err := sd.Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !delivered {
		t.Error("expected notification to be delivered")
	}
}

func TestSuppressDispatcher_SuppressesWhenInWindow(t *testing.T) {
	delivered := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delivered = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	inner := makeSuppressDispatcher(t, srv)
	clock := func() time.Time { return time.Date(2024, 1, 1, 23, 30, 0, 0, time.UTC) }
	suppressor := NewSuppressor(SuppressPolicy{
		Windows: []TimeWindow{makeWindow(22, 0, 6, 0)},
	}, clock)

	sd := NewSuppressDispatcher(inner, suppressor)
	n := Notification{Port: 9090, State: "closed", OccurredAt: time.Now()}
	if err := sd.Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delivered {
		t.Error("expected notification to be suppressed")
	}
}

func TestSuppressDispatcher_NilDispatcherPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Dispatcher")
		}
	}()
	NewSuppressDispatcher(nil, NewSuppressor(DefaultSuppressPolicy(), nil))
}

func TestSuppressDispatcher_NilSuppressorPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Suppressor")
		}
	}()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	NewSuppressDispatcher(makeSuppressDispatcher(t, srv), nil)
}
