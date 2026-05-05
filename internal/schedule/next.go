package schedule

import "time"

// Next returns the next scheduled time after 'from' for the given Schedule.
// It advances minute by minute up to one year to find the next match.
func Next(s Schedule, from time.Time) (time.Time, bool) {
	// Truncate to the start of the next minute
	t := from.Truncate(time.Minute).Add(time.Minute)

	deadline := from.Add(366 * 24 * time.Hour)

	for t.Before(deadline) {
		if matches(s, t) {
			return t, true
		}
		t = t.Add(time.Minute)
	}

	return time.Time{}, false
}

// matches reports whether t satisfies all fields of the schedule.
func matches(s Schedule, t time.Time) bool {
	if !contains(s.Minute, t.Minute()) {
		return false
	}
	if !contains(s.Hour, t.Hour()) {
		return false
	}
	if !contains(s.DayOfMonth, t.Day()) {
		return false
	}
	if !contains(s.Month, int(t.Month())) {
		return false
	}
	if !contains(s.DayOfWeek, int(t.Weekday())) {
		return false
	}
	return true
}
