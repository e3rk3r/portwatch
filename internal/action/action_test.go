package action

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/config"
)

func makeConfig(port int, actionType, on, url, path string) *config.Config {
	return &config.Config{
		Ports: []config.PortConfig{
			{
				Port: port,
				Actions: []config.Action{
					{Type: actionType, On: on, URL: url, Path: path},
				},
			},
		},
	}
}

func TestRun_WebhookOnOpen(t *testing.T) {
	var got webhookPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := makeConfig(8080, "webhook", "open", ts.URL, "")
	ex := NewExecutor(cfg)
	if err := ex.Run(context.Background(), 8080, "open"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Port != 8080 || got.State != "open" {
		t.Errorf("payload mismatch: got %+v", got)
	}
}

func TestRun_SkipsNonMatchingState(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := makeConfig(9090, "webhook", "closed", ts.URL, "")
	ex := NewExecutor(cfg)
	_ = ex.Run(context.Background(), 9090, "open")
	if called {
		t.Error("webhook should not have been called for non-matching state")
	}
}

func TestRun_UnknownPort(t *testing.T) {
	cfg := makeConfig(1234, "webhook", "open", "http://localhost", "")
	ex := NewExecutor(cfg)
	if err := ex.Run(context.Background(), 9999, "open"); err == nil {
		t.Error("expected error for unknown port")
	}
}
