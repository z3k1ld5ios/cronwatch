package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/internal/api"
	"github.com/cronwatch/internal/schedule"
)

func newTestServer(t *testing.T) *api.Server {
	t.Helper()
	tracker := schedule.NewTracker()
	history := schedule.NewHistory(10)
	return api.NewServer(":0", tracker, history)
}

func TestHeartbeat_Success(t *testing.T) {
	srv := newTestServer(t)
	body, _ := json.Marshal(api.HeartbeatRequest{
		JobName:   "backup",
		Timestamp: time.Now().UTC(),
	})
	req := httptest.NewRequest(http.MethodPost, "/heartbeat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}

func TestHeartbeat_MissingJobName(t *testing.T) {
	srv := newTestServer(t)
	body, _ := json.Marshal(api.HeartbeatRequest{Timestamp: time.Now().UTC()})
	req := httptest.NewRequest(http.MethodPost, "/heartbeat", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHeartbeat_WrongMethod(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/heartbeat", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestStatus_ReturnsJobs(t *testing.T) {
	srv := newTestServer(t)
	// seed a heartbeat first
	body, _ := json.Marshal(api.HeartbeatRequest{JobName: "sync", Timestamp: time.Now().UTC()})
	req := httptest.NewRequest(http.MethodPost, "/heartbeat", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	req2 := httptest.NewRequest(http.MethodGet, "/status", nil)
	rr2 := httptest.NewRecorder()
	srv.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	var resp api.StatusResponse
	if err := json.NewDecoder(rr2.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Jobs) != 1 || resp.Jobs[0].JobName != "sync" {
		t.Fatalf("unexpected jobs: %+v", resp.Jobs)
	}
	if resp.Jobs[0].LastSeen == nil {
		t.Fatal("expected LastSeen to be set")
	}
}

func TestStatus_WrongMethod(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/status", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
