package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronExpression represents a parsed cron schedule
type CronExpression struct {
	Raw     string
	Minute  []int
	Hour    []int
	Day     []int
	Month   []int
	Weekday []int
}

// Parse parses a standard 5-field cron expression
func Parse(expr string) (*CronExpression, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid cron expression %q: expected 5 fields, got %d", expr, len(fields))
	}

	c := &CronExpression{Raw: expr}
	var err error

	limits := [5][2]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 6}}
	ptrs := []*[]int{&c.Minute, &c.Hour, &c.Day, &c.Month, &c.Weekday}

	for i, field := range fields {
		*ptrs[i], err = parseField(field, limits[i][0], limits[i][1])
		if err != nil {
			return nil, fmt.Errorf("field %d: %w", i+1, err)
		}
	}

	return c, nil
}

// Next returns the next scheduled time after t
func (c *CronExpression) Next(t time.Time) time.Time {
	t = t.Add(time.Minute).Truncate(time.Minute)
	for i := 0; i < 366*24*60; i++ {
		if contains(c.Month, int(t.Month())) &&
			contains(c.Day, t.Day()) &&
			contains(c.Weekday, int(t.Weekday())) &&
			contains(c.Hour, t.Hour()) &&
			contains(c.Minute, t.Minute()) {
			return t
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}
}

func parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		return makeRange(min, max, 1), nil
	}
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(field[2:])
		if err != nil || step <= 0 {
			return nil, fmt.Errorf("invalid step %q", field)
		}
		return makeRange(min, max, step), nil
	}
	var result []int
	for _, part := range strings.Split(field, ",") {
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			lo, err1 := strconv.Atoi(bounds[0])
			hi, err2 := strconv.Atoi(bounds[1])
			if err1 != nil || err2 != nil || lo > hi {
				return nil, fmt.Errorf("invalid range %q", part)
			}
			result = append(result, makeRange(lo, hi, 1)...)
		} else {
			v, err := strconv.Atoi(part)
			if err != nil || v < min || v > max {
				return nil, fmt.Errorf("value %q out of range [%d,%d]", part, min, max)
			}
			result = append(result, v)
		}
	}
	return result, nil
}

func makeRange(lo, hi, step int) []int {
	var r []int
	for i := lo; i <= hi; i += step {
		r = append(r, i)
	}
	return r
}

func contains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
