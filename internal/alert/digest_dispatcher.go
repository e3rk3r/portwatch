package alert

import "fmt"

// DigestDispatcher wraps a Digester and implements the Dispatcher interface
// so it can be composed with other middleware (suppress, escalation, etc.).
type DigestDispatcher struct {
	digester *Digester
}

// NewDigestDispatcher creates a DigestDispatcher using the given policy and
// downstream Dispatcher.
func NewDigestDispatcher(policy DigestPolicy, next Dispatcher) *DigestDispatcher {
	return &DigestDispatcher{
		digester: NewDigester(policy, next),
	}
}

// Send accumulates the notification into the current digest window.
func (d *DigestDispatcher) Send(n Notification) error {
	if d.digester == nil {
		return fmt.Errorf("digest dispatcher: digester is nil")
	}
	d.digester.Add(n)
	return nil
}

// Flush forces the underlying Digester to flush all pending notifications.
func (d *DigestDispatcher) Flush() {
	d.digester.Flush()
}

// digestKey returns a string key that uniquely identifies a port+state pair
// for use in digest bucketing.
func digestKey(n Notification) string {
	return fmt.Sprintf("%d:%s", n.Port, n.State)
}
