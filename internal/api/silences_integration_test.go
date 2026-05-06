package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/monitor"
)

// TestSilenceRoundtrip exercises add → list → delete as a full lifecycle.
func TestSilenceRoundtrip(t *testing.T) {
	sm := monitor.NewSilenceManager(nil)
	mux := http.NewServeMux()
	registerSilenceRoutes(mux, sm)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Add a silence
	body := map[string]interface{}{
		"label":     "roundtrip",
		"job_names": []string{"nightly-backup"},
		"start":     time.Now().Add(-time.Minute).Format(time.RFC3339),
		"end":       time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"reason":    "integration test",
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/silences", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("add request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// Confirm it is active
	if !sm.IsSilenced("nightly-backup") {
		t.Fatal("expected nightly-backup to be silenced")
	}

	// List should return 1 entry
	listResp, _ := http.Get(srv.URL + "/silences")
	var silences []monitor.Silence
	json.NewDecoder(listResp.Body).Decode(&silences)
	if len(silences) != 1 {
		t.Fatalf("expected 1 silence, got %d", len(silences))
	}

	// Delete it
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/silences/delete?label=roundtrip", nil)
	delResp, _ := http.DefaultClient.Do(req)
	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", delResp.StatusCode)
	}

	// Confirm no longer silenced
	if sm.IsSilenced("nightly-backup") {
		t.Fatal("expected silence removed after delete")
	}
}

// TestAddSilence_DuplicateConflict ensures duplicate labels are rejected.
func TestAddSilence_DuplicateConflict(t *testing.T) {
	sm := monitor.NewSilenceManager(nil)
	mux := http.NewServeMux()
	registerSilenceRoutes(mux, sm)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	body := map[string]interface{}{
		"label": "dup-label",
		"start": time.Now().Add(-time.Minute).Format(time.RFC3339),
		"end":   time.Now().Add(time.Hour).Format(time.RFC3339),
	}
	b, _ := json.Marshal(body)

	http.Post(srv.URL+"/silences", "application/json", bytes.NewReader(b))
	b, _ = json.Marshal(body)
	resp, _ := http.Post(srv.URL+"/silences", "application/json", bytes.NewReader(b))
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}
