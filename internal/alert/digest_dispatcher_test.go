package alert

import (
	"testing"
)

func makeDigestDispatcher(t *testing.T) (*DigestDispatcher, *captureDispatcher) {
	t.Helper()
	cap := &captureDispatcher{}
	dd := NewDigestDispatcher(DefaultDigestPolicy(), cap)
	return dd, cap
}

func TestDigestDispatcher_SendAccumulates(t *testing.T) {
	dd, cap := makeDigestDispatcher(t)

	_ = dd.Send(Notification{Port: 8080, State: "open", Message: "opened"})

	// Not flushed yet — nothing should be delivered.
	if len(cap.Received()) != 0 {
		t.Errorf("expected 0 delivered before flush, got %d", len(cap.Received()))
	}

	dd.Flush()

	if len(cap.Received()) != 1 {
		t.Fatalf("expected 1 after flush, got %d", len(cap.Received()))
	}
}

func TestDigestDispatcher_MultiplePortsDeliveredSeparately(t *testing.T) {
	dd, cap := makeDigestDispatcher(t)

	_ = dd.Send(Notification{Port: 22, State: "closed", Message: "ssh gone"})
	_ = dd.Send(Notification{Port: 80, State: "open", Message: "http up"})
	dd.Flush()

	if len(cap.Received()) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(cap.Received()))
	}
}

func TestDigestKey_Format(t *testing.T) {
	n := Notification{Port: 443, State: "open"}
	key := digestKey(n)
	if key != "443:open" {
		t.Errorf("unexpected digest key: %s", key)
	}
}

func TestDigestDispatcher_FlushTwiceNoDuplicates(t *testing.T) {
	dd, cap := makeDigestDispatcher(t)

	_ = dd.Send(Notification{Port: 9000, State: "open", Message: "test"})
	dd.Flush()
	dd.Flush()

	if len(cap.Received()) != 1 {
		t.Errorf("expected 1 notification, got %d after double flush", len(cap.Received()))
	}
}
