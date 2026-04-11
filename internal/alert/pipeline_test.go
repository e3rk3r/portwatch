package alert_test

import (
	"context"
	"errors"
	"testing"

	"github.com/user/portwatch/internal/alert"
)

// recordDispatcher records every notification it receives.
type recordDispatcher struct {
	calls []alert.Notification
	err   error
}

func (r *recordDispatcher) Send(_ context.Context, n alert.Notification) error {
	r.calls = append(r.calls, n)
	return r.err
}

func TestPipeline_NilFirstDispatcher(t *testing.T) {
	_, err := alert.NewPipeline(nil)
	if err == nil {
		t.Fatal("expected error for nil first dispatcher")
	}
}

func TestPipeline_NilMiddleDispatcher(t *testing.T) {
	a := &recordDispatcher{}
	_, err := alert.NewPipeline(a, nil)
	if err == nil {
		t.Fatal("expected error for nil middle dispatcher")
	}
}

func TestPipeline_AllStagesCalled(t *testing.T) {
	a, b, c := &recordDispatcher{}, &recordDispatcher{}, &recordDispatcher{}
	p, err := alert.NewPipeline(a, b, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	n := alert.Notification{Port: 8080, State: "open"}
	if err := p.Send(context.Background(), n); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	for i, d := range []*recordDispatcher{a, b, c} {
		if len(d.calls) != 1 {
			t.Errorf("stage %d: expected 1 call, got %d", i, len(d.calls))
		}
	}
}

func TestPipeline_AbortsOnError(t *testing.T) {
	sentinel := errors.New("stage error")
	a := &recordDispatcher{}
	b := &recordDispatcher{err: sentinel}
	c := &recordDispatcher{}

	p, _ := alert.NewPipeline(a, b, c)
	n := alert.Notification{Port: 9090, State: "closed"}

	err := p.Send(context.Background(), n)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if len(c.calls) != 0 {
		t.Errorf("stage after error should not be called")
	}
}

func TestPipeline_Len(t *testing.T) {
	a, b := &recordDispatcher{}, &recordDispatcher{}
	p, _ := alert.NewPipeline(a, b)
	if p.Len() != 2 {
		t.Errorf("expected Len 2, got %d", p.Len())
	}
}
