package model

import "time"

// DependencyType specifies how tasks relate
type DependencyType string

const (
	FinishToStart  DependencyType = "finish-to-start"
	FinishToFinish DependencyType = "finish-to-finish"
	StartToStart   DependencyType = "start-to-start"
	StartToFinish  DependencyType = "start-to-finish"
)

// Dependency represents a task dependency
type Dependency struct {
	TaskName string
	Type     DependencyType
}

// Task represents a task or milestone
type Task struct {
	Name         string
	Level        int        // Heading level (2=H2, 3=H3, etc) or 0 for milestone
	IsMilestone  bool
	Start        *time.Time // Explicit start date
	End          *time.Time // Explicit end date (only for date ranges)
	Date         *time.Time // Explicit date for milestones
	Duration     int        // Duration in days
	Link         string
	CalendarName string
	Dependencies []Dependency

	// Calculated fields (filled by resolver)
	CalculatedStart *time.Time
	CalculatedEnd   *time.Time
}

// IsCalculated returns true if the task timing is determined by dependencies
func (t *Task) IsCalculated() bool {
	return t.Start == nil && t.Date == nil && len(t.Dependencies) > 0
}

// Calendar represents working days configuration
type Calendar struct {
	Name      string
	IsDefault bool
	Weekends  []time.Weekday
	Holidays  []time.Time
}

// Project represents the entire parsed document
type Project struct {
	Name      string
	Tasks     []Task
	Calendars []Calendar
}
