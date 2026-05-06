package monitor

import (
	"sync"
	"time"
)

// DedupKey uniquely identifies an alert event.
type DedupKey struct {
	JobName string
	Kind    string
}

// DedupEntry records when an alert was last fired and its fingerprint.
type DedupEntry struct {
	Fingerprint string
	FiredAt     time.Time
}

// DedupManager prevents duplicate alerts from being sent within a
// configurable window when the alert fingerprint has not changed.
type DedupManager struct {
	mu      sync.Mutex
	window  time.Duration
	entries map[DedupKey]DedupEntry
	clock   func() time.Time
}

// NewDedupManager creates a DedupManager with the given deduplication window.
func NewDedupManager(window time.Duration) *DedupManager {
	return &DedupManager{
		window:  window,
		entries: make(map[DedupKey]DedupEntry),
		clock:   time.Now,
	}
}

// IsDuplicate returns true when an identical alert (same job, kind, and
// fingerprint) was already sent within the deduplication window.
func (d *DedupManager) IsDuplicate(jobName, kind, fingerprint string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := DedupKey{JobName: jobName, Kind: kind}
	entry, ok := d.entries[key]
	if !ok {
		return false
	}
	if entry.Fingerprint != fingerprint {
		return false
	}
	return d.clock().Sub(entry.FiredAt) < d.window
}

// Record stores the alert as the most recent firing for the given key.
func (d *DedupManager) Record(jobName, kind, fingerprint string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := DedupKey{JobName: jobName, Kind: kind}
	d.entries[key] = DedupEntry{
		Fingerprint: fingerprint,
		FiredAt:     d.clock(),
	}
}

// Reset clears the deduplication record for a specific job and kind.
func (d *DedupManager) Reset(jobName, kind string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.entries, DedupKey{JobName: jobName, Kind: kind})
}

// Len returns the number of tracked dedup entries.
func (d *DedupManager) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries)
}
