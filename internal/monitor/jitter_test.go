package monitor

import (
	"testing"
	"time"
)

func seedJitter(a *JitterAnalyzer, job string, base time.Time, intervals []time.Duration) {
	t := base
	a.Record(job, t)
	for _, d := range intervals {
		t = t.Add(d)
		a.Record(job, t)
	}
}

func TestJitter_NilWithFewSamples(t *testing.T) {
	a := NewJitterAnalyzer(DefaultJitterPolicy())
	base := time.Now()
	// Only 3 deltas — below MinSamples=5
	seedJitter(a, "job1", base, []time.Duration{
		time.Minute, time.Minute, time.Minute,
	})
	if got := a.Analyze("job1"); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestJitter_LowJitter(t *testing.T) {
	a := NewJitterAnalyzer(DefaultJitterPolicy())
	base := time.Now()
	// Very consistent 60-second intervals
	seedJitter(a, "job1", base, []time.Duration{
		60 * time.Second,
		60 * time.Second,
		60 * time.Second,
		60 * time.Second,
		60 * time.Second,
	})
	res := a.Analyze("job1")
	if res == nil {
		t.Fatal("expected result, got nil")
	}
	if res.High {
		t.Errorf("expected High=false, got true (CV=%.4f)", res.CV)
	}
	if res.CV != 0 {
		t.Errorf("expected CV=0 for uniform intervals, got %.4f", res.CV)
	}
}

func TestJitter_HighJitter(t *testing.T) {
	a := NewJitterAnalyzer(DefaultJitterPolicy())
	base := time.Now()
	// Highly variable intervals
	seedJitter(a, "job2", base, []time.Duration{
		10 * time.Second,
		120 * time.Second,
		5 * time.Second,
		200 * time.Second,
		15 * time.Second,
	})
	res := a.Analyze("job2")
	if res == nil {
		t.Fatal("expected result, got nil")
	}
	if !res.High {
		t.Errorf("expected High=true, got false (CV=%.4f)", res.CV)
	}
}

func TestJitter_ResetClearsSamples(t *testing.T) {
	a := NewJitterAnalyzer(DefaultJitterPolicy())
	base := time.Now()
	seedJitter(a, "job3", base, []time.Duration{
		time.Minute, time.Minute, time.Minute,
		time.Minute, time.Minute,
	})
	if a.Analyze("job3") == nil {
		t.Fatal("expected result before reset")
	}
	a.Reset("job3")
	if got := a.Analyze("job3"); got != nil {
		t.Fatalf("expected nil after reset, got %+v", got)
	}
}

func TestJitter_IndependentJobs(t *testing.T) {
	a := NewJitterAnalyzer(DefaultJitterPolicy())
	base := time.Now()
	seedJitter(a, "alpha", base, []time.Duration{
		time.Minute, time.Minute, time.Minute,
		time.Minute, time.Minute,
	})
	if a.Analyze("beta") != nil {
		t.Error("beta should have no data")
	}
	if a.Analyze("alpha") == nil {
		t.Error("alpha should have data")
	}
}
