package alert

import (
	"testing"
	"time"
)

func fixedSuppressClock(h, m, s int) func() time.Time {
	return func() time.Time {
		return time.Date(2024, 1, 1, h, m, s, 0, time.UTC)
	}
}

func makeWindow(startH, startM, endH, endM int) TimeWindow {
	return TimeWindow{
		Start: time.Date(0, 1, 1, startH, startM, 0, 0, time.UTC),
		End:   time.Date(0, 1, 1, endH, endM, 0, 0, time.UTC),
	}
}

func TestSuppressor_NoWindows(t *testing.T) {
	s := NewSuppressor(DefaultSuppressPolicy(), fixedSuppressClock(14, 0, 0))
	if s.IsSuppressed() {
		t.Error("expected not suppressed with no windows")
	}
}

func TestSuppressor_WithinWindow(t *testing.T) {
	policy := SuppressPolicy{
		Windows: []TimeWindow{makeWindow(22, 0, 6, 0)},
	}
	// 23:00 — inside the overnight window
	s := NewSuppressor(policy, fixedSuppressClock(23, 0, 0))
	if !s.IsSuppressed() {
		t.Error("expected suppressed at 23:00 in 22:00-06:00 window")
	}
}

func TestSuppressor_OutsideWindow(t *testing.T) {
	policy := SuppressPolicy{
		Windows: []TimeWindow{makeWindow(22, 0, 6, 0)},
	}
	// 14:00 — outside the overnight window
	s := NewSuppressor(policy, fixedSuppressClock(14, 0, 0))
	if s.IsSuppressed() {
		t.Error("expected not suppressed at 14:00")
	}
}

func TestSuppressor_SimpleWindow(t *testing.T) {
	policy := SuppressPolicy{
		Windows: []TimeWindow{makeWindow(9, 0, 17, 0)},
	}
	s := NewSuppressor(policy, fixedSuppressClock(12, 30, 0))
	if !s.IsSuppressed() {
		t.Error("expected suppressed at 12:30 in 09:00-17:00 window")
	}

	s2 := NewSuppressor(policy, fixedSuppressClock(8, 59, 59))
	if s2.IsSuppressed() {
		t.Error("expected not suppressed at 08:59:59")
	}
}

func TestSuppressor_UpdatePolicy(t *testing.T) {
	s := NewSuppressor(DefaultSuppressPolicy(), fixedSuppressClock(12, 0, 0))
	if s.IsSuppressed() {
		t.Fatal("should not be suppressed initially")
	}

	s.UpdatePolicy(SuppressPolicy{
		Windows: []TimeWindow{makeWindow(9, 0, 17, 0)},
	})
	if !s.IsSuppressed() {
		t.Error("expected suppressed after policy update")
	}
}

func TestSuppressor_MultipleWindows(t *testing.T) {
	policy := SuppressPolicy{
		Windows: []TimeWindow{
			makeWindow(0, 0, 1, 0),
			makeWindow(13, 0, 14, 0),
		},
	}
	s := NewSuppressor(policy, fixedSuppressClock(13, 30, 0))
	if !s.IsSuppressed() {
		t.Error("expected suppressed at 13:30")
	}
	s2 := NewSuppressor(policy, fixedSuppressClock(10, 0, 0))
	if s2.IsSuppressed() {
		t.Error("expected not suppressed at 10:00")
	}
}
