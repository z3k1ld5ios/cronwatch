package monitor

import (
	"testing"
	"time"
)

func seedBaseline(b *BaselineManager, job string, durations ...time.Duration) {
	for _, d := range durations {
		b.Record(job, d)
	}
}

func TestBaseline_NilWithFewSamples(t *testing.T) {
	bm := NewBaselineManager(50, 2.0)
	bm.Record("job", 5*time.Second)
	bm.Record("job", 6*time.Second)
	if got := bm.Check("job", 10*time.Second); got != nil {
		t.Errorf("expected nil with < 3 samples, got %+v", got)
	}
}

func TestBaseline_NormalValueNotFlagged(t *testing.T) {
	bm := NewBaselineManager(50, 2.0)
	for i := 0; i < 10; i++ {
		bm.Record("job", 10*time.Second)
	}
	res := bm.Check("job", 10*time.Second)
	if res == nil {
		t.Fatal("expected result, got nil")
	}
	if res.Anomalous {
		t.Errorf("expected normal, got anomalous (z=%.2f)", res.ZScore)
	}
}

func TestBaseline_OutlierIsFlagged(t *testing.T) {
	bm := NewBaselineManager(50, 2.0)
	for i := 0; i < 20; i++ {
		bm.Record("job", 10*time.Second)
	}
	res := bm.Check("job", 120*time.Second)
	if res == nil {
		t.Fatal("expected result, got nil")
	}
	if !res.Anomalous {
		t.Errorf("expected anomalous, got normal (z=%.2f)", res.ZScore)
	}
}

func TestBaseline_ResetClearsSamples(t *testing.T) {
	bm := NewBaselineManager(50, 2.0)
	seedBaseline(bm, "job",
		10*time.Second, 10*time.Second, 10*time.Second, 10*time.Second)
	bm.Reset("job")
	if got := bm.Check("job", 10*time.Second); got != nil {
		t.Errorf("expected nil after reset, got %+v", got)
	}
}

func TestBaseline_BoundedSamples(t *testing.T) {
	bm := NewBaselineManager(5, 2.0)
	for i := 0; i < 20; i++ {
		bm.Record("job", time.Duration(i)*time.Second)
	}
	bm.mu.Lock()
	count := len(bm.samples["job"])
	bm.mu.Unlock()
	if count != 5 {
		t.Errorf("expected 5 samples, got %d", count)
	}
}

func TestBaseline_AllStats(t *testing.T) {
	bm := NewBaselineManager(50, 2.0)
	seedBaseline(bm, "alpha",
		8*time.Second, 10*time.Second, 12*time.Second)
	seedBaseline(bm, "beta",
		5*time.Second, 5*time.Second, 5*time.Second)
	stats := bm.AllStats()
	if _, ok := stats["alpha"]; !ok {
		t.Error("expected alpha in stats")
	}
	if _, ok := stats["beta"]; !ok {
		t.Error("expected beta in stats")
	}
	if stats["beta"][1] != 0 {
		t.Errorf("expected zero stddev for constant samples, got %.4f", stats["beta"][1])
	}
}

func TestBaseline_IndependentJobs(t *testing.T) {
	bm := NewBaselineManager(50, 2.0)
	for i := 0; i < 10; i++ {
		bm.Record("fast", 1*time.Second)
		bm.Record("slow", 60*time.Second)
	}
	if r := bm.Check("fast", 60*time.Second); r == nil || !r.Anomalous {
		t.Error("expected fast job to flag 60s as anomalous")
	}
	if r := bm.Check("slow", 60*time.Second); r == nil || r.Anomalous {
		t.Error("expected slow job to accept 60s as normal")
	}
}
