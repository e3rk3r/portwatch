package alert

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// captureDispatcher records every notification it receives and can be
// configured to return a fixed error.
type captureDispatcher struct {
	mu   sync.Mutex
	msgs []Notification
	err  error
}

func (c *captureDispatcher) Send(n Notification) error {
	c.mu.Lock()
	c.msgs = append(c.msgs, n)
	c.mu.Unlock()
	return c.err
}

func (c *captureDispatcher) received() []Notification {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Notification, len(c.msgs))
	copy(out, c.msgs)
	return out
}

func shadowNotif(port int) Notification {
	return Notification{Port: port, State: "open", Title: "test"}
}

func TestShadowDispatcher_NilPrimaryPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil primary")
		}
	}()
	NewShadowDispatcher(nil, &captureDispatcher{}, DefaultShadowPolicy())
}

func TestShadowDispatcher_NilShadowPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil shadow")
		}
	}()
	NewShadowDispatcher(&captureDispatcher{}, nil, DefaultShadowPolicy())
}

func TestShadowDispatcher_BothReceiveNotification(t *testing.T) {
	primary := &captureDispatcher{}
	shadow := &captureDispatcher{}
	sd := NewShadowDispatcher(primary, shadow, DefaultShadowPolicy())

	n := shadowNotif(8080)
	if err := sd.Send(n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give the goroutine time to complete.
	time.Sleep(20 * time.Millisecond)

	if got := primary.received(); len(got) != 1 || got[0].Port != 8080 {
		t.Errorf("primary did not receive notification, got %v", got)
	}
	if got := shadow.received(); len(got) != 1 || got[0].Port != 8080 {
		t.Errorf("shadow did not receive notification, got %v", got)
	}
}

func TestShadowDispatcher_PrimaryErrorReturned(t *testing.T) {
	primErr := errors.New("primary failure")
	primary := &captureDispatcher{err: primErr}
	shadow := &captureDispatcher{}
	sd := NewShadowDispatcher(primary, shadow, DefaultShadowPolicy())

	if err := sd.Send(shadowNotif(9090)); !errors.Is(err, primErr) {
		t.Errorf("expected primary error, got %v", err)
	}
}

func TestShadowDispatcher_DivergenceRecorded(t *testing.T) {
	primary := &captureDispatcher{}
	shadow := &captureDispatcher{err: errors.New("shadow failure")}
	sd := NewShadowDispatcher(primary, shadow, DefaultShadowPolicy())

	sd.Send(shadowNotif(3000)) //nolint:errcheck
	time.Sleep(30 * time.Millisecond)

	divs := sd.Divergences()
	if len(divs) != 1 {
		t.Fatalf("expected 1 divergence, got %d", len(divs))
	}
	if divs[0].Primary != nil {
		t.Errorf("expected primary nil, got %v", divs[0].Primary)
	}
	if divs[0].Shadow == nil {
		t.Error("expected shadow error, got nil")
	}
}

func TestShadowDispatcher_ResetClearsDivergences(t *testing.T) {
	primary := &captureDispatcher{}
	shadow := &captureDispatcher{err: errors.New("err")}
	sd := NewShadowDispatcher(primary, shadow, DefaultShadowPolicy())

	sd.Send(shadowNotif(4000)) //nolint:errcheck
	time.Sleep(30 * time.Millisecond)

	sd.Reset()
	if got := sd.Divergences(); len(got) != 0 {
		t.Errorf("expected empty after reset, got %d", len(got))
	}
}
