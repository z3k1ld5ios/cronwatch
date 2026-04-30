package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job defines a single monitored cron job.
type Job struct {
	Name     string        `yaml:"name"`
	Schedule string        `yaml:"schedule"`
	Tolerance time.Duration `yaml:"tolerance"`
	Webhook  string        `yaml:"webhook"`
}

// Config holds the full cronwatch configuration.
type Config struct {
	Jobs           []Job         `yaml:"jobs"`
	DefaultWebhook string        `yaml:"default_webhook"`
	CheckInterval  time.Duration `yaml:"check_interval"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 30 * time.Second
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("no jobs defined")
	}
	for i, j := range c.Jobs {
		if j.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if j.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", j.Name)
		}
		wh := j.Webhook
		if wh == "" {
			wh = c.DefaultWebhook
		}
		if wh == "" {
			return fmt.Errorf("job %q: webhook is required (no default set)", j.Name)
		}
	}
	return nil
}
