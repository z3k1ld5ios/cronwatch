package monitor

import (
	"testing"
	"time"
)

func fixedCooldownClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestCooldown_AllowsFirstAlert(t *testing.T) {
	cm := NewCooldownManager(DefaultCooldownPolicy())
	if !cm.Allow("job-a") {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestCooldown_SuppressesWithinWindow(t *testing.T) {
	now := time.Now()
	cm := NewCooldownManager(DefaultCooldownPolicy())
	cm.clock = fixedCooldownClock(now)
	cm.Allow("job-a")

	// Still within base cooldown window.
	cm.clock = fixedCooldownClock(now.Add(1 * time.Minute))
	if cm.Allow("job-a") {
		t.Fatal("expected alert to be suppressed within cooldown window")
	}
}

func TestCooldown_AllowsAfterWindow(t *testing.T) {
	now := time.Now()
	cm := NewCooldownManager(DefaultCooldownPolicy())
	cm.clock = fixedCooldownClock(now)
	cm.Allow("job-a")

	// Advance past the base cooldown (5 minutes).
	cm.clock = fixedCooldownClock(now.Add(6 * time.Minute))
	if !cm.Allow("job-a") {
		t.Fatal("expected alert to be allowed after cooldown window")
	}
}

func TestCooldown_ExponentialBackoff(t *testing.T) {
	policy := CooldownPolicy{
		Base:       1 * time.Minute,
		Multiplier: 2.0,
		Max:        10 * time.Minute,
	}
	now := time.Now()
	cm := NewCooldownManager(policy)
	cm.clock = fixedCooldownClock(now)

	// First alert allowed; next delay = 1m * 2^1 = 2m.
	cm.Allow("job-b")

	// After 1m5s (past base but within 2m window), still suppressed.
	cm.clock = fixedCooldownClock(now.Add(65 * time.Second))
	if cm.Allow("job-b") {
		t.Fatal("expected suppression on second attempt within 2m window")
	}

	// After 2m5s, second alert allowed; next delay = 1m * 2^2 = 4m.
	cm.clock = fixedCooldownClock(now.Add(125 * time.Second))
	if !cm.Allow("job-b") {
		t.Fatal("expected second alert after 2m window")
	}
}

func TestCooldown_MaxCapApplied(t *testing.T) {
	policy := CooldownPolicy{
		Base:       1 * time.Minute,
		Multiplier: 10.0,
		Max:        5 * time.Minute,
	}
	d := (&CooldownManager{policy: policy}).nextDelay(10)
	if d != 5*time.Minute {
		t.Fatalf("expected max cap of 5m, got %v", d)
	}
}

func TestCooldown_ResetClearsState(t *testing.T) {
	now := time.Now()
	cm := NewCooldownManager(DefaultCooldownPolicy())
	cm.clock = fixedCooldownClock(now)
	cm.Allow("job-c")

	cm.Reset("job-c")

	// After reset, should be allowed again immediately.
	if !cm.Allow("job-c") {
		t.Fatal("expected alert to be allowed after reset")
	}
}

func TestCooldown_IndependentJobs(t *testing.T) {
	now := time.Now()
	cm := NewCooldownManager(DefaultCooldownPolicy())
	cm.clock = fixedCooldownClock(now)

	cm.Allow("job-x")

	// job-y has not been seen; should be allowed.
	if !cm.Allow("job-y") {
		t.Fatal("expected independent job to be allowed")
	}
}
