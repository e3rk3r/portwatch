package alert

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func quorumNotif() Notification {
	return Notification{Port: 8080, State: "open"}
}

func TestDefaultQuorumPolicy(t *testing.T) {
	p := DefaultQuorumPolicy(5)
	if p.Total != 5 {
		t.Fatalf("expected total 5, got %d", p.Total)
	}
	if p.Required != 3 {
		t.Fatalf("expected required 3, got %d", p.Required)
	}
}

func TestQuorumPolicy_Validate(t *testing.T) {
	tests := []struct {
		policy  QuorumPolicy
		wantErr bool
	}{
		{QuorumPolicy{Total: 3, Required: 2}, false},
		{QuorumPolicy{Total: 0, Required: 1}, true},
		{QuorumPolicy{Total: 3, Required: 0}, true},
		{QuorumPolicy{Total: 3, Required: 4}, true},
	}
	for _, tt := range tests {
		err := tt.policy.Validate()
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate(%+v) error=%v, wantErr=%v", tt.policy, err, tt.wantErr)
		}
	}
}

func TestQuorumDispatcher_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty targets")
		}
	}()
	NewQuorumDispatcher(DefaultQuorumPolicy(0), nil)
}

func TestQuorumDispatcher_AllSucceed(t *testing.T) {
	targets := []Dispatcher{
		&captureDispatcher{},
		&captureDispatcher{},
		&captureDispatcher{},
	}
	q := NewQuorumDispatcher(QuorumPolicy{Total: 3, Required: 2}, targets)
	if err := q.Send(context.Background(), quorumNotif()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestQuorumDispatcher_QuorumMet(t *testing.T) {
	targets := []Dispatcher{
		&captureDispatcher{},
		&captureDispatcher{err: errors.New("fail")},
		&captureDispatcher{},
	}
	q := NewQuorumDispatcher(QuorumPolicy{Total: 3, Required: 2}, targets)
	if err := q.Send(context.Background(), quorumNotif()); err != nil {
		t.Fatalf("expected quorum met, got %v", err)
	}
}

func TestQuorumDispatcher_QuorumNotMet(t *testing.T) {
	targets := []Dispatcher{
		&captureDispatcher{err: errors.New("fail")},
		&captureDispatcher{err: errors.New("fail")},
		&captureDispatcher{},
	}
	q := NewQuorumDispatcher(QuorumPolicy{Total: 3, Required: 2}, targets)
	if err := q.Send(context.Background(), quorumNotif()); err == nil {
		t.Fatal("expected quorum failure")
	}
}

func TestQuorumDispatcher_AllFail(t *testing.T) {
	targets := []Dispatcher{
		&captureDispatcher{err: fmt.Errorf("e1")},
		&captureDispatcher{err: fmt.Errorf("e2")},
	}
	q := NewQuorumDispatcher(QuorumPolicy{Total: 2, Required: 1}, targets)
	if err := q.Send(context.Background(), quorumNotif()); err == nil {
		t.Fatal("expected error when all fail")
	}
}
