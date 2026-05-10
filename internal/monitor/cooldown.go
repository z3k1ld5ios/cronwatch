package monitor

import (
	"sync"
	"time"
)

// CooldownPolicy defines how long to wait between repeated alerts for a job.
type CooldownPolicy struct {
	Base       time.Duration
	Multiplier float64
	Max        time.Duration
}

// DefaultCooldownPolicy returns a sensible default exponential cooldown policy.
func DefaultCooldownPolicy() CooldownPolicy {
	return CooldownPolicy{
		Base:       5 * time.Minute,
		Multiplier: 2.0,
		Max:        2 * time.Hour,
	}
}

// cooldownState tracks per-job cooldown state.
type cooldownState struct {
	attempts  int
	nextAllow time.Time
}

// CooldownManager enforces exponential backoff between repeated alerts per job.
type CooldownManager struct {
	mu     sync.Mutex
	policy CooldownPolicy
	state  map[string]*cooldownState
	clock  func() time.Time
}

// NewCooldownManager creates a CooldownManager with the given policy.
func NewCooldownManager(policy CooldownPolicy) *CooldownManager {
	return &CooldownManager{
		policy: policy,
		state:  make(map[string]*cooldownState),
		clock:  time.Now,
	}
}

// Allow returns true if an alert for the given job should be sent now.
// If allowed, it advances the cooldown window for the next call.
func (c *CooldownManager) Allow(job string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	s, ok := c.state[job]
	if !ok {
		c.state[job] = &cooldownState{
			attempts:  1,
			nextAllow: now.Add(c.nextDelay(0)),
		}
		return true
	}
	if now.Before(s.nextAllow) {
		return false
	}
	s.attempts++
	s.nextAllow = now.Add(c.nextDelay(s.attempts))
	return true
}

// Reset clears the cooldown state for the given job.
func (c *CooldownManager) Reset(job string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.state, job)
}

// nextDelay computes the delay for the given attempt count.
func (c *CooldownManager) nextDelay(attempts int) time.Duration {
	if attempts <= 0 {
		return c.policy.Base
	}
	d := float64(c.policy.Base)
	for i := 0; i < attempts; i++ {
		d *= c.policy.Multiplier
	}
	if d > float64(c.policy.Max) {
		return c.policy.Max
	}
	return time.Duration(d)
}
