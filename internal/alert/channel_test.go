package alert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func sampleNotification() Notification {
	return Notification{
		Port:      8080,
		State:     "open",
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestNewChannel_UnknownType(t *testing.T) {
	_, err := NewChannel(ChannelConfig{Type: "email"})
	if err == nil {
		t.Fatal("expected error for unknown channel type")
	}
}

func TestNewChannel_LogType(t *testing.T) {
	ch, err := NewChannel(ChannelConfig{Type: ChannelLog})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ch.Send(context.Background(), sampleNotification()); err != nil {
		t.Fatalf("log channel send failed: %v", err)
	}
}

func TestNewChannel_WebhookSuccess(t *testing.T) {
	var received Notification
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ch, err := NewChannel(ChannelConfig{Type: ChannelWebhook, Target: ts.URL, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewChannel: %v", err)
	}

	n := sampleNotification()
	if err := ch.Send(context.Background(), n); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if received.Port != n.Port {
		t.Errorf("port mismatch: got %d want %d", received.Port, n.Port)
	}
}

func TestNewChannel_WebhookNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	ch, _ := NewChannel(ChannelConfig{Type: ChannelWebhook, Target: ts.URL})
	if err := ch.Send(context.Background(), sampleNotification()); err == nil {
		t.Fatal("expected error on 500 response")
	}
}

func TestScriptChannel_EmptyPath(t *testing.T) {
	ch, err := NewChannel(ChannelConfig{Type: ChannelScript, Target: ""})
	if err != nil {
		t.Fatalf("NewChannel: %v", err)
	}
	if err := ch.Send(context.Background(), sampleNotification()); err == nil {
		t.Fatal("expected error for empty script path")
	}
}
