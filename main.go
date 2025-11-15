package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"gantt-gen/parser"
	"gantt-gen/renderer"
	"gantt-gen/resolver"
)

func main() {
	// Define flags
	format := flag.String("format", "svg", "Output format: svg, html, or confluence")
	flag.Parse()

	// Check remaining arguments
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--format=svg|html|confluence] <input.md|-> <output-file|->\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Use '-' for stdin (input) or stdout (output)\n")
		os.Exit(1)
	}

	inputPath := args[0]
	outputPath := args[1]

	// Validate format
	outputFormat := strings.ToLower(*format)
	if outputFormat != "svg" && outputFormat != "html" && outputFormat != "confluence" {
		fmt.Fprintf(os.Stderr, "Error: Invalid format '%s'. Use 'svg', 'html', or 'confluence'\n", outputFormat)
		os.Exit(1)
	}

	// Read input file or stdin
	var input []byte
	if inputPath == "-" {
		var err error
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	} else {
		var err error
		input, err = os.ReadFile(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
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

	// Generate output based on format
	var output string
	switch outputFormat {
	case "html":
		output, err = renderer.RenderHTML(project)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering HTML: %v\n", err)
			os.Exit(1)
		}
	case "confluence":
		output, err = renderer.RenderConfluence(project)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering Confluence: %v\n", err)
			os.Exit(1)
		}
	default: // svg
		output, err = renderer.RenderSVG(project)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering SVG: %v\n", err)
			os.Exit(1)
		}
	}

	// Write output file or stdout
	if outputPath == "-" {
		if _, err := os.Stdout.WriteString(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "✓ Generated Gantt chart (%s): %s\n", outputFormat, outputPath)
	}
}
