package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/monitor"
)

func newEscalationServer() (*httptest.Server, *monitor.EscalationManager) {
	policy := monitor.EscalationPolicy{
		WarningAfter:  2,
		CriticalAfter: 4,
		ResetAfter:    10 * time.Minute,
	}
	em := monitor.NewEscalationManager(policy)
	mux := http.NewServeMux()
	registerEscalationRoutes(mux, em)
	return httptest.NewServer(mux), em
}

func TestGetEscalation_NoneInitially(t *testing.T) {
	srv, _ := newEscalationServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/escalation?job=backup")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["label"] != "none" {
		t.Fatalf("expected label=none, got %v", body["label"])
	}
}

func TestGetEscalation_MissingJob(t *testing.T) {
	srv, _ := newEscalationServer()
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/escalation")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetEscalation_WrongMethod(t *testing.T) {
	srv, _ := newEscalationServer()
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/escalation?job=x", "", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestResetEscalation_ClearsLevel(t *testing.T) {
	srv, em := newEscalationServer()
	defer srv.Close()

	for i := 0; i < 4; i++ {
		em.Record("backup")
	}

	resp, _ := http.Post(srv.URL+"/escalation/reset?job=backup", "", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if em.Level("backup") != monitor.LevelNone {
		t.Fatal("expected level to be reset to none")
	}
}

func TestResetEscalation_MissingJob(t *testing.T) {
	srv, _ := newEscalationServer()
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/escalation/reset", "", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
