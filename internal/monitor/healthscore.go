package monitor

import (
	"sync"
	"time"
)

// HealthScore represents a computed reliability score for a job (0–100).
type HealthScore struct {
	JobName   string    `json:"job_name"`
	Score     int       `json:"score"`
	Missed    int       `json:"missed"`
	Drifted   int       `json:"drifted"`
	Total     int       `json:"total"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HealthScoreManager tracks per-job health scores based on missed and drifted events.
type HealthScoreManager struct {
	mu      sync.RWMutex
	records map[string]*healthRecord
	clock   func() time.Time
}

type healthRecord struct {
	missed  int
	drifted int
	total   int
}

// NewHealthScoreManager creates a new HealthScoreManager.
func NewHealthScoreManager(clock func() time.Time) *HealthScoreManager {
	if clock == nil {
		clock = time.Now
	}
	return &HealthScoreManager{
		records: make(map[string]*healthRecord),
		clock:   clock,
	}
}

// RecordRun records a job execution outcome.
func (h *HealthScoreManager) RecordRun(job string, missed bool, drifted bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	r := h.getOrCreate(job)
	r.total++
	if missed {
		r.missed++
	}
	if drifted {
		r.drifted++
	}
}

// Score returns the current HealthScore for a job.
func (h *HealthScoreManager) Score(job string) HealthScore {
	h.mu.RLock()
	defer h.mu.RUnlock()
	r, ok := h.records[job]
	if !ok {
		return HealthScore{JobName: job, Score: 100, UpdatedAt: h.clock()}
	}
	return HealthScore{
		JobName:   job,
		Score:     computeScore(r),
		Missed:    r.missed,
		Drifted:   r.drifted,
		Total:     r.total,
		UpdatedAt: h.clock(),
	}
}

// Reset clears the health record for a job.
func (h *HealthScoreManager) Reset(job string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.records, job)
}

// All returns health scores for every tracked job.
func (h *HealthScoreManager) All() []HealthScore {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]HealthScore, 0, len(h.records))
	for job, r := range h.records {
		out = append(out, HealthScore{
			JobName:   job,
			Score:     computeScore(r),
			Missed:    r.missed,
			Drifted:   r.drifted,
			Total:     r.total,
			UpdatedAt: h.clock(),
		})
	}
	return out
}

func (h *HealthScoreManager) getOrCreate(job string) *healthRecord {
	if _, ok := h.records[job]; !ok {
		h.records[job] = &healthRecord{}
	}
	return h.records[job]
}

func computeScore(r *healthRecord) int {
	if r.total == 0 {
		return 100
	}
	penalty := (r.missed*2 + r.drifted) * 100 / (r.total * 3)
	score := 100 - penalty
	if score < 0 {
		return 0
	}
	return score
}
