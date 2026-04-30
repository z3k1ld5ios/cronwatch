package schedule

import (
	"testing"
	"time"
)

func TestParse_Wildcard(t *testing.T) {
	c, err := Parse("* * * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Minute) != 60 {
		t.Errorf("expected 60 minutes, got %d", len(c.Minute))
	}
}

func TestParse_InvalidFieldCount(t *testing.T) {
	_, err := Parse("* * * *")
	if err == nil {
		t.Fatal("expected error for 4-field expression")
	}
}

func TestParse_StepExpression(t *testing.T) {
	c, err := Parse("*/15 * * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []int{0, 15, 30, 45}
	for i, v := range expected {
		if c.Minute[i] != v {
			t.Errorf("minute[%d]: want %d, got %d", i, v, c.Minute[i])
		}
	}
}

func TestParse_RangeExpression(t *testing.T) {
	c, err := Parse("0 9-17 * * 1-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Hour) != 9 {
		t.Errorf("expected 9 hours (9-17), got %d", len(c.Hour))
	}
	if len(c.Weekday) != 5 {
		t.Errorf("expected 5 weekdays (1-5), got %d", len(c.Weekday))
	}
}

func TestParse_ListExpression(t *testing.T) {
	c, err := Parse("0 8,12,18 * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Hour) != 3 {
		t.Errorf("expected 3 hours, got %d", len(c.Hour))
	}
}

func TestNext_EveryMinute(t *testing.T) {
	c, _ := Parse("* * * * *")
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	next := c.Next(now)
	expected := time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next() = %v, want %v", next, expected)
	}
}

func TestNext_HourlyAt30(t *testing.T) {
	c, _ := Parse("30 * * * *")
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	next := c.Next(now)
	expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next() = %v, want %v", next, expected)
	}
}

func TestNext_CrossesMidnight(t *testing.T) {
	c, _ := Parse("0 2 * * *")
	now := time.Date(2024, 1, 15, 3, 0, 0, 0, time.UTC)
	next := c.Next(now)
	expected := time.Date(2024, 1, 16, 2, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next() = %v, want %v", next, expected)
	}
}
