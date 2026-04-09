package poller

import (
	"net"
	"strconv"
	"testing"
	"time"
)

// startListener opens a TCP listener on an ephemeral port and returns the port.
func startListener(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	port, _ := strconv.Atoi(ln.Addr().(*net.TCPAddr).Port.String())
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port64, _ := strconv.ParseInt(portStr, 10, 64)
	return int(port64), func() { ln.Close() }
}

func TestStateString(t *testing.T) {
	if StateOpen.String() != "open" {
		t.Errorf("expected 'open', got %q", StateOpen.String())
	}
	if StateClosed.String() != "closed" {
		t.Errorf("expected 'closed', got %q", StateClosed.String())
	}
}

func TestPoll_OpenPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	p := New(time.Second)
	r := p.Poll("127.0.0.1", port)
	if r.State != StateOpen {
		t.Errorf("expected StateOpen, got %v (err: %v)", r.State, r.Err)
	}
}

func TestPoll_ClosedPort(t *testing.T) {
	p := New(300 * time.Millisecond)
	// Port 1 is almost certainly not open in test environments.
	r := p.Poll("127.0.0.1", 1)
	if r.State != StateClosed {
		t.Errorf("expected StateClosed, got %v", r.State)
	}
	if r.Err == nil {
		t.Error("expected non-nil error for closed port")
	}
}

func TestPollAll_Mixed(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	openPort, _ := strconv.Atoi(portStr)

	targets := []Target{
		{Host: "127.0.0.1", Port: openPort},
		{Host: "127.0.0.1", Port: 1},
	}

	p := New(300 * time.Millisecond)
	results := p.PollAll(targets)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].State != StateOpen {
		t.Errorf("target 0: expected open, got %v", results[0].State)
	}
	if results[1].State != StateClosed {
		t.Errorf("target 1: expected closed, got %v", results[1].State)
	}
}
