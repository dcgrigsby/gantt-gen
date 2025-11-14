package main

import (
	"fmt"
	"os"

	"gantt-gen/parser"
	"gantt-gen/renderer"
	"gantt-gen/resolver"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.md> <output.svg>\n", os.Args[0])
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	// Read input file
	input, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Parse markdown
	project, err := parser.Parse(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing markdown: %v\n", err)
		os.Exit(1)
	}

	// Validate project structure
	if err := project.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	// Resolve dependencies and calculate dates
	if err := resolver.Resolve(project); err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving dependencies: %v\n", err)
		os.Exit(1)
	}

	// Generate SVG output
	output, err := renderer.RenderSVG(project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering SVG: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated Gantt chart: %s\n", outputPath)
}
