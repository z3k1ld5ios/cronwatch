package monitor

import (
	"fmt"
	"time"
)

// AlertPayload represents the webhook payload sent on a job anomaly.
type AlertPayload struct {
	Job          string    `json:"job"`
	Kind         string    `json:"kind"` // "drift" or "missed"
	LastRun      time.Time `json:"last_run"`
	ExpectedNext time.Time `json:"expected_next"`
	CheckedAt    time.Time `json:"checked_at"`
	Message      string    `json:"message"`
}

// buildAlert constructs an AlertPayload for the given anomaly.
func buildAlert(job, kind string, lastRun, expectedNext, now time.Time) AlertPayload {
	var msg string
	switch kind {
	case "missed":
		msg = fmt.Sprintf("Job %q missed its scheduled run at %s (last seen: %s)",
			job, expectedNext.Format(time.RFC3339), lastRun.Format(time.RFC3339))
	case "drift":
		drift := now.Sub(expectedNext)
		msg = fmt.Sprintf("Job %q is drifting by %s (expected at %s)",
			job, drift.Round(time.Second), expectedNext.Format(time.RFC3339))
	default:
		msg = fmt.Sprintf("Job %q anomaly: %s", job, kind)
	}
	return AlertPayload{
		Job:          job,
		Kind:         kind,
		LastRun:      lastRun,
		ExpectedNext: expectedNext,
		CheckedAt:    now,
		Message:      msg,
	}
}
