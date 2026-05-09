package monitor

import (
	"testing"
	"time"
)

var fixedAnomalyNow = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func seedDetector(t *testing.T, a *AnomalyDetector, job string, durations []time.Duration) {
	t.Helper()
	for _, d := range durations {
		a.Record(job, d, fixedAnomalyNow)
	}
}

func TestAnomaly_NoDetectionWithFewSamples(t *testing.T) {
	a := NewAnomalyDetector(2.0, 60)
	for i := 0; i < 4; i++ {
		r := a.Record("job1", time.Second, fixedAnomalyNow)
		if r.Anomaly {
			t.Errorf("expected no anomaly with only %d samples", i+1)
		}
	}
}

func TestAnomaly_NormalValueNotFlagged(t *testing.T) {
	a := NewAnomalyDetector(2.0, 60)
	base := []time.Duration{
		10 * time.Second, 11 * time.Second, 10 * time.Second,
		10 * time.Second, 11 * time.Second,
	}
	seedDetector(t, a, "job1", base)
	r := a.Record("job1", 10500*time.Millisecond, fixedAnomalyNow)
	if r.Anomaly {
		t.Errorf("expected no anomaly, got z=%.2f", r.ZScore)
	}
}

func TestAnomaly_OutlierIsFlagged(t *testing.T) {
	a := NewAnomalyDetector(2.0, 60)
	base := []time.Duration{
		10 * time.Second, 10 * time.Second, 10 * time.Second,
		10 * time.Second, 10 * time.Second,
	}
	seedDetector(t, a, "job1", base)
	// inject a tiny variance so stddev != 0
	a.samples["job1"][0] = 9.9
	a.samples["job1"][1] = 10.1

	r := a.Record("job1", 60*time.Second, fixedAnomalyNow)
	if !r.Anomaly {
		t.Errorf("expected anomaly for large outlier, got z=%.2f", r.ZScore)
	}
}

func TestAnomaly_ResetClearsSamples(t *testing.T) {
	a := NewAnomalyDetector(2.0, 60)
	base := []time.Duration{
		10 * time.Second, 10 * time.Second, 10 * time.Second,
		10 * time.Second, 10 * time.Second,
	}
	seedDetector(t, a, "job1", base)
	a.Reset("job1")
	r := a.Record("job1", 10*time.Second, fixedAnomalyNow)
	if r.Anomaly {
		t.Error("expected no anomaly after reset")
	}
	if r.ZScore != 0 {
		t.Errorf("expected zero z-score after reset, got %.2f", r.ZScore)
	}
}

func TestAnomaly_WindowBounded(t *testing.T) {
	a := NewAnomalyDetector(2.0, 5)
	for i := 0; i < 20; i++ {
		a.Record("job1", time.Duration(i)*time.Second, fixedAnomalyNow)
	}
	if len(a.samples["job1"]) > 5 {
		t.Errorf("expected at most 5 samples, got %d", len(a.samples["job1"]))
	}
}

func TestAnomaly_AllStats(t *testing.T) {
	a := NewAnomalyDetector(2.0, 60)
	base := []time.Duration{
		10 * time.Second, 12 * time.Second, 10 * time.Second,
		11 * time.Second, 10 * time.Second,
	}
	seedDetector(t, a, "jobA", base)
	stats := a.AllStats()
	s, ok := stats["jobA"]
	if !ok {
		t.Fatal("expected jobA in AllStats")
	}
	if s[0] < 10 || s[0] > 12 {
		t.Errorf("unexpected mean %.2f", s[0])
	}
}
