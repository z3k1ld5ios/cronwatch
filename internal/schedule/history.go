package schedule

import (
	"sync"
	"time"
)

// RunRecord stores metadata about a single cron job execution.
type RunRecord struct {
	JobName   string
	Scheduled time.Time
	Actual    time.Time
	Drift     time.Duration
}

// History maintains a bounded ring-buffer of recent run records per job.
type History struct {
	mu      sync.RWMutex
	records map[string][]RunRecord
	maxSize int
}

// NewHistory creates a History that retains up to maxSize records per job.
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &History{
		records: make(map[string][]RunRecord),
		maxSize: maxSize,
	}
}

// Record appends a new run record for the given job.
func (h *History) Record(jobName string, scheduled, actual time.Time) RunRecord {
	rr := RunRecord{
		JobName:   jobName,
		Scheduled: scheduled,
		Actual:    actual,
		Drift:     actual.Sub(scheduled),
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	buf := h.records[jobName]
	buf = append(buf, rr)
	if len(buf) > h.maxSize {
		buf = buf[len(buf)-h.maxSize:]
	}
	h.records[jobName] = buf
	return rr
}

// Latest returns the most recent RunRecord for jobName, and false if none exist.
func (h *History) Latest(jobName string) (RunRecord, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	buf := h.records[jobName]
	if len(buf) == 0 {
		return RunRecord{}, false
	}
	return buf[len(buf)-1], true
}

// All returns a copy of all run records for jobName.
func (h *History) All(jobName string) []RunRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()
	buf := h.records[jobName]
	out := make([]RunRecord, len(buf))
	copy(out, buf)
	return out
}
