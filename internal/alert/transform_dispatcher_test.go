package alert

import (
	"errors"
	"testing"
)

// captureDispatcher records the last notification it received.
type captureDispatcher struct {
	last Notification
	err  error
}

func (c *captureDispatcher) Send(n Notification) error {
	c.last = n
	return c.err
}

func makeTransformDispatcher(policy TransformPolicy) (*captureDispatcher, *TransformDispatcher) {
	cap := &captureDispatcher{}
	td := NewTransformDispatcher(cap, NewTransformer(policy))
	return cap, td
}

func TestTransformDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewTransformDispatcher(nil, NewTransformer(DefaultTransformPolicy()))
}

func TestTransformDispatcher_NilTransformerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil transformer")
		}
	}()
	NewTransformDispatcher(&captureDispatcher{}, nil)
}

func TestTransformDispatcher_NoOpPassesThrough(t *testing.T) {
	cap, td := makeTransformDispatcher(DefaultTransformPolicy())
	n := baseNotification()
	if err := td.Send(n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.last.Port != n.Port {
		t.Errorf("port mismatch: got %d want %d", cap.last.Port, n.Port)
	}
}

func TestTransformDispatcher_TitleApplied(t *testing.T) {
	policy := DefaultTransformPolicy()
	policy.TitleTemplate = "port {{.Port}} is {{.State}}"
	cap, td := makeTransformDispatcher(policy)
	n := baseNotification()
	if err := td.Send(n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "port 8080 is open"
	if cap.last.Title != want {
		t.Errorf("title: got %q want %q", cap.last.Title, want)
	}
}

func TestTransformDispatcher_DownstreamErrorPropagated(t *testing.T) {
	cap, td := makeTransformDispatcher(DefaultTransformPolicy())
	sentinel := errors.New("downstream failure")
	cap.err = sentinel
	if err := td.Send(baseNotification()); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestTransformDispatcher_LabelsAdded(t *testing.T) {
	policy := DefaultTransformPolicy()
	policy.AddLabels = map[string]string{"env": "prod"}
	cap, td := makeTransformDispatcher(policy)
	if err := td.Send(baseNotification()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.last.Labels["env"] != "prod" {
		t.Errorf("label not applied: %v", cap.last.Labels)
	}
}
