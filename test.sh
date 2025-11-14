#!/bin/bash
set -e

echo "Running unit tests..."
go test ./... -v

echo ""
echo "Building binary..."
go build -o gantt-gen

echo ""
echo "Testing with example file..."
./gantt-gen examples/sample-project.md examples/test-output.svg

echo ""
echo "Checking output file exists..."
test -f examples/test-output.svg && echo "✓ SVG output created"

echo ""
echo "All tests passed!"
