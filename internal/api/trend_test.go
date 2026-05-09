package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/internal/monitor"
	"github.com/cronwatch/internal/schedule"
)

func newTrendServer(t *testing.T) (*httptest.Server, *monitor.TrendAnalyzer) {
	t.Helper()
	analyzer := monitor.NewTrendAnalyzer(10)
	mux := http.NewServeMux()
	registerTrendRoutes(mux, analyzer)
	return httptest.NewServer(mux), analyzer
}

func TestListTrends_Empty(t *testing.T) {
	srv, _ := newTrendServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/trends")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	trends, ok := body["trends"]
	if !ok {
		t.Fatal("expected trends key")
	}
	if len(trends.([]interface{})) != 0 {
		t.Fatalf("expected empty trends, got %v", trends)
	}
}

func TestListTrends_AfterRecording(t *testing.T) {
	srv, analyzer := newTrendServer(t)
	defer srv.Close()

	base := time.Now()
	for i := 0; i < 5; i++ {
		analyzer.Record("backup", base.Add(time.Duration(i)*time.Minute), 10.0+float64(i)*2)
	}

	resp, err := http.Get(srv.URL + "/trends")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	trends := body["trends"].([]interface{})
	if len(trends) != 1 {
		t.Fatalf("expected 1 trend entry, got %d", len(trends))
	}
}

func TestJobTrend_MissingParam(t *testing.T) {
	srv, _ := newTrendServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/trends/job")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestJobTrend_WrongMethod(t *testing.T) {
	srv, _ := newTrendServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/trends/job?name=backup", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestJobTrend_KnownJob(t *testing.T) {
	srv, analyzer := newTrendServer(t)
	defer srv.Close()

	base := time.Now()
	for i := 0; i < 4; i++ {
		analyzer.Record("nightly", base.Add(time.Duration(i)*time.Hour), float64(i)*5)
	}

	resp, err := http.Get(srv.URL + "/trends/job?name=nightly")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if body["job"] != "nightly" {
		t.Fatalf("expected job=nightly, got %v", body["job"])
	}
}

// Ensure schedule package is imported to avoid unused import errors in build.
var _ = schedule.Parse
