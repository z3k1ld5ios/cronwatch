package monitor

import (
	"testing"
	"time"
)

func fixedBurnClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestBurnRate_NoneWhenNoEvents(t *testing.T) {
	bm := NewBurnRateManager(DefaultBurnRatePolicy())
	res := bm.Compute("myjob")
	if res.Level != BurnRateNone {
		t.Errorf("expected none, got %s", res.Level)
	}
	if res.Rate != 0 {
		t.Errorf("expected rate 0, got %f", res.Rate)
	}
}

func TestBurnRate_WarningLevel(t *testing.T) {
	now := time.Now()
	policy := DefaultBurnRatePolicy()
	// 0.5 failures/min * 30 min = 15 failures => warning
	bm := NewBurnRateManager(policy)
	bm.clock = fixedBurnClock(now)
	for i := 0; i < 15; i++ {
		bm.RecordFailure("job")
	}
	res := bm.Compute("job")
	if res.Level != BurnRateWarning {
		t.Errorf("expected warning, got %s (rate=%f)", res.Level, res.Rate)
	}
}

func TestBurnRate_CriticalLevel(t *testing.T) {
	now := time.Now()
	policy := DefaultBurnRatePolicy()
	// 2.0 failures/min * 30 min = 60 failures => critical
	bm := NewBurnRateManager(policy)
	bm.clock = fixedBurnClock(now)
	for i := 0; i < 60; i++ {
		bm.RecordFailure("job")
	}
	res := bm.Compute("job")
	if res.Level != BurnRateCritical {
		t.Errorf("expected critical, got %s (rate=%f)", res.Level, res.Rate)
	}
}

func TestBurnRate_EventsPrunedOutsideWindow(t *testing.T) {
	policy := DefaultBurnRatePolicy()
	policy.Window = 10 * time.Minute
	bm := NewBurnRateManager(policy)
	old := time.Now().Add(-20 * time.Minute)
	bm.clock = fixedBurnClock(old)
	for i := 0; i < 50; i++ {
		bm.RecordFailure("job")
	}
	// advance clock beyond window
	now := time.Now()
	bm.clock = fixedBurnClock(now)
	res := bm.Compute("job")
	if res.Level != BurnRateNone {
		t.Errorf("expected none after pruning, got %s (rate=%f)", res.Level, res.Rate)
	}
}

func TestBurnRate_ResetClearsEvents(t *testing.T) {
	now := time.Now()
	bm := NewBurnRateManager(DefaultBurnRatePolicy())
	bm.clock = fixedBurnClock(now)
	for i := 0; i < 60; i++ {
		bm.RecordFailure("job")
	}
	bm.Reset("job")
	res := bm.Compute("job")
	if res.Level != BurnRateNone {
		t.Errorf("expected none after reset, got %s", res.Level)
	}
}

func TestBurnRate_IndependentJobs(t *testing.T) {
	now := time.Now()
	bm := NewBurnRateManager(DefaultBurnRatePolicy())
	bm.clock = fixedBurnClock(now)
	for i := 0; i < 60; i++ {
		bm.RecordFailure("jobA")
	}
	resB := bm.Compute("jobB")
	if resB.Level != BurnRateNone {
		t.Errorf("jobB should be unaffected, got %s", resB.Level)
	}
	resA := bm.Compute("jobA")
	if resA.Level != BurnRateCritical {
		t.Errorf("jobA should be critical, got %s", resA.Level)
	}
}
