package monitor

import (
	"sync"
	"time"
)

// BudgetPolicy defines thresholds for error budget consumption.
type BudgetPolicy struct {
	WindowDuration  time.Duration
	TotalAllowed    int     // total allowed failures in window
	WarnThreshold   float64 // fraction consumed before warning (e.g. 0.5)
	CritThreshold   float64 // fraction consumed before critical (e.g. 0.9)
}

// DefaultBudgetPolicy returns sensible defaults.
func DefaultBudgetPolicy() BudgetPolicy {
	return BudgetPolicy{
		WindowDuration: 24 * time.Hour,
		TotalAllowed:   10,
		WarnThreshold:  0.5,
		CritThreshold:  0.9,
	}
}

// BudgetEvent records a single failure event.
type BudgetEvent struct {
	At time.Time
}

// BudgetStatus represents the current state of an error budget.
type BudgetStatus struct {
	Job       string
	Consumed  int
	Remaining int
	Fraction  float64
	Level     string // "ok", "warn", "critical", "exhausted"
}

// BudgetManager tracks error budget consumption per job.
type BudgetManager struct {
	mu     sync.Mutex
	policy BudgetPolicy
	events map[string][]BudgetEvent
	clock  func() time.Time
}

// NewBudgetManager creates a BudgetManager with the given policy.
func NewBudgetManager(policy BudgetPolicy) *BudgetManager {
	return &BudgetManager{
		policy: policy,
		events: make(map[string][]BudgetEvent),
		clock:  time.Now,
	}
}

// RecordFailure records a budget-consuming failure for a job.
func (b *BudgetManager) RecordFailure(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clock()
	b.events[job] = append(b.prune(b.events[job], now), BudgetEvent{At: now})
}

// Status returns the current budget status for a job.
func (b *BudgetManager) Status(job string) BudgetStatus {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clock()
	evs := b.prune(b.events[job], now)
	b.events[job] = evs
	consumed := len(evs)
	remaining := b.policy.TotalAllowed - consumed
	if remaining < 0 {
		remaining = 0
	}
	var fraction float64
	if b.policy.TotalAllowed > 0 {
		fraction = float64(consumed) / float64(b.policy.TotalAllowed)
	}
	level := "ok"
	switch {
	case fraction >= 1.0:
		level = "exhausted"
	case fraction >= b.policy.CritThreshold:
		level = "critical"
	case fraction >= b.policy.WarnThreshold:
		level = "warn"
	}
	return BudgetStatus{Job: job, Consumed: consumed, Remaining: remaining, Fraction: fraction, Level: level}
}

// Reset clears all events for a job.
func (b *BudgetManager) Reset(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.events, job)
}

// AllStatuses returns the budget status for every tracked job.
func (b *BudgetManager) AllStatuses() []BudgetStatus {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clock()
	var out []BudgetStatus
	for job, evs := range b.events {
		evs = b.prune(evs, now)
		b.events[job] = evs
		consumed := len(evs)
		remaining := b.policy.TotalAllowed - consumed
		if remaining < 0 {
			remaining = 0
		}
		var fraction float64
		if b.policy.TotalAllowed > 0 {
			fraction = float64(consumed) / float64(b.policy.TotalAllowed)
		}
		level := "ok"
		switch {
		case fraction >= 1.0:
			level = "exhausted"
		case fraction >= b.policy.CritThreshold:
			level = "critical"
		case fraction >= b.policy.WarnThreshold:
			level = "warn"
		}
		out = append(out, BudgetStatus{Job: job, Consumed: consumed, Remaining: remaining, Fraction: fraction, Level: level})
	}
	return out
}

func (b *BudgetManager) prune(evs []BudgetEvent, now time.Time) []BudgetEvent {
	cutoff := now.Add(-b.policy.WindowDuration)
	var kept []BudgetEvent
	for _, e := range evs {
		if e.At.After(cutoff) {
			kept = append(kept, e)
		}
	}
	return kept
}
