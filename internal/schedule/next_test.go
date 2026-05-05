package schedule

import (
	"testing"
	"time"
)

func mustParse(t *testing.T, expr string) Schedule {
	t.Helper()
	s, err := Parse(expr)
	if err != nil {
		t.Fatalf("Parse(%q) unexpected error: %v", expr, err)
	}
	return s
}

func TestNext_EveryMinute(t *testing.T) {
	s := mustParse(t, "* * * * *")
	from := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	next, ok := Next(s, from)
	if !ok {
		t.Fatal("expected a next time, got none")
	}
	want := time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Errorf("Next() = %v, want %v", next, want)
	}
}

func TestNext_HourlyAt30(t *testing.T) {
	s := mustParse(t, "30 * * * *")
	from := time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC)
	next, ok := Next(s, from)
	if !ok {
		t.Fatal("expected a next time, got none")
	}
	want := time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Errorf("Next() = %v, want %v", next, want)
	}
}

func TestNext_DailyMidnight(t *testing.T) {
	s := mustParse(t, "0 0 * * *")
	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	next, ok := Next(s, from)
	if !ok {
		t.Fatal("expected a next time, got none")
	}
	want := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Errorf("Next() = %v, want %v", next, want)
	}
}

func TestNext_WeekdayOnly(t *testing.T) {
	// Monday = 1; find next Monday from a Wednesday
	s := mustParse(t, "0 9 * * 1")
	// 2024-01-17 is a Wednesday
	from := time.Date(2024, 1, 17, 9, 0, 0, 0, time.UTC)
	next, ok := Next(s, from)
	if !ok {
		t.Fatal("expected a next time, got none")
	}
	// Next Monday is 2024-01-22
	want := time.Date(2024, 1, 22, 9, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Errorf("Next() = %v, want %v", next, want)
	}
}

func TestNext_AlreadyOnBoundary(t *testing.T) {
	// from is exactly on a scheduled minute; Next should return the NEXT occurrence
	s := mustParse(t, "0 12 * * *")
	from := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	next, ok := Next(s, from)
	if !ok {
		t.Fatal("expected a next time, got none")
	}
	want := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Errorf("Next() = %v, want %v", next, want)
	}
}
