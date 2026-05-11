package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatch/internal/monitor"
)

func registerJitterRoutes(mux *http.ServeMux, analyzer *monitor.JitterAnalyzer) {
	mux.HandleFunc("/jitter/record", handleRecordJitter(analyzer))
	mux.HandleFunc("/jitter/stats", handleJitterStats(analyzer))
	mux.HandleFunc("/jitter/reset", handleResetJitter(analyzer))
}

func handleRecordJitter(a *monitor.JitterAnalyzer) http.HandlerFunc {
	type request struct {
		Job string `json:"job"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
			http.Error(w, "missing job", http.StatusBadRequest)
			return
		}
		a.Record(req.Job, time.Now())
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleJitterStats(a *monitor.JitterAnalyzer) http.HandlerFunc {
	type response struct {
		Job    string  `json:"job"`
		MeanS  float64 `json:"mean_seconds"`
		StdDevS float64 `json:"stddev_seconds"`
		CV     float64 `json:"cv"`
		High   bool    `json:"high_jitter"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing job", http.StatusBadRequest)
			return
		}
		res := a.Analyze(job)
		if res == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"job": job, "available": false})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{
			Job:     res.Job,
			MeanS:   res.Mean.Seconds(),
			StdDevS: res.StdDev.Seconds(),
			CV:      res.CV,
			High:    res.High,
		})
	}
}

func handleResetJitter(a *monitor.JitterAnalyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing job", http.StatusBadRequest)
			return
		}
		a.Reset(job)
		w.WriteHeader(http.StatusNoContent)
	}
}
