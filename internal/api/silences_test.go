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

func newSilenceServer() (*httptest.Server, *monitor.SilenceManager) {
	sm := monitor.NewSilenceManager(nil)
	mux := http.NewServeMux()
	registerSilenceRoutes(mux, sm)
	return httptest.NewServer(mux), sm
}

func TestAddSilence_Success(t *testing.T) {
	srv, _ := newSilenceServer()
	defer srv.Close()
	body := map[string]interface{}{
		"label":  "maint",
		"start":  time.Now().Add(-time.Minute).Format(time.RFC3339),
		"end":    time.Now().Add(time.Hour).Format(time.RFC3339),
		"reason": "planned maintenance",
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/silences", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAddSilence_MissingLabel(t *testing.T) {
	srv, _ := newSilenceServer()
	defer srv.Close()
	body := map[string]interface{}{
		"start": time.Now().Format(time.RFC3339),
		"end":   time.Now().Add(time.Hour).Format(time.RFC3339),
	}
	b, _ := json.Marshal(body)
	resp, _ := http.Post(srv.URL+"/silences", "application/json", bytes.NewReader(b))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAddSilence_EndBeforeStart(t *testing.T) {
	srv, _ := newSilenceServer()
	defer srv.Close()
	body := map[string]interface{}{
		"label": "bad",
		"start": time.Now().Add(time.Hour).Format(time.RFC3339),
		"end":   time.Now().Format(time.RFC3339),
	}
	b, _ := json.Marshal(body)
	resp, _ := http.Post(srv.URL+"/silences", "application/json", bytes.NewReader(b))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestListSilences_Empty(t *testing.T) {
	srv, _ := newSilenceServer()
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/silences")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDeleteSilence_Success(t *testing.T) {
	srv, sm := newSilenceServer()
	defer srv.Close()
	sm.Add(monitor.Silence{
		Label: "todelete",
		Start: time.Now().Add(-time.Minute),
		End:   time.Now().Add(time.Hour),
	})
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/silences/delete?label=todelete", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if sm.IsSilenced("anyjob") {
		t.Fatal("expected silence removed")
	}
}
