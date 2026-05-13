package monitor

import (
	"testing"
	"time"
)

func fixedSLOClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSLO_PerfectSuccessRate(t *testing.T) {
	m := NewSLOManager(DefaultSLOPolicy())
	now := time.Now()
	m.clock = fixedSLOClock(now)
	for i := 0; i < 10; i++ {
		m.Record("backup", true)
	}
	s := m.Status("backup")
	if s.SuccessRate != 100.0 {
		t.Errorf("expected 100%%, got %.2f", s.SuccessRate)
	}
	if s.Breaching {
		t.Error("should not be breaching")
	}
}

func TestSLO_BreachingWhenBelowTarget(t *testing.T) {
	policy := SLOPolicy{TargetPercent: 99.0, Window: 7 * 24 * time.Hour}
	m := NewSLOManager(policy)
	now := time.Now()
	m.clock = fixedSLOClock(now)
	for i := 0; i < 90; i++ {
		m.Record("sync", true)
	}
	for i := 0; i < 10; i++ {
		m.Record("sync", false)
	}
	s := m.Status("sync")
	if !s.Breaching {
		t.Errorf("expected SLO breach at %.2f%% (target %.2f%%)", s.SuccessRate, s.Target)
	}
	if s.FailedEvents != 10 {
		t.Errorf("expected 10 failures, got %d", s.FailedEvents)
	}
}

func TestSLO_NoEventsIsNotBreaching(t *testing.T) {
	m := NewSLOManager(DefaultSLOPolicy())
	s := m.Status("unknown")
	if s.Breaching {
		t.Error("no events should not be breaching")
	}
	if s.SuccessRate != 100.0 {
		t.Errorf("expected 100%% with no events, got %.2f", s.SuccessRate)
	}
}

func TestSLO_EventsPrunedOutsideWindow(t *testing.T) {
	policy := SLOPolicy{TargetPercent: 99.0, Window: time.Hour}
	m := NewSLOManager(policy)
	old := time.Now().Add(-2 * time.Hour)
	m.clock = fixedSLOClock(old)
	for i := 0; i < 5; i++ {
		m.Record("prune", false)
	}
	now := time.Now()
	m.clock = fixedSLOClock(now)
	m.Record("prune", true)
	s := m.Status("prune")
	if s.TotalEvents != 1 {
		t.Errorf("expected 1 event after pruning, got %d", s.TotalEvents)
	}
	if s.Breaching {
		t.Error("should not breach with only the recent success")
	}
}

func TestSLO_ResetClearsHistory(t *testing.T) {
	m := NewSLOManager(DefaultSLOPolicy())
	now := time.Now()
	m.clock = fixedSLOClock(now)
	for i := 0; i < 5; i++ {
		m.Record("job", false)
	}
	m.Reset("job")
	s := m.Status("job")
	if s.TotalEvents != 0 {
		t.Errorf("expected 0 events after reset, got %d", s.TotalEvents)
	}
}

func TestSLO_AllStatuses(t *testing.T) {
	m := NewSLOManager(DefaultSLOPolicy())
	now := time.Now()
	m.clock = fixedSLOClock(now)
	m.Record("jobA", true)
	m.Record("jobB", false)
	all := m.AllStatuses()
	if len(all) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(all))
	}
}
