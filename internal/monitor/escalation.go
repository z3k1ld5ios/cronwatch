package monitor

import (
	"sync"
	"time"
)

// EscalationLevel represents how severe an alert has become.
type EscalationLevel int

const (
	LevelNone    EscalationLevel = 0
	LevelWarning EscalationLevel = 1
	LevelCritical EscalationLevel = 2
)

// EscalationPolicy defines thresholds for escalating alerts.
type EscalationPolicy struct {
	// WarningAfter is how many consecutive failures trigger a warning.
	WarningAfter int
	// CriticalAfter is how many consecutive failures trigger critical.
	CriticalAfter int
	// ResetAfter is the idle duration after which the counter resets.
	ResetAfter time.Duration
}

type escalationState struct {
	count     int
	lastSeen  time.Time
	current   EscalationLevel
}

// EscalationManager tracks consecutive failures and computes alert levels.
type EscalationManager struct {
	mu     sync.Mutex
	policy EscalationPolicy
	states map[string]*escalationState
	clock  func() time.Time
}

// NewEscalationManager creates an EscalationManager with the given policy.
func NewEscalationManager(policy EscalationPolicy) *EscalationManager {
	return &EscalationManager{
		policy: policy,
		states: make(map[string]*escalationState),
		clock:  time.Now,
	}
}

// Record registers a failure for the given job and returns the new level.
func (e *EscalationManager) Record(job string) EscalationLevel {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := e.clock()
	s, ok := e.states[job]
	if !ok {
		s = &escalationState{}
		e.states[job] = s
	}

	if !s.lastSeen.IsZero() && now.Sub(s.lastSeen) > e.policy.ResetAfter {
		s.count = 0
	}
	s.count++
	s.lastSeen = now

	switch {
	case s.count >= e.policy.CriticalAfter:
		s.current = LevelCritical
	case s.count >= e.policy.WarningAfter:
		s.current = LevelWarning
	default:
		s.current = LevelNone
	}
	return s.current
}

// Reset clears the escalation state for a job (e.g. on success).
func (e *EscalationManager) Reset(job string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.states, job)
}

// Level returns the current escalation level for a job without recording.
func (e *EscalationManager) Level(job string) EscalationLevel {
	e.mu.Lock()
	defer e.mu.Unlock()
	if s, ok := e.states[job]; ok {
		return s.current
	}
	return LevelNone
}
