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

func newLoggingServer(t *testing.T) *api.Server {
	t.Helper()
	tracker := schedule.NewTracker()
	history := schedule.NewHistory(10)
	return api.NewServer(":0", tracker, history).WithLogging()
}

func TestWithLogging_HeartbeatStillAccepted(t *testing.T) {
	srv := newLoggingServer(t)
	body, _ := json.Marshal(api.HeartbeatRequest{
		JobName:   "report",
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

func TestWithLogging_StatusStillOK(t *testing.T) {
	srv := newLoggingServer(t)
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestWithLogging_UnknownRoute(t *testing.T) {
	srv := newLoggingServer(t)
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
