package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/webhook"
)

// fakeTracker implements Tracker for testing.
type fakeTracker struct {
	runs map[string]time.Time
}

func (f *fakeTracker) LastRun(name string) (time.Time, bool) {
	t, ok := f.runs[name]
	return t, ok
}

func TestChecker_DetectsMissedJob(t *testing.T) {
	var received AlertPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	notifier := webhook.NewNotifier(ts.URL)
	tracker := &fakeTracker{
		runs: map[string]time.Time{
			"backup": time.Now().Add(-2 * time.Hour),
		},
	}
	cfg := []JobConfig{{
		Name:            "backup",
		CronExpr:        "0 * * * *",
		DriftThreshold:  2 * time.Minute,
		MissedThreshold: 10 * time.Minute,
	}}
	checker := NewChecker(cfg, tracker, notifier, time.Minute)
	// Trigger a manual check at a time well past the expected next run.
	checker.checkAll(time.Now())

	if received.Kind != "missed" {
		t.Errorf("expected alert kind 'missed', got %q", received.Kind)
	}
	if received.Job != "backup" {
		t.Errorf("expected job 'backup', got %q", received.Job)
	}
}

func TestChecker_NoAlertWhenOnTime(t *testing.T) {
	alertSent := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		alertSent = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	notifier := webhook.NewNotifier(ts.URL)
	now := time.Now()
	tracker := &fakeTracker{
		runs: map[string]time.Time{
			"heartbeat": now.Add(-30 * time.Second),
		},
	}
	cfg := []JobConfig{{
		Name:            "heartbeat",
		CronExpr:        "* * * * *",
		DriftThreshold:  30 * time.Second,
		MissedThreshold: 2 * time.Minute,
	}}
	checker := NewChecker(cfg, tracker, notifier, time.Minute)
	checker.checkAll(now)

	if alertSent {
		t.Error("expected no alert for on-time job, but one was sent")
	}
}
