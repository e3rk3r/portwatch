package alert

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func makeTestDispatcher(t *testing.T, url string) *Dispatcher {
	t.Helper()
	return NewDispatcher(url, DefaultPolicy(), nil)
}

func TestEscalationDispatcher_NoEscalationBelowThreshold(t *testing.T) {
	var escalated atomic.Int32

	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer primary.Close()

	escSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		escalated.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer escSrv.Close()

	policy := EscalationPolicy{Threshold: 3, MinDuration: 0}
	escCh, _ := NewChannel(ChannelConfig{Type: "webhook", Endpoint: escSrv.URL})
	ed := NewEscalationDispatcher(makeTestDispatcher(t, primary.URL), escCh, policy)

	n := Notification{Port: 9090, State: "closed", Timestamp: time.Now()}
	_ = ed.Send(context.Background(), n)
	_ = ed.Send(context.Background(), n)

	if escalated.Load() != 0 {
		t.Fatalf("expected 0 escalations, got %d", escalated.Load())
	}
}

func TestEscalationDispatcher_EscalatesAtThreshold(t *testing.T) {
	var escalated atomic.Int32

	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer primary.Close()

	escSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		escalated.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer escSrv.Close()

	policy := EscalationPolicy{Threshold: 2, MinDuration: 0}
	escCh, _ := NewChannel(ChannelConfig{Type: "webhook", Endpoint: escSrv.URL})
	ed := NewEscalationDispatcher(makeTestDispatcher(t, primary.URL), escCh, policy)

	n := Notification{Port: 9091, State: "closed", Timestamp: time.Now()}
	_ = ed.Send(context.Background(), n)
	_ = ed.Send(context.Background(), n)

	if escalated.Load() != 1 {
		t.Fatalf("expected 1 escalation, got %d", escalated.Load())
	}
}

func TestEscalationDispatcher_ResetOnOpen(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer primary.Close()

	escSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer escSrv.Close()

	policy := EscalationPolicy{Threshold: 2, MinDuration: 0}
	escCh, _ := NewChannel(ChannelConfig{Type: "webhook", Endpoint: escSrv.URL})
	ed := NewEscalationDispatcher(makeTestDispatcher(t, primary.URL), escCh, policy)

	key := "port:9092:closed"
	ed.escalator.Record(key)

	open := Notification{Port: 9092, State: "open", Timestamp: time.Now()}
	_ = ed.Send(context.Background(), open)

	if ed.escalator.Count(key) != 0 {
		t.Fatal("expected escalator state reset after open notification")
	}
}
