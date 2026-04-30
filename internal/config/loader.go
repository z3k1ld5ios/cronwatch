package config

import (
	"fmt"
	"os"
)

// EnvOrDefault returns the value of the environment variable key,
// or fallback if the variable is unset or empty.
func EnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// LoadFromEnv loads configuration from the path specified by the
// CRONWATCH_CONFIG environment variable, falling back to defaultPath.
func LoadFromEnv(defaultPath string) (*Config, error) {
	path := EnvOrDefault("CRONWATCH_CONFIG", defaultPath)
	cfg, err := Load(path)
	if err != nil {
		return nil, fmt.Errorf("LoadFromEnv(%q): %w", path, err)
	}
	return cfg, nil
}

// EffectiveWebhook returns the job-level webhook if set,
// otherwise falls back to the config-level default.
func EffectiveWebhook(cfg *Config, job Job) string {
	if job.Webhook != "" {
		return job.Webhook
	}
	return cfg.DefaultWebhook
}
