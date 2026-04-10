// Package alert provides rate-limiting and notification dispatch for port state changes.
package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/portwatch/internal/config"
)

// Notification represents a structured alert payload sent to notification endpoints.
type Notification struct {
	Port      int       `json:"port"`
	State     string    `json:"state"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// Dispatcher sends notifications for port state changes, subject to rate limiting.
type Dispatcher struct {
	limiter *Limiter
	client  *http.Client
}

// NewDispatcher creates a Dispatcher using the provided Limiter and a default HTTP client.
func NewDispatcher(l *Limiter) *Dispatcher {
	return &Dispatcher{
		limiter: l,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Send dispatches a Notification to the given URL if the rate limiter permits it.
// Returns an error if the request fails or is rate-limited.
func (d *Dispatcher) Send(url string, n Notification) error {
	key := fmt.Sprintf("%d:%s", n.Port, n.State)
	if !d.limiter.Allow(key) {
		return fmt.Errorf("alert suppressed for %s (cooldown active)", key)
	}

	body, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	resp, err := d.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("notification endpoint returned %d", resp.StatusCode)
	}
	return nil
}

// BuildNotification constructs a Notification from a config entry and current state string.
func BuildNotification(port int, state string, cfg config.Action) Notification {
	return Notification{
		Port:      port,
		State:     state,
		Timestamp: time.Now().UTC(),
		Message:   fmt.Sprintf("port %d is now %s", port, state),
	}
}
