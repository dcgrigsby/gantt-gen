package resolver

import (
	"fmt"
	"time"

	"gantt-gen/calendar"
	"gantt-gen/model"
)

// Resolve calculates all task dates based on dependencies and calendars
func Resolve(project *model.Project) error {
	// Build task map for lookup
	taskMap := make(map[string]*model.Task)
	for i := range project.Tasks {
		taskMap[project.Tasks[i].Name] = &project.Tasks[i]
	}

	// Get default calendar
	var defaultCal *model.Calendar
	for i := range project.Calendars {
		if project.Calendars[i].IsDefault {
			defaultCal = &project.Calendars[i]
			break
		}
	}

	// Build calendar map
	calMap := make(map[string]*model.Calendar)
	for i := range project.Calendars {
		calMap[project.Calendars[i].Name] = &project.Calendars[i]
	}

	// Resolve each task (topological order handled by recursive resolution)
	for i := range project.Tasks {
		if err := resolveTask(&project.Tasks[i], taskMap, calMap, defaultCal, make(map[string]bool)); err != nil {
			return err
		}
	}

	return nil
}

func resolveTask(task *model.Task, taskMap map[string]*model.Task, calMap map[string]*model.Calendar, defaultCal *model.Calendar, visiting map[string]bool) error {
	// Already resolved?
	if task.CalculatedStart != nil && task.CalculatedEnd != nil {
		return nil
	}

	// Cycle detection
	if visiting[task.Name] {
		return fmt.Errorf("circular dependency detected involving task: %s", task.Name)
	}
	visiting[task.Name] = true
	defer delete(visiting, task.Name)

	// Get calendar
	cal := defaultCal
	if task.CalendarName != "" {
		if c, ok := calMap[task.CalendarName]; ok {
			cal = c
		}
	}

	// Case 1: Explicit start date
	if task.Start != nil {
		start := *task.Start
		task.CalculatedStart = &start

		if task.Duration > 0 {
			end := calendar.AddBusinessDays(start, task.Duration, cal)
			task.CalculatedEnd = &end
		} else if task.End != nil {
			task.CalculatedEnd = task.End
		} else {
			// Milestone with explicit date
			task.CalculatedEnd = &start
		}
		return nil
	}

	// Case 2: Explicit date range
	if task.End != nil && task.Start != nil {
		// Already handled above
		return nil
	}

	// Case 3: Explicit milestone date
	if task.Date != nil {
		task.CalculatedStart = task.Date
		task.CalculatedEnd = task.Date
		return nil
	}

	// Case 4: Calculate from dependencies
	if len(task.Dependencies) > 0 {
		var startConstraint time.Time
		var endConstraint time.Time
		hasStartConstraint := false
		hasEndConstraint := false

		for _, dep := range task.Dependencies {
			depTask, ok := taskMap[dep.TaskName]
			if !ok {
				return fmt.Errorf("dependency not found: %s", dep.TaskName)
			}

			// Resolve dependency first
			if err := resolveTask(depTask, taskMap, calMap, defaultCal, visiting); err != nil {
				return err
			}

			switch dep.Type {
			case model.FinishToStart:
				// Task starts when dependency finishes
				if depTask.CalculatedEnd != nil {
					if !hasStartConstraint || depTask.CalculatedEnd.After(startConstraint) {
						startConstraint = *depTask.CalculatedEnd
						hasStartConstraint = true
					}
				}

			case model.StartToStart:
				// Task starts when dependency starts
				if depTask.CalculatedStart != nil {
					if !hasStartConstraint || depTask.CalculatedStart.After(startConstraint) {
						startConstraint = *depTask.CalculatedStart
						hasStartConstraint = true
					}
				}

			case model.FinishToFinish:
				// Task finishes when dependency finishes
				if depTask.CalculatedEnd != nil {
					if !hasEndConstraint || depTask.CalculatedEnd.After(endConstraint) {
						endConstraint = *depTask.CalculatedEnd
						hasEndConstraint = true
					}
				}

			case model.StartToFinish:
				// Task finishes when dependency starts
				if depTask.CalculatedStart != nil {
					if !hasEndConstraint || depTask.CalculatedStart.After(endConstraint) {
						endConstraint = *depTask.CalculatedStart
						hasEndConstraint = true
					}
				}

			default:
				// Treat unknown types as finish-to-start
				if depTask.CalculatedEnd != nil {
					if !hasStartConstraint || depTask.CalculatedEnd.After(startConstraint) {
						startConstraint = *depTask.CalculatedEnd
						hasStartConstraint = true
					}
				}
			}
		}

		// Resolve based on constraint types
		if hasStartConstraint && hasEndConstraint {
			// Both constraints: use start constraint, calculate end from duration
			// (This is a simplification; real MS Project would check for conflicts)
			task.CalculatedStart = &startConstraint
			if task.Duration > 0 {
				end := calendar.AddBusinessDays(startConstraint, task.Duration, cal)
				task.CalculatedEnd = &end
			} else {
				task.CalculatedEnd = &endConstraint
			}
		} else if hasStartConstraint {
			// Only start constraint: calculate normally
			task.CalculatedStart = &startConstraint
			if task.Duration > 0 {
				end := calendar.AddBusinessDays(startConstraint, task.Duration, cal)
				task.CalculatedEnd = &end
			} else {
				task.CalculatedEnd = &startConstraint
			}
		} else if hasEndConstraint {
			// Only end constraint: calculate backwards from end
			task.CalculatedEnd = &endConstraint
			if task.Duration > 0 {
				// Calculate start by subtracting duration (rough approximation)
				start := endConstraint.AddDate(0, 0, -task.Duration)
				task.CalculatedStart = &start
			} else {
				task.CalculatedStart = &endConstraint
			}
		} else {
			return fmt.Errorf("task %s has dependencies but none could be resolved", task.Name)
		}

		return nil
	}

	return fmt.Errorf("task %s has no start date, date range, or dependencies", task.Name)
}
