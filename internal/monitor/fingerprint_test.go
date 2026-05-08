package monitor

import (
	"testing"
	"time"
)

func TestFingerprint_SameInputSameResult(t *testing.T) {
	at := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	bucket := 5 * time.Minute

	a := FingerprintFor("backup", "0 * * * *", AlertKindMissed, at, bucket)
	b := FingerprintFor("backup", "0 * * * *", AlertKindMissed, at, bucket)

	if a != b {
		t.Errorf("expected identical fingerprints, got %q and %q", a, b)
	}
}

func TestFingerprint_DifferentKindDiffers(t *testing.T) {
	at := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	bucket := 5 * time.Minute

	missed := FingerprintFor("backup", "0 * * * *", AlertKindMissed, at, bucket)
	drift := FingerprintFor("backup", "0 * * * *", AlertKindDrift, at, bucket)

	if missed == drift {
		t.Error("expected different fingerprints for different kinds")
	}
}

func TestFingerprint_DifferentJobDiffers(t *testing.T) {
	at := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	bucket := 5 * time.Minute

	a := FingerprintFor("backup", "0 * * * *", AlertKindMissed, at, bucket)
	b := FingerprintFor("cleanup", "0 * * * *", AlertKindMissed, at, bucket)

	if a == b {
		t.Error("expected different fingerprints for different jobs")
	}
}

func TestFingerprint_SameBucketGroupsNearbyTimes(t *testing.T) {
	bucket := 5 * time.Minute
	t1 := time.Date(2024, 1, 15, 10, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 10, 3, 59, 0, time.UTC) // same 5-min bucket

	a := FingerprintFor("backup", "0 * * * *", AlertKindMissed, t1, bucket)
	b := FingerprintFor("backup", "0 * * * *", AlertKindMissed, t2, bucket)

	if a != b {
		t.Errorf("expected same fingerprint within bucket, got %q and %q", a, b)
	}
}

func TestFingerprint_DifferentBucketsDiffer(t *testing.T) {
	bucket := 5 * time.Minute
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 10, 5, 0, 0, time.UTC) // next bucket

	a := FingerprintFor("backup", "0 * * * *", AlertKindMissed, t1, bucket)
	b := FingerprintFor("backup", "0 * * * *", AlertKindMissed, t2, bucket)

	if a == b {
		t.Error("expected different fingerprints for different buckets")
	}
}

func TestFingerprint_LengthIs16Hex(t *testing.T) {
	at := time.Now()
	fp := FingerprintFor("job", "* * * * *", AlertKindDrift, at, time.Minute)
	if len(fp) != 16 {
		t.Errorf("expected 16-char hex fingerprint, got len %d: %q", len(fp), fp)
	}
}
