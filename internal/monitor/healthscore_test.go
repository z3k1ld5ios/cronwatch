package monitor

import (
	"testing"
	"time"
)

func fixedHealthClock() func() time.Time {
	t := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

func TestHealthScore_DefaultIs100(t *testing.T) {
	m := NewHealthScoreManager(fixedHealthClock())
	s := m.Score("backup")
	if s.Score != 100 {
		t.Errorf("expected 100, got %d", s.Score)
	}
}

func TestHealthScore_PerfectRuns(t *testing.T) {
	m := NewHealthScoreManager(fixedHealthClock())
	for i := 0; i < 10; i++ {
		m.RecordRun("backup", false, false)
	}
	s := m.Score("backup")
	if s.Score != 100 {
		t.Errorf("expected 100, got %d", s.Score)
	}
	if s.Total != 10 {
		t.Errorf("expected total 10, got %d", s.Total)
	}
}

func TestHealthScore_MissedRunsReduceScore(t *testing.T) {
	m := NewHealthScoreManager(fixedHealthClock())
	m.RecordRun("sync", true, false)
	m.RecordRun("sync", true, false)
	m.RecordRun("sync", false, false)
	s := m.Score("sync")
	if s.Score >= 100 {
		t.Errorf("expected score < 100, got %d", s.Score)
	}
	if s.Missed != 2 {
		t.Errorf("expected 2 missed, got %d", s.Missed)
	}
}

func TestHealthScore_DriftReducesScore(t *testing.T) {
	m := NewHealthScoreManager(fixedHealthClock())
	for i := 0; i < 5; i++ {
		m.RecordRun("report", false, true)
	}
	s := m.Score("report")
	if s.Score >= 100 {
		t.Errorf("expected score < 100, got %d", s.Score)
	}
	if s.Drifted != 5 {
		t.Errorf("expected 5 drifted, got %d", s.Drifted)
	}
}

func TestHealthScore_ResetRestores100(t *testing.T) {
	m := NewHealthScoreManager(fixedHealthClock())
	m.RecordRun("cleanup", true, true)
	m.Reset("cleanup")
	s := m.Score("cleanup")
	if s.Score != 100 {
		t.Errorf("expected 100 after reset, got %d", s.Score)
	}
}

func TestHealthScore_AllReturnsAllJobs(t *testing.T) {
	m := NewHealthScoreManager(fixedHealthClock())
	m.RecordRun("a", false, false)
	m.RecordRun("b", true, false)
	all := m.All()
	if len(all) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(all))
	}
}

func TestHealthScore_ScoreFloorIsZero(t *testing.T) {
	m := NewHealthScoreManager(fixedHealthClock())
	for i := 0; i < 20; i++ {
		m.RecordRun("bad", true, true)
	}
	s := m.Score("bad")
	if s.Score < 0 {
		t.Errorf("score should not be negative, got %d", s.Score)
	}
}
