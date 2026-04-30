package schedule

import (
	"fmt"
	"sync"
	"time"
)

// JobStatus tracks the last known state of a monitored cron job
type JobStatus struct {
	Name        string
	Expression  *CronExpression
	LastSeen    time.Time
	NextExpected time.Time
	Missed      bool
	Drift       time.Duration
}

// Tracker maintains state for all monitored jobs
type Tracker struct {
	mu      sync.RWMutex
	jobs    map[string]*JobStatus
	driftTolerance time.Duration
}

// NewTracker creates a Tracker with the given drift tolerance window
func NewTracker(driftTolerance time.Duration) *Tracker {
	return &Tracker{
		jobs:           make(map[string]*JobStatus),
		driftTolerance: driftTolerance,
	}
}

// Register adds a new job to be tracked
func (t *Tracker) Register(name, expr string) error {
	parsed, err := Parse(expr)
	if err != nil {
		return fmt.Errorf("register job %q: %w", name, err)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.jobs[name] = &JobStatus{
		Name:         name,
		Expression:   parsed,
		NextExpected: parsed.Next(time.Now()),
	}
	return nil
}

// RecordRun records that a job ran at the given time and returns drift
func (t *Tracker) RecordRun(name string, at time.Time) (time.Duration, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	status, ok := t.jobs[name]
	if !ok {
		return 0, fmt.Errorf("unknown job %q", name)
	}

	drift := at.Sub(status.NextExpected)
	if drift < 0 {
		drift = -drift
	}

	status.LastSeen = at
	status.Drift = drift
	status.Missed = false
	status.NextExpected = status.Expression.Next(at)
	return drift, nil
}

// CheckMissed marks jobs that have not run past their expected time + tolerance
func (t *Tracker) CheckMissed(now time.Time) []*JobStatus {
	t.mu.Lock()
	defer t.mu.Unlock()

	var missed []*JobStatus
	for _, status := range t.jobs {
		if !status.NextExpected.IsZero() && now.After(status.NextExpected.Add(t.driftTolerance)) {
			status.Missed = true
			missed = append(missed, status)
			// Advance to next expected so we don't re-alert every tick
			status.NextExpected = status.Expression.Next(now)
		}
	}
	return missed
}

// Status returns a snapshot of the given job's current state
func (t *Tracker) Status(name string) (*JobStatus, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	status, ok := t.jobs[name]
	if !ok {
		return nil, fmt.Errorf("unknown job %q", name)
	}
	copy := *status
	return &copy, nil
}
