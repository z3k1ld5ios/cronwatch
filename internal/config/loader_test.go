package config

import (
	"os"
	"testing"
)

func TestEnvOrDefault_UsesEnv(t *testing.T) {
	t.Setenv("TEST_KEY", "from-env")
	if got := EnvOrDefault("TEST_KEY", "fallback"); got != "from-env" {
		t.Errorf("expected 'from-env', got %q", got)
	}
}

func TestEnvOrDefault_UsesFallback(t *testing.T) {
	os.Unsetenv("TEST_KEY_MISSING")
	if got := EnvOrDefault("TEST_KEY_MISSING", "fallback"); got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}
}

func TestLoadFromEnv_UsesEnvVar(t *testing.T) {
	path := writeTempConfig(t, `
default_webhook: "https://hooks.example.com/x"
jobs:
  - name: envjob
    schedule: "* * * * *"
`)
	t.Setenv("CRONWATCH_CONFIG", path)
	cfg, err := LoadFromEnv("/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Jobs[0].Name != "envjob" {
		t.Errorf("expected 'envjob', got %q", cfg.Jobs[0].Name)
	}
}

func TestEffectiveWebhook_JobOverridesDefault(t *testing.T) {
	cfg := &Config{DefaultWebhook: "https://default.example.com"}
	job := Job{Name: "j", Webhook: "https://job-specific.example.com"}
	got := EffectiveWebhook(cfg, job)
	if got != "https://job-specific.example.com" {
		t.Errorf("expected job webhook, got %q", got)
	}
}

func TestEffectiveWebhook_FallsBackToDefault(t *testing.T) {
	cfg := &Config{DefaultWebhook: "https://default.example.com"}
	job := Job{Name: "j"}
	got := EffectiveWebhook(cfg, job)
	if got != "https://default.example.com" {
		t.Errorf("expected default webhook, got %q", got)
	}
}
