package monitor

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// AlertKind represents the type of alert being fingerprinted.
type AlertKind string

const (
	AlertKindMissed AlertKind = "missed"
	AlertKindDrift  AlertKind = "drift"
)

// FingerprintInput holds the fields used to compute an alert fingerprint.
type FingerprintInput struct {
	JobName  string
	Kind     AlertKind
	Schedule string
	// Bucket rounds time to the nearest interval to group related alerts.
	Bucket time.Duration
	At     time.Time
}

// Compute returns a stable hex fingerprint for the given alert input.
// Alerts with the same job, kind, schedule, and time bucket share a fingerprint.
func Compute(in FingerprintInput) string {
	bucketedAt := in.At.Truncate(in.Bucket)
	raw := fmt.Sprintf("%s|%s|%s|%d", in.JobName, in.Kind, in.Schedule, bucketedAt.Unix())
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum[:8])
}

// FingerprintFor is a convenience wrapper used by the checker to produce
// a fingerprint directly from alert metadata.
func FingerprintFor(jobName, schedule string, kind AlertKind, at time.Time, bucket time.Duration) string {
	return Compute(FingerprintInput{
		JobName:  jobName,
		Kind:     kind,
		Schedule: schedule,
		Bucket:   bucket,
		At:       at,
	})
}
