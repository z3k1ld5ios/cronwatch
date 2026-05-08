package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatch/internal/monitor"
)

type trendResponse struct {
	Job       string `json:"job"`
	Direction string `json:"direction"`
	Slope     float64 `json:"slope_seconds_per_observation"`
	Samples   int    `json:"samples"`
}

func registerTrendRoutes(mux *http.ServeMux, analyzer *monitor.TrendAnalyzer) {
	mux.HandleFunc("/trends", handleListTrends(analyzer))
	mux.HandleFunc("/trends/job", handleJobTrend(analyzer))
}

func handleListTrends(analyzer *monitor.TrendAnalyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		trends := analyzer.AllTrends()
		out := make([]trendResponse, 0, len(trends))
		for _, tr := range trends {
			out = append(out, trendResponse{
				Job:       tr.Job,
				Direction: string(tr.Direction),
				Slope:     tr.Slope,
				Samples:   tr.Samples,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}

func handleJobTrend(analyzer *monitor.TrendAnalyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing job parameter", http.StatusBadRequest)
			return
		}
		tr := analyzer.Analyze(job)
		out := trendResponse{
			Job:       tr.Job,
			Direction: string(tr.Direction),
			Slope:     tr.Slope,
			Samples:   tr.Samples,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}
