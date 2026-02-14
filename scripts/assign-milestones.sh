#!/bin/bash
# Wrapper script for the milestone assignment tool

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.24 or higher."
    exit 1
fi

# Check if GITHUB_TOKEN is set
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Error: GITHUB_TOKEN environment variable is not set."
    echo "Please set it with: export GITHUB_TOKEN='your_token_here'"
    exit 1
fi

# Build if binary doesn't exist
if [ ! -f "assign-milestones" ]; then
    echo "Building assign-milestones..."
    go build -o assign-milestones assign-milestones.go
fi

# Run the tool
./assign-milestones "$@"
