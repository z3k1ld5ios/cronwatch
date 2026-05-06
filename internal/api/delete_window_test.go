package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/internal/monitor"
)

func TestDeleteWindow_MissingLabel(t *testing.T) {
	wm := monitor.NewWindowManager()
	h := handleDeleteWindow(wm)

	req := httptest.NewRequest(http.MethodDelete, "/windows", nil)
	rw := httptest.NewRecorder()
	h(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestDeleteWindow_WrongMethod(t *testing.T) {
	wm := monitor.NewWindowManager()
	h := handleDeleteWindow(wm)

	req := httptest.NewRequest(http.MethodGet, "/windows?label=foo", nil)
	rw := httptest.NewRecorder()
	h(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}

func TestDeleteWindow_RemovesExisting(t *testing.T) {
	wm := monitor.NewWindowManager()
	now := time.Now().UTC()
	wm.AddWindow("maint", "", now, now.Add(time.Hour))

	h := handleDeleteWindow(wm)
	req := httptest.NewRequest(http.MethodDelete, "/windows?label=maint", nil)
	rw := httptest.NewRecorder()
	h(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}

	// Confirm it is gone by checking suppression is lifted
	if wm.IsSuppressed("any-job") {
		t.Fatal("window should have been removed")
	}
}

func TestDeleteWindow_NonExistentLabelIsNoOp(t *testing.T) {
	wm := monitor.NewWindowManager()
	h := handleDeleteWindow(wm)

	req := httptest.NewRequest(http.MethodDelete, "/windows?label=ghost", nil)
	rw := httptest.NewRecorder()
	h(rw, req)

	// Removing a non-existent window should still succeed gracefully
	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}
}
