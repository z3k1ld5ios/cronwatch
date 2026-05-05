package schedule

import (
	"testing"
	"time"
)

func TestHistory_RecordAndLatest(t *testing.T) {
	h := NewHistory(10)
	scheduled := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	actual := scheduled.Add(3 * time.Second)

	rr := h.Record("backup", scheduled, actual)
	if rr.Drift != 3*time.Second {
		t.Errorf("expected drift 3s, got %v", rr.Drift)
	}

	latest, ok := h.Latest("backup")
	if !ok {
		t.Fatal("expected a latest record")
	}
	if latest.Drift != 3*time.Second {
		t.Errorf("latest drift mismatch: got %v", latest.Drift)
	}
}

func TestHistory_LatestMissing(t *testing.T) {
	h := NewHistory(10)
	_, ok := h.Latest("nonexistent")
	if ok {
		t.Error("expected no record for unknown job")
	}
}

func TestHistory_BoundedSize(t *testing.T) {
	h := NewHistory(3)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		h.Record("job", base.Add(time.Duration(i)*time.Minute), base.Add(time.Duration(i)*time.Minute))
	}
	all := h.All("job")
	if len(all) != 3 {
		t.Errorf("expected 3 records (ring buffer), got %d", len(all))
	}
	// Oldest should be the 3rd run (index 2 of original)
	expected := base.Add(2 * time.Minute)
	if !all[0].Scheduled.Equal(expected) {
		t.Errorf("expected oldest scheduled %v, got %v", expected, all[0].Scheduled)
	}
}

func TestHistory_AllReturnsCopy(t *testing.T) {
	h := NewHistory(10)
	base := time.Now().UTC()
	h.Record("job", base, base)
	all := h.All("job")
	all[0].JobName = "mutated"

	latest, _ := h.Latest("job")
	if latest.JobName == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}

func TestHistory_DefaultMaxSize(t *testing.T) {
	h := NewHistory(0)
	if h.maxSize != 100 {
		t.Errorf("expected default maxSize 100, got %d", h.maxSize)
	}
}
