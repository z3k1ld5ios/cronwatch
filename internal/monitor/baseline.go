package monitor

import (
	"math"
	"sync"
	"time"
)

// BaselineManager tracks expected execution durations per job and detects
// when a job's runtime deviates significantly from its historical baseline.
type BaselineManager struct {
	mu      sync.Mutex
	samples map[string][]float64
	maxSize int
	threshold float64 // z-score threshold for flagging deviation
}

// BaselineResult holds the outcome of a baseline check.
type BaselineResult struct {
	JobName   string
	Actual    time.Duration
	Mean      time.Duration
	StdDev    time.Duration
	ZScore    float64
	Anomalous bool
}

// NewBaselineManager creates a BaselineManager with the given sample cap and
// z-score threshold (e.g. 2.0 flags values more than 2 std-devs from mean).
func NewBaselineManager(maxSize int, threshold float64) *BaselineManager {
	if maxSize <= 0 {
		maxSize = 100
	}
	if threshold <= 0 {
		threshold = 2.0
	}
	return &BaselineManager{
		samples:   make(map[string][]float64),
		maxSize:   maxSize,
		threshold: threshold,
	}
}

// Record adds a duration sample for the named job.
func (b *BaselineManager) Record(job string, d time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	v := d.Seconds()
	s := append(b.samples[job], v)
	if len(s) > b.maxSize {
		s = s[len(s)-b.maxSize:]
	}
	b.samples[job] = s
}

// Check evaluates the given duration against the job's baseline.
// Returns nil if fewer than 3 samples exist (not enough data).
func (b *BaselineManager) Check(job string, d time.Duration) *BaselineResult {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.samples[job]
	if len(s) < 3 {
		return nil
	}
	mean, stddev := meanStddev(s)
	actual := d.Seconds()
	var z float64
	if stddev > 0 {
		z = math.Abs(actual-mean) / stddev
	}
	return &BaselineResult{
		JobName:   job,
		Actual:    d,
		Mean:      time.Duration(mean * float64(time.Second)),
		StdDev:    time.Duration(stddev * float64(time.Second)),
		ZScore:    z,
		Anomalous: z >= b.threshold,
	}
}

// Reset clears all samples for a job.
func (b *BaselineManager) Reset(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.samples, job)
}

// AllStats returns a snapshot of mean/stddev for every tracked job.
func (b *BaselineManager) AllStats() map[string][2]float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make(map[string][2]float64, len(b.samples))
	for job, s := range b.samples {
		if len(s) == 0 {
			continue
		}
		m, sd := meanStddev(s)
		out[job] = [2]float64{m, sd}
	}
	return out
}

func meanStddev(s []float64) (float64, float64) {
	var sum float64
	for _, v := range s {
		sum += v
	}
	mean := sum / float64(len(s))
	var variance float64
	for _, v := range s {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(s))
	return mean, math.Sqrt(variance)
}
