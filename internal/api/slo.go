package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatch/internal/monitor"
)

func registerSLORoutes(mux *http.ServeMux, mgr *monitor.SLOManager) {
	mux.HandleFunc("/slo/record", handleRecordSLO(mgr))
	mux.HandleFunc("/slo/status", handleSLOStatus(mgr))
	mux.HandleFunc("/slo/list", handleListSLO(mgr))
	mux.HandleFunc("/slo/reset", handleResetSLO(mgr))
}

func handleRecordSLO(mgr *monitor.SLOManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Job     string `json:"job"`
			Success bool   `json:"success"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
			http.Error(w, "invalid request: job required", http.StatusBadRequest)
			return
		}
		mgr.Record(req.Job, req.Success)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleSLOStatus(mgr *monitor.SLOManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing job param", http.StatusBadRequest)
			return
		}
		status := mgr.Status(job)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}

func handleListSLO(mgr *monitor.SLOManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		all := mgr.AllStatuses()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(all)
	}
}

func handleResetSLO(mgr *monitor.SLOManager) http.HandlerFunc {
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
		mgr.Reset(job)
		w.WriteHeader(http.StatusNoContent)
	}
}
