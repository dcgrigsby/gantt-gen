package model

import (
	"testing"
	"time"
)

func TestTask_IsCalculated(t *testing.T) {
	tests := []struct {
		name string
		task Task
		want bool
	}{
		{
			name: "task with explicit start date",
			task: Task{Start: ptr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))},
			want: false,
		},
		{
			name: "task with dependency only",
			task: Task{Dependencies: []Dependency{{TaskName: "other"}}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.task.IsCalculated(); got != tt.want {
				t.Errorf("Task.IsCalculated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		project Project
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid project",
			project: Project{
				Name: "Test",
				Tasks: []Task{
					{Name: "Task A", Level: 2},
					{Name: "Task B", Level: 2},
				},
			},
			wantErr: false,
		},
		{
			name: "empty task name",
			project: Project{
				Tasks: []Task{
					{Name: "", Level: 2},
				},
			},
			wantErr: true,
			errMsg:  "task has empty name",
		},
		{
			name: "duplicate task names",
			project: Project{
				Tasks: []Task{
					{Name: "Task A", Level: 2},
					{Name: "Task A", Level: 3},
				},
			},
			wantErr: true,
			errMsg:  "duplicate task name: Task A",
		},
		{
			name: "dependency on non-existent task",
			project: Project{
				Tasks: []Task{
					{
						Name:  "Task A",
						Level: 2,
						Dependencies: []Dependency{
							{TaskName: "Task B", Type: FinishToStart},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "task \"Task A\" depends on non-existent task: Task B",
		},
		{
			name: "calendar reference to non-existent calendar",
			project: Project{
				Tasks: []Task{
					{Name: "Task A", Level: 2, CalendarName: "US-2024"},
				},
				Calendars: []Calendar{},
			},
			wantErr: true,
			errMsg:  "task \"Task A\" references unknown calendar: US-2024",
		},
		{
			name: "task name too long",
			project: Project{
				Tasks: []Task{
					{Name: string(make([]byte, 201)), Level: 2},
				},
			},
			wantErr: true,
			errMsg:  "task name exceeds 200 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Project.Validate() expected error containing %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Project.Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Project.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func ptr(t time.Time) *time.Time {
	return &t
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
