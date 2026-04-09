package poller

import (
	"fmt"
	"net"
	"time"
)

// State represents whether a port is open or closed.
type State int

const (
	StateClosed State = iota
	StateOpen
)

func (s State) String() string {
	if s == StateOpen {
		return "open"
	}
	return "closed"
}

// Result holds the outcome of a single port poll.
type Result struct {
	Host  string
	Port  int
	State State
	Err   error
}

// Poller checks TCP port reachability.
type Poller struct {
	timeout time.Duration
}

// New creates a Poller with the given dial timeout.
func New(timeout time.Duration) *Poller {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &Poller{timeout: timeout}
}

// Poll attempts a TCP connection to host:port and returns the result.
func (p *Poller) Poll(host string, port int) Result {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, p.timeout)
	if err != nil {
		return Result{Host: host, Port: port, State: StateClosed, Err: err}
	}
	conn.Close()
	return Result{Host: host, Port: port, State: StateOpen}
}

// PollAll polls each (host, port) pair concurrently and returns all results.
func (p *Poller) PollAll(targets []Target) []Result {
	results := make([]Result, len(targets))
	ch := make(chan indexed, len(targets))

	for i, t := range targets {
		go func(idx int, tgt Target) {
			ch <- indexed{idx: idx, result: p.Poll(tgt.Host, tgt.Port)}
		}(i, t)
	}

	for range targets {
		ir := <-ch
		results[ir.idx] = ir.result
	}
	return results
}

// Target is a host/port pair to poll.
type Target struct {
	Host string
	Port int
}

type indexed struct {
	idx    int
	result Result
}
