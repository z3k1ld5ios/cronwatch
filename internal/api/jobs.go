package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// JobSummary represents a snapshot of a monitored job's state.
type JobSummary struct {
	Name        string     `json:"name"`
	Schedule    string     `json:"schedule"`
	LastSeen    *time.Time `json:"last_seen,omitempty"`
	NextExpected *time.Time `json:"next_expected,omitempty"`
	Missed      bool       `json:"missed"`
	DriftMs     *int64     `json:"drift_ms,omitempty"`
}

// registerJobRoutes wires job-related endpoints onto the given mux.
func registerJobRoutes(mux *http.ServeMux, s *Server) {
	mux.HandleFunc("/jobs", s.handleListJobs)
	mux.HandleFunc("/jobs/reset", s.handleResetJob)
}

// handleListJobs returns a JSON array of all tracked job summaries.
func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs := s.tracker.Jobs()
	summaries := make([]JobSummary, 0, len(jobs))

	for _, name := range jobs {
		sum := JobSummary{Name: name}

		if cfg, ok := s.config.JobByName(name); ok {
			sum.Schedule = cfg.Schedule
		}

		if latest, ok := s.tracker.Latest(name); ok {
			t := latest
			sum.LastSeen = &t
		}

		if next, ok := s.tracker.NextExpected(name); ok {
			n := next
			sum.NextExpected = &n
			sum.Missed = time.Now().After(next)
		}

		if drift, ok := s.tracker.LastDrift(name); ok {
			ms := drift.Milliseconds()
			sum.DriftMs = &ms
		}

		summaries = append(summaries, sum)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// handleResetJob clears the tracking history for a specific job.
func (s *Server) handleResetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("job")
	if name == "" {
		http.Error(w, "missing job query parameter", http.StatusBadRequest)
		return
	}

	s.tracker.Reset(name)
	w.WriteHeader(http.StatusNoContent)
}
