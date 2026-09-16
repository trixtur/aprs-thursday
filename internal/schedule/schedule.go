// Package schedule calculates occurrences in an explicitly configured timezone.
package schedule

import (
	"fmt"
	"time"
)

// NextOccurrence returns the first occurrence of weekday and local hour/minute
// strictly after now. A call made at or after the scheduled time on that day
// therefore returns the following week's occurrence.
func NextOccurrence(now time.Time, weekday time.Weekday, hour, minute int, location *time.Location) (time.Time, error) {
	if weekday < time.Sunday || weekday > time.Saturday {
		return time.Time{}, fmt.Errorf("invalid weekday %d", weekday)
	}
	if hour < 0 || hour > 23 {
		return time.Time{}, fmt.Errorf("hour must be between 0 and 23")
	}
	if minute < 0 || minute > 59 {
		return time.Time{}, fmt.Errorf("minute must be between 0 and 59")
	}
	if location == nil {
		return time.Time{}, fmt.Errorf("schedule timezone is required")
	}

	localNow := now.In(location)
	daysUntil := (int(weekday) - int(localNow.Weekday()) + 7) % 7
	date := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location)
	date = date.AddDate(0, 0, daysUntil)
	candidate := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, location)
	if !candidate.After(localNow) {
		date = date.AddDate(0, 0, 7)
		candidate = time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, location)
	}
	return candidate, nil
}
