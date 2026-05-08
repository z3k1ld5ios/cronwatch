package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/monitor"
)

func newHealthScoreServer() (*monitor.HealthScoreManager, *http.ServeMux) {
	fixed := func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }
	mgr := monitor.NewHealthScoreManager(fixed)
	mux := http.NewServeMux()
	registerHealthScoreRoutes(mux, mgr)
	return mgr, mux
}

func TestListHealthScores_Empty(t *testing.T) {
	_, mux := newHealthScoreServer()
	req := httptest.NewRequest(http.MethodGet, "/health-scores", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListHealthScores_SingleJob(t *testing.T) {
	mgr, mux := newHealthScoreServer()
	mgr.RecordRun("backup", false, false)
	req := httptest.NewRequest(http.MethodGet, "/health-scores?job=backup", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var s monitor.HealthScore
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatal(err)
	}
	if s.JobName != "backup" {
		t.Errorf("expected job backup, got %s", s.JobName)
	}
	if s.Score != 100 {
		t.Errorf("expected score 100, got %d", s.Score)
	}
}

func TestListHealthScores_WrongMethod(t *testing.T) {
	_, mux := newHealthScoreServer()
	req := httptest.NewRequest(http.MethodPost, "/health-scores", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestResetHealthScore_Success(t *testing.T) {
	mgr, mux := newHealthScoreServer()
	mgr.RecordRun("sync", true, true)
	req := httptest.NewRequest(http.MethodPost, "/health-scores/reset?job=sync", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	s := mgr.Score("sync")
	if s.Score != 100 {
		t.Errorf("expected 100 after reset, got %d", s.Score)
	}
}

func TestResetHealthScore_MissingJob(t *testing.T) {
	_, mux := newHealthScoreServer()
	req := httptest.NewRequest(http.MethodPost, "/health-scores/reset", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestResetHealthScore_WrongMethod(t *testing.T) {
	_, mux := newHealthScoreServer()
	req := httptest.NewRequest(http.MethodGet, "/health-scores/reset?job=x", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
