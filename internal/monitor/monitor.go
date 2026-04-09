package monitor

import (
	"fmt"
	"net"
	"time"
)

// State represents the availability state of a port.
type State int

const (
	StateUnknown State = iota
	StateOpen
	StateClosed
)

func (s State) String() string {
	switch s {
	case StateOpen:
		return "open"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// PortStatus holds the current state of a monitored port.
type PortStatus struct {
	Host  string
	Port  int
	State State
}

// Checker defines the interface for checking port availability.
type Checker interface {
	Check(host string, port int, timeout time.Duration) State
}

// TCPChecker implements Checker using TCP dial.
type TCPChecker struct{}

// Check attempts a TCP connection to host:port within the given timeout.
func (c *TCPChecker) Check(host string, port int, timeout time.Duration) State {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return StateClosed
	}
	conn.Close()
	return StateOpen
}

// StateChange represents a transition between two port states.
type StateChange struct {
	Host     string
	Port     int
	Previous State
	Current  State
}

// Tracker keeps track of previous port states and emits changes.
type Tracker struct {
	states  map[string]State
	checker Checker
	timeout time.Duration
}

// NewTracker creates a Tracker with the given Checker and dial timeout.
func NewTracker(checker Checker, timeout time.Duration) *Tracker {
	return &Tracker{
		states:  make(map[string]State),
		checker: checker,
		timeout: timeout,
	}
}

// Poll checks the given host:port and returns a StateChange if the state
// has changed since the last call. Returns nil when there is no change.
func (t *Tracker) Poll(host string, port int) *StateChange {
	key := fmt.Sprintf("%s:%d", host, port)
	current := t.checker.Check(host, port, t.timeout)
	previous, exists := t.states[key]
	t.states[key] = current
	if !exists || previous == current {
		return nil
	}
	return &StateChange{
		Host:     host,
		Port:     port,
		Previous: previous,
		Current:  current,
	}
}
