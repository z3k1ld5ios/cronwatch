package monitor

import (
	"math"
	"sync"
	"time"
)

// JitterPolicy configures jitter detection thresholds.
type JitterPolicy struct {
	// MinSamples is the number of runs required before jitter is reported.
	MinSamples int
	// HighThreshold is the coefficient of variation (stddev/mean) above which
	// jitter is considered high.
	HighThreshold float64
}

// DefaultJitterPolicy returns sensible defaults.
func DefaultJitterPolicy() JitterPolicy {
	return JitterPolicy{
		MinSamples:    5,
		HighThreshold: 0.2,
	}
}

// JitterResult holds the computed jitter metrics for a single job.
type JitterResult struct {
	Job    string
	Mean   time.Duration
	StdDev time.Duration
	CV     float64 // coefficient of variation
	High   bool
}

// JitterAnalyzer tracks inter-arrival deltas and computes scheduling jitter.
type JitterAnalyzer struct {
	mu      sync.Mutex
	policy  JitterPolicy
	samples map[string][]float64 // seconds between successive runs
	last    map[string]time.Time
}

// NewJitterAnalyzer creates a JitterAnalyzer with the given policy.
func NewJitterAnalyzer(policy JitterPolicy) *JitterAnalyzer {
	return &JitterAnalyzer{
		policy:  policy,
		samples: make(map[string][]float64),
		last:    make(map[string]time.Time),
	}
}

// Record registers a heartbeat for a job at the given time.
func (a *JitterAnalyzer) Record(job string, at time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if prev, ok := a.last[job]; ok {
		delta := at.Sub(prev).Seconds()
		if delta > 0 {
			a.samples[job] = append(a.samples[job], delta)
		}
	}
	a.last[job] = at
}

// Analyze returns jitter metrics for the given job, or nil if insufficient data.
func (a *JitterAnalyzer) Analyze(job string) *JitterResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	samples := a.samples[job]
	if len(samples) < a.policy.MinSamples {
		return nil
	}
	mean, stddev := meanStddev(samples)
	var cv float64
	if mean > 0 {
		cv = stddev / mean
	}
	return &JitterResult{
		Job:    job,
		Mean:   time.Duration(mean * float64(time.Second)),
		StdDev: time.Duration(stddev * float64(time.Second)),
		CV:     cv,
		High:   cv > a.policy.HighThreshold,
	}
}

// Reset clears all recorded samples for a job.
func (a *JitterAnalyzer) Reset(job string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.samples, job)
	delete(a.last, job)
}

func meanStddev(vals []float64) (mean, stddev float64) {
	for _, v := range vals {
		mean += v
	}
	mean /= float64(len(vals))
	var variance float64
	for _, v := range vals {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(vals))
	stddev = math.Sqrt(variance)
	return
}
