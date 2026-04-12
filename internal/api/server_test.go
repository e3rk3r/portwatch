package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/api"
	"github.com/yourorg/portwatch/internal/history"
	"github.com/yourorg/portwatch/internal/monitor"
)

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freeAddr: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func startServer(t *testing.T) (string, context.CancelFunc) {
	t.Helper()
	addr := freeAddr(t)
	ring := history.NewRing(0)
	tracker := monitor.NewTracker()
	srv := api.New(addr, ring, tracker)

	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = srv.Start(ctx) }()
	time.Sleep(30 * time.Millisecond)
	return addr, cancel
}

// getJSON is a helper that performs a GET request and decodes the JSON response
// body into dst. It fails the test if the request fails or decoding fails.
func getJSON(t *testing.T, url string, dst any) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		resp.Body.Close()
		t.Fatalf("decode response from %s: %v", url, err)
	}
	return resp
}

func TestHealthz(t *testing.T) {
	addr, cancel := startServer(t)
	defer cancel()

	resp, err := http.Get(fmt.Sprintf("http://%s/healthz", addr))
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestStatus_EmptyTracker(t *testing.T) {
	addr, cancel := startServer(t)
	defer cancel()

	resp, err := http.Get(fmt.Sprintf("http://%s/status", addr))
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHistory_InvalidPort(t *testing.T) {
	addr, cancel := startServer(t)
	defer cancel()

	resp, err := http.Get(fmt.Sprintf("http://%s/history?port=notanumber", addr))
	if err != nil {
		t.Fatalf("GET /history: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHistory_ReturnsJSON(t *testing.T) {
	addr, cancel := startServer(t)
	defer cancel()

	var out []any
	resp := getJSON(t, fmt.Sprintf("http://%s/history", addr), &out)
	defer resp.Body.Close()
}
