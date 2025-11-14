package calendar

import (
	"testing"
	"time"

	"gantt-gen/model"
)

func TestAddBusinessDays(t *testing.T) {
	cal := &model.Calendar{
		Weekends: []time.Weekday{time.Saturday, time.Sunday},
		Holidays: []time.Time{
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // Monday holiday
		},
	}

	tests := []struct {
		name  string
		start time.Time
		days  int
		want  time.Time
	}{
		{
			name:  "add 1 day on Friday",
			start: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), // Friday
			days:  1,
			want:  time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), // Monday (skip weekend)
		},
		{
			name:  "add 1 day before holiday",
			start: time.Date(2023, 12, 29, 0, 0, 0, 0, time.UTC), // Friday
			days:  1,
			want:  time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), // Tuesday (skip weekend and holiday)
		},
		{
			name:  "add 5 days",
			start: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			days:  5,
			want:  time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddBusinessDays(tt.start, tt.days, cal)
			if !got.Equal(tt.want) {
				t.Errorf("AddBusinessDays() = %v, want %v", got, tt.want)
			}
		})
	}
}
