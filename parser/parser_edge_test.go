package parser

import (
	"strings"
	"testing"
)

func TestParse_EmptyDocument(t *testing.T) {
	input := ""
	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if project.Name != "" {
		t.Errorf("project.Name = %q, want empty string", project.Name)
	}

	if len(project.Tasks) != 0 {
		t.Errorf("len(tasks) = %d, want 0", len(project.Tasks))
	}
}

func TestParse_OnlyTitle(t *testing.T) {
	input := "# Project Name\n"
	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if project.Name != "Project Name" {
		t.Errorf("project.Name = %q, want %q", project.Name, "Project Name")
	}

	if len(project.Tasks) != 0 {
		t.Errorf("len(tasks) = %d, want 0", len(project.Tasks))
	}
}

func TestParse_UnicodeTaskNames(t *testing.T) {
	input := `# Project

## 设计阶段 🎨

| Property | Value |
|----------|-------|
| Duration | 5d |

## Архитектура

| Property | Value |
|----------|-------|
| Duration | 3d |
`

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(project.Tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(project.Tasks))
	}

	if project.Tasks[0].Name != "设计阶段 🎨" {
		t.Errorf("task[0].Name = %q, want %q", project.Tasks[0].Name, "设计阶段 🎨")
	}

	if project.Tasks[1].Name != "Архитектура" {
		t.Errorf("task[1].Name = %q, want %q", project.Tasks[1].Name, "Архитектура")
	}
}

func TestParse_MalformedTable_MissingPipe(t *testing.T) {
	// Goldmark should handle this gracefully
	input := `# Project

## Task A

| Property | Value
|----------|-------|
| Duration | 5d |
`

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should still parse task, table parsing may fail silently
	if len(project.Tasks) != 1 {
		t.Errorf("len(tasks) = %d, want 1", len(project.Tasks))
	}
}

func TestParse_MultiplePropertiesAndDependencies(t *testing.T) {
	input := `# Project

## Task A

| Property | Value |
|----------|-------|
| Start | 2024-01-01 |
| Duration | 5d |

## Task B

| Property | Value |
|----------|-------|
| Duration | 3d |
| Link | https://example.com |

| Depends On | Type |
|------------|------|
| Task A | finish-to-start |

## Task C

| Property | Value |
|----------|-------|
| Duration | 2d |

| Depends On | Type |
|------------|------|
| Task A | finish-to-start |
| Task B | finish-to-start |
`

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(project.Tasks) != 3 {
		t.Fatalf("len(tasks) = %d, want 3", len(project.Tasks))
	}

	taskC := project.Tasks[2]
	if len(taskC.Dependencies) != 2 {
		t.Errorf("task C has %d dependencies, want 2", len(taskC.Dependencies))
	}
}

func TestParse_TaskWithNoProperties(t *testing.T) {
	input := `# Project

## Task A

## Task B
`

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(project.Tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(project.Tasks))
	}

	// Tasks should have zero values
	if project.Tasks[0].Duration != 0 {
		t.Errorf("task A duration = %d, want 0", project.Tasks[0].Duration)
	}
}

func TestParse_VeryLongTaskName(t *testing.T) {
	longName := strings.Repeat("A", 300)
	input := "# Project\n\n## " + longName + "\n"

	project, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(project.Tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(project.Tasks))
	}

	if project.Tasks[0].Name != longName {
		t.Errorf("task name length = %d, want %d", len(project.Tasks[0].Name), len(longName))
	}
}
