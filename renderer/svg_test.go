package renderer

import (
	"strings"
	"testing"
	"time"

	"gantt-gen/model"
)

func TestRenderSVG(t *testing.T) {
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

	svg, err := RenderSVG(project)
	if err != nil {
		t.Fatalf("RenderSVG() error = %v", err)
	}

	// Basic structure checks
	if !strings.Contains(svg, "<svg") {
		t.Error("Output should contain SVG tag")
	}

	if !strings.Contains(svg, "Task A") {
		t.Error("SVG should contain task name")
	}

	if !strings.Contains(svg, "<rect") {
		t.Error("SVG should contain rectangles for tasks")
	}
}
