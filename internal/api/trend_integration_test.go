package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/internal/monitor"
)

func TestTrendRoundtrip_RecordAndList(t *testing.T) {
	analyzer := monitor.NewTrendAnalyzer(20)
	mux := http.NewServeMux()
	registerTrendRoutes(mux, analyzer)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	base := time.Now()
	jobs := []string{"alpha", "beta"}
	for _, job := range jobs {
		for i := 0; i < 6; i++ {
			analyzer.Record(job, base.Add(time.Duration(i)*time.Minute), float64(i)*3)
		}
	}

	// List all trends
	resp, err := http.Get(srv.URL + "/trends")
	if err != nil {
		t.Fatalf("list request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var listBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	trends, ok := listBody["trends"].([]interface{})
	if !ok {
		t.Fatal("trends key missing or wrong type")
	}
	if len(trends) != 2 {
		t.Fatalf("expected 2 trend entries, got %d", len(trends))
	}

	// Query individual job trend
	resp2, err := http.Get(srv.URL + "/trends/job?name=alpha")
	if err != nil {
		t.Fatalf("job request failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for job trend, got %d", resp2.StatusCode)
	}

	var jobBody map[string]interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&jobBody); err != nil {
		t.Fatalf("decode job body failed: %v", err)
	}
	if jobBody["job"] != "alpha" {
		t.Fatalf("expected job=alpha, got %v", jobBody["job"])
	}
	if _, ok := jobBody["slope"]; !ok {
		t.Fatal("expected slope field in job trend response")
	}
}

func TestTrendRoundtrip_UnknownJobReturnsEmpty(t *testing.T) {
	analyzer := monitor.NewTrendAnalyzer(10)
	mux := http.NewServeMux()
	registerTrendRoutes(mux, analyzer)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/trends/job?name=ghost")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for unknown job, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if body["job"] != "ghost" {
		t.Fatalf("expected job=ghost, got %v", body["job"])
	}
	samples, ok := body["samples"]
	if !ok {
		t.Fatal("expected samples field")
	}
	if samples.(float64) != 0 {
		t.Fatalf("expected 0 samples for unknown job, got %v", samples)
	}
}
