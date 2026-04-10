package alert_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/alert"
)

// TestLimiter_MultiPortScenario simulates multiple ports firing alerts and
// verifies that each port's cooldown is tracked independently.
func TestLimiter_MultiPortScenario(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := base

	l := alert.NewLimiter(alert.Policy{Cooldown: 5 * time.Second})
	// Inject controllable clock via the exported field path used in unit tests
	// is unexported; use the public API only for integration coverage.

	ports := []int{8080, 9090, 3000}
	allowed := 0

	for _, p := range ports {
		key := fmt.Sprintf("localhost:%d:open", p)
		if l.Allow(key) {
			allowed++
		}
		_ = clock // clock used conceptually; real time is fine here
	}

	if allowed != len(ports) {
		t.Fatalf("expected %d first-time allows, got %d", len(ports), allowed)
	}

	// Immediate second pass — all should be suppressed.
	suppressed := 0
	for _, p := range ports {
		key := fmt.Sprintf("localhost:%d:open", p)
		if !l.Allow(key) {
			suppressed++
		}
	}

	if suppressed != len(ports) {
		t.Fatalf("expected all %d keys suppressed, got %d suppressed", len(ports), suppressed)
	}
}

// TestLimiter_ResetRestoresAllKeys verifies batch-reset behaviour.
func TestLimiter_ResetRestoresAllKeys(t *testing.T) {
	l := alert.NewLimiter(alert.Policy{Cooldown: 24 * time.Hour})
	keys := []string{"a", "b", "c"}

	for _, k := range keys {
		l.Allow(k)
	}
	for _, k := range keys {
		l.Reset(k)
	}
	for _, k := range keys {
		if !l.Allow(k) {
			t.Fatalf("key %q should be allowed after reset", k)
		}
	}
}
