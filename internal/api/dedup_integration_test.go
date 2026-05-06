package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func TestDedupRoundtrip_ResetClearsSuppression(t *testing.T) {
	dm := monitor.NewDedupManager(5*time.Minute, nil)
	mux := http.NewServeMux()
	registerDedupRoutes(mux, dm)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Record a fingerprint for job "sync" so it appears in stats
	const job = "sync"
	const fp = "abc123"
	if !dm.IsNew(job, fp) {
		t.Fatal("expected first occurrence to be new")
	}
	if dm.IsNew(job, fp) {
		t.Fatal("expected second occurrence to be duplicate")
	}

	// Verify stats reflect the entry
	resp, err := http.Get(srv.URL + "/dedup/stats")
	if err != nil {
		t.Fatalf("stats request failed: %v", err)
	}
	defer resp.Body.Close()

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode stats: %v", err)
	}
	entries, ok := stats["entries"].(float64)
	if !ok || entries < 1 {
		t.Errorf("expected at least 1 entry in stats, got %v", stats["entries"])
	}

	// Reset dedup state for the job
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/dedup/reset?job="+job, nil)
	resetResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("reset request failed: %v", err)
	}
	defer resetResp.Body.Close()
	if resetResp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204 after reset, got %d", resetResp.StatusCode)
	}

	// After reset, the same fingerprint should be treated as new again
	if !dm.IsNew(job, fp) {
		t.Error("expected fingerprint to be new after reset")
	}
}
