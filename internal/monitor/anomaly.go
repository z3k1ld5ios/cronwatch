package monitor

import (
	"math"
	"sync"
	"time"
)

// AnomalyDetector uses a rolling z-score to flag execution times that deviate
// significantly from a job's historical mean.
type AnomalyDetector struct {
	mu        sync.Mutex
	threshold float64
	samples   map[string][]float64
	maxWindow int
}

// AnomalyResult holds the outcome of an anomaly check.
type AnomalyResult struct {
	JobName   string
	ZScore    float64
	Mean      float64
	StdDev    float64
	Value     float64
	Anomaly   bool
	DetectedAt time.Time
}

// NewAnomalyDetector creates a detector with the given z-score threshold and
// rolling window size. A threshold of 2.0 is a reasonable default.
func NewAnomalyDetector(threshold float64, maxWindow int) *AnomalyDetector {
	if maxWindow <= 0 {
		maxWindow = 60
	}
	if threshold <= 0 {
		threshold = 2.0
	}
	return &AnomalyDetector{
		threshold: threshold,
		samples:   make(map[string][]float64),
		maxWindow: maxWindow,
	}
}

// Record adds a new duration sample for the given job and returns an
// AnomalyResult. At least 5 samples are required before anomaly detection
// is active; earlier calls always return Anomaly=false.
func (a *AnomalyDetector) Record(job string, d time.Duration, now time.Time) AnomalyResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	v := d.Seconds()
	a.samples[job] = append(a.samples[job], v)
	if len(a.samples[job]) > a.maxWindow {
		a.samples[job] = a.samples[job][len(a.samples[job])-a.maxWindow:]
	}

	result := AnomalyResult{JobName: job, Value: v, DetectedAt: now}
	if len(a.samples[job]) < 5 {
		return result
	}

	mean, stddev := stats(a.samples[job])
	result.Mean = mean
	result.StdDev = stddev

	if stddev == 0 {
		return result
	}

	z := math.Abs(v-mean) / stddev
	result.ZScore = z
	result.Anomaly = z >= a.threshold
	return result
}

// Reset clears all samples for a job.
func (a *AnomalyDetector) Reset(job string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.samples, job)
}

// AllStats returns mean and stddev for every tracked job.
func (a *AnomalyDetector) AllStats() map[string][2]float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make(map[string][2]float64, len(a.samples))
	for job, s := range a.samples {
		m, sd := stats(s)
		out[job] = [2]float64{m, sd}
	}
	return out
}

func stats(xs []float64) (mean, stddev float64) {
	for _, x := range xs {
		mean += x
	}
	mean /= float64(len(xs))
	var variance float64
	for _, x := range xs {
		d := x - mean
		variance += d * d
	}
	variance /= float64(len(xs))
	return mean, math.Sqrt(variance)
}
