package model

import (
	"fmt"
	"time"
)

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

const (
	MaxTaskNameLength = 200
	MaxTasks          = 1000
)

// Validate checks the project for common errors and invariants
func (p *Project) Validate() error {
	if len(p.Tasks) > MaxTasks {
		return fmt.Errorf("too many tasks: %d (max %d)", len(p.Tasks), MaxTasks)
	}

	// Build task name set for uniqueness and dependency checking
	taskNames := make(map[string]bool)
	for _, task := range p.Tasks {
		if task.Name == "" {
			return fmt.Errorf("task has empty name")
		}

		if len(task.Name) > MaxTaskNameLength {
			return fmt.Errorf("task name exceeds %d characters: %q", MaxTaskNameLength, truncate(task.Name, 50))
		}

		if taskNames[task.Name] {
			return fmt.Errorf("duplicate task name: %s", task.Name)
		}
		taskNames[task.Name] = true
	}

	// Build calendar name set
	calNames := make(map[string]bool)
	for _, cal := range p.Calendars {
		calNames[cal.Name] = true
	}

	// Validate dependencies and calendar references
	for _, task := range p.Tasks {
		for _, dep := range task.Dependencies {
			if !taskNames[dep.TaskName] {
				return fmt.Errorf("task %q depends on non-existent task: %s", task.Name, dep.TaskName)
			}
		}

		if task.CalendarName != "" && !calNames[task.CalendarName] {
			return fmt.Errorf("task %q references unknown calendar: %s", task.Name, task.CalendarName)
		}
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
