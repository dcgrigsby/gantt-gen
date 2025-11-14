package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gantt-gen/parser"
	"gantt-gen/renderer"
	"gantt-gen/resolver"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.md> <output.html|output.svg>\n", os.Args[0])
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

	// Resolve dependencies and calculate dates
	if err := resolver.Resolve(project); err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving dependencies: %v\n", err)
		os.Exit(1)
	}

	// Generate output based on file extension
	var output string
	ext := strings.ToLower(filepath.Ext(outputPath))

	switch ext {
	case ".html":
		output, err = renderer.RenderHTML(project)
	case ".svg":
		output, err = renderer.RenderSVG(project)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported output format: %s (use .html or .svg)\n", ext)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering output: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated Gantt chart: %s\n", outputPath)
}
