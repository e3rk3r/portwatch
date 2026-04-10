package alert

import (
	"sync"
	"testing"
	"time"
)

// captureDispatcher records every notification sent to it.
type captureDispatcher struct {
	mu    sync.Mutex
	sent  []Notification
}

func (c *captureDispatcher) Send(n Notification) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, n)
	return nil
}

func (c *captureDispatcher) Received() []Notification {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Notification, len(c.sent))
	copy(out, c.sent)
	return out
}

func TestDigester_SingleEventFlushed(t *testing.T) {
	cap := &captureDispatcher{}
	d := NewDigester(DefaultDigestPolicy(), cap)

	n := Notification{Port: 8080, State: "open", Message: "port opened"}
	d.Add(n)
	d.Flush()

	got := cap.Received()
	if len(got) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(got))
	}
	if got[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", got[0].Port)
	}
}

func TestDigester_DuplicatesAggregated(t *testing.T) {
	cap := &captureDispatcher{}
	d := NewDigester(DefaultDigestPolicy(), cap)

	n := Notification{Port: 9090, State: "closed", Message: "port closed"}
	d.Add(n)
	d.Add(n)
	d.Add(n)
	d.Flush()

	got := cap.Received()
	if len(got) != 1 {
		t.Fatalf("expected 1 aggregated notification, got %d", len(got))
	}
	if got[0].Message == "port closed" {
		t.Errorf("expected digest suffix in message, got: %s", got[0].Message)
	}
}

func TestDigester_DistinctKeysFlushSeparately(t *testing.T) {
	cap := &captureDispatcher{}
	d := NewDigester(DefaultDigestPolicy(), cap)

	d.Add(Notification{Port: 80, State: "open", Message: "80 open"})
	d.Add(Notification{Port: 443, State: "closed", Message: "443 closed"})
	d.Flush()

	got := cap.Received()
	if len(got) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(got))
	}
}

func TestDigester_FlushClearsBucket(t *testing.T) {
	cap := &captureDispatcher{}
	d := NewDigester(DefaultDigestPolicy(), cap)

	d.Add(Notification{Port: 8080, State: "open", Message: "first"})
	d.Flush()
	d.Flush() // second flush should send nothing

	got := cap.Received()
	if len(got) != 1 {
		t.Errorf("expected 1 notification after double flush, got %d", len(got))
	}
}

func TestDigester_TimerFlush(t *testing.T) {
	cap := &captureDispatcher{}
	policy := DigestPolicy{Window: 50 * time.Millisecond}
	d := NewDigester(policy, cap)

	d.Add(Notification{Port: 3000, State: "open", Message: "timer test"})

	time.Sleep(120 * time.Millisecond)

	got := cap.Received()
	if len(got) != 1 {
		t.Errorf("expected timer to flush 1 notification, got %d", len(got))
	}
}
