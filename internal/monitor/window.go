package monitor

import (
	"sync"
	"time"
)

// WindowConfig defines the parameters for a maintenance window.
type WindowConfig struct {
	Name      string
	Start     time.Time
	End       time.Time
	JobNames  []string // empty means all jobs
}

// WindowManager tracks active maintenance windows during which alerts are suppressed.
type WindowManager struct {
	mu      sync.RWMutex
	windows []WindowConfig
	now     func() time.Time
}

// NewWindowManager creates a WindowManager with an optional clock override.
func NewWindowManager(now func() time.Time) *WindowManager {
	if now == nil {
		now = time.Now
	}
	return &WindowManager{now: now}
}

// Add registers a new maintenance window.
func (w *WindowManager) Add(wc WindowConfig) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.windows = append(w.windows, wc)
}

// Remove deletes a maintenance window by name.
func (w *WindowManager) Remove(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	updated := w.windows[:0]
	for _, wc := range w.windows {
		if wc.Name != name {
			updated = append(updated, wc)
		}
	}
	w.windows = updated
}

// IsSuppressed returns true if the given job is currently within a maintenance window.
func (w *WindowManager) IsSuppressed(jobName string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	now := w.now()
	for _, wc := range w.windows {
		if now.Before(wc.Start) || now.After(wc.End) {
			continue
		}
		if len(wc.JobNames) == 0 {
			return true
		}
		for _, jn := range wc.JobNames {
			if jn == jobName {
				return true
			}
		}
	}
	return false
}

// ActiveWindows returns a snapshot of all currently active windows.
func (w *WindowManager) ActiveWindows() []WindowConfig {
	w.mu.RLock()
	defer w.mu.RUnlock()
	now := w.now()
	var active []WindowConfig
	for _, wc := range w.windows {
		if !now.Before(wc.Start) && !now.After(wc.End) {
			active = append(active, wc)
		}
	}
	return active
}
