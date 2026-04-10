package alert

import (
	"context"
	"fmt"
	"log"
)

// SuppressDispatcher wraps a Dispatcher and skips delivery when the
// Suppressor indicates the current time is within a suppression window.
type SuppressDispatcher struct {
	inner      *Dispatcher
	suppressor *Suppressor
}

// NewSuppressDispatcher creates a SuppressDispatcher.
func NewSuppressDispatcher(inner *Dispatcher, suppressor *Suppressor) *SuppressDispatcher {
	if inner == nil {
		panic("alert: SuppressDispatcher requires a non-nil Dispatcher")
	}
	if suppressor == nil {
		panic("alert: SuppressDispatcher requires a non-nil Suppressor")
	}
	return &SuppressDispatcher{inner: inner, suppressor: suppressor}
}

// Send delivers the notification via the inner Dispatcher unless the current
// time falls within a configured suppression window, in which case it logs
// and silently drops the notification.
func (sd *SuppressDispatcher) Send(ctx context.Context, n Notification) error {
	if sd.suppressor.IsSuppressed() {
		log.Printf("[portwatch] alert suppressed for port %d (suppression window active)", n.Port)
		return nil
	}
	return sd.inner.Send(ctx, n)
}

// suppressKey returns a string key used to identify a suppression context.
func suppressKey(port int, state string) string {
	return fmt.Sprintf("%d:%s", port, state)
}
