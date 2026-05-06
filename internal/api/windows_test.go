package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/internal/monitor"
)

func TestAddWindow_Success(t *testing.T) {
	wm := monitor.NewWindowManager()
	mux := http.NewServeMux()
	registerWindowRoutes(mux, wm)

	body := map[string]interface{}{
		"label": "deploy",
		"start": time.Now().UTC().Format(time.RFC3339),
		"end":   time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/windows", bytes.NewReader(b))
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rw.Code)
	}
}

func TestAddWindow_MissingLabel(t *testing.T) {
	wm := monitor.NewWindowManager()
	mux := http.NewServeMux()
	registerWindowRoutes(mux, wm)

	body := map[string]interface{}{
		"start": time.Now().UTC().Format(time.RFC3339),
		"end":   time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/windows", bytes.NewReader(b))
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestAddWindow_EndBeforeStart(t *testing.T) {
	wm := monitor.NewWindowManager()
	mux := http.NewServeMux()
	registerWindowRoutes(mux, wm)

	body := map[string]interface{}{
		"label": "bad-window",
		"start": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		"end":   time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/windows", bytes.NewReader(b))
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestListWindows_Empty(t *testing.T) {
	wm := monitor.NewWindowManager()
	mux := http.NewServeMux()
	registerWindowRoutes(mux, wm)

	req := httptest.NewRequest(http.MethodGet, "/windows", nil)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var result []interface{}
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d items", len(result))
	}
}

func TestRemoveWindow_Success(t *testing.T) {
	wm := monitor.NewWindowManager()
	mux := http.NewServeMux()
	registerWindowRoutes(mux, wm)

	now := time.Now().UTC()
	wm.AddWindow("cleanup", "", now, now.Add(time.Hour))

	req := httptest.NewRequest(http.MethodDelete, "/windows?label=cleanup", nil)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}
}
