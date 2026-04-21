package alert

import "context"

// captureDispatcher is a test helper that records the last notification
// received and optionally returns a pre-configured error.
type captureDispatcher struct {
	last Notification
	err  error
}

func (c *captureDispatcher) Send(_ context.Context, n Notification) error {
	c.last = n
	return c.err
}
