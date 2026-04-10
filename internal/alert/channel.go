// Package alert provides rate-limiting, notification dispatch, and
// delivery channel abstractions for portwatch state-change alerts.
package alert

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ChannelType identifies the delivery mechanism for an alert.
type ChannelType string

const (
	ChannelWebhook ChannelType = "webhook"
	ChannelScript  ChannelType = "script"
	ChannelLog     ChannelType = "log"
)

// ChannelConfig holds the configuration for a single delivery channel.
type ChannelConfig struct {
	Type    ChannelType   `yaml:"type"`
	Target  string        `yaml:"target"` // URL for webhook, path for script
	Timeout time.Duration `yaml:"timeout"`
}

// Channel is the interface every delivery backend must satisfy.
type Channel interface {
	// Send delivers the notification. Returns an error if delivery fails.
	Send(ctx context.Context, n Notification) error
}

// NewChannel constructs the appropriate Channel for the given config.
func NewChannel(cfg ChannelConfig) (Channel, error) {
	switch cfg.Type {
	case ChannelWebhook:
		return &webhookChannel{target: cfg.Target, timeout: cfg.Timeout}, nil
	case ChannelScript:
		return &scriptChannel{path: cfg.Target}, nil
	case ChannelLog:
		return &logChannel{}, nil
	default:
		return nil, fmt.Errorf("unknown channel type: %q", cfg.Type)
	}
}

// --- log channel (always available, useful for debugging) ---

type logChannel struct{}

func (l *logChannel) Send(_ context.Context, n Notification) error {
	log.Printf("[portwatch] alert port=%d state=%s ts=%s",
		n.Port, n.State, n.Timestamp.Format(time.RFC3339))
	return nil
}

// --- webhook channel ---

type webhookChannel struct {
	target  string
	timeout time.Duration
}

func (w *webhookChannel) Send(ctx context.Context, n Notification) error {
	timeout := w.timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return fireWebhookCtx(ctx, w.target, n)
}

// --- script channel ---

type scriptChannel struct {
	path string
}

func (s *scriptChannel) Send(ctx context.Context, n Notification) error {
	return fireScriptCtx(ctx, s.path, n)
}
