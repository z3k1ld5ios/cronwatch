package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatch/internal/monitor"
)

func registerBaselineRoutes(mux *http.ServeMux, bm *monitor.BaselineManager) {
	mux.HandleFunc("/baseline/record", handleRecordBaseline(bm))
	mux.HandleFunc("/baseline/check", handleCheckBaseline(bm))
	mux.HandleFunc("/baseline/stats", handleBaselineStats(bm))
	mux.HandleFunc("/baseline/reset", handleResetBaseline(bm))
}

func handleRecordBaseline(bm *monitor.BaselineManager) http.HandlerFunc {
	type request struct {
		Job      string  `json:"job"`
		Duration float64 `json:"duration_seconds"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" || req.Duration <= 0 {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		bm.Record(req.Job, time.Duration(req.Duration*float64(time.Second)))
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleCheckBaseline(bm *monitor.BaselineManager) http.HandlerFunc {
	type request struct {
		Job      string  `json:"job"`
		Duration float64 `json:"duration_seconds"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		res := bm.Check(req.Job, time.Duration(req.Duration*float64(time.Second)))
		if res == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "insufficient_data"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"job":              res.JobName,
			"actual_seconds":   res.Actual.Seconds(),
			"mean_seconds":     res.Mean.Seconds(),
			"stddev_seconds":   res.StdDev.Seconds(),
			"z_score":          res.ZScore,
			"anomalous":        res.Anomalous,
		})
	}
}

func handleBaselineStats(bm *monitor.BaselineManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		stats := bm.AllStats()
		out := make(map[string]map[string]float64, len(stats))
		for job, v := range stats {
			out[job] = map[string]float64{"mean_seconds": v[0], "stddev_seconds": v[1]}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}

func handleResetBaseline(bm *monitor.BaselineManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing job param", http.StatusBadRequest)
			return
		}
		bm.Reset(job)
		w.WriteHeader(http.StatusNoContent)
	}
}
