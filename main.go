package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.md> <output.html|output.svg>\n", os.Args[0])
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	fmt.Printf("Input: %s\nOutput: %s\n", inputPath, outputPath)
	fmt.Println("Gantt chart generator - coming soon!")
}
