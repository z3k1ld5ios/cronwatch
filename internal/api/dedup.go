package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatch/internal/monitor"
)

// registerDedupRoutes attaches deduplication management endpoints to mux.
func registerDedupRoutes(mux *http.ServeMux, dm *monitor.DedupManager) {
	mux.HandleFunc("/dedup/reset", handleResetDedup(dm))
	mux.HandleFunc("/dedup/stats", handleDedupStats(dm))
}

type resetDedupRequest struct {
	JobName string `json:"job_name"`
	Kind    string `json:"kind"`
}

func handleResetDedup(dm *monitor.DedupManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req resetDedupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.JobName == "" || req.Kind == "" {
			http.Error(w, "job_name and kind are required", http.StatusBadRequest)
			return
		}
		dm.Reset(req.JobName, req.Kind)
		w.WriteHeader(http.StatusNoContent)
	}
}

type dedupStatsResponse struct {
	TrackedEntries int `json:"tracked_entries"`
}

func handleDedupStats(dm *monitor.DedupManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp := dedupStatsResponse{TrackedEntries: dm.Len()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
