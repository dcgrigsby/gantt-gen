package resolver

import (
	"testing"
	"time"

	"gantt-gen/model"
)

func TestResolve_EmptyProject(t *testing.T) {
	project := &model.Project{
		Tasks: []model.Task{},
	}

	err := Resolve(project)
	if err != nil {
		t.Errorf("Resolve() error = %v, want nil for empty project", err)
	}
}

func TestResolve_TaskWithNoDatesOrDependencies(t *testing.T) {
	project := &model.Project{
		Tasks: []model.Task{
			{Name: "Task A", Duration: 5},
		},
	}

	err := Resolve(project)
	if err == nil {
		t.Error("Resolve() expected error for task with no dates or dependencies")
	}

	expectedMsg := "task Task A has no start date, date range, or dependencies"
	if err.Error() != expectedMsg {
		t.Errorf("Resolve() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestResolve_MilestoneWithDate(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	project := &model.Project{
		Tasks: []model.Task{
			{
				Name:        "Launch",
				IsMilestone: true,
				Date:        &date,
			},
		},
	}

	err := Resolve(project)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	task := project.Tasks[0]
	if task.CalculatedStart == nil || !task.CalculatedStart.Equal(date) {
		t.Errorf("milestone start = %v, want %v", task.CalculatedStart, date)
	}
	if task.CalculatedEnd == nil || !task.CalculatedEnd.Equal(date) {
		t.Errorf("milestone end = %v, want %v", task.CalculatedEnd, date)
	}
}

func TestResolve_MultipleDependencySameTask(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	project := &model.Project{
		Tasks: []model.Task{
			{Name: "Task A", Start: &start, Duration: 5},
			{Name: "Task B", Start: &start, Duration: 10},
			{
				Name:     "Task C",
				Duration: 3,
				Dependencies: []model.Dependency{
					{TaskName: "Task A", Type: model.FinishToStart},
					{TaskName: "Task B", Type: model.FinishToStart},
				},
			},
		},
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

	// Task C should start after BOTH A and B finish
	// B finishes later (Jan 11), so C should start Jan 11
	taskC := project.Tasks[2]
	wantStart := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)
	if !taskC.CalculatedStart.Equal(wantStart) {
		t.Errorf("Task C start = %v, want %v", taskC.CalculatedStart, wantStart)
	}
}

func TestResolve_ZeroDurationTask(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	project := &model.Project{
		Tasks: []model.Task{
			{Name: "Task A", Start: &start, Duration: 0},
		},
	}

	err := Resolve(project)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// Zero duration should result in start == end
	task := project.Tasks[0]
	if !task.CalculatedStart.Equal(*task.CalculatedEnd) {
		t.Errorf("zero duration task: start %v != end %v", task.CalculatedStart, task.CalculatedEnd)
	}
}
