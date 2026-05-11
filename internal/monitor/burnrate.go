package monitor

import (
	"sync"
	"time"
)

// BurnRatePolicy configures burn rate alerting thresholds.
type BurnRatePolicy struct {
	Window        time.Duration
	CriticalRate  float64 // failures per minute to be considered critical
	WarningRate   float64 // failures per minute to be considered warning
}

// DefaultBurnRatePolicy returns sensible defaults.
func DefaultBurnRatePolicy() BurnRatePolicy {
	return BurnRatePolicy{
		Window:       30 * time.Minute,
		CriticalRate: 2.0,
		WarningRate:  0.5,
	}
}

// BurnRateLevel represents the severity of the current burn rate.
type BurnRateLevel string

const (
	BurnRateNone     BurnRateLevel = "none"
	BurnRateWarning  BurnRateLevel = "warning"
	BurnRateCritical BurnRateLevel = "critical"
)

// BurnRateResult holds the computed burn rate and its level.
type BurnRateResult struct {
	Job        string
	Rate       float64
	Level      BurnRateLevel
	WindowUsed time.Duration
}

type burnEvent struct {
	at time.Time
}

// BurnRateManager tracks failure events per job and computes burn rates.
type BurnRateManager struct {
	mu     sync.Mutex
	policy BurnRatePolicy
	events map[string][]burnEvent
	clock  func() time.Time
}

// NewBurnRateManager creates a new BurnRateManager with the given policy.
func NewBurnRateManager(policy BurnRatePolicy) *BurnRateManager {
	return &BurnRateManager{
		policy: policy,
		events: make(map[string][]burnEvent),
		clock:  time.Now,
	}
}

// RecordFailure records a failure event for the given job at the current time.
func (b *BurnRateManager) RecordFailure(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clock()
	b.events[job] = append(b.events[job], burnEvent{at: now})
	b.prune(job, now)
}

// Compute returns the current burn rate result for the given job.
func (b *BurnRateManager) Compute(job string) BurnRateResult {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clock()
	b.prune(job, now)
	evs := b.events[job]
	minutes := b.policy.Window.Minutes()
	rate := float64(len(evs)) / minutes
	level := BurnRateNone
	switch {
	case rate >= b.policy.CriticalRate:
		level = BurnRateCritical
	case rate >= b.policy.WarningRate:
		level = BurnRateWarning
	}
	return BurnRateResult{Job: job, Rate: rate, Level: level, WindowUsed: b.policy.Window}
}

// Reset clears all recorded events for the given job.
func (b *BurnRateManager) Reset(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.events, job)
}

// prune removes events outside the current window. Must be called with lock held.
func (b *BurnRateManager) prune(job string, now time.Time) {
	cutoff := now.Add(-b.policy.Window)
	evs := b.events[job]
	i := 0
	for i < len(evs) && evs[i].at.Before(cutoff) {
		i++
	}
	b.events[job] = evs[i:]
}
