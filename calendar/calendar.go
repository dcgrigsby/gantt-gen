package calendar

import (
	"time"

	"gantt-gen/model"
)

// AddBusinessDays adds the specified number of business days to start date
func AddBusinessDays(start time.Time, days int, cal *model.Calendar) time.Time {
	if cal == nil {
		cal = DefaultCalendar()
	}

	current := start
	remaining := days

	for remaining > 0 {
		current = current.AddDate(0, 0, 1)

		if IsBusinessDay(current, cal) {
			remaining--
		}
	}

	return current
}

// IsBusinessDay checks if a date is a business day
func IsBusinessDay(date time.Time, cal *model.Calendar) bool {
	if cal == nil {
		cal = DefaultCalendar()
	}

	// Check weekends
	for _, weekend := range cal.Weekends {
		if date.Weekday() == weekend {
			return false
		}
	}

	// Check holidays
	for _, holiday := range cal.Holidays {
		if sameDate(date, holiday) {
			return false
		}
	}

	return true
}

// DefaultCalendar returns a calendar with Sat/Sun weekends and no holidays
func DefaultCalendar() *model.Calendar {
	return &model.Calendar{
		Name:     "default",
		Weekends: []time.Weekday{time.Saturday, time.Sunday},
	}
}

func sameDate(d1, d2 time.Time) bool {
	y1, m1, day1 := d1.Date()
	y2, m2, day2 := d2.Date()
	return y1 == y2 && m1 == m2 && day1 == day2
}
