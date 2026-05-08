package monitor

import (
	"testing"
	"time"
)

func fixedRateClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func defaultRatePolicy() RateLimitPolicy {
	return RateLimitPolicy{
		MaxAlerts: 3,
		Window:    10 * time.Minute,
	}
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(defaultRatePolicy())
	for i := 0; i < 3; i++ {
		if !rl.Allow("job-a") {
			t.Fatalf("expected Allow on attempt %d", i+1)
		}
	}
}

func TestRateLimit_SuppressesAtLimit(t *testing.T) {
	rl := NewRateLimiter(defaultRatePolicy())
	for i := 0; i < 3; i++ {
		rl.Allow("job-a")
	}
	if rl.Allow("job-a") {
		t.Fatal("expected suppression after max alerts")
	}
}

func TestRateLimit_ResetsAfterWindow(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(defaultRatePolicy())
	rl.now = fixedRateClock(base)

	for i := 0; i < 3; i++ {
		rl.Allow("job-a")
	}
	// advance past window
	rl.now = fixedRateClock(base.Add(11 * time.Minute))
	if !rl.Allow("job-a") {
		t.Fatal("expected Allow after window reset")
	}
}

func TestRateLimit_IndependentJobs(t *testing.T) {
	rl := NewRateLimiter(defaultRatePolicy())
	for i := 0; i < 3; i++ {
		rl.Allow("job-a")
	}
	if !rl.Allow("job-b") {
		t.Fatal("job-b should not be affected by job-a rate limit")
	}
}

func TestRateLimit_ResetClearsEntry(t *testing.T) {
	rl := NewRateLimiter(defaultRatePolicy())
	for i := 0; i < 3; i++ {
		rl.Allow("job-a")
	}
	rl.Reset("job-a")
	if !rl.Allow("job-a") {
		t.Fatal("expected Allow after Reset")
	}
}

func TestRateLimit_StatsReflectCount(t *testing.T) {
	rl := NewRateLimiter(defaultRatePolicy())
	rl.Allow("job-a")
	rl.Allow("job-a")
	count, windowEnd := rl.Stats("job-a")
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
	if windowEnd.IsZero() {
		t.Fatal("expected non-zero windowEnd")
	}
}

func TestRateLimit_StatsZeroForUnknown(t *testing.T) {
	rl := NewRateLimiter(defaultRatePolicy())
	count, windowEnd := rl.Stats("unknown")
	if count != 0 || !windowEnd.IsZero() {
		t.Fatal("expected zero stats for unknown job")
	}
}
