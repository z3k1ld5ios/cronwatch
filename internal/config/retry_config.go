package config

import (
	"fmt"
	"time"
)

// RetryConfig holds configuration for alert delivery retries.
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Backoff     time.Duration `yaml:"backoff"`
}

// DefaultRetryConfig returns sensible retry defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Backoff:     30 * time.Second,
	}
}

// validateRetry checks that the RetryConfig values are within acceptable bounds.
func validateRetry(rc RetryConfig) error {
	if rc.MaxAttempts < 0 {
		return fmt.Errorf("retry max_attempts must be >= 0, got %d", rc.MaxAttempts)
	}
	if rc.MaxAttempts > 10 {
		return fmt.Errorf("retry max_attempts must be <= 10, got %d", rc.MaxAttempts)
	}
	if rc.Backoff < 0 {
		return fmt.Errorf("retry backoff must be >= 0, got %v", rc.Backoff)
	}
	if rc.Backoff > 10*time.Minute {
		return fmt.Errorf("retry backoff must be <= 10m, got %v", rc.Backoff)
	}
	return nil
}

// ApplyRetryDefaults fills in zero values with defaults.
func ApplyRetryDefaults(rc RetryConfig) RetryConfig {
	defaults := DefaultRetryConfig()
	if rc.MaxAttempts == 0 {
		rc.MaxAttempts = defaults.MaxAttempts
	}
	if rc.Backoff == 0 {
		rc.Backoff = defaults.Backoff
	}
	return rc
}
