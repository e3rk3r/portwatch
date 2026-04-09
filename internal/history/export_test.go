package history

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func sampleEvents() []Event {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	return []Event{
		{Port: 8080, State: StateOpen, Timestamp: base},
		{Port: 9090, State: StateClosed, Timestamp: base.Add(time.Minute)},
	}
}

func TestExportJSON_ValidOutput(t *testing.T) {
	var buf bytes.Buffer
	events := sampleEvents()
	if err := ExportJSON(&buf, events); err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in JSON output, got: %s", out)
	}
	if !strings.Contains(out, "open") {
		t.Errorf("expected state 'open' in JSON output, got: %s", out)
	}
}

func TestExportJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportJSON(&buf, nil); err != nil {
		t.Fatalf("ExportJSON error on nil: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "[]" {
		t.Errorf("expected empty JSON array, got: %s", out)
	}
}

func TestExportCSV_HeaderAlwaysPresent(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportCSV(&buf, nil); err != nil {
		t.Fatalf("ExportCSV error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (header only), got %d", len(lines))
	}
	if lines[0] != "timestamp,port,state" {
		t.Errorf("unexpected header: %s", lines[0])
	}
}

func TestExportCSV_ValidRows(t *testing.T) {
	var buf bytes.Buffer
	events := sampleEvents()
	if err := ExportCSV(&buf, events); err != nil {
		t.Fatalf("ExportCSV error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d", len(lines))
	}
	if !strings.Contains(lines[1], "8080") {
		t.Errorf("row 1 missing port 8080: %s", lines[1])
	}
	if !strings.Contains(lines[2], "closed") {
		t.Errorf("row 2 missing state closed: %s", lines[2])
	}
}

func TestExport_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Export(&buf, nil, ExportFormat("xml"))
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("unexpected error message: %v", err)
	}
}
