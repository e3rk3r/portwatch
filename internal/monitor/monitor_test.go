package monitor

import (
	"testing"
	"time"
)

// stubChecker allows tests to control what Check returns.
type stubChecker struct {
	results map[string]State
}

func (s *stubChecker) Check(host string, port int, _ time.Duration) State {
	key := stateKey(host, port)
	if st, ok := s.results[key]; ok {
		return st
	}
	return StateClosed
}

func stateKey(host string, port int) string {
	return host + ":" + string(rune('0'+port)) // simple key for test ports 1-9
}

func newStub(host string, port int, state State) *stubChecker {
	sc := &stubChecker{results: make(map[string]State)}
	sc.results[stateKey(host, port)] = state
	return sc
}

func TestStateString(t *testing.T) {
	cases := []struct {
		state State
		want  string
	}{
		{StateOpen, "open"},
		{StateClosed, "closed"},
		{StateUnknown, "unknown"},
	}
	for _, tc := range cases {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", tc.state, got, tc.want)
		}
	}
}

func TestTracker_FirstPollNoChange(t *testing.T) {
	sc := newStub("localhost", 3, StateOpen)
	tracker := NewTracker(sc, time.Second)
	change := tracker.Poll("localhost", 3)
	if change != nil {
		t.Errorf("expected nil on first poll, got %+v", change)
	}
}

func TestTracker_DetectsOpenToClosed(t *testing.T) {
	sc := newStub("localhost", 3, StateOpen)
	tracker := NewTracker(sc, time.Second)
	tracker.Poll("localhost", 3) // seed state as open

	sc.results[stateKey("localhost", 3)] = StateClosed
	change := tracker.Poll("localhost", 3)
	if change == nil {
		t.Fatal("expected a StateChange, got nil")
	}
	if change.Previous != StateOpen || change.Current != StateClosed {
		t.Errorf("unexpected change: %+v", change)
	}
}

func TestTracker_DetectsClosedToOpen(t *testing.T) {
	sc := newStub("localhost", 4, StateClosed)
	tracker := NewTracker(sc, time.Second)
	tracker.Poll("localhost", 4)

	sc.results[stateKey("localhost", 4)] = StateOpen
	change := tracker.Poll("localhost", 4)
	if change == nil {
		t.Fatal("expected a StateChange, got nil")
	}
	if change.Previous != StateClosed || change.Current != StateOpen {
		t.Errorf("unexpected change: %+v", change)
	}
}

func TestTracker_NoChangeWhenStateSame(t *testing.T) {
	sc := newStub("localhost", 5, StateOpen)
	tracker := NewTracker(sc, time.Second)
	tracker.Poll("localhost", 5)
	change := tracker.Poll("localhost", 5)
	if change != nil {
		t.Errorf("expected nil when state unchanged, got %+v", change)
	}
}
