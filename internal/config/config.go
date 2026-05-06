package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// JobConfig holds configuration for a single monitored cron job.
type JobConfig struct {
	Name            string        `yaml:"name"`
	Schedule        string        `yaml:"schedule"`
	WebhookURL      string        `yaml:"webhook_url"`
	DriftThreshold  time.Duration `yaml:"drift_threshold"`
	AlertCooldown   time.Duration `yaml:"alert_cooldown"`
}

// Config is the top-level configuration structure for cronwatch.
type Config struct {
	WebhookURL      string        `yaml:"webhook_url"`
	CheckInterval   time.Duration `yaml:"check_interval"`
	AlertCooldown   time.Duration `yaml:"alert_cooldown"`
	Jobs            []JobConfig   `yaml:"jobs"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}
	if len(cfg.Jobs) == 0 {
		return fmt.Errorf("at least one job must be configured")
	}
	for _, j := range cfg.Jobs {
		if j.Name == "" {
			return fmt.Errorf("each job must have a name")
		}
		if j.Schedule == "" {
			return fmt.Errorf("job %q must have a schedule", j.Name)
		}
	}
	return nil
}

func applyDefaults(cfg *Config) {
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 30 * time.Second
	}
	if cfg.AlertCooldown == 0 {
		cfg.AlertCooldown = 15 * time.Minute
	}
	for i := range cfg.Jobs {
		if cfg.Jobs[i].DriftThreshold == 0 {
			cfg.Jobs[i].DriftThreshold = 2 * time.Minute
		}
		if cfg.Jobs[i].AlertCooldown == 0 {
			cfg.Jobs[i].AlertCooldown = cfg.AlertCooldown
		}
	}
}
