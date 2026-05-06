package monitor

import (
	"sync"
	"time"
)

// Suppression tracks recently fired alerts to avoid duplicate notifications
// within a configurable cooldown window.
type Suppression struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastFired map[string]time.Time
}

// NewSuppression creates a Suppression with the given cooldown duration.
// Alerts for the same job will be suppressed if fired within the cooldown window.
func NewSuppression(cooldown time.Duration) *Suppression {
	return &Suppression{
		cooldown:  cooldown,
		lastFired: make(map[string]time.Time),
	}
}

// Allow returns true if the alert for jobName should be sent (not suppressed).
// It records the current time as the last fired time when returning true.
func (s *Suppression) Allow(jobName string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if last, ok := s.lastFired[jobName]; ok {
		if now.Sub(last) < s.cooldown {
			return false
		}
	}
	s.lastFired[jobName] = now
	return true
}

// Reset clears the suppression record for a specific job.
// Useful when a job recovers and the next alert should not be suppressed.
func (s *Suppression) Reset(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.lastFired, jobName)
}

// Snapshot returns a copy of all last-fired times, keyed by job name.
func (s *Suppression) Snapshot() map[string]time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]time.Time, len(s.lastFired))
	for k, v := range s.lastFired {
		out[k] = v
	}
	return out
}
