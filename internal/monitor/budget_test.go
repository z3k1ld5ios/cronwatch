package monitor

import (
	"testing"
	"time"
)

func fixedBudgetClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func defaultBudgetPolicy() BudgetPolicy {
	return BudgetPolicy{
		WindowDuration: time.Hour,
		TotalAllowed:   10,
		WarnThreshold:  0.5,
		CritThreshold:  0.9,
	}
}

func TestBudget_OkWhenNoFailures(t *testing.T) {
	bm := NewBudgetManager(defaultBudgetPolicy())
	s := bm.Status("jobA")
	if s.Level != "ok" {
		t.Fatalf("expected ok, got %s", s.Level)
	}
	if s.Consumed != 0 || s.Remaining != 10 {
		t.Fatalf("unexpected counts: consumed=%d remaining=%d", s.Consumed, s.Remaining)
	}
}

func TestBudget_WarnLevel(t *testing.T) {
	now := time.Now()
	bm := NewBudgetManager(defaultBudgetPolicy())
	bm.clock = fixedBudgetClock(now)
	for i := 0; i < 5; i++ {
		bm.RecordFailure("jobA")
	}
	s := bm.Status("jobA")
	if s.Level != "warn" {
		t.Fatalf("expected warn, got %s", s.Level)
	}
	if s.Consumed != 5 {
		t.Fatalf("expected 5 consumed, got %d", s.Consumed)
	}
}

func TestBudget_CriticalLevel(t *testing.T) {
	now := time.Now()
	bm := NewBudgetManager(defaultBudgetPolicy())
	bm.clock = fixedBudgetClock(now)
	for i := 0; i < 9; i++ {
		bm.RecordFailure("jobA")
	}
	s := bm.Status("jobA")
	if s.Level != "critical" {
		t.Fatalf("expected critical, got %s", s.Level)
	}
}

func TestBudget_ExhaustedLevel(t *testing.T) {
	now := time.Now()
	bm := NewBudgetManager(defaultBudgetPolicy())
	bm.clock = fixedBudgetClock(now)
	for i := 0; i < 10; i++ {
		bm.RecordFailure("jobA")
	}
	s := bm.Status("jobA")
	if s.Level != "exhausted" {
		t.Fatalf("expected exhausted, got %s", s.Level)
	}
	if s.Remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", s.Remaining)
	}
}

func TestBudget_EventsPrunedOutsideWindow(t *testing.T) {
	now := time.Now()
	bm := NewBudgetManager(defaultBudgetPolicy())
	bm.clock = fixedBudgetClock(now.Add(-2 * time.Hour))
	for i := 0; i < 9; i++ {
		bm.RecordFailure("jobA")
	}
	bm.clock = fixedBudgetClock(now)
	s := bm.Status("jobA")
	if s.Level != "ok" {
		t.Fatalf("expected ok after pruning, got %s", s.Level)
	}
	if s.Consumed != 0 {
		t.Fatalf("expected 0 consumed after pruning, got %d", s.Consumed)
	}
}

func TestBudget_ResetClearsEvents(t *testing.T) {
	now := time.Now()
	bm := NewBudgetManager(defaultBudgetPolicy())
	bm.clock = fixedBudgetClock(now)
	for i := 0; i < 8; i++ {
		bm.RecordFailure("jobA")
	}
	bm.Reset("jobA")
	s := bm.Status("jobA")
	if s.Level != "ok" || s.Consumed != 0 {
		t.Fatalf("expected clean state after reset, got level=%s consumed=%d", s.Level, s.Consumed)
	}
}

func TestBudget_AllStatuses(t *testing.T) {
	now := time.Now()
	bm := NewBudgetManager(defaultBudgetPolicy())
	bm.clock = fixedBudgetClock(now)
	bm.RecordFailure("jobA")
	bm.RecordFailure("jobB")
	bm.RecordFailure("jobB")
	statuses := bm.AllStatuses()
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
}
