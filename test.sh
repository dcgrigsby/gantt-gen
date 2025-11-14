#!/bin/bash
set -e

echo "Running unit tests..."
go test ./... -v

echo ""
echo "Building binary..."
go build -o gantt-gen

echo ""
echo "Testing with example file..."
./gantt-gen examples/sample-project.md examples/test-output.html
./gantt-gen examples/sample-project.md examples/test-output.svg

echo ""
echo "Checking output files exist..."
test -f examples/test-output.html && echo "✓ HTML output created"
test -f examples/test-output.svg && echo "✓ SVG output created"

echo ""
echo "All tests passed!"
