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
