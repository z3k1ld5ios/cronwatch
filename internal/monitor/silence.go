package monitor

import (
	"sync"
	"time"
)

// Silence represents a named period during which alerts for specific jobs are muted.
type Silence struct {
	Label    string
	JobNames []string // empty means all jobs
	Start    time.Time
	End      time.Time
	Reason   string
}

// SilenceManager tracks active silences and determines whether a job alert should be muted.
type SilenceManager struct {
	mu      sync.RWMutex
	silences map[string]Silence
	clock   func() time.Time
}

// NewSilenceManager creates a SilenceManager with an optional clock override.
func NewSilenceManager(clock func() time.Time) *SilenceManager {
	if clock == nil {
		clock = time.Now
	}
	return &SilenceManager{
		silences: make(map[string]Silence),
		clock:    clock,
	}
}

// Add registers a new silence. Returns false if a silence with that label already exists.
func (sm *SilenceManager) Add(s Silence) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exists := sm.silences[s.Label]; exists {
		return false
	}
	sm.silences[s.Label] = s
	return true
}

// Remove deletes a silence by label.
func (sm *SilenceManager) Remove(label string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.silences, label)
}

// IsSilenced returns true if the given job currently has an active silence.
func (sm *SilenceManager) IsSilenced(jobName string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	now := sm.clock()
	for _, s := range sm.silences {
		if now.Before(s.Start) || now.After(s.End) {
			continue
		}
		if len(s.JobNames) == 0 {
			return true
		}
		for _, jn := range s.JobNames {
			if jn == jobName {
				return true
			}
		}
	}
	return false
}

// List returns a snapshot of all registered silences.
func (sm *SilenceManager) List() []Silence {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	out := make([]Silence, 0, len(sm.silences))
	for _, s := range sm.silences {
		out = append(out, s)
	}
	return out
}
