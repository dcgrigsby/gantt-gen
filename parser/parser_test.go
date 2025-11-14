package parser

import (
	"testing"
)

func TestParse_Headers(t *testing.T) {
	input := `# Project Name

## Design Phase

## Implementation

### Code Review
`

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if project.Name != "Project Name" {
		t.Errorf("project.Name = %q, want %q", project.Name, "Project Name")
	}

	if len(project.Tasks) != 3 {
		t.Fatalf("len(tasks) = %d, want 3", len(project.Tasks))
	}

	want := []struct{name string; level int}{
		{"Design Phase", 2},
		{"Implementation", 2},
		{"Code Review", 3},
	}

	for i, w := range want {
		if project.Tasks[i].Name != w.name {
			t.Errorf("task[%d].Name = %q, want %q", i, project.Tasks[i].Name, w.name)
		}
		if project.Tasks[i].Level != w.level {
			t.Errorf("task[%d].Level = %d, want %d", i, project.Tasks[i].Level, w.level)
		}
	}
}

func TestParse_Milestones(t *testing.T) {
	input := `# Project

## Regular Task

**Launch Milestone**
`

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(project.Tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(project.Tasks))
	}

	if !project.Tasks[1].IsMilestone {
		t.Error("second task should be milestone")
	}

	if project.Tasks[1].Name != "Launch Milestone" {
		t.Errorf("milestone.Name = %q, want %q", project.Tasks[1].Name, "Launch Milestone")
	}
}

func TestParse_PropertyTable(t *testing.T) {
	input := `# Project

## Design Phase

| Property | Value |
|----------|-------|
| Start | 2024-01-01 |
| Duration | 5d |
| Link | https://jira.com/PROJ-123 |
`

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(project.Tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(project.Tasks))
	}

	task := project.Tasks[0]

	if task.Start == nil {
		t.Error("task.Start should not be nil")
	}

	if task.Duration != 5 {
		t.Errorf("task.Duration = %d, want 5", task.Duration)
	}

	if task.Link != "https://jira.com/PROJ-123" {
		t.Errorf("task.Link = %q, want %q", task.Link, "https://jira.com/PROJ-123")
	}
}

func TestParse_DependencyTable(t *testing.T) {
	input := `# Project

## Task A

## Task B

| Depends On | Type |
|------------|------|
| Task A | finish-to-start |
`

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(project.Tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(project.Tasks))
	}

	task := project.Tasks[1]

	if len(task.Dependencies) != 1 {
		t.Fatalf("len(dependencies) = %d, want 1", len(task.Dependencies))
	}

	dep := task.Dependencies[0]
	if dep.TaskName != "Task A" {
		t.Errorf("dep.TaskName = %q, want %q", dep.TaskName, "Task A")
	}

	if dep.Type != "finish-to-start" {
		t.Errorf("dep.Type = %q, want %q", dep.Type, "finish-to-start")
	}
}

func TestParse_CalendarTable(t *testing.T) {
	input := `# Project

## Calendar: US-2024

| Type | Value |
|------|-------|
| Default | true |
| Weekends | Sat, Sun |
| Holiday | 2024-01-01 |
| Holiday | 2024-07-04 |
`

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(project.Calendars) != 1 {
		t.Fatalf("len(calendars) = %d, want 1", len(project.Calendars))
	}

	cal := project.Calendars[0]

	if cal.Name != "US-2024" {
		t.Errorf("calendar.Name = %q, want %q", cal.Name, "US-2024")
	}

	if !cal.IsDefault {
		t.Error("calendar should be default")
	}

	if len(cal.Weekends) != 2 {
		t.Errorf("len(weekends) = %d, want 2", len(cal.Weekends))
	}

	if len(cal.Holidays) != 2 {
		t.Errorf("len(holidays) = %d, want 2", len(cal.Holidays))
	}
}
