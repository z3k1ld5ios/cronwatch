package monitor

import (
	"strings"
	"testing"
	"time"
)

func TestBuildAlert_Missed(t *testing.T) {
	now := time.Now()
	lastRun := now.Add(-2 * time.Hour)
	expected := now.Add(-time.Hour)

	alert := buildAlert("db-backup", "missed", lastRun, expected, now)

	if alert.Kind != "missed" {
		t.Errorf("expected kind 'missed', got %q", alert.Kind)
	}
	if alert.Job != "db-backup" {
		t.Errorf("expected job 'db-backup', got %q", alert.Job)
	}
	if !strings.Contains(alert.Message, "missed") {
		t.Errorf("expected message to contain 'missed', got: %s", alert.Message)
	}
}

func TestBuildAlert_Drift(t *testing.T) {
	now := time.Now()
	lastRun := now.Add(-70 * time.Minute)
	expected := now.Add(-10 * time.Minute)

	alert := buildAlert("report-gen", "drift", lastRun, expected, now)

	if alert.Kind != "drift" {
		t.Errorf("expected kind 'drift', got %q", alert.Kind)
	}
	if !strings.Contains(alert.Message, "drifting") {
		t.Errorf("expected message to mention drift, got: %s", alert.Message)
	}
	if alert.CheckedAt != now {
		t.Errorf("expected CheckedAt to match now")
	}
}

func TestBuildAlert_UnknownKind(t *testing.T) {
	now := time.Now()
	alert := buildAlert("job", "unknown", now, now, now)
	if !strings.Contains(alert.Message, "anomaly") {
		t.Errorf("expected fallback message to contain 'anomaly', got: %s", alert.Message)
	}
}
