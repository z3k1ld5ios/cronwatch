package monitor

import (
	"sync"
	"time"
)

// RateLimitPolicy defines the maximum number of alerts allowed within a window.
type RateLimitPolicy struct {
	MaxAlerts int
	Window    time.Duration
}

type rateLimitEntry struct {
	count     int
	windowEnd time.Time
}

// RateLimiter tracks per-job alert rates and suppresses when the limit is exceeded.
type RateLimiter struct {
	mu      sync.Mutex
	policy  RateLimitPolicy
	entries map[string]*rateLimitEntry
	now     func() time.Time
}

// NewRateLimiter creates a RateLimiter with the given policy.
func NewRateLimiter(policy RateLimitPolicy) *RateLimiter {
	return &RateLimiter{
		policy:  policy,
		entries: make(map[string]*rateLimitEntry),
		now:     time.Now,
	}
}

// Allow returns true if the alert for jobName is within the rate limit.
// It increments the counter and resets the window when expired.
func (r *RateLimiter) Allow(jobName string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	e, ok := r.entries[jobName]
	if !ok || now.After(e.windowEnd) {
		r.entries[jobName] = &rateLimitEntry{
			count:     1,
			windowEnd: now.Add(r.policy.Window),
		}
		return true
	}

	if e.count >= r.policy.MaxAlerts {
		return false
	}
	e.count++
	return true
}

// Reset clears the rate limit state for a specific job.
func (r *RateLimiter) Reset(jobName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, jobName)
}

// Stats returns the current alert count and window-end time for a job.
// Returns zero values if no entry exists.
func (r *RateLimiter) Stats(jobName string) (count int, windowEnd time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.entries[jobName]; ok {
		return e.count, e.windowEnd
	}
	return 0, time.Time{}
}
