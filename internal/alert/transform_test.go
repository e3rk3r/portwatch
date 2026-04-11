package alert

import (
	"errors"
	"testing"
	"time"
)

// captureDispatcher records the last Notification it received.
type captureDispatcher struct {
	last Notification
	err  error
}

func (c *captureDispatcher) Send(n Notification) error {
	c.last = n
	return c.err
}

func baseNotification() Notification {
	return Notification{
		Port:      8080,
		Host:      "localhost",
		State:     "open",
		Title:     "original title",
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Labels:    map[string]string{"env": "prod"},
	}
}

func TestTransformer_NoOpPolicy(t *testing.T) {
	cap := &captureDispatcher{}
	tr := NewTransformer(DefaultTransformPolicy(), cap)
	n := baseNotification()
	if err := tr.Send(n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.last.Title != "original title" {
		t.Errorf("expected title unchanged, got %q", cap.last.Title)
	}
}

func TestTransformer_TitleTemplate(t *testing.T) {
	cap := &captureDispatcher{}
	policy := TransformPolicy{TitleTemplate: "port {port} is {state} on {host}"}
	tr := NewTransformer(policy, cap)
	if err := tr.Send(baseNotification()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "port 8080 is open on localhost"
	if cap.last.Title != want {
		t.Errorf("title: got %q, want %q", cap.last.Title, want)
	}
}

func TestTransformer_AddLabels(t *testing.T) {
	cap := &captureDispatcher{}
	policy := TransformPolicy{AddLabels: map[string]string{"region": "us-east-1", "env": "override"}}
	tr := NewTransformer(policy, cap)
	if err := tr.Send(baseNotification()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.last.Labels["region"] != "us-east-1" {
		t.Errorf("expected region label, got %v", cap.last.Labels)
	}
	if cap.last.Labels["env"] != "override" {
		t.Errorf("expected env label overridden, got %v", cap.last.Labels["env"])
	}
}

func TestTransformer_StripSensitiveHeaders(t *testing.T) {
	cap := &captureDispatcher{}
	policy := TransformPolicy{StripSensitiveHeaders: true}
	tr := NewTransformer(policy, cap)
	n := baseNotification()
	n.Labels["Authorization"] = "Bearer secret"
	n.Labels["Cookie"] = "session=abc"
	n.Labels["X-Custom"] = "keep-me"
	if err := tr.Send(n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cap.last.Labels["Authorization"]; ok {
		t.Error("Authorization header should have been stripped")
	}
	if _, ok := cap.last.Labels["Cookie"]; ok {
		t.Error("Cookie header should have been stripped")
	}
	if cap.last.Labels["X-Custom"] != "keep-me" {
		t.Error("X-Custom label should be preserved")
	}
}

func TestTransformer_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	cap := &captureDispatcher{err: sentinel}
	tr := NewTransformer(DefaultTransformPolicy(), cap)
	if err := tr.Send(baseNotification()); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestNewTransformer_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil next dispatcher")
		}
	}()
	NewTransformer(DefaultTransformPolicy(), nil)
}
