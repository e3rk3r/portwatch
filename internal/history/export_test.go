package history

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

var exportTime = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func sampleEvents() []Event {
	return []Event{
		{Port: 8080, State: "open", Timestamp: exportTime},
		{Port: 443, State: "closed", Timestamp: exportTime.Add(time.Minute)},
	}
}

func TestExportJSON_ValidOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportJSON(&buf, sampleEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out []Event
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 events, got %d", len(out))
	}
	if out[0].Port != 8080 || out[0].State != "open" {
		t.Errorf("unexpected first event: %+v", out[0])
	}
}

func TestExportJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportJSON(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "null") && !strings.Contains(buf.String(), "[]") {
		t.Errorf("expected empty JSON array or null, got: %s", buf.String())
	}
}

func TestExportCSV_HeaderAlwaysPresent(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportCSV(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "timestamp,port,state") {
		t.Errorf("CSV header missing, got: %s", buf.String())
	}
}

func TestExportCSV_ValidRows(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportCSV(&buf, sampleEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d", len(lines))
	}
	if !strings.Contains(lines[1], "8080") || !strings.Contains(lines[1], "open") {
		t.Errorf("unexpected row: %s", lines[1])
	}
}

func TestExport_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Export(&buf, sampleEvents(), Format("xml"))
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("unexpected error message: %v", err)
	}
}
