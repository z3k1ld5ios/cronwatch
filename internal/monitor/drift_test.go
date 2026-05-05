package monitor

import (
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/schedule"
)

func newAnalyzerWithRecord(t *testing.T, job string, at time.Time) *DriftAnalyzer {
	t.Helper()
	h := schedule.NewHistory(10)
	h.Record(job, at)
	return NewDriftAnalyzer(h)
}

func TestDriftAnalyzer_NoDrift(t *testing.T) {
	expected := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	a := newAnalyzerWithRecord(t, "backup", expected)

	res, ok := a.Analyze("backup", expected, 30*time.Second)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if res.Drift != 0 {
		t.Errorf("expected zero drift, got %v", res.Drift)
	}
	if res.IsSignificant {
		t.Error("expected drift not significant")
	}
}

func TestDriftAnalyzer_SignificantDrift(t *testing.T) {
	expected := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	actual := expected.Add(2 * time.Minute)
	a := newAnalyzerWithRecord(t, "backup", actual)

	res, ok := a.Analyze("backup", expected, 30*time.Second)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if res.Drift != 2*time.Minute {
		t.Errorf("unexpected drift: %v", res.Drift)
	}
	if !res.IsSignificant {
		t.Error("expected drift to be significant")
	}
}

func TestDriftAnalyzer_NegativeDrift(t *testing.T) {
	expected := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	actual := expected.Add(-45 * time.Second)
	a := newAnalyzerWithRecord(t, "backup", actual)

	res, ok := a.Analyze("backup", expected, 30*time.Second)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if res.Drift != -45*time.Second {
		t.Errorf("unexpected drift: %v", res.Drift)
	}
	if res.AbsDrift != 45*time.Second {
		t.Errorf("unexpected abs drift: %v", res.AbsDrift)
	}
	if !res.IsSignificant {
		t.Error("expected negative drift to be significant")
	}
}

func TestDriftAnalyzer_NoHistory(t *testing.T) {
	h := schedule.NewHistory(10)
	a := NewDriftAnalyzer(h)

	_, ok := a.Analyze("missing", time.Now(), 30*time.Second)
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}
