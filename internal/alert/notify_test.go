package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/config"
)

func TestBuildNotification_Fields(t *testing.T) {
	n := alert.BuildNotification(8080, "open", config.Action{})
	if n.Port != 8080 {
		t.Errorf("expected port 8080, got %d", n.Port)
	}
	if n.State != "open" {
		t.Errorf("expected state 'open', got %s", n.State)
	}
	if n.Message == "" {
		t.Error("expected non-empty message")
	}
	if n.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestDispatcher_Send_Success(t *testing.T) {
	received := make(chan alert.Notification, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var n alert.Notification
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			t.Errorf("decode body: %v", err)
		}
		received <- n
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	policy := alert.DefaultPolicy()
	policy.Cooldown = 0
	limiter := alert.NewLimiter(policy, func() time.Time { return time.Now() })
	d := alert.NewDispatcher(limiter)

	n := alert.Notification{Port: 9090, State: "closed", Timestamp: time.Now().UTC(), Message: "port 9090 is now closed"}
	if err := d.Send(ts.URL, n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case got := <-received:
		if got.Port != 9090 {
			t.Errorf("expected port 9090, got %d", got.Port)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for notification")
	}
}

func TestDispatcher_Send_RateLimited(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	fixed := time.Now()
	policy := alert.DefaultPolicy()
	policy.Cooldown = 5 * time.Minute
	limiter := alert.NewLimiter(policy, func() time.Time { return fixed })
	d := alert.NewDispatcher(limiter)

	n := alert.Notification{Port: 3000, State: "open"}
	_ = d.Send(ts.URL, n)
	err := d.Send(ts.URL, n)
	if err == nil {
		t.Error("expected rate-limit error on second send")
	}
	if calls != 1 {
		t.Errorf("expected 1 HTTP call, got %d", calls)
	}
}

func TestDispatcher_Send_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	policy := alert.DefaultPolicy()
	policy.Cooldown = 0
	limiter := alert.NewLimiter(policy, func() time.Time { return time.Now() })
	d := alert.NewDispatcher(limiter)

	n := alert.Notification{Port: 5000, State: "open"}
	if err := d.Send(ts.URL, n); err == nil {
		t.Error("expected error on 500 response")
	}
}
