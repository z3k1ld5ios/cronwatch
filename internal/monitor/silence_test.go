package monitor

import (
	"testing"
	"time"
)

func fixedSilenceClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSilence_AllowsOutsideWindow(t *testing.T) {
	now := time.Now()
	sm := NewSilenceManager(fixedSilenceClock(now))
	sm.Add(Silence{
		Label: "maint",
		Start: now.Add(1 * time.Hour),
		End:   now.Add(2 * time.Hour),
	})
	if sm.IsSilenced("myjob") {
		t.Fatal("expected job not silenced outside window")
	}
}

func TestSilence_SilencesAllJobsWhenNoJobFilter(t *testing.T) {
	now := time.Now()
	sm := NewSilenceManager(fixedSilenceClock(now))
	sm.Add(Silence{
		Label: "global",
		Start: now.Add(-1 * time.Minute),
		End:   now.Add(1 * time.Hour),
	})
	if !sm.IsSilenced("anyjob") {
		t.Fatal("expected all jobs silenced during global window")
	}
}

func TestSilence_SilencesSpecificJob(t *testing.T) {
	now := time.Now()
	sm := NewSilenceManager(fixedSilenceClock(now))
	sm.Add(Silence{
		Label:    "targeted",
		JobNames: []string{"backup"},
		Start:    now.Add(-1 * time.Minute),
		End:      now.Add(1 * time.Hour),
	})
	if !sm.IsSilenced("backup") {
		t.Fatal("expected backup to be silenced")
	}
	if sm.IsSilenced("deploy") {
		t.Fatal("expected deploy not to be silenced")
	}
}

func TestSilence_RemoveClearsEntry(t *testing.T) {
	now := time.Now()
	sm := NewSilenceManager(fixedSilenceClock(now))
	sm.Add(Silence{
		Label: "temp",
		Start: now.Add(-1 * time.Minute),
		End:   now.Add(1 * time.Hour),
	})
	sm.Remove("temp")
	if sm.IsSilenced("anyjob") {
		t.Fatal("expected silence cleared after remove")
	}
}

func TestSilence_DuplicateLabelRejected(t *testing.T) {
	now := time.Now()
	sm := NewSilenceManager(fixedSilenceClock(now))
	s := Silence{Label: "dup", Start: now, End: now.Add(time.Hour)}
	if !sm.Add(s) {
		t.Fatal("first add should succeed")
	}
	if sm.Add(s) {
		t.Fatal("duplicate add should fail")
	}
}

func TestSilence_ListReturnsCopy(t *testing.T) {
	now := time.Now()
	sm := NewSilenceManager(fixedSilenceClock(now))
	sm.Add(Silence{Label: "s1", Start: now, End: now.Add(time.Hour)})
	sm.Add(Silence{Label: "s2", Start: now, End: now.Add(time.Hour)})
	list := sm.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 silences, got %d", len(list))
	}
}
