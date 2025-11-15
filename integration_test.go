package main

import (
	"os"
	"strings"
	"testing"

	"gantt-gen/parser"
	"gantt-gen/renderer"
	"gantt-gen/resolver"
)

func TestFullPipeline_SimpleProject(t *testing.T) {
	input := []byte(`# Software Project

## Design Phase

| Property | Value |
|----------|-------|
| Start | 2024-01-01 |
| Duration | 10d |

## Implementation

| Property | Value |
|----------|-------|
| Duration | 15d |

| Depends On | Type |
|------------|------|
| Design Phase | finish-to-start |

**Launch**

| Property | Value |
|----------|-------|
| Date | 2024-02-01 |
`)

	// Parse
	project, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Validate
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Resolve
	if err := resolver.Resolve(project); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// Render
	svg, err := renderer.RenderSVG(project)
	if err != nil {
		t.Fatalf("RenderSVG() error = %v", err)
	}

	// Verify output
	if !strings.Contains(svg, "Software Project") {
		t.Error("SVG missing project name")
	}
	if !strings.Contains(svg, "Design Phase") {
		t.Error("SVG missing task name")
	}
	if !strings.Contains(svg, "Implementation") {
		t.Error("SVG missing dependent task")
	}
	if !strings.Contains(svg, "Launch") {
		t.Error("SVG missing milestone")
	}
}

func TestFullPipeline_ValidationFailure(t *testing.T) {
	input := []byte(`# Project

## Task A
## Task A
`)

	project, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should fail validation due to duplicate names
	err = project.Validate()
	if err == nil {
		t.Fatal("Validate() expected error for duplicate task names")
	}

	if !strings.Contains(err.Error(), "duplicate task name") {
		t.Errorf("Validate() error = %q, want error about duplicate names", err.Error())
	}
}

func TestFullPipeline_RealFile(t *testing.T) {
	// Test with actual example file
	input, err := os.ReadFile("examples/sample-project.md")
	if err != nil {
		t.Skipf("Skipping: examples/sample-project.md not found: %v", err)
	}

	project, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if err := project.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if err := resolver.Resolve(project); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	svg, err := renderer.RenderSVG(project)
	if err != nil {
		t.Fatalf("RenderSVG() error = %v", err)
	}

	if len(svg) < 1000 {
		t.Errorf("SVG output seems too small: %d bytes", len(svg))
	}
}
