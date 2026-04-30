package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "cronwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
default_webhook: "https://hooks.example.com/alert"
check_interval: 1m
jobs:
  - name: backup
    schedule: "0 2 * * *"
    tolerance: 5m
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job name 'backup', got %q", cfg.Jobs[0].Name)
	}
	if cfg.CheckInterval != time.Minute {
		t.Errorf("expected 1m interval, got %v", cfg.CheckInterval)
	}
}

func TestLoad_DefaultCheckInterval(t *testing.T) {
	path := writeTempConfig(t, `
default_webhook: "https://hooks.example.com/alert"
jobs:
  - name: sync
    schedule: "*/5 * * * *"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 30*time.Second {
		t.Errorf("expected default 30s, got %v", cfg.CheckInterval)
	}
}

func TestLoad_MissingWebhook(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: nightly
    schedule: "0 0 * * *"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing webhook, got nil")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeTempConfig(t, `default_webhook: "https://hooks.example.com/alert"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for no jobs, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/cronwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
