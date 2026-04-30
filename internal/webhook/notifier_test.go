package webhook_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cronwatch/internal/webhook"
)

func TestSend_Success(t *testing.T) {
	var received webhook.Payload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := webhook.NewNotifier(srv.URL)
	p := webhook.Payload{
		JobName:   "backup",
		AlertType: webhook.AlertMissedRun,
		Message:   "job did not run",
		OccuredAt: time.Now().UTC(),
	}

	if err := n.Send(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.JobName != "backup" {
		t.Errorf("job_name: got %q, want %q", received.JobName, "backup")
	}
	if received.AlertType != webhook.AlertMissedRun {
		t.Errorf("alert_type: got %q, want %q", received.AlertType, webhook.AlertMissedRun)
	}
}

func TestSend_NonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := webhook.NewNotifier(srv.URL)
	err := n.Send(webhook.Payload{JobName: "test", AlertType: webhook.AlertDrift})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestSend_EmptyURL(t *testing.T) {
	n := webhook.NewNotifier("")
	err := n.Send(webhook.Payload{JobName: "test"})
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}
