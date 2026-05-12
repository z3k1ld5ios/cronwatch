package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatch/internal/monitor"
)

// registerBurnRateRoutes attaches burn rate endpoints to the given mux.
func registerBurnRateRoutes(mux *http.ServeMux, mgr *monitor.BurnRateManager) {
	mux.HandleFunc("/burnrate/record", handleRecordBurnRate(mgr))
	mux.HandleFunc("/burnrate/stats", handleBurnRateStats(mgr))
	mux.HandleFunc("/burnrate/reset", handleResetBurnRate(mgr))
}

// handleRecordBurnRate records a burn rate event (missed run or alert) for a job.
//
// POST /burnrate/record
// Body: { "job": "name", "kind": "missed|drift" }
func handleRecordBurnRate(mgr *monitor.BurnRateManager) http.HandlerFunc {
	type request struct {
		Job  string `json:"job"`
		Kind string `json:"kind"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Job == "" {
			http.Error(w, "missing field: job", http.StatusBadRequest)
			return
		}
		if req.Kind == "" {
			http.Error(w, "missing field: kind", http.StatusBadRequest)
			return
		}
		mgr.Record(req.Job, req.Kind, time.Now())
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleBurnRateStats returns the current burn rate level and event count for a job.
//
// GET /burnrate/stats?job=<name>
func handleBurnRateStats(mgr *monitor.BurnRateManager) http.HandlerFunc {
	type response struct {
		Job        string  `json:"job"`
		Level      string  `json:"level"`
		Rate       float64 `json:"rate_per_hour"`
		EventCount int     `json:"event_count"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing query param: job", http.StatusBadRequest)
			return
		}
		level, rate, count := mgr.Stats(job, time.Now())
		resp := response{
			Job:        job,
			Level:      level,
			Rate:       rate,
			EventCount: count,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// handleResetBurnRate clears all recorded events for a job.
//
// POST /burnrate/reset
// Body: { "job": "name" }
func handleResetBurnRate(mgr *monitor.BurnRateManager) http.HandlerFunc {
	type request struct {
		Job string `json:"job"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Job == "" {
			http.Error(w, "missing field: job", http.StatusBadRequest)
			return
		}
		mgr.Reset(req.Job)
		w.WriteHeader(http.StatusNoContent)
	}
}
