---
title: "Backend Development"
weight: 20
---

# Backend Development

This guide covers Go-specific development practices for ytdl-go's backend.

## Table of Contents

- [Go Standards](#go-standards)
- [Testing](#testing)
- [Common Tasks](#common-tasks)
- [Debugging](#debugging)
- [Best Practices](#best-practices)

## Go Standards

### Code Formatting

Use `gofmt` for consistent formatting:

```bash
# Check formatting
gofmt -l .

# Format all files
gofmt -w .

# Or use go fmt
go fmt ./...
```

### Code Vetting

Use `go vet` to catch common mistakes:

```bash
go vet ./...
```

### Linting

Use `golangci-lint` for comprehensive linting:

```bash
# Run linter
golangci-lint run

# Auto-fix issues where possible
golangci-lint run -fix
```

### Pre-Commit Checklist

Before committing, ensure:
- [ ] Code is formatted: `go fmt ./...`
- [ ] No vet warnings: `go vet ./...`
- [ ] Linter passes: `golangci-lint run`
- [ ] Tests pass: `go test ./...`
- [ ] Binary builds: `go build .`

## Testing

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test -v ./internal/downloader

# Run specific test
go test -v -run TestFunctionName ./internal/downloader
```

### Writing Tests

Tests belong in `*_test.go` files:

```go
// internal/downloader/path_test.go
package downloader

import "testing"

func TestSanitizeFilename(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"basic", "file.mp4", "file.mp4"},
        {"with spaces", "my file.mp4", "my file.mp4"},
        {"invalid chars", "file<>?.mp4", "file___.mp4"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := SanitizeFilename(tt.input)
            if result != tt.expected {
                t.Errorf("expected %q, got %q", tt.expected, result)
            }
        })
    }
}
```

### Test Coverage

Generate coverage reports:

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

### Benchmarking

Write benchmarks for performance-critical code:

```go
func BenchmarkSanitizeFilename(b *testing.B) {
    input := "test<>file?.mp4"
    for i := 0; i < b.N; i++ {
        SanitizeFilename(input)
    }
}
```

Run benchmarks:

```bash
go test -bench=. ./...
```

## Common Tasks

### Adding a New Download Strategy

1. **Define the strategy** in `downloader.go`:
   ```go
   type Strategy int
   
   const (
       StrategyStandard Strategy = iota
       StrategyRetry
       StrategyFFmpeg
       StrategyYourNew  // Add here
   )
   ```

2. **Implement the download logic** in appropriate file (e.g., `youtube.go`):
   ```go
   func downloadWithYourNew(ctx context.Context, video *Video) error {
       // Implementation
   }
   ```

3. **Update strategy selection** in `downloader.go`:
   ```go
   func selectStrategy(format Format) Strategy {
       // Add logic to select your new strategy
   }
   ```

4. **Test thoroughly** with various video types

### Adding a New CLI Flag

1. **Define the flag** in `main.go`:
   ```go
   var myNewFlag = flag.String("my-flag", "default", "Description")
   ```

2. **Pass to options struct**:
   ```go
   opts := app.Options{
       MyNewFlag: *myNewFlag,
   }
   ```

3. **Use in implementation** (e.g., `runner.go` or `downloader.go`)

4. **Update documentation** in `docs/FLAGS.md`

### Adding a New Output Template Variable

1. **Define variable** in `output.go`:
   ```go
   func expandTemplate(template string, video *Video) string {
       replacements := map[string]string{
           "{title}":    video.Title,
           "{id}":       video.ID,
           "{mynew}":    video.MyNew,  // Add here
       }
       // ...
   }
   ```

2. **Ensure metadata extraction** populates the field

3. **Update template documentation**

### Modifying Progress Display

Progress display is coordinated by `ProgressManager` and rendered by `unified_tui.go`.

**To change progress bar appearance:**
- Modify `unified_tui.go` rendering logic
- Adjust styles in lipgloss definitions

**To add new progress metrics:**
- Update `ProgressState` struct in `progress_manager.go`
- Modify `Update()` to accept new metrics
- Update rendering in `unified_tui.go`

## Debugging

### Enable Debug Logging

```bash
ytdl-go -log-level debug [URL]
```

### Use Delve Debugger

Install Delve:
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Debug a test:
```bash
dlv test ./internal/downloader -- -test.run TestSanitizeFilename
```

Debug the binary:
```bash
dlv debug . -- -info https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

### Print Debugging

Use `fmt.Printf` or the logger:

```go
// Temporary debug print
fmt.Printf("DEBUG: value=%v\n", myValue)

// Using logger (if available)
log.Debugf("Processing video: %s", video.ID)
```

> **Note:** Remove debug prints before committing!

### Inspect Network Requests

The codebase uses standard Go `net/http`. Add debug prints to `http.go`:

```go
// In http.go
resp, err := client.Do(req)
fmt.Printf("Request: %s %s\n", req.Method, req.URL)
fmt.Printf("Response: %d\n", resp.StatusCode)
```

### Test with Specific Videos

For testing, use:
- **Public videos** - Avoid copyrighted content
- **Short videos** - Speed up testing
- **Various formats** - Test different scenarios

Example test videos:
```bash
# Standard video
ytdl-go -info https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Audio-only
ytdl-go -audio https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Playlist
ytdl-go -info https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf
```

## Best Practices

### 1. Always Use Contexts

Pass `context.Context` through the call chain:

```go
func Download(ctx context.Context, url string) error {
    return downloadWithContext(ctx, url)
}

func downloadWithContext(ctx context.Context, url string) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process
    }
}
```

### 2. Defer Cleanup

Always defer cleanup operations:

```go
f, err := os.Create(tempFile)
if err != nil {
    return err
}
defer os.Remove(tempFile)  // Cleanup on error
defer f.Close()            // Always close file
```

### 3. Check Cancellation in Loops

Check context cancellation in long-running operations:

```go
for _, item := range largeList {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        process(item)
    }
}
```

### 4. Use Atomic File Operations

Write to temp files and atomically rename:

```go
// Write to temp file
tempFile := outputFile + ".tmp"
if err := writeToFile(tempFile, data); err != nil {
    return err
}

// Atomic rename
return os.Rename(tempFile, outputFile)
```

### 5. Thread-Safe Progress Updates

Use `ProgressManager` for thread-safe updates:

```go
// Don't: Direct shared state mutation
// progressMap[id] = current  // Race condition!

// Do: Use ProgressManager
pm.Update(id, current, total)
```

### 6. Handle Errors Gracefully

Wrap errors with context:

```go
if err := downloadFile(url); err != nil {
    return fmt.Errorf("failed to download %s: %w", url, err)
}
```

Categorize errors for appropriate handling:

```go
if isNetworkError(err) {
    return ErrorTypeNetwork
} else if isPermissionError(err) {
    return ErrorTypePermission
}
```

### 7. Document Exported Functions

Add comments to all exported functions and types:

```go
// SanitizeFilename removes or replaces invalid characters from a filename.
// It ensures the resulting filename is valid on all major operating systems.
func SanitizeFilename(filename string) string {
    // Implementation
}
```

### 8. Keep Functions Focused

Each function should do one thing well:

```go
// Bad: Does too much
func downloadAndProcessAndSave(url string) error {
    // 100+ lines of mixed concerns
}

// Good: Separate concerns
func download(url string) ([]byte, error) { ... }
func process(data []byte) ([]byte, error) { ... }
func save(data []byte, path string) error { ... }
```

### 9. Use Constants for Magic Values

```go
// Bad
if retries > 3 { ... }

// Good
const maxRetries = 3
if retries > maxRetries { ... }
```

### 10. Test Edge Cases

Consider edge cases in tests:
- Empty inputs
- Nil values
- Very large values
- Unicode characters
- Network failures

## Common Issues

### Issue: Import Cycle

**Error:** `import cycle not allowed`

**Solution:** Refactor to break the cycle:
- Move shared types to a separate package
- Use interfaces to decouple packages
- Reconsider package boundaries

### Issue: 403 Errors in Tests

**Problem:** YouTube API changes frequently

**Solution:**
- Update `kkdai/youtube` library: `go get github.com/kkdai/youtube/v2@latest`
- Check upstream issues: https://github.com/kkdai/youtube/issues
- Implement retry logic in tests

### Issue: Race Conditions

**Detection:** Run tests with race detector:
```bash
go test -race ./...
```

**Solution:**
- Use mutexes for shared state
- Prefer channels for communication
- Use atomic operations for counters

## Related Documentation

- [Getting Started](getting-started) - Development environment setup
- [Code Style](code-style) - Go coding standards
- [Backend Structure](../architecture/backend-structure) - Architecture deep dive
- [Pull Request Process](pull-request-process) - PR workflow
