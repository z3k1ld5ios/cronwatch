package monitor

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestWindow_NotSuppressedOutsideWindow(t *testing.T) {
	now := time.Now()
	wm := NewWindowManager(fixedClock(now))
	wm.Add(WindowConfig{
		Name:  "maintenance",
		Start: now.Add(1 * time.Hour),
		End:   now.Add(2 * time.Hour),
	})
	if wm.IsSuppressed("job1") {
		t.Error("expected job1 to not be suppressed before window starts")
	}
}

func TestWindow_SuppressedInsideWindowAllJobs(t *testing.T) {
	now := time.Now()
	wm := NewWindowManager(fixedClock(now))
	wm.Add(WindowConfig{
		Name:  "global",
		Start: now.Add(-1 * time.Minute),
		End:   now.Add(1 * time.Hour),
	})
	if !wm.IsSuppressed("any-job") {
		t.Error("expected any-job to be suppressed during global window")
	}
}

func TestWindow_SuppressedForSpecificJob(t *testing.T) {
	now := time.Now()
	wm := NewWindowManager(fixedClock(now))
	wm.Add(WindowConfig{
		Name:     "targeted",
		Start:    now.Add(-1 * time.Minute),
		End:      now.Add(1 * time.Hour),
		JobNames: []string{"backup"},
	})
	if !wm.IsSuppressed("backup") {
		t.Error("expected backup to be suppressed")
	}
	if wm.IsSuppressed("report") {
		t.Error("expected report to not be suppressed")
	}
}

func TestWindow_RemoveWindow(t *testing.T) {
	now := time.Now()
	wm := NewWindowManager(fixedClock(now))
	wm.Add(WindowConfig{
		Name:  "temp",
		Start: now.Add(-1 * time.Minute),
		End:   now.Add(1 * time.Hour),
	})
	wm.Remove("temp")
	if wm.IsSuppressed("job1") {
		t.Error("expected job1 to not be suppressed after window removed")
	}
}

func TestWindow_ActiveWindows(t *testing.T) {
	now := time.Now()
	wm := NewWindowManager(fixedClock(now))
	wm.Add(WindowConfig{
		Name:  "active",
		Start: now.Add(-5 * time.Minute),
		End:   now.Add(5 * time.Minute),
	})
	wm.Add(WindowConfig{
		Name:  "future",
		Start: now.Add(1 * time.Hour),
		End:   now.Add(2 * time.Hour),
	})
	active := wm.ActiveWindows()
	if len(active) != 1 {
		t.Fatalf("expected 1 active window, got %d", len(active))
	}
	if active[0].Name != "active" {
		t.Errorf("expected window named 'active', got %q", active[0].Name)
	}
}
