package alert

import (
	"fmt"
	"sync"
	"time"
)

// DigestPolicy controls how digests are batched and flushed.
type DigestPolicy struct {
	// Window is how long to accumulate events before flushing.
	Window time.Duration
}

// DefaultDigestPolicy returns a sensible default digest policy.
func DefaultDigestPolicy() DigestPolicy {
	return DigestPolicy{
		Window: 5 * time.Minute,
	}
}

// DigestEntry holds a single accumulated notification.
type DigestEntry struct {
	Notification Notification
	Count        int
}

// Digester batches notifications within a time window and flushes them
// as a single summary to an underlying Dispatcher.
type Digester struct {
	policy     DigestPolicy
	next       Dispatcher
	mu         sync.Mutex
	bucket     map[string]*DigestEntry
	flushTimer *time.Timer
	clock      func() time.Time
}

// NewDigester creates a Digester that flushes batched alerts via next.
func NewDigester(policy DigestPolicy, next Dispatcher) *Digester {
	d := &Digester{
		policy: policy,
		next:   next,
		bucket: make(map[string]*DigestEntry),
		clock:  time.Now,
	}
	return d
}

// Add accumulates a notification into the current digest window.
// The first notification in a window starts the flush timer.
func (d *Digester) Add(n Notification) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := fmt.Sprintf("%d:%s", n.Port, n.State)
	if entry, ok := d.bucket[key]; ok {
		entry.Count++
	} else {
		d.bucket[key] = &DigestEntry{Notification: n, Count: 1}
	}

	if d.flushTimer == nil {
		d.flushTimer = time.AfterFunc(d.policy.Window, d.flush)
	}
}

// Flush forces an immediate flush of accumulated notifications.
func (d *Digester) Flush() {
	d.flush()
}

func (d *Digester) flush() {
	d.mu.Lock()
	entries := d.bucket
	d.bucket = make(map[string]*DigestEntry)
	if d.flushTimer != nil {
		d.flushTimer.Stop()
		d.flushTimer = nil
	}
	d.mu.Unlock()

	for _, entry := range entries {
		n := entry.Notification
		if entry.Count > 1 {
			n.Message = fmt.Sprintf("%s (x%d in digest window)", n.Message, entry.Count)
		}
		_ = d.next.Send(n)
	}
}
