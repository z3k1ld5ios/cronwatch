package config

import (
	"fmt"
	"time"

	"github.com/example/cronwatch/internal/monitor"
)

// EscalationConfig holds YAML-decoded escalation settings.
type EscalationConfig struct {
	WarningAfter  int    `yaml:"warning_after"`
	CriticalAfter int    `yaml:"critical_after"`
	ResetAfter    string `yaml:"reset_after"`
}

// DefaultEscalationConfig returns sensible defaults.
func DefaultEscalationConfig() EscalationConfig {
	return EscalationConfig{
		WarningAfter:  3,
		CriticalAfter: 5,
		ResetAfter:    "30m",
	}
}

// ApplyEscalationDefaults fills zero values with defaults.
func ApplyEscalationDefaults(c *EscalationConfig) {
	def := DefaultEscalationConfig()
	if c.WarningAfter <= 0 {
		c.WarningAfter = def.WarningAfter
	}
	if c.CriticalAfter <= 0 {
		c.CriticalAfter = def.CriticalAfter
	}
	if c.ResetAfter == "" {
		c.ResetAfter = def.ResetAfter
	}
}

// validateEscalation returns an error if the config is logically invalid.
func validateEscalation(c EscalationConfig) error {
	if c.WarningAfter >= c.CriticalAfter {
		return fmt.Errorf("escalation: warning_after (%d) must be less than critical_after (%d)",
			c.WarningAfter, c.CriticalAfter)
	}
	if _, err := time.ParseDuration(c.ResetAfter); err != nil {
		return fmt.Errorf("escalation: invalid reset_after %q: %w", c.ResetAfter, err)
	}
	return nil
}

// ToPolicy converts an EscalationConfig into a monitor.EscalationPolicy.
func (c EscalationConfig) ToPolicy() (monitor.EscalationPolicy, error) {
	d, err := time.ParseDuration(c.ResetAfter)
	if err != nil {
		return monitor.EscalationPolicy{}, fmt.Errorf("escalation: invalid reset_after: %w", err)
	}
	return monitor.EscalationPolicy{
		WarningAfter:  c.WarningAfter,
		CriticalAfter: c.CriticalAfter,
		ResetAfter:    d,
	}, nil
}
