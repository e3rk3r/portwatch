package alert

import (
	"context"
	"errors"
	"testing"
)

func labelNotif() Notification {
	return Notification{Port: 8080, State: "open", Labels: map[string]string{"env": "prod"}}
}

func TestLabeler_DefaultPolicy(t *testing.T) {
	l := NewLabeler(DefaultLabelPolicy())
	if len(l.Snapshot()) != 0 {
		t.Fatal("expected empty label map")
	}
}

func TestLabeler_ApplyMergesLabels(t *testing.T) {
	l := NewLabeler(LabelPolicy{Labels: map[string]string{"region": "us-east", "env": "staging"}})
	n := labelNotif() // already has env=prod
	l.Apply(&n)
	if n.Labels["region"] != "us-east" {
		t.Errorf("expected region=us-east, got %q", n.Labels["region"])
	}
	// existing key must NOT be overwritten
	if n.Labels["env"] != "prod" {
		t.Errorf("expected env=prod (not overwritten), got %q", n.Labels["env"])
	}
}

func TestLabeler_Prefix(t *testing.T) {
	l := NewLabeler(LabelPolicy{Labels: map[string]string{"team": "ops"}, Prefix: "pw_"})
	n := Notification{Labels: map[string]string{}}
	l.Apply(&n)
	if n.Labels["pw_team"] != "ops" {
		t.Errorf("expected pw_team=ops, got %q", n.Labels["pw_team"])
	}
}

func TestLabeler_Set(t *testing.T) {
	l := NewLabeler(DefaultLabelPolicy())
	l.Set("dc", "lon1")
	snap := l.Snapshot()
	if snap["dc"] != "lon1" {
		t.Errorf("expected dc=lon1, got %q", snap["dc"])
	}
}

func TestLabelDispatcher_NilLabelerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil labeler")
		}
	}()
	NewLabelDispatcher(nil, &captureDispatcher{})
}

func TestLabelDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewLabelDispatcher(NewLabeler(DefaultLabelPolicy()), nil)
}

func TestLabelDispatcher_ForwardsWithLabels(t *testing.T) {
	cap := &captureDispatcher{}
	l := NewLabeler(LabelPolicy{Labels: map[string]string{"source": "portwatch"}})
	d := NewLabelDispatcher(l, cap)

	n := labelNotif()
	if err := d.Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.last.Labels["source"] != "portwatch" {
		t.Errorf("expected source=portwatch, got %q", cap.last.Labels["source"])
	}
}

func TestLabelDispatcher_PropagatesError(t *testing.T) {
	errDispatcher := &errorDispatcher{err: errors.New("downstream fail")}
	l := NewLabeler(DefaultLabelPolicy())
	d := NewLabelDispatcher(l, errDispatcher)
	if err := d.Send(context.Background(), labelNotif()); err == nil {
		t.Fatal("expected error from downstream")
	}
}

func TestLabelDispatcher_DoesNotMutateOriginal(t *testing.T) {
	cap := &captureDispatcher{}
	l := NewLabeler(LabelPolicy{Labels: map[string]string{"injected": "yes"}})
	d := NewLabelDispatcher(l, cap)

	orig := labelNotif()
	_ = d.Send(context.Background(), orig)

	if _, ok := orig.Labels["injected"]; ok {
		t.Error("original notification should not be mutated")
	}
}
