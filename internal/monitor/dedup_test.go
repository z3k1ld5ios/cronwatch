package monitor

import (
	"testing"
	"time"
)

func fixedDedupClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestDedup_NotDuplicateFirstTime(t *testing.T) {
	d := NewDedupManager(5 * time.Minute)
	if d.IsDuplicate("backup", "missed", "fp1") {
		t.Fatal("expected not duplicate on first check")
	}
}

func TestDedup_DuplicateWithinWindow(t *testing.T) {
	now := time.Now()
	d := NewDedupManager(5 * time.Minute)
	d.clock = fixedDedupClock(now)

	d.Record("backup", "missed", "fp1")
	d.clock = fixedDedupClock(now.Add(2 * time.Minute))

	if !d.IsDuplicate("backup", "missed", "fp1") {
		t.Fatal("expected duplicate within window")
	}
}

func TestDedup_NotDuplicateAfterWindow(t *testing.T) {
	now := time.Now()
	d := NewDedupManager(5 * time.Minute)
	d.clock = fixedDedupClock(now)

	d.Record("backup", "missed", "fp1")
	d.clock = fixedDedupClock(now.Add(6 * time.Minute))

	if d.IsDuplicate("backup", "missed", "fp1") {
		t.Fatal("expected not duplicate after window expires")
	}
}

func TestDedup_NotDuplicateOnFingerprintChange(t *testing.T) {
	now := time.Now()
	d := NewDedupManager(10 * time.Minute)
	d.clock = fixedDedupClock(now)

	d.Record("backup", "drift", "fp-old")
	d.clock = fixedDedupClock(now.Add(1 * time.Minute))

	if d.IsDuplicate("backup", "drift", "fp-new") {
		t.Fatal("expected not duplicate when fingerprint changes")
	}
}

func TestDedup_IndependentKinds(t *testing.T) {
	now := time.Now()
	d := NewDedupManager(5 * time.Minute)
	d.clock = fixedDedupClock(now)

	d.Record("backup", "missed", "fp1")

	if d.IsDuplicate("backup", "drift", "fp1") {
		t.Fatal("different kind should not be a duplicate")
	}
}

func TestDedup_ResetClearsEntry(t *testing.T) {
	now := time.Now()
	d := NewDedupManager(5 * time.Minute)
	d.clock = fixedDedupClock(now)

	d.Record("backup", "missed", "fp1")
	d.Reset("backup", "missed")

	if d.IsDuplicate("backup", "missed", "fp1") {
		t.Fatal("expected not duplicate after reset")
	}
	if d.Len() != 0 {
		t.Fatalf("expected 0 entries after reset, got %d", d.Len())
	}
}

func TestDedup_LenTracksEntries(t *testing.T) {
	d := NewDedupManager(5 * time.Minute)
	d.Record("job-a", "missed", "fp1")
	d.Record("job-b", "missed", "fp2")
	d.Record("job-a", "drift", "fp3")

	if d.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", d.Len())
	}
}
