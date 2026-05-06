package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/cronwatch/internal/monitor"
)

type silenceRequest struct {
	Label    string   `json:"label"`
	JobNames []string `json:"job_names"`
	Start    string   `json:"start"`
	End      string   `json:"end"`
	Reason   string   `json:"reason"`
}

func registerSilenceRoutes(mux *http.ServeMux, sm *monitor.SilenceManager) {
	mux.HandleFunc("/silences", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleAddSilence(w, r, sm)
		case http.MethodGet:
			handleListSilences(w, r, sm)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/silences/delete", func(w http.ResponseWriter, r *http.Request) {
		handleDeleteSilence(w, r, sm)
	})
}

func handleAddSilence(w http.ResponseWriter, r *http.Request, sm *monitor.SilenceManager) {
	var req silenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Label == "" {
		http.Error(w, "label is required", http.StatusBadRequest)
		return
	}
	start, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		http.Error(w, "invalid start time", http.StatusBadRequest)
		return
	}
	end, err := time.Parse(time.RFC3339, req.End)
	if err != nil {
		http.Error(w, "invalid end time", http.StatusBadRequest)
		return
	}
	if !end.After(start) {
		http.Error(w, "end must be after start", http.StatusBadRequest)
		return
	}
	s := monitor.Silence{
		Label:    req.Label,
		JobNames: req.JobNames,
		Start:    start,
		End:      end,
		Reason:   req.Reason,
	}
	if !sm.Add(s) {
		http.Error(w, "silence with that label already exists", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func handleListSilences(w http.ResponseWriter, _ *http.Request, sm *monitor.SilenceManager) {
	list := sm.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func handleDeleteSilence(w http.ResponseWriter, r *http.Request, sm *monitor.SilenceManager) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	label := r.URL.Query().Get("label")
	if label == "" {
		http.Error(w, "label query param required", http.StatusBadRequest)
		return
	}
	sm.Remove(label)
	w.WriteHeader(http.StatusNoContent)
}
