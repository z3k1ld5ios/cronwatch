package monitor

import (
	"math"
	"sync"
	"time"
)

// TrendDirection indicates whether drift is improving, stable, or worsening.
type TrendDirection string

const (
	TrendStable   TrendDirection = "stable"
	TrendWorsening TrendDirection = "worsening"
	TrendImproving TrendDirection = "improving"
)

// TrendSummary holds the computed trend for a single job.
type TrendSummary struct {
	Job       string
	Direction TrendDirection
	Slope     float64 // seconds per observation
	Samples   int
}

// TrendAnalyzer tracks recent drift samples and computes linear trend.
type TrendAnalyzer struct {
	mu      sync.Mutex
	samples map[string][]driftSample
	maxSamples int
}

type driftSample struct {
	at    time.Time
	drift float64 // seconds
}

// NewTrendAnalyzer returns a TrendAnalyzer keeping up to maxSamples per job.
func NewTrendAnalyzer(maxSamples int) *TrendAnalyzer {
	if maxSamples <= 0 {
		maxSamples = 10
	}
	return &TrendAnalyzer{
		samples:    make(map[string][]driftSample),
		maxSamples: maxSamples,
	}
}

// Record adds a drift observation for the given job.
func (t *TrendAnalyzer) Record(job string, at time.Time, driftSeconds float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.samples[job] = append(t.samples[job], driftSample{at: at, drift: driftSeconds})
	if len(t.samples[job]) > t.maxSamples {
		t.samples[job] = t.samples[job][len(t.samples[job])-t.maxSamples:]
	}
}

// Analyze returns a TrendSummary for the given job.
func (t *TrendAnalyzer) Analyze(job string) TrendSummary {
	t.mu.Lock()
	defer t.mu.Unlock()
	samples := t.samples[job]
	if len(samples) < 2 {
		return TrendSummary{Job: job, Direction: TrendStable, Samples: len(samples)}
	}
	slope := leastSquaresSlope(samples)
	dir := TrendStable
	if slope > 1.0 {
		dir = TrendWorsening
	} else if slope < -1.0 {
		dir = TrendImproving
	}
	return TrendSummary{Job: job, Direction: dir, Slope: slope, Samples: len(samples)}
}

// AllTrends returns trend summaries for all tracked jobs.
func (t *TrendAnalyzer) AllTrends() []TrendSummary {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]TrendSummary, 0, len(t.samples))
	for job, samples := range t.samples {
		if len(samples) < 2 {
			out = append(out, TrendSummary{Job: job, Direction: TrendStable, Samples: len(samples)})
			continue
		}
		slope := leastSquaresSlope(samples)
		dir := TrendStable
		if slope > 1.0 {
			dir = TrendWorsening
		} else if slope < -1.0 {
			dir = TrendImproving
		}
		out = append(out, TrendSummary{Job: job, Direction: dir, Slope: slope, Samples: len(samples)})
	}
	return out
}

func leastSquaresSlope(samples []driftSample) float64 {
	n := float64(len(samples))
	base := samples[0].at
	var sumX, sumY, sumXY, sumX2 float64
	for _, s := range samples {
		x := s.at.Sub(base).Seconds()
		y := s.drift
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if math.Abs(denom) < 1e-9 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}
