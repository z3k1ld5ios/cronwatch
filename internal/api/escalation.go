package api

import (
	"encoding/json"
	"net/http"

	"github.com/example/cronwatch/internal/monitor"
)

type escalationResponse struct {
	Job   string `json:"job"`
	Level int    `json:"level"`
	Label string `json:"label"`
}

func levelLabel(l monitor.EscalationLevel) string {
	switch l {
	case monitor.LevelWarning:
		return "warning"
	case monitor.LevelCritical:
		return "critical"
	default:
		return "none"
	}
}

func registerEscalationRoutes(mux *http.ServeMux, em *monitor.EscalationManager) {
	mux.HandleFunc("/escalation", func(w http.ResponseWriter, r *http.Request) {
		handleGetEscalation(w, r, em)
	})
	mux.HandleFunc("/escalation/reset", func(w http.ResponseWriter, r *http.Request) {
		handleResetEscalation(w, r, em)
	})
}

func handleGetEscalation(w http.ResponseWriter, r *http.Request, em *monitor.EscalationManager) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	lvl := em.Level(job)
	resp := escalationResponse{Job: job, Level: int(lvl), Label: levelLabel(lvl)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleResetEscalation(w http.ResponseWriter, r *http.Request, em *monitor.EscalationManager) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	em.Reset(job)
	w.WriteHeader(http.StatusNoContent)
}
