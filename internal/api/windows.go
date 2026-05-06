package api

import (
	"encoding/json"
	"net/http"
	"time"

	"cronwatch/internal/monitor"
)

type windowRequest struct {
	Name     string   `json:"name"`
	Start    string   `json:"start"`     // RFC3339
	End      string   `json:"end"`       // RFC3339
	JobNames []string `json:"job_names"` // optional
}

type windowResponse struct {
	Name     string   `json:"name"`
	Start    string   `json:"start"`
	End      string   `json:"end"`
	JobNames []string `json:"job_names"`
}

func registerWindowRoutes(mux *http.ServeMux, wm *monitor.WindowManager) {
	mux.HandleFunc("/windows", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleAddWindow(w, r, wm)
		case http.MethodGet:
			handleListWindows(w, r, wm)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/windows/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "missing name", http.StatusBadRequest)
			return
		}
		wm.Remove(name)
		w.WriteHeader(http.StatusNoContent)
	})
}

func handleAddWindow(w http.ResponseWriter, r *http.Request, wm *monitor.WindowManager) {
	var req windowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
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
	wm.Add(monitor.WindowConfig{
		Name:     req.Name,
		Start:    start,
		End:      end,
		JobNames: req.JobNames,
	})
	w.WriteHeader(http.StatusCreated)
}

func handleListWindows(w http.ResponseWriter, _ *http.Request, wm *monitor.WindowManager) {
	active := wm.ActiveWindows()
	resp := make([]windowResponse, 0, len(active))
	for _, wc := range active {
		resp = append(resp, windowResponse{
			Name:     wc.Name,
			Start:    wc.Start.Format(time.RFC3339),
			End:      wc.End.Format(time.RFC3339),
			JobNames: wc.JobNames,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
