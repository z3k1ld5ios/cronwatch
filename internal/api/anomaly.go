package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/cronwatch/internal/monitor"
)

func registerAnomalyRoutes(mux *http.ServeMux, detector *monitor.AnomalyDetector) {
	mux.HandleFunc("/anomaly/stats", handleAnomalyStats(detector))
	mux.HandleFunc("/anomaly/reset", handleResetAnomaly(detector))
	mux.HandleFunc("/anomaly/record", handleRecordAnomaly(detector))
}

func handleAnomalyStats(detector *monitor.AnomalyDetector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		type statEntry struct {
			Job    string  `json:"job"`
			Mean   float64 `json:"mean_seconds"`
			StdDev float64 `json:"stddev_seconds"`
		}
		all := detector.AllStats()
		out := make([]statEntry, 0, len(all))
		for job, s := range all {
			out = append(out, statEntry{Job: job, Mean: s[0], StdDev: s[1]})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}

func handleResetAnomaly(detector *monitor.AnomalyDetector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing job parameter", http.StatusBadRequest)
			return
		}
		detector.Reset(job)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleRecordAnomaly(detector *monitor.AnomalyDetector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Job          string  `json:"job"`
			DurationSecs float64 `json:"duration_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		result := detector.Record(req.Job, time.Duration(req.DurationSecs*float64(time.Second)), time.Now())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
