package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/cronwatch/internal/monitor"
)

func registerBudgetRoutes(mux *http.ServeMux, bm *monitor.BudgetManager) {
	mux.HandleFunc("/budget/record", handleRecordBudget(bm))
	mux.HandleFunc("/budget/status", handleBudgetStatus(bm))
	mux.HandleFunc("/budget/list", handleListBudget(bm))
	mux.HandleFunc("/budget/reset", handleResetBudget(bm))
}

func handleRecordBudget(bm *monitor.BudgetManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Job string `json:"job"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
			http.Error(w, "missing job", http.StatusBadRequest)
			return
		}
		bm.RecordFailure(req.Job)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleBudgetStatus(bm *monitor.BudgetManager) http.HandlerFunc {
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
		s := bm.Status(job)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	}
}

func handleListBudget(bm *monitor.BudgetManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		statuses := bm.AllStatuses()
		if statuses == nil {
			statuses = []monitor.BudgetStatus{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statuses)
	}
}

func handleResetBudget(bm *monitor.BudgetManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Job string `json:"job"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
			http.Error(w, "missing job", http.StatusBadRequest)
			return
		}
		bm.Reset(req.Job)
		w.WriteHeader(http.StatusNoContent)
	}
}
