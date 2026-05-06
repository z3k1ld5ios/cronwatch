package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListJobs_EmptyTracker(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var jobs []JobSummary
	if err := json.NewDecoder(rec.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected empty list, got %d jobs", len(jobs))
	}
}

func TestListJobs_WrongMethod(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/jobs", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestListJobs_AfterHeartbeat(t *testing.T) {
	srv := newTestServer(t)

	// Record a heartbeat first
	heartbeat := httptest.NewRequest(http.MethodPost, "/heartbeat?job=backup", nil)
	srv.ServeHTTP(httptest.NewRecorder(), heartbeat)

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var jobs []JobSummary
	if err := json.NewDecoder(rec.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Name != "backup" {
		t.Errorf("expected job name 'backup', got %q", jobs[0].Name)
	}
	if jobs[0].LastSeen == nil {
		t.Error("expected LastSeen to be set")
	}
}

func TestResetJob_Success(t *testing.T) {
	srv := newTestServer(t)

	// Seed a heartbeat
	heartbeat := httptest.NewRequest(http.MethodPost, "/heartbeat?job=cleanup", nil)
	srv.ServeHTTP(httptest.NewRecorder(), heartbeat)

	req := httptest.NewRequest(http.MethodPost, "/jobs/reset?job=cleanup", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestResetJob_MissingParam(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/jobs/reset", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestResetJob_WrongMethod(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/jobs/reset?job=cleanup", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
