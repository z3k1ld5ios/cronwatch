package monitor

import (
	"testing"
	"time"
)

func defaultPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		Backoff:     10 * time.Second,
	}
}

func TestRetry_NoStateInitially(t *testing.T) {
	rm := NewRetryManager(defaultPolicy())
	if rm.ShouldRetry("job-a", time.Now()) {
		t.Fatal("expected no retry for unknown job")
	}
	if rm.Attempts("job-a") != 0 {
		t.Fatal("expected 0 attempts for unknown job")
	}
}

func TestRetry_RecordFailureEnablesRetry(t *testing.T) {
	now := time.Now()
	rm := NewRetryManager(defaultPolicy())
	rm.RecordFailure("job-a", now)

	if rm.Attempts("job-a") != 1 {
		t.Fatalf("expected 1 attempt, got %d", rm.Attempts("job-a"))
	}
	// next retry is now + 1*backoff = now+10s, so not yet ready
	if rm.ShouldRetry("job-a", now) {
		t.Fatal("should not retry before backoff window")
	}
	// advance past backoff
	if !rm.ShouldRetry("job-a", now.Add(11*time.Second)) {
		t.Fatal("expected retry after backoff window")
	}
}

func TestRetry_ExceedsMaxAttempts(t *testing.T) {
	now := time.Now()
	rm := NewRetryManager(defaultPolicy())
	for i := 0; i < 3; i++ {
		rm.RecordFailure("job-b", now)
	}
	if rm.ShouldRetry("job-b", now.Add(time.Hour)) {
		t.Fatal("should not retry after max attempts exceeded")
	}
}

func TestRetry_RecordSuccessClearsState(t *testing.T) {
	now := time.Now()
	rm := NewRetryManager(defaultPolicy())
	rm.RecordFailure("job-c", now)
	rm.RecordSuccess("job-c")

	if rm.Attempts("job-c") != 0 {
		t.Fatal("expected attempts reset after success")
	}
	if rm.ShouldRetry("job-c", now.Add(time.Hour)) {
		t.Fatal("should not retry after success clears state")
	}
}

func TestRetry_BackoffGrowsWithAttempts(t *testing.T) {
	now := time.Now()
	rm := NewRetryManager(defaultPolicy())

	rm.RecordFailure("job-d", now)             // attempt 1, next = now+10s
	rm.RecordFailure("job-d", now.Add(11*time.Second)) // attempt 2, next = now+11s+20s

	if rm.Attempts("job-d") != 2 {
		t.Fatalf("expected 2 attempts, got %d", rm.Attempts("job-d"))
	}
	// 20s after second failure means we need now+11s+20s = now+31s
	if rm.ShouldRetry("job-d", now.Add(25*time.Second)) {
		t.Fatal("should not retry before second backoff window")
	}
	if !rm.ShouldRetry("job-d", now.Add(32*time.Second)) {
		t.Fatal("expected retry after second backoff window")
	}
}

func TestRetry_DefaultPolicyApplied(t *testing.T) {
	rm := NewRetryManager(RetryPolicy{})
	if rm.policy.MaxAttempts != 3 {
		t.Fatalf("expected default MaxAttempts=3, got %d", rm.policy.MaxAttempts)
	}
	if rm.policy.Backoff != 30*time.Second {
		t.Fatalf("expected default Backoff=30s, got %v", rm.policy.Backoff)
	}
}
