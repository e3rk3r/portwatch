package alert

import (
	"context"
	"fmt"
	"sync"
)

// DefaultQuorumPolicy returns a QuorumPolicy requiring a simple majority.
func DefaultQuorumPolicy(total int) QuorumPolicy {
	return QuorumPolicy{
		Total:    total,
		Required: (total/2 + 1),
	}
}

// QuorumPolicy defines how many dispatchers must succeed for a quorum send.
type QuorumPolicy struct {
	Total    int
	Required int
}

// Validate returns an error if the policy is misconfigured.
func (p QuorumPolicy) Validate() error {
	if p.Total <= 0 {
		return fmt.Errorf("quorum: total must be > 0, got %d", p.Total)
	}
	if p.Required <= 0 || p.Required > p.Total {
		return fmt.Errorf("quorum: required must be between 1 and %d, got %d", p.Total, p.Required)
	}
	return nil
}

// QuorumDispatcher sends a notification to all targets concurrently and
// succeeds only when at least Policy.Required targets respond without error.
type QuorumDispatcher struct {
	policy  QuorumPolicy
	targets []Dispatcher
}

// NewQuorumDispatcher creates a QuorumDispatcher. It panics if targets is
// empty or if the policy is invalid.
func NewQuorumDispatcher(policy QuorumPolicy, targets []Dispatcher) *QuorumDispatcher {
	if len(targets) == 0 {
		panic("quorum: targets must not be empty")
	}
	policy.Total = len(targets)
	if err := policy.Validate(); err != nil {
		panic(err)
	}
	return &QuorumDispatcher{policy: policy, targets: targets}
}

// Send dispatches n to all targets in parallel and returns nil when the
// required quorum of successes is reached, or an error otherwise.
func (q *QuorumDispatcher) Send(ctx context.Context, n Notification) error {
	type result struct{ err error }
	results := make(chan result, len(q.targets))

	var wg sync.WaitGroup
	for _, d := range q.targets {
		wg.Add(1)
		go func(d Dispatcher) {
			defer wg.Done()
			results <- result{err: d.Send(ctx, n)}
		}(d)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	successes, failures := 0, 0
	for r := range results {
		if r.err == nil {
			successes++
			if successes >= q.policy.Required {
				return nil
			}
		} else {
			failures++
			if failures > q.policy.Total-q.policy.Required {
				return fmt.Errorf("quorum: failed to reach quorum (%d/%d succeeded)",
					successes, q.policy.Required)
			}
		}
	}
	return fmt.Errorf("quorum: failed to reach quorum (%d/%d succeeded)",
		successes, q.policy.Required)
}
