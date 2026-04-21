package alert_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/yourorg/portwatch/internal/alert"
)

type countDispatcher struct {
	called int64
	err    error
}

func (c *countDispatcher) Send(_ context.Context, _ alert.Notification) error {
	atomic.AddInt64(&c.called, 1)
	return c.err
}

func TestQuorum_PipelineIntegration(t *testing.T) {
	d1 := &countDispatcher{}
	d2 := &countDispatcher{}
	d3 := &countDispatcher{err: errors.New("down")}

	policy := alert.QuorumPolicy{Total: 3, Required: 2}
	q := alert.NewQuorumDispatcher(policy, []alert.Dispatcher{d1, d2, d3})

	obs := alert.NewObserver(alert.DefaultObservePolicy())
	pipe := alert.NewPipeline(
		alert.NewObserveDispatcher(obs, q),
	)

	n := alert.Notification{Port: 9090, State: "closed"}
	if err := pipe.Send(context.Background(), n); err != nil {
		t.Fatalf("pipeline quorum should succeed: %v", err)
	}

	snap := obs.Snapshot()
	if snap.Success != 1 {
		t.Fatalf("expected 1 success recorded, got %d", snap.Success)
	}
}

func TestQuorum_AllTargetsCalledEvenOnQuorumMet(t *testing.T) {
	d1 := &countDispatcher{}
	d2 := &countDispatcher{}
	d3 := &countDispatcher{}

	policy := alert.QuorumPolicy{Total: 3, Required: 1}
	q := alert.NewQuorumDispatcher(policy, []alert.Dispatcher{d1, d2, d3})

	_ = q.Send(context.Background(), alert.Notification{Port: 80, State: "open"})

	total := atomic.LoadInt64(&d1.called) +
		atomic.LoadInt64(&d2.called) +
		atomic.LoadInt64(&d3.called)
	if total != 3 {
		t.Fatalf("expected all 3 dispatchers called, got %d", total)
	}
}
