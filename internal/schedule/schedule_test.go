package schedule_test

import (
	"testing"
	"time"

	"github.com/trixtur/aprs-thursday/internal/schedule"
)

func TestNextOccurrenceReturnsNextScheduledThursday(t *testing.T) {
	loc := time.UTC
	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "before scheduled time on Thursday",
			now:  time.Date(2026, time.January, 1, 8, 59, 0, 0, loc),
			want: time.Date(2026, time.January, 1, 9, 0, 0, 0, loc),
		},
		{
			name: "at scheduled time advances a week",
			now:  time.Date(2026, time.January, 1, 9, 0, 0, 0, loc),
			want: time.Date(2026, time.January, 8, 9, 0, 0, 0, loc),
		},
		{
			name: "after scheduled time advances a week",
			now:  time.Date(2026, time.January, 1, 9, 1, 0, 0, loc),
			want: time.Date(2026, time.January, 8, 9, 0, 0, 0, loc),
		},
		{
			name: "from Wednesday schedules tomorrow",
			now:  time.Date(2026, time.January, 7, 12, 0, 0, 0, loc),
			want: time.Date(2026, time.January, 8, 9, 0, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := schedule.NextOccurrence(tt.now, time.Thursday, 9, 0, loc)
			if err != nil {
				t.Fatalf("NextOccurrence() error = %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("NextOccurrence() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNextOccurrenceKeepsLocalTimeAcrossDaylightSaving(t *testing.T) {
	loc, err := time.LoadLocation("America/Denver")
	if err != nil {
		t.Fatalf("LoadLocation(): %v", err)
	}
	now := time.Date(2026, time.March, 5, 8, 0, 0, 0, loc)
	got, err := schedule.NextOccurrence(now, time.Thursday, 9, 0, loc)
	if err != nil {
		t.Fatalf("NextOccurrence() error = %v", err)
	}
	want := time.Date(2026, time.March, 5, 9, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("NextOccurrence() = %s, want %s", got, want)
	}

	got, err = schedule.NextOccurrence(time.Date(2026, time.March, 5, 9, 1, 0, 0, loc), time.Thursday, 9, 0, loc)
	if err != nil {
		t.Fatalf("NextOccurrence() following week error = %v", err)
	}
	want = time.Date(2026, time.March, 12, 9, 0, 0, 0, loc)
	if !got.Equal(want) || got.Hour() != 9 || got.Location() != loc {
		t.Fatalf("following occurrence = %s, want %s in configured location", got, want)
	}
}

func TestNextOccurrenceRejectsInvalidSchedule(t *testing.T) {
	tests := []struct {
		name    string
		weekday time.Weekday
		hour    int
		minute  int
		loc     *time.Location
	}{
		{name: "invalid weekday", weekday: time.Weekday(7), hour: 9, loc: time.UTC},
		{name: "invalid hour", weekday: time.Thursday, hour: 24, loc: time.UTC},
		{name: "invalid minute", weekday: time.Thursday, hour: 9, minute: 60, loc: time.UTC},
		{name: "missing timezone", weekday: time.Thursday, hour: 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := schedule.NextOccurrence(time.Now(), tt.weekday, tt.hour, tt.minute, tt.loc); err == nil {
				t.Fatal("NextOccurrence() error = nil, want validation error")
			}
		})
	}
}
