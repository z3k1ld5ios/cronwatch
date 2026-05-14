package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronwatch/internal/monitor"
)

func newBudgetServer() (*httptest.Server, *monitor.BudgetManager) {
	bm := monitor.NewBudgetManager(monitor.DefaultBudgetPolicy())
	mux := http.NewServeMux()
	registerBudgetRoutes(mux, bm)
	return httptest.NewServer(mux), bm
}

func TestBudgetRecord_Success(t *testing.T) {
	srv, _ := newBudgetServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]string{"job": "jobA"})
	resp, err := http.Post(srv.URL+"/budget/record", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestBudgetRecord_MissingJob(t *testing.T) {
	srv, _ := newBudgetServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]string{})
	resp, err := http.Post(srv.URL+"/budget/record", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestBudgetStatus_ReturnsOkInitially(t *testing.T) {
	srv, _ := newBudgetServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/budget/status?job=jobA")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var s monitor.BudgetStatus
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		t.Fatal(err)
	}
	if s.Level != "ok" {
		t.Fatalf("expected ok, got %s", s.Level)
	}
}

func TestBudgetStatus_MissingJob(t *testing.T) {
	srv, _ := newBudgetServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/budget/status")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestBudgetList_Empty(t *testing.T) {
	srv, _ := newBudgetServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/budget/list")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var statuses []monitor.BudgetStatus
	if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 0 {
		t.Fatalf("expected empty list, got %d", len(statuses))
	}
}

func TestBudgetReset_ClearsFailures(t *testing.T) {
	srv, bm := newBudgetServer()
	defer srv.Close()
	for i := 0; i < 8; i++ {
		bm.RecordFailure("jobA")
	}
	body, _ := json.Marshal(map[string]string{"job": "jobA"})
	resp, err := http.Post(srv.URL+"/budget/reset", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	s := bm.Status("jobA")
	if s.Consumed != 0 {
		t.Fatalf("expected 0 after reset, got %d", s.Consumed)
	}
}

func TestBudgetRecord_WrongMethod(t *testing.T) {
	srv, _ := newBudgetServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/budget/record")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
