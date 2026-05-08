package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/monitor"
)

func TestHealthScoreRoundtrip_RecordAndQuery(t *testing.T) {
	fixed := func() time.Time { return time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC) }
	mgr := monitor.NewHealthScoreManager(fixed)
	mux := http.NewServeMux()
	registerHealthScoreRoutes(mux, mgr)

	mgr.RecordRun("nightly", false, false)
	mgr.RecordRun("nightly", false, false)
	mgr.RecordRun("nightly", true, false)

	req := httptest.NewRequest(http.MethodGet, "/health-scores?job=nightly", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var s monitor.HealthScore
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatal(err)
	}
	if s.Total != 3 {
		t.Errorf("expected total 3, got %d", s.Total)
	}
	if s.Missed != 1 {
		t.Errorf("expected 1 missed, got %d", s.Missed)
	}
	if s.Score == 100 {
		t.Error("expected score below 100 after a missed run")
	}
}

func TestHealthScoreRoundtrip_AllJobsListed(t *testing.T) {
	fixed := func() time.Time { return time.Now() }
	mgr := monitor.NewHealthScoreManager(fixed)
	mux := http.NewServeMux()
	registerHealthScoreRoutes(mux, mgr)

	mgr.RecordRun("jobA", false, false)
	mgr.RecordRun("jobB", true, false)

	req := httptest.NewRequest(http.MethodGet, "/health-scores", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var scores []monitor.HealthScore
	if err := json.NewDecoder(rec.Body).Decode(&scores); err != nil {
		t.Fatal(err)
	}
	if len(scores) != 2 {
		t.Errorf("expected 2 scores, got %d", len(scores))
	}
}
