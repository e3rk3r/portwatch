package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/history"
	"github.com/user/portwatch/internal/monitor"
)

func TestExport_DefaultJSON(t *testing.T) {
	ring := history.NewRing(0)
	ring.Record(history.Event{Port: 8080, State: monitor.StateOpen})

	_, addr := startServer(t, nil, ring)
	resp, err := http.Get("http://" + addr + "/export")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("expected JSON content-type, got %q", ct)
	}
	var rows []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

func TestExport_CSV(t *testing.T) {
	ring := history.NewRing(0)
	ring.Record(history.Event{Port: 9090, State: monitor.StateClosed})

	_, addr := startServer(t, nil, ring)
	resp, err := http.Get("http://" + addr + "/export?format=csv")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.HasPrefix(string(body), "port,state,timestamp") {
		t.Fatalf("expected CSV header, got: %s", body)
	}
}

func TestExport_InvalidPort(t *testing.T) {
	ring := history.NewRing(0)
	_, addr := startServer(t, nil, ring)
	resp, err := http.Get("http://" + addr + "/export?port=notanumber")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestExport_MethodNotAllowed(t *testing.T) {
	ring := history.NewRing(0)
	_, addr := startServer(t, nil, ring)
	resp, err := http.Post("http://"+addr+"/export", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
