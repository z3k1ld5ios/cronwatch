package monitor

import (
	"testing"
	"time"
)

func TestSuppression_AllowFirstAlert(t *testing.T) {
	s := NewSuppression(5 * time.Minute)
	now := time.Now()

	if !s.Allow("job-a", now) {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestSuppression_SuppressesWithinCooldown(t *testing.T) {
	s := NewSuppression(5 * time.Minute)
	now := time.Now()

	s.Allow("job-a", now)

	// Second call within cooldown window
	if s.Allow("job-a", now.Add(2*time.Minute)) {
		t.Fatal("expected alert to be suppressed within cooldown")
	}
}

func TestSuppression_AllowsAfterCooldown(t *testing.T) {
	s := NewSuppression(5 * time.Minute)
	now := time.Now()

	s.Allow("job-a", now)

	// Call after cooldown has elapsed
	if !s.Allow("job-a", now.Add(6*time.Minute)) {
		t.Fatal("expected alert to be allowed after cooldown")
	}
}

func TestSuppression_IndependentJobs(t *testing.T) {
	s := NewSuppression(5 * time.Minute)
	now := time.Now()

	s.Allow("job-a", now)

	// Different job should not be suppressed
	if !s.Allow("job-b", now.Add(1*time.Minute)) {
		t.Fatal("expected independent job to be allowed")
	}
}

func TestSuppression_ResetClearsRecord(t *testing.T) {
	s := NewSuppression(5 * time.Minute)
	now := time.Now()

	s.Allow("job-a", now)
	s.Reset("job-a")

	// After reset, alert should be allowed again immediately
	if !s.Allow("job-a", now.Add(1*time.Minute)) {
		t.Fatal("expected alert to be allowed after reset")
	}
}

func TestSuppression_Snapshot(t *testing.T) {
	s := NewSuppression(5 * time.Minute)
	now := time.Now()

	s.Allow("job-a", now)
	s.Allow("job-b", now.Add(time.Second))

	snap := s.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries in snapshot, got %d", len(snap))
	}
	if _, ok := snap["job-a"]; !ok {
		t.Error("expected job-a in snapshot")
	}
	if _, ok := snap["job-b"]; !ok {
		t.Error("expected job-b in snapshot")
	}
}
