package monitor

import (
	"testing"
	"time"
)

var trendBase = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestTrend_StableWithFewSamples(t *testing.T) {
	a := NewTrendAnalyzer(10)
	result := a.Analyze("job1")
	if result.Direction != TrendStable {
		t.Errorf("expected stable with no samples, got %s", result.Direction)
	}
	a.Record("job1", trendBase, 5.0)
	result = a.Analyze("job1")
	if result.Direction != TrendStable {
		t.Errorf("expected stable with one sample, got %s", result.Direction)
	}
}

func TestTrend_WorsensOverTime(t *testing.T) {
	a := NewTrendAnalyzer(10)
	for i := 0; i < 5; i++ {
		at := trendBase.Add(time.Duration(i) * time.Minute)
		a.Record("job1", at, float64(i*10))
	}
	result := a.Analyze("job1")
	if result.Direction != TrendWorsening {
		t.Errorf("expected worsening, got %s (slope=%.4f)", result.Direction, result.Slope)
	}
	if result.Slope <= 0 {
		t.Errorf("expected positive slope, got %.4f", result.Slope)
	}
}

func TestTrend_ImprovesOverTime(t *testing.T) {
	a := NewTrendAnalyzer(10)
	for i := 0; i < 5; i++ {
		at := trendBase.Add(time.Duration(i) * time.Minute)
		a.Record("job1", at, float64(50-i*10))
	}
	result := a.Analyze("job1")
	if result.Direction != TrendImproving {
		t.Errorf("expected improving, got %s (slope=%.4f)", result.Direction, result.Slope)
	}
	if result.Slope >= 0 {
		t.Errorf("expected negative slope, got %.4f", result.Slope)
	}
}

func TestTrend_BoundedSamples(t *testing.T) {
	a := NewTrendAnalyzer(3)
	for i := 0; i < 10; i++ {
		at := trendBase.Add(time.Duration(i) * time.Minute)
		a.Record("job1", at, float64(i))
	}
	result := a.Analyze("job1")
	if result.Samples != 3 {
		t.Errorf("expected 3 samples, got %d", result.Samples)
	}
}

func TestTrend_AllTrendsMultipleJobs(t *testing.T) {
	a := NewTrendAnalyzer(10)
	for i := 0; i < 3; i++ {
		at := trendBase.Add(time.Duration(i) * time.Minute)
		a.Record("jobA", at, float64(i*5))
		a.Record("jobB", at, float64(20-i*5))
	}
	trends := a.AllTrends()
	if len(trends) != 2 {
		t.Fatalf("expected 2 trends, got %d", len(trends))
	}
	byJob := make(map[string]TrendSummary)
	for _, tr := range trends {
		byJob[tr.Job] = tr
	}
	if byJob["jobA"].Direction != TrendWorsening {
		t.Errorf("jobA: expected worsening, got %s", byJob["jobA"].Direction)
	}
	if byJob["jobB"].Direction != TrendImproving {
		t.Errorf("jobB: expected improving, got %s", byJob["jobB"].Direction)
	}
}

func TestTrend_DefaultMaxSamples(t *testing.T) {
	a := NewTrendAnalyzer(0) // should default to 10
	for i := 0; i < 15; i++ {
		at := trendBase.Add(time.Duration(i) * time.Minute)
		a.Record("job1", at, float64(i))
	}
	result := a.Analyze("job1")
	if result.Samples != 10 {
		t.Errorf("expected default max 10, got %d", result.Samples)
	}
}
