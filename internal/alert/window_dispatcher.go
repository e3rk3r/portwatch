package alert

import "fmt"

// WindowDispatcher wraps a Dispatcher and gates delivery through a
// WindowCounter so that at most MaxEvents notifications per key are
// forwarded within the configured sliding window.
type WindowDispatcher struct {
	next    Dispatcher
	counter *WindowCounter
}

// NewWindowDispatcher creates a WindowDispatcher.
// Panics if next or counter is nil.
func NewWindowDispatcher(next Dispatcher, counter *WindowCounter) *WindowDispatcher {
	if next == nil {
		panic("alert: WindowDispatcher next dispatcher must not be nil")
	}
	if counter == nil {
		panic("alert: WindowDispatcher counter must not be nil")
	}
	return &WindowDispatcher{next: next, counter: counter}
}

// Send forwards n to the next dispatcher only when the sliding window
// allows it. Returns a descriptive error when the limit is exceeded.
func (d *WindowDispatcher) Send(n Notification) error {
	key := windowKey(n)
	if !d.counter.Allow(key) {
		return fmt.Errorf("alert: window limit reached for key %q", key)
	}
	return d.next.Send(n)
}

func windowKey(n Notification) string {
	return fmt.Sprintf("%d:%s", n.Port, n.State)
}
