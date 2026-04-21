package alert

import (
	"context"
	"testing"
)

func muxNotif(port int, state string) Notification {
	return Notification{Port: port, State: state, Title: "test"}
}

func TestMux_DefaultPolicy(t *testing.T) {
	p := DefaultMuxPolicy()
	if p.MaxKeys <= 0 {
		t.Fatalf("expected positive MaxKeys, got %d", p.MaxKeys)
	}
}

func TestNewMux_NilPolicyUsesDefault(t *testing.T) {
	m := NewMux(nil)
	if m == nil {
		t.Fatal("expected non-nil Mux")
	}
}
