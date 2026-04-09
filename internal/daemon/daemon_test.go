package daemon_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/daemon"
)

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func makeCfg(port int) *config.Config {
	return &config.Config{
		Interval: 1,
		Ports: []config.PortConfig{
			{
				Port: port,
				Host: "127.0.0.1",
				Actions: []config.Action{},
			},
		},
	}
}

func TestDaemon_RunCancels(t *testing.T) {
	port := freePort(t)
	cfg := makeCfg(port)
	d := daemon.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := d.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestDaemon_TickDetectsOpen(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port

	_ = strconv.Itoa(port) // ensure port is used
	cfg := makeCfg(port)
	d := daemon.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Should complete without panic
	d.Run(ctx) //nolint:errcheck
}
