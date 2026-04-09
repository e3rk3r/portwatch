package api_test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestMethodNotAllowed_Status(t *testing.T) {
	addr, cancel := startServer(t)
	defer cancel()

	resp, err := http.Post(fmt.Sprintf("http://%s/status", addr), "application/json", nil)
	if err != nil {
		t.Fatalf("POST /status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestMethodNotAllowed_History(t *testing.T) {
	addr, cancel := startServer(t)
	defer cancel()

	resp, err := http.Post(fmt.Sprintf("http://%s/history", addr), "application/json", nil)
	if err != nil {
		t.Fatalf("POST /history: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHistory_InvalidSince(t *testing.T) {
	addr, cancel := startServer(t)
	defer cancel()

	resp, err := http.Get(fmt.Sprintf("http://%s/history?since=not-a-time", addr))
	if err != nil {
		t.Fatalf("GET /history: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHistory_ValidPortFilter(t *testing.T) {
	addr, cancel := startServer(t)
	defer cancel()

	resp, err := http.Get(fmt.Sprintf("http://%s/history?port=8080", addr))
	if err != nil {
		t.Fatalf("GET /history?port=8080: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
