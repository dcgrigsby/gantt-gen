package renderer

import (
	"strings"
	"testing"
	"time"

	"gantt-gen/model"
)

func TestRenderHTML(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	project := &model.Project{
		Name: "Test Project",
		Tasks: []model.Task{
			{
				Name:            "Task A",
				Level:           2,
				CalculatedStart: &start,
				CalculatedEnd:   &end,
			},
		},
	}

	html, err := RenderHTML(project)
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}

	// Basic structure checks
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE")
	}

	if !strings.Contains(html, "Test Project") {
		t.Error("HTML should contain project name")
	}

	if !strings.Contains(html, "Task A") {
		t.Error("HTML should contain task name")
	}
}
