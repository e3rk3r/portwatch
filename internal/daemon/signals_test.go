package daemon_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/daemon"
)

func TestDaemon_StopsOnContextCancel(t *testing.T) {
	cfg := &config.Config{
		Interval: 1,
		Ports:    []config.PortConfig{},
	}
	d := daemon.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- d.Run(ctx)
	}()

	// Give the goroutine time to start
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("daemon did not stop after context cancel")
	}
}

func TestDaemon_New_NotNil(t *testing.T) {
	cfg := &config.Config{
		Interval: 5,
		Ports:    []config.PortConfig{},
	}
	d := daemon.New(cfg)
	if d == nil {
		t.Fatal("expected non-nil daemon")
	}
}
