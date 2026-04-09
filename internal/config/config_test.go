package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "portwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	yaml := `
interval: 5s
ports:
  - host: localhost
    port: 8080
    on_open:
      - type: webhook
        url: http://example.com/hook
    on_close:
      - type: script
        command: echo closed
`
	path := writeTempConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 5*time.Second {
		t.Errorf("expected interval 5s, got %v", cfg.Interval)
	}
	if len(cfg.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(cfg.Ports))
	}
	if cfg.Ports[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Ports[0].Port)
	}
}

func TestLoad_DefaultInterval(t *testing.T) {
	yaml := `ports:
  - port: 9090
`
	path := writeTempConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 10*time.Second {
		t.Errorf("expected default interval 10s, got %v", cfg.Interval)
	}
	if cfg.Ports[0].Host != "localhost" {
		t.Errorf("expected default host localhost, got %q", cfg.Ports[0].Host)
	}
}

func TestLoad_InvalidPortRange(t *testing.T) {
	yaml := `ports:
  - port: 99999
`
	path := writeTempConfig(t, yaml)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for out-of-range port, got nil")
	}
}

func TestLoad_UnknownActionType(t *testing.T) {
	yaml := `ports:
  - port: 3000
    on_open:
      - type: magic
`
	path := writeTempConfig(t, yaml)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown action type, got nil")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/portwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
