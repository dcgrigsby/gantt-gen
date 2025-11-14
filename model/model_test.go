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

func ptr(t time.Time) *time.Time {
	return &t
}
