package monitor

import (
	"sync"
	"time"
)

// SLOPolicy defines the target availability for a job.
type SLOPolicy struct {
	// TargetPercent is the desired success rate, e.g. 99.9
	TargetPercent float64
	// Window is the rolling time window to evaluate
	Window time.Duration
}

// DefaultSLOPolicy returns a sensible default: 99% over 7 days.
func DefaultSLOPolicy() SLOPolicy {
	return SLOPolicy{
		TargetPercent: 99.0,
		Window:        7 * 24 * time.Hour,
	}
}

// sloEvent records a single execution outcome.
type sloEvent struct {
	at      time.Time
	success bool
}

// SLOStatus summarises current SLO standing for a job.
type SLOStatus struct {
	Job           string  `json:"job"`
	SuccessRate   float64 `json:"success_rate"`
	Target        float64 `json:"target"`
	Breaching     bool    `json:"breaching"`
	TotalEvents   int     `json:"total_events"`
	FailedEvents  int     `json:"failed_events"`
}

// SLOManager tracks per-job SLO compliance over a rolling window.
type SLOManager struct {
	mu     sync.Mutex
	policy SLOPolicy
	events map[string][]sloEvent
	clock  func() time.Time
}

// NewSLOManager creates a new SLOManager with the given policy.
func NewSLOManager(policy SLOPolicy) *SLOManager {
	return &SLOManager{
		policy: policy,
		events: make(map[string][]sloEvent),
		clock:  time.Now,
	}
}

// Record adds an execution outcome for the given job.
func (m *SLOManager) Record(job string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.clock()
	m.events[job] = append(m.prune(m.events[job], now), sloEvent{at: now, success: success})
}

// Status returns the current SLO status for a job.
func (m *SLOManager) Status(job string) SLOStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.clock()
	evts := m.prune(m.events[job], now)
	m.events[job] = evts

	total := len(evts)
	failed := 0
	for _, e := range evts {
		if !e.success {
			failed++
		}
	}
	var rate float64
	if total > 0 {
		rate = float64(total-failed) / float64(total) * 100.0
	} else {
		rate = 100.0
	}
	return SLOStatus{
		Job:          job,
		SuccessRate:  rate,
		Target:       m.policy.TargetPercent,
		Breaching:    total > 0 && rate < m.policy.TargetPercent,
		TotalEvents:  total,
		FailedEvents: failed,
	}
}

// AllStatuses returns SLO status for every tracked job.
func (m *SLOManager) AllStatuses() []SLOStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.clock()
	out := make([]SLOStatus, 0, len(m.events))
	for job, evts := range m.events {
		evts = m.prune(evts, now)
		m.events[job] = evts
		total := len(evts)
		failed := 0
		for _, e := range evts {
			if !e.success {
				failed++
			}
		}
		var rate float64
		if total > 0 {
			rate = float64(total-failed) / float64(total) * 100.0
		} else {
			rate = 100.0
		}
		out = append(out, SLOStatus{
			Job:          job,
			SuccessRate:  rate,
			Target:       m.policy.TargetPercent,
			Breaching:    total > 0 && rate < m.policy.TargetPercent,
			TotalEvents:  total,
			FailedEvents: failed,
		})
	}
	return out
}

// Reset clears all recorded events for a job.
func (m *SLOManager) Reset(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.events, job)
}

func (m *SLOManager) prune(evts []sloEvent, now time.Time) []sloEvent {
	cutoff := now.Add(-m.policy.Window)
	i := 0
	for i < len(evts) && evts[i].at.Before(cutoff) {
		i++
	}
	return evts[i:]
}
