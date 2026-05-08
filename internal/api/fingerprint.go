package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatch/cronwatch/internal/monitor"
)

// fingerprintRequest is the payload for POST /fingerprint.
type fingerprintRequest struct {
	JobName  string `json:"job_name"`
	Kind     string `json:"kind"`
	Schedule string `json:"schedule"`
	At       string `json:"at"` // RFC3339
	BucketS  int    `json:"bucket_seconds"`
}

// fingerprintResponse is returned by POST /fingerprint.
type fingerprintResponse struct {
	Fingerprint string `json:"fingerprint"`
	BucketedAt  string `json:"bucketed_at"`
}

func registerFingerprintRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/fingerprint", handleComputeFingerprint)
}

func handleComputeFingerprint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req fingerprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.JobName == "" || req.Kind == "" || req.Schedule == "" {
		http.Error(w, "job_name, kind, and schedule are required", http.StatusBadRequest)
		return
	}

	at := time.Now().UTC()
	if req.At != "" {
		parsed, err := time.Parse(time.RFC3339, req.At)
		if err != nil {
			http.Error(w, "invalid at timestamp (use RFC3339)", http.StatusBadRequest)
			return
		}
		at = parsed
	}

	bucket := 5 * time.Minute
	if req.BucketS > 0 {
		bucket = time.Duration(req.BucketS) * time.Second
	}

	fp := monitor.FingerprintFor(req.JobName, req.Schedule, monitor.AlertKind(req.Kind), at, bucket)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(fingerprintResponse{
		Fingerprint: fp,
		BucketedAt:  at.Truncate(bucket).Format(time.RFC3339),
	})
}
