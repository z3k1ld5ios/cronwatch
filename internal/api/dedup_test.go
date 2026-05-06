package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func newDedupServer(t *testing.T) (*httptest.Server, *monitor.DedupManager) {
	t.Helper()
	dm := monitor.NewDedupManager(5*time.Minute, nil)
	mux := http.NewServeMux()
	registerDedupRoutes(mux, dm)
	return httptest.NewServer(mux), dm
}

func TestResetDedup_Success(t *testing.T) {
	srv, _ := newDedupServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/dedup/reset?job=backup", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestResetDedup_MissingJob(t *testing.T) {
	srv, _ := newDedupServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/dedup/reset", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestResetDedup_WrongMethod(t *testing.T) {
	srv, _ := newDedupServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/dedup/reset?job=backup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestDedupStats_Empty(t *testing.T) {
	srv, _ := newDedupServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/dedup/stats")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := body["entries"]; !ok {
		t.Error("expected 'entries' key in response")
	}
}

func TestDedupStats_WrongMethod(t *testing.T) {
	srv, _ := newDedupServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/dedup/stats", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}
