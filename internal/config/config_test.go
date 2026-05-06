package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "cronwatch.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempConfig: %v", err)
	}
	return p
}

func TestLoad_ValidConfig(t *testing.T) {
	p := writeTempConfig(t, `
webhook_url: https://hooks.example.com/alert
check_interval: 1m
jobs:
  - name: backup
    schedule: "0 2 * * *"
`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WebhookURL != "https://hooks.example.com/alert" {
		t.Errorf("unexpected webhook_url: %s", cfg.WebhookURL)
	}
	if cfg.CheckInterval != time.Minute {
		t.Errorf("unexpected check_interval: %v", cfg.CheckInterval)
	}
}

func TestLoad_DefaultCheckInterval(t *testing.T) {
	p := writeTempConfig(t, `
webhook_url: https://hooks.example.com/alert
jobs:
  - name: sync
    schedule: "*/5 * * * *"
`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 30*time.Second {
		t.Errorf("expected default check_interval 30s, got %v", cfg.CheckInterval)
	}
}

func TestLoad_MissingWebhook(t *testing.T) {
	p := writeTempConfig(t, `
jobs:
  - name: sync
    schedule: "*/5 * * * *"
`)
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for missing webhook_url")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	p := writeTempConfig(t, `
webhook_url: https://hooks.example.com/alert
`)
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for missing jobs")
	}
}

func TestLoad_DefaultDriftAndCooldownPerJob(t *testing.T) {
	p := writeTempConfig(t, `
webhook_url: https://hooks.example.com/alert
alert_cooldown: 10m
jobs:
  - name: report
    schedule: "0 6 * * 1"
`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Jobs[0].DriftThreshold != 2*time.Minute {
		t.Errorf("expected default drift_threshold 2m, got %v", cfg.Jobs[0].DriftThreshold)
	}
	if cfg.Jobs[0].AlertCooldown != 10*time.Minute {
		t.Errorf("expected job cooldown inherited from global 10m, got %v", cfg.Jobs[0].AlertCooldown)
	}
}
