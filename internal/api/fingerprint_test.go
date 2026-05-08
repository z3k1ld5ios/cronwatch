package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newFingerprintServer() *httptest.Server {
	mux := http.NewServeMux()
	registerFingerprintRoutes(mux)
	return httptest.NewServer(mux)
}

func TestFingerprint_Success(t *testing.T) {
	srv := newFingerprintServer()
	defer srv.Close()

	body := `{"job_name":"backup","kind":"missed","schedule":"0 * * * *","at":"2024-01-15T10:00:00Z","bucket_seconds":300}`
	resp, err := http.Post(srv.URL+"/fingerprint", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["fingerprint"] == "" {
		t.Error("expected non-empty fingerprint")
	}
	if result["bucketed_at"] == "" {
		t.Error("expected non-empty bucketed_at")
	}
}

func TestFingerprint_MissingFields(t *testing.T) {
	srv := newFingerprintServer()
	defer srv.Close()

	body := `{"job_name":"backup"}`
	resp, err := http.Post(srv.URL+"/fingerprint", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestFingerprint_WrongMethod(t *testing.T) {
	srv := newFingerprintServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/fingerprint")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestFingerprint_InvalidJSON(t *testing.T) {
	srv := newFingerprintServer()
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/fingerprint", "application/json", bytes.NewReader([]byte("not-json")))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestFingerprint_SameInputConsistent(t *testing.T) {
	srv := newFingerprintServer()
	defer srv.Close()

	body := `{"job_name":"cleanup","kind":"drift","schedule":"*/5 * * * *","at":"2024-03-01T08:00:00Z","bucket_seconds":60}`

	get := func() string {
		resp, err := http.Post(srv.URL+"/fingerprint", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var result map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&result)
		return result["fingerprint"]
	}

	if a, b := get(), get(); a != b {
		t.Errorf("expected stable fingerprint, got %q and %q", a, b)
	}
}
