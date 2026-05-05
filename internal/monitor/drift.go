package monitor

import (
	"time"

	"github.com/cronwatch/cronwatch/internal/schedule"
)

// DriftThreshold is the maximum acceptable drift before an alert is raised.
const DefaultDriftThreshold = 30 * time.Second

// DriftAnalyzer evaluates run history to detect drift trends.
type DriftAnalyzer struct {
	history   *schedule.History
	threshold time.Duration
}

// NewDriftAnalyzer creates a DriftAnalyzer using the provided History and threshold.
func NewDriftAnalyzer(h *schedule.History, threshold time.Duration) *DriftAnalyzer {
	if threshold <= 0 {
		threshold = DefaultDriftThreshold
	}
	return &DriftAnalyzer{history: h, threshold: threshold}
}

// DriftResult holds the outcome of a drift evaluation.
type DriftResult struct {
	JobName       string
	LatestDrift   time.Duration
	AverageDrift  time.Duration
	ExceedsLimit  bool
}

// Evaluate checks the recent run history for jobName and returns a DriftResult.
func (d *DriftAnalyzer) Evaluate(jobName string) (DriftResult, bool) {
	records := d.history.All(jobName)
	if len(records) == 0 {
		return DriftResult{}, false
	}

	latest := records[len(records)-1]
	var total time.Duration
	for _, r := range records {
		total += r.Drift
	}
	avg := total / time.Duration(len(records))

	return DriftResult{
		JobName:      jobName,
		LatestDrift:  latest.Drift,
		AverageDrift: avg,
		ExceedsLimit: latest.Drift > d.threshold || latest.Drift < -d.threshold,
	}, true
}
