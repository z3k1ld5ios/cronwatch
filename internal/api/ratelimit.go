package api

import (
	"encoding/json"
	"net/http"
	"time"

	"cronwatch/internal/monitor"
)

func registerRateLimitRoutes(mux *http.ServeMux, rl *monitor.RateLimiter) {
	mux.HandleFunc("/ratelimit/stats", handleRateLimitStats(rl))
	mux.HandleFunc("/ratelimit/reset", handleResetRateLimit(rl))
}

type rateLimitStatsResponse struct {
	Job       string    `json:"job"`
	Count     int       `json:"count"`
	WindowEnd time.Time `json:"window_end"`
}

func handleRateLimitStats(rl *monitor.RateLimiter) http.HandlerFunc {
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
		count, windowEnd := rl.Stats(job)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rateLimitStatsResponse{
			Job:       job,
			Count:     count,
			WindowEnd: windowEnd,
		})
	}
}

func handleResetRateLimit(rl *monitor.RateLimiter) http.HandlerFunc {
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
		rl.Reset(job)
		w.WriteHeader(http.StatusNoContent)
	}
}
