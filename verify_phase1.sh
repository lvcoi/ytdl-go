#!/bin/bash
set -e

echo "=== Building ytdl-go ==="
go build -o bin/yt main.go

echo "=== Running Unit Tests ==="
go test ./internal/downloader/...

echo "=== Running Race Detector ==="
go test -race ./internal/downloader/...

echo "=== Running Smoke Test (Safe Video) ==="
# Me at the zoo (shortest, oldest, safe content)
VIDEO_URL="https://www.youtube.com/watch?v=jNQXAC9IVRw"

# Create a clear test output directory
mkdir -p test_output
rm -rf test_output/*

echo "Downloading video to test_output/..."
./bin/yt --output-dir test_output --quality 360p "$VIDEO_URL"

if [ -f "test_output/Me at the zoo.mp4" ]; then
    echo "✅ Video downloaded successfully!"
else
    echo "❌ Video download failed!"
    exit 1
fi

echo "=== Phase 1 Verification Complete ==="
