package alert

import (
	"context"
	"time"
)

// AuditDispatcher wraps a Dispatcher and records every Send outcome
// into an AuditLog.
type AuditDispatcher struct {
	next    Dispatcher
	log     *AuditLog
	channel string
	clock   func() time.Time
}

// NewAuditDispatcher creates an AuditDispatcher that records outcomes for
// notifications forwarded to next. channel is a label stored in each entry.
func NewAuditDispatcher(next Dispatcher, log *AuditLog, channel string) *AuditDispatcher {
	if next == nil {
		panic("audit: next dispatcher must not be nil")
	}
	if log == nil {
		panic("audit: log must not be nil")
	}
	return &AuditDispatcher{
		next:    next,
		log:     log,
		channel: channel,
		clock:   time.Now,
	}
}

// Send forwards the notification to the wrapped dispatcher and logs the result.
func (a *AuditDispatcher) Send(ctx context.Context, n Notification) error {
	err := a.next.Send(ctx, n)
	entry := AuditEntry{
		Timestamp: a.clock(),
		Port:      n.Port,
		State:     n.State,
		Channel:   a.channel,
		Success:   err == nil,
	}
	if err != nil {
		entry.Err = err.Error()
	}
	a.log.Record(entry)
	return err
}
