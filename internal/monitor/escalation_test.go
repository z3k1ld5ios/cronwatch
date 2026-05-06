package monitor

import (
	"testing"
	"time"
)

func defaultEscalationPolicy() EscalationPolicy {
	return EscalationPolicy{
		WarningAfter:  2,
		CriticalAfter: 4,
		ResetAfter:    10 * time.Minute,
	}
}

func TestEscalation_LevelNoneInitially(t *testing.T) {
	em := NewEscalationManager(defaultEscalationPolicy())
	if got := em.Level("job1"); got != LevelNone {
		t.Fatalf("expected LevelNone, got %d", got)
	}
}

func TestEscalation_WarningAfterThreshold(t *testing.T) {
	em := NewEscalationManager(defaultEscalationPolicy())
	em.Record("job1") // count=1 → none
	lvl := em.Record("job1") // count=2 → warning
	if lvl != LevelWarning {
		t.Fatalf("expected LevelWarning after 2 failures, got %d", lvl)
	}
}

func TestEscalation_CriticalAfterThreshold(t *testing.T) {
	em := NewEscalationManager(defaultEscalationPolicy())
	for i := 0; i < 4; i++ {
		em.Record("job1")
	}
	if got := em.Level("job1"); got != LevelCritical {
		t.Fatalf("expected LevelCritical after 4 failures, got %d", got)
	}
}

func TestEscalation_ResetClearsState(t *testing.T) {
	em := NewEscalationManager(defaultEscalationPolicy())
	for i := 0; i < 4; i++ {
		em.Record("job1")
	}
	em.Reset("job1")
	if got := em.Level("job1"); got != LevelNone {
		t.Fatalf("expected LevelNone after reset, got %d", got)
	}
}

func TestEscalation_ResetsAfterIdlePeriod(t *testing.T) {
	policy := defaultEscalationPolicy()
	policy.ResetAfter = 5 * time.Minute
	em := NewEscalationManager(policy)

	now := time.Now()
	em.clock = func() time.Time { return now }
	em.Record("job1")
	em.Record("job1") // warning

	// Advance clock past reset window
	em.clock = func() time.Time { return now.Add(6 * time.Minute) }
	lvl := em.Record("job1") // should reset count to 1 → none
	if lvl != LevelNone {
		t.Fatalf("expected LevelNone after idle reset, got %d", lvl)
	}
}

func TestEscalation_IndependentJobs(t *testing.T) {
	em := NewEscalationManager(defaultEscalationPolicy())
	for i := 0; i < 4; i++ {
		em.Record("jobA")
	}
	if got := em.Level("jobB"); got != LevelNone {
		t.Fatalf("jobB should be unaffected, got %d", got)
	}
}
