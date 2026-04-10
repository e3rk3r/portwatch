package alert

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstCallAlwaysPasses(t *testing.T) {
	l := NewLimiter(DefaultPolicy())
	if !l.Allow("host:8080:open") {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestAllow_BlockedWithinCooldown(t *testing.T) {
	base := time.Now()
	l := NewLimiter(Policy{Cooldown: 10 * time.Second})
	l.now = fixedClock(base)

	l.Allow("key")

	// Advance only 5 s — still within cooldown.
	l.now = fixedClock(base.Add(5 * time.Second))
	if l.Allow("key") {
		t.Fatal("expected alert to be suppressed within cooldown")
	}
}

func TestAllow_PassesAfterCooldown(t *testing.T) {
	base := time.Now()
	l := NewLimiter(Policy{Cooldown: 10 * time.Second})
	l.now = fixedClock(base)

	l.Allow("key")

	// Advance past cooldown.
	l.now = fixedClock(base.Add(11 * time.Second))
	if !l.Allow("key") {
		t.Fatal("expected alert to pass after cooldown expired")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	base := time.Now()
	l := NewLimiter(Policy{Cooldown: 60 * time.Second})
	l.now = fixedClock(base)

	l.Allow("keyA")

	// keyB has never fired — should pass even though keyA is suppressed.
	if !l.Allow("keyB") {
		t.Fatal("expected independent key to pass")
	}
}

func TestReset_AllowsImmediateRetry(t *testing.T) {
	base := time.Now()
	l := NewLimiter(Policy{Cooldown: 60 * time.Second})
	l.now = fixedClock(base)

	l.Allow("key")
	l.Reset("key")

	// Same timestamp, but reset cleared the record.
	if !l.Allow("key") {
		t.Fatal("expected allow after reset")
	}
}
