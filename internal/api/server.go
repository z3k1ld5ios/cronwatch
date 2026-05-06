package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cronwatch/internal/schedule"
)

// Server exposes an HTTP API for receiving heartbeats from cron jobs
// and querying job status.
type Server struct {
	tracker  *schedule.Tracker
	history  *schedule.History
	mux      *http.ServeMux
	addr     string
}

// NewServer creates a new API server bound to addr.
func NewServer(addr string, tracker *schedule.Tracker, history *schedule.History) *Server {
	s := &Server{
		tracker: tracker,
		history: history,
		mux:     http.NewServeMux(),
		addr:    addr,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/heartbeat", s.handleHeartbeat)
	s.mux.HandleFunc("/status", s.handleStatus)
}

// Start begins listening and serving HTTP requests.
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return srv.ListenAndServe()
}

// HeartbeatRequest is the payload sent by a cron job on completion.
type HeartbeatRequest struct {
	JobName   string    `json:"job_name"`
	Timestamp time.Time `json:"timestamp"`
}

// StatusResponse summarises the last recorded run for each tracked job.
type StatusResponse struct {
	Jobs []JobStatus `json:"jobs"`
}

// JobStatus holds the latest run info for a single job.
type JobStatus struct {
	JobName   string     `json:"job_name"`
	LastSeen  *time.Time `json:"last_seen,omitempty"`
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid body: %v", err), http.StatusBadRequest)
		return
	}
	if req.JobName == "" {
		http.Error(w, "job_name is required", http.StatusBadRequest)
		return
	}
	t := req.Timestamp
	if t.IsZero() {
		t = time.Now().UTC()
	}
	s.history.Record(req.JobName, t)
	s.tracker.Saw(req.JobName, t)
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	jobs := s.tracker.Jobs()
	resp := StatusResponse{Jobs: make([]JobStatus, 0, len(jobs))}
	for _, name := range jobs {
		js := JobStatus{JobName: name}
		if latest, ok := s.history.Latest(name); ok {
			t := latest
			js.LastSeen = &t
		}
		resp.Jobs = append(resp.Jobs, js)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
