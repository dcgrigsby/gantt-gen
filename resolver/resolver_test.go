package resolver

import (
	"testing"
	"time"

	"gantt-gen/model"
)

func TestResolve_FinishToStart(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tasks := []model.Task{
		{
			Name:     "Task A",
			Start:    &start,
			Duration: 5,
		},
		{
			Name:     "Task B",
			Duration: 3,
			Dependencies: []model.Dependency{
				{TaskName: "Task A", Type: model.FinishToStart},
			},
		},
	}

	// Use a calendar with no weekends for simple date arithmetic
	project := &model.Project{
		Tasks: tasks,
		Calendars: []model.Calendar{
			{
				Name:      "no-weekends",
				IsDefault: true,
				Weekends:  []time.Weekday{}, // No weekends
			},
		},
	}

	err := Resolve(project)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// Task A: Jan 1 + 5 days = Jan 6 (end)
	// Task B: Jan 6 (start) + 3 days = Jan 9 (end)

	taskB := project.Tasks[1]

	wantStart := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)
	if taskB.CalculatedStart == nil {
		t.Fatal("Task B CalculatedStart is nil")
	}
	if !taskB.CalculatedStart.Equal(wantStart) {
		t.Errorf("Task B start = %v, want %v", taskB.CalculatedStart, wantStart)
	}

	wantEnd := time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC)
	if taskB.CalculatedEnd == nil {
		t.Fatal("Task B CalculatedEnd is nil")
	}
	if !taskB.CalculatedEnd.Equal(wantEnd) {
		t.Errorf("Task B end = %v, want %v", taskB.CalculatedEnd, wantEnd)
	}
}

func TestResolve_CircularDependency(t *testing.T) {
	tasks := []model.Task{
		{
			Name:     "Task A",
			Duration: 5,
			Dependencies: []model.Dependency{
				{TaskName: "Task B", Type: model.FinishToStart},
			},
		},
		{
			Name:     "Task B",
			Duration: 3,
			Dependencies: []model.Dependency{
				{TaskName: "Task A", Type: model.FinishToStart},
			},
		},
	}

	project := &model.Project{
		Tasks: tasks,
	}

	err := Resolve(project)
	if err == nil {
		t.Error("Expected error for circular dependency, got nil")
	}

	if err != nil && err.Error() != "circular dependency detected involving task: Task A" &&
		err.Error() != "circular dependency detected involving task: Task B" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestResolve_FinishToFinish(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tasks := []model.Task{
		{
			Name:     "Task A",
			Start:    &start,
			Duration: 10, // Jan 1 -> Jan 11
		},
		{
			Name:     "Task B",
			Duration: 5, // Should end when A ends (Jan 11), so starts Jan 6
			Dependencies: []model.Dependency{
				{TaskName: "Task A", Type: model.FinishToFinish},
			},
		},
	}

	project := &model.Project{
		Tasks: tasks,
		Calendars: []model.Calendar{
			{
				Name:      "no-weekends",
				IsDefault: true,
				Weekends:  []time.Weekday{}, // No weekends
			},
		},
	}

	err := Resolve(project)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	taskB := project.Tasks[1]

	// Task B should end on Jan 11 (same as Task A)
	wantEnd := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)
	if taskB.CalculatedEnd == nil {
		t.Fatal("Task B CalculatedEnd is nil")
	}
	if !taskB.CalculatedEnd.Equal(wantEnd) {
		t.Errorf("Task B end = %v, want %v", taskB.CalculatedEnd, wantEnd)
	}

	// Task B duration is 5 days, so should start on Jan 6
	wantStart := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)
	if taskB.CalculatedStart == nil {
		t.Fatal("Task B CalculatedStart is nil")
	}
	if !taskB.CalculatedStart.Equal(wantStart) {
		t.Errorf("Task B start = %v, want %v", taskB.CalculatedStart, wantStart)
	}
}

func TestResolve_StartToFinish(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tasks := []model.Task{
		{
			Name:     "Task A",
			Start:    &start,
			Duration: 10, // Jan 1 -> Jan 11
		},
		{
			Name:     "Task B",
			Duration: 5, // Should finish when A starts (Jan 1), so runs Dec 27 -> Jan 1
			Dependencies: []model.Dependency{
				{TaskName: "Task A", Type: model.StartToFinish},
			},
		},
	}

	project := &model.Project{
		Tasks: tasks,
		Calendars: []model.Calendar{
			{
				Name:      "no-weekends",
				IsDefault: true,
				Weekends:  []time.Weekday{}, // No weekends
			},
		},
	}

	err := Resolve(project)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	taskB := project.Tasks[1]

	// Task B should end on Jan 1 (when Task A starts)
	wantEnd := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if taskB.CalculatedEnd == nil {
		t.Fatal("Task B CalculatedEnd is nil")
	}
	if !taskB.CalculatedEnd.Equal(wantEnd) {
		t.Errorf("Task B end = %v, want %v", taskB.CalculatedEnd, wantEnd)
	}

	// Task B duration is 5 days, so should start on Dec 27
	wantStart := time.Date(2023, 12, 27, 0, 0, 0, 0, time.UTC)
	if taskB.CalculatedStart == nil {
		t.Fatal("Task B CalculatedStart is nil")
	}
	if !taskB.CalculatedStart.Equal(wantStart) {
		t.Errorf("Task B start = %v, want %v", taskB.CalculatedStart, wantStart)
	}
}
