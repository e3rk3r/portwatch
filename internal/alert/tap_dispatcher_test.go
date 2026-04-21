package alert

import (
	"context"
	"sync"
	"testing"
)

// TestTapDispatcher_ConcurrentSafe verifies that concurrent sends do not
// race on the tap's internal buffer.
func TestTapDispatcher_ConcurrentSafe(t *testing.T) {
	tap := NewTap(&TapPolicy{MaxCapacity: 64})
	cap := &captureDispatcher{}
	d := NewTapDispatcher(tap, cap)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(port int) {
			defer wg.Done()
			_ = d.Send(context.Background(), tapNotif(port))
		}(i)
	}
	wg.Wait()

	if tap.Len() == 0 {
		t.Fatal("expected at least one recorded notification")
	}
}

// TestTapDispatcher_SnapshotIsolation confirms that mutating the returned
// snapshot does not affect the tap's internal state.
func TestTapDispatcher_SnapshotIsolation(t *testing.T) {
	tap := NewTap(nil)
	cap := &captureDispatcher{}
	d := NewTapDispatcher(tap, cap)

	_ = d.Send(context.Background(), tapNotif(1111))
	snap := tap.Snapshot()
	snap[0].Port = 9999 // mutate the copy

	if tap.Snapshot()[0].Port == 9999 {
		t.Fatal("snapshot mutation affected internal tap buffer")
	}
}

// TestTapDispatcher_ZeroCapacityFallsBackToDefault ensures a zero-value
// MaxCapacity is treated as invalid and replaced with the default.
func TestTapDispatcher_ZeroCapacityFallsBackToDefault(t *testing.T) {
	tap := NewTap(&TapPolicy{MaxCapacity: 0})
	if tap.policy.MaxCapacity != DefaultTapPolicy().MaxCapacity {
		t.Fatalf("expected default capacity, got %d", tap.policy.MaxCapacity)
	}
}
