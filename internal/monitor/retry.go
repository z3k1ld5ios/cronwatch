package monitor

import (
	"sync"
	"time"
)

// RetryPolicy defines how alert retries are handled when a webhook delivery fails.
type RetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
}

// RetryState tracks the retry state for a single job alert.
type RetryState struct {
	Attempts  int
	LastTried time.Time
	NextRetry time.Time
}

// RetryManager manages retry state for failed alert deliveries across jobs.
type RetryManager struct {
	mu     sync.Mutex
	policy RetryPolicy
	states map[string]*RetryState
}

// NewRetryManager creates a RetryManager with the given policy.
func NewRetryManager(policy RetryPolicy) *RetryManager {
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 3
	}
	if policy.Backoff <= 0 {
		policy.Backoff = 30 * time.Second
	}
	return &RetryManager{
		policy: policy,
		states: make(map[string]*RetryState),
	}
}

// ShouldRetry returns true if the job alert should be retried now.
func (r *RetryManager) ShouldRetry(jobName string, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	state, ok := r.states[jobName]
	if !ok {
		return false
	}
	if state.Attempts >= r.policy.MaxAttempts {
		return false
	}
	return !now.Before(state.NextRetry)
}

// RecordFailure records a failed delivery attempt for the given job.
func (r *RetryManager) RecordFailure(jobName string, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	state, ok := r.states[jobName]
	if !ok {
		state = &RetryState{}
		r.states[jobName] = state
	}
	state.Attempts++
	state.LastTried = now
	backoff := time.Duration(state.Attempts) * r.policy.Backoff
	state.NextRetry = now.Add(backoff)
}

// RecordSuccess clears the retry state for a successfully delivered alert.
func (r *RetryManager) RecordSuccess(jobName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.states, jobName)
}

// Attempts returns the current attempt count for a job (0 if none).
func (r *RetryManager) Attempts(jobName string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if state, ok := r.states[jobName]; ok {
		return state.Attempts
	}
	return 0
}
