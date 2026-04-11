package alert

import "fmt"

// DedupDispatcher wraps a Dispatcher and drops notifications whose key was
// already delivered within the dedup window.
type DedupDispatcher struct {
	next  Dispatcher
	dedup *Deduplicator
}

// NewDedupDispatcher creates a DedupDispatcher that forwards to next only
// when the Deduplicator permits the notification.
func NewDedupDispatcher(next Dispatcher, d *Deduplicator) *DedupDispatcher {
	if next == nil {
		panic("dedup: next dispatcher must not be nil")
	}
	if d == nil {
		panic("dedup: deduplicator must not be nil")
	}
	return &DedupDispatcher{next: next, dedup: d}
}

// Send forwards n to the wrapped Dispatcher only if the deduplicator allows
// it. Suppressed notifications are silently dropped.
func (dd *DedupDispatcher) Send(n Notification) error {
	key := dedupKey(n)
	if !dd.dedup.Allow(key) {
		return nil
	}
	return dd.next.Send(n)
}

// dedupKey builds a string key from the notification's port and state.
func dedupKey(n Notification) string {
	return fmt.Sprintf("%d:%s", n.Port, n.State)
}
