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
		var maxDate time.Time

		for _, dep := range task.Dependencies {
			depTask, ok := taskMap[dep.TaskName]
			if !ok {
				return fmt.Errorf("dependency not found: %s", dep.TaskName)
			}

			// Resolve dependency first
			if err := resolveTask(depTask, taskMap, calMap, defaultCal, visiting); err != nil {
				return err
			}

			var candidateDate time.Time

			switch dep.Type {
			case model.FinishToStart:
				if depTask.CalculatedEnd != nil {
					candidateDate = *depTask.CalculatedEnd
				}
			case model.StartToStart:
				if depTask.CalculatedStart != nil {
					candidateDate = *depTask.CalculatedStart
				}
			case model.FinishToFinish:
				// Task must finish when dependency finishes
				// So start = finish - duration
				if depTask.CalculatedEnd != nil && task.Duration > 0 {
					// Calculate backwards
					candidateDate = depTask.CalculatedEnd.AddDate(0, 0, -task.Duration)
				}
			case model.StartToFinish:
				// Task must finish when dependency starts
				if depTask.CalculatedStart != nil {
					candidateDate = *depTask.CalculatedStart
					// This is the end date, not start
				}
			default:
				candidateDate = *depTask.CalculatedEnd
			}

			if candidateDate.After(maxDate) {
				maxDate = candidateDate
			}
		}

		// For most dependency types, maxDate is the start
		// Special handling for finish-based dependencies done above
		task.CalculatedStart = &maxDate

		if task.Duration > 0 {
			end := calendar.AddBusinessDays(maxDate, task.Duration, cal)
			task.CalculatedEnd = &end
		} else {
			// Milestone
			task.CalculatedEnd = &maxDate
		}

		return nil
	}

	return fmt.Errorf("task %s has no start date, date range, or dependencies", task.Name)
}
