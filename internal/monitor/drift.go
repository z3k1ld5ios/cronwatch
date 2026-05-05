package monitor

import (
	"time"

	"github.com/yourorg/cronwatch/internal/schedule"
)

// DriftAnalyzer computes execution drift for a named job.
type DriftAnalyzer struct {
	history *schedule.History
}

// NewDriftAnalyzer returns a DriftAnalyzer backed by the given History.
func NewDriftAnalyzer(h *schedule.History) *DriftAnalyzer {
	return &DriftAnalyzer{history: h}
}

// DriftResult holds the result of a drift analysis.
type DriftResult struct {
	JobName      string
	Expected     time.Time
	Actual       time.Time
	Drift        time.Duration
	AbsDrift     time.Duration
	IsSignificant bool
}

// Analyze computes the drift between the last recorded execution and the
// expected execution time derived from the schedule. threshold is the minimum
// absolute drift to consider significant.
func (d *DriftAnalyzer) Analyze(jobName string, expected time.Time, threshold time.Duration) (DriftResult, bool) {
	latest, ok := d.history.Latest(jobName)
	if !ok {
		return DriftResult{}, false
	}

	drift := latest.Sub(expected)
	abs := drift
	if abs < 0 {
		abs = -abs
	}

	return DriftResult{
		JobName:       jobName,
		Expected:      expected,
		Actual:        latest,
		Drift:         drift,
		AbsDrift:      abs,
		IsSignificant: abs >= threshold,
	}, true
}
