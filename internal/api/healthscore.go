package api

import (
	"encoding/json"
	"net/http"

	"github.com/example/cronwatch/internal/monitor"
)

func registerHealthScoreRoutes(mux *http.ServeMux, mgr *monitor.HealthScoreManager) {
	mux.HandleFunc("/health-scores", handleListHealthScores(mgr))
	mux.HandleFunc("/health-scores/reset", handleResetHealthScore(mgr))
}

func handleListHealthScores(mgr *monitor.HealthScoreManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job != "" {
			s := mgr.Score(job)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(s)
			return
		}
		all := mgr.All()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(all)
	}
}

func handleResetHealthScore(mgr *monitor.HealthScoreManager) http.HandlerFunc {
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
		mgr.Reset(job)
		w.WriteHeader(http.StatusNoContent)
	}
}
