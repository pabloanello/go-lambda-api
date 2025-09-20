#!/bin/bash

set -e

echo "Generating coverage profiles..."

# Generate a single coverage profile for all packages
go test -v -timeout 5m -coverprofile=coverage.out -covermode=atomic ./...

echo "Combined coverage written to coverage.out"

# Generate HTML report
echo "Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "Coverage report generated: coverage.html"
