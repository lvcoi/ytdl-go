---
title: "Backend Structure"
weight: 20
---

# Backend Structure

This document provides a deep dive into ytdl-go's Go backend architecture, focusing on module organization, concurrency patterns, and error handling strategies.

## Table of Contents

- [Module Organization](#module-organization)
- [Concurrency Patterns](#concurrency-patterns)
- [Error Handling Strategy](#error-handling-strategy)
- [Key Data Structures](#key-data-structures)
- [Download Pipeline](#download-pipeline)

## Module Organization

The backend follows Go's standard project layout with all implementation code in the `internal/` directory to prevent external imports.

### Package Structure

```
internal/
├── app/                    # Application orchestration layer
│   └── runner.go           # Coordinates workflow and concurrency
│
└── downloader/             # Core download implementation
    ├── downloader.go       # Main download logic and strategy selection
    ├── youtube.go          # YouTube-specific extraction and downloads
    ├── direct.go           # Direct URL download handling
    ├── segment_downloader.go  # HLS/DASH segment operations
    ├── unified_tui.go      # Terminal UI coordination
    ├── progress_manager.go # Concurrent progress tracking
    ├── output.go           # File I/O and path management
    ├── metadata.go         # Metadata extraction and sidecar JSON
    ├── tags.go             # Audio tag embedding (ID3)
    ├── prompt.go           # User interaction prompts
    ├── path.go             # Path sanitization and validation
    ├── http.go             # HTTP client configuration
    ├── errors.go           # Error categorization and exit codes
    └── music.go            # YouTube Music URL handling
```

### Module Responsibilities

#### `internal/app/runner.go`

The **orchestration layer** that coordinates the entire download workflow:

- Parses command-line flags and validates inputs
- Manages job-level concurrency (`-jobs N`)
- Coordinates playlist expansion and parallel downloads
- Initializes progress tracking and TUI
- Handles graceful shutdown on Ctrl+C

**Key Functions:**
- `Run(ctx context.Context, opts Options) error` - Main entry point
- `processURL(ctx context.Context, url string) error` - Per-URL processing
- `handlePlaylist(ctx context.Context, playlist []Entry) error` - Playlist coordination

#### `internal/downloader/downloader.go`

The **core download engine** that implements strategy selection:

- Determines download strategy based on URL type
- Implements retry logic and error handling
- Coordinates between metadata extraction and file writing
- Manages download lifecycle

**Key Functions:**
- `Download(ctx context.Context, url string, opts DownloadOptions) error`
- `selectBestFormat(formats []Format, opts Options) Format`
- `determineStrategy(format Format) Strategy`

#### `internal/downloader/youtube.go`

**YouTube-specific implementation**:

- Extracts video/playlist metadata using `kkdai/youtube` library
- Implements chunked download with progress tracking
- Handles YouTube API quirks and rate limiting
- Implements FFmpeg fallback for audio-only downloads

**Key Functions:**
- `ExtractMetadata(ctx context.Context, url string) (*Video, error)`
- `DownloadYouTube(ctx context.Context, video *Video, format Format) error`
- `downloadWithFFmpegFallback(ctx context.Context, video *Video) error`

**Download Strategy Chain:**
1. Standard chunked download (`Client.GetStreamContext()`)
2. Single-request retry (on 403 error)
3. FFmpeg fallback (audio-only, requires FFmpeg)

#### `internal/downloader/direct.go`

**Direct URL downloads** (non-YouTube):

- Handles plain HTTP/HTTPS URLs
- Detects HLS/DASH manifests
- Delegates to segment downloader when needed
- Simple HTTP GET for regular files

**Supported Formats:**
- MP4, WebM, MKV (direct download)
- HLS (`.m3u8` playlists)
- DASH (`.mpd` manifests)

#### `internal/downloader/segment_downloader.go`

**Parallel segment downloads** for HLS/DASH streams:

- Parses HLS manifests (`.m3u8`)
- Downloads segments concurrently
- Concatenates segments in order
- Provides per-segment progress tracking

**Concurrency Control:**
- `-segment-concurrency N` flag
- Worker pool pattern
- Preserves segment order during concatenation

#### `internal/downloader/unified_tui.go`

**Terminal UI coordinator** using Bubble Tea:

- Provides format selection interface
- Shows real-time progress bars
- Handles interactive prompts
- Implements seamless view transitions

**View Types:**
- `SeamlessViewFormatSelector` - Interactive format browser
- `SeamlessViewProgress` - Multi-download progress display

**Message Types:**
- `seamlessRegisterMsg` - Register new download task
- `seamlessUpdateMsg` - Update progress
- `seamlessFinishMsg` - Mark task complete
- `seamlessLogMsg` - Display log message
- `seamlessPromptMsg` - Interactive prompt

#### `internal/downloader/progress_manager.go`

**Concurrent progress coordination**:

- Thread-safe progress tracking across goroutines
- Manages multiple simultaneous progress bars
- Routes updates to TUI via channels
- Prevents race conditions on shared state

**Thread Safety Mechanisms:**
- Mutex-protected internal state
- Atomic counters for progress values
- Channel-based message passing to TUI
- No shared mutable state exposed to callers

#### `internal/downloader/output.go`

**File I/O and path management**:

- Expands output templates (`{title}`, `{ext}`, etc.)
- Handles file conflict resolution
- Implements atomic file writes
- Manages temp file cleanup

**Template Variables:**
- `{title}` - Video title
- `{ext}` - File extension
- `{id}` - Video ID
- `{uploader}` - Channel name
- `{date}` - Upload date

#### `internal/downloader/metadata.go`

**Metadata extraction and persistence**:

- Extracts video/audio metadata
- Writes sidecar JSON files
- Formats metadata for embedding
- Handles metadata cleanup

**Sidecar JSON Structure:**
```json
{
  "id": "video_id",
  "title": "Video Title",
  "uploader": "Channel Name",
  "duration": 300,
  "upload_date": "2024-01-01",
  "formats": [...]
}
```

#### `internal/downloader/tags.go`

**Audio tag embedding**:

- Embeds ID3v2 tags in MP3 files
- Supports title, artist, album, year
- Handles cover art embedding
- Logs unsupported formats

**Supported Tags:**
- TIT2 - Title
- TPE1 - Artist
- TALB - Album
- TYER - Year
- APIC - Cover art

#### `internal/downloader/prompt.go`

**Interactive user prompts**:

- File conflict resolution (overwrite/skip/rename/quit)
- TTY detection for automatic fallback
- Clear, actionable prompt messages
- Applies "all" choices to remaining files

**Prompt Actions:**
- Overwrite - Replace existing file
- Skip - Skip this download
- Rename - Auto-rename with suffix
- Quit - Abort all remaining downloads

#### `internal/downloader/path.go`

**Path sanitization and validation**:

- Sanitizes filenames (removes invalid characters)
- Prevents path traversal attacks
- Implements auto-rename logic
- Validates output directories

**Sanitization Rules:**
- Remove/replace invalid filename characters
- Prevent absolute paths in templates
- Block `..` directory traversal
- Limit filename length

## Concurrency Patterns

### 1. Worker Pool Pattern

Used for job-level and playlist-level concurrency:

```go
// Worker pool with semaphore pattern
sem := make(chan struct{}, maxWorkers)

for _, url := range urls {
    sem <- struct{}{}  // Acquire
    go func(u string) {
        defer func() { <-sem }()  // Release
        processURL(ctx, u)
    }(url)
}

// Wait for all workers
for i := 0; i < maxWorkers; i++ {
    sem <- struct{}{}
}
```

### 2. Fan-Out, Fan-In Pattern

Used for segment downloads:

```go
// Fan-out: Launch workers
results := make(chan SegmentResult, len(segments))
for _, seg := range segments {
    go func(s Segment) {
        data, err := downloadSegment(ctx, s)
        results <- SegmentResult{s.Index, data, err}
    }(seg)
}

// Fan-in: Collect results in order
for i := 0; i < len(segments); i++ {
    result := <-results
    // Process in order
}
```

### 3. Message Passing Pattern

Used for TUI updates:

```go
// Producer (download goroutine)
pm.Update(taskID, current, total)

// Consumer (TUI)
for msg := range updateChan {
    switch m := msg.(type) {
    case seamlessUpdateMsg:
        updateProgressBar(m)
    }
}
```

### 4. Context-Based Cancellation

Used throughout for graceful shutdown:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Listen for Ctrl+C
go func() {
    <-sigChan
    cancel()  // Triggers cleanup in all goroutines
}()

// In worker goroutines
select {
case <-ctx.Done():
    return ctx.Err()
case data := <-workChan:
    // Process work
}
```

## Error Handling Strategy

### Error Categorization

Errors are categorized by type for appropriate handling:

```go
// Error types (defined in errors.go)
type ErrorType int

const (
    ErrorTypeNetwork      ErrorType = iota  // Network failures
    ErrorTypePermission                     // File permission errors
    ErrorTypeNotFound                       // URL/file not found
    ErrorTypeInvalidInput                   // Invalid user input
    ErrorTypeExtraction                     // Metadata extraction failed
    ErrorTypeDownload                       // Download failed
    ErrorTypeFFmpeg                         // FFmpeg errors
)
```

### Exit Codes

Different exit codes for different failure types:

| Exit Code | Meaning | Example |
|-----------|---------|---------|
| 0 | Success | All downloads completed |
| 1 | General error | Unexpected failure |
| 2 | Invalid input | Malformed URL |
| 3 | Network error | Connection timeout |
| 4 | Permission error | Cannot write file |
| 5 | Not found | Video unavailable |
| 6 | Extraction error | Cannot parse metadata |
| 7 | Download error | 403 Forbidden |
| 8 | FFmpeg error | FFmpeg not found |

### Error Wrapping

Errors are wrapped with context for better debugging:

```go
if err := downloadFile(url); err != nil {
    return fmt.Errorf("failed to download %s: %w", url, err)
}
```

### Retry Logic

Automatic retries for transient failures:

```go
const maxRetries = 3

for attempt := 1; attempt <= maxRetries; attempt++ {
    err := tryDownload(ctx, url)
    if err == nil {
        return nil
    }
    
    if !isRetryable(err) {
        return err
    }
    
    if attempt < maxRetries {
        time.Sleep(backoff(attempt))
    }
}
```

### Graceful Degradation

When optional features fail, the tool continues:

```go
// FFmpeg not available - log but continue
if !ffmpegAvailable() {
    log.Warn("FFmpeg not found, audio fallback unavailable")
    // Continue without FFmpeg
}

// Tag embedding fails - log but continue
if err := embedTags(file, metadata); err != nil {
    log.Warnf("Failed to embed tags: %v", err)
    // File still downloaded successfully
}
```

## Key Data Structures

### Video Struct

Represents extracted video metadata:

```go
type Video struct {
    ID          string
    Title       string
    Author      string
    Duration    time.Duration
    Formats     []Format
    Thumbnails  []Thumbnail
    Description string
}
```

### Format Struct

Represents a downloadable format:

```go
type Format struct {
    ITag      int
    Quality   string
    MimeType  string
    Bitrate   int
    URL       string
    Width     int
    Height    int
    FPS       int
    AudioBitrate int
    AudioCodec   string
}
```

### DownloadOptions Struct

Configuration for downloads:

```go
type DownloadOptions struct {
    OutputTemplate   string
    Quality          string
    AudioOnly        bool
    Format           string
    Jobs             int
    Timeout          time.Duration
    OnDuplicate      DuplicatePolicy
    WriteMetadata    bool
    EmbedThumbnail   bool
}
```

### ProgressState Struct

Tracks download progress:

```go
type ProgressState struct {
    TaskID       string
    Description  string
    Current      int64
    Total        int64
    StartTime    time.Time
    LastUpdate   time.Time
    Finished     bool
}
```

## Download Pipeline

### Complete Download Flow

```
1. Parse URL
   ├─→ Detect type (YouTube, playlist, direct)
   └─→ Validate format

2. Extract Metadata
   ├─→ Fetch video info
   ├─→ Parse available formats
   └─→ Select best format

3. Prepare Output
   ├─→ Expand output template
   ├─→ Check for duplicates
   └─→ Resolve conflicts

4. Download Content
   ├─→ Select strategy (chunked/direct/segments)
   ├─→ Create progress tracker
   ├─→ Stream to temp file
   └─→ Handle errors (retry/fallback)

5. Post-Process
   ├─→ Move temp file to final location
   ├─→ Write sidecar metadata
   ├─→ Embed tags (audio only)
   └─→ Cleanup temp files

6. Report Status
   ├─→ Update progress bar
   ├─→ Log completion
   └─→ Return result
```

### Error Recovery Flow

```
Download Attempt
   ├─→ Success → Continue to post-processing
   │
   ├─→ 403 Error → Retry with single request
   │   ├─→ Success → Continue
   │   └─→ Still 403 + Audio → FFmpeg fallback
   │       ├─→ Success → Continue
   │       └─→ Failure → Report error
   │
   ├─→ Network Error → Retry with backoff
   │   ├─→ Success → Continue
   │   └─→ Max retries → Report error
   │
   └─→ Other Error → Report immediately
```

## Go Module Dependencies

Core dependencies and their purposes:

| Dependency | Purpose | Update Priority |
|------------|---------|----------------|
| `github.com/kkdai/youtube/v2` | YouTube API client | **Critical** |
| `github.com/charmbracelet/bubbletea` | TUI framework | Medium |
| `github.com/charmbracelet/bubbles` | TUI components | Medium |
| `github.com/charmbracelet/lipgloss` | TUI styling | Low |
| `github.com/bogem/id3v2/v2` | ID3 tag embedding | Low |
| `github.com/u2takey/ffmpeg-go` | FFmpeg bindings | Medium |

> **Note:** The `kkdai/youtube` library is the most critical dependency. YouTube frequently changes their API, so this library requires regular updates when downloads start failing.

## Best Practices

### 1. Always Use Contexts

Pass context through the call chain for cancellation:

```go
func Download(ctx context.Context, url string) error {
    return downloadWithContext(ctx, url)
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
defer f.Close()
```

### 3. Check Cancellation

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

### 4. Atomic File Operations

Use temp files and atomic renames:

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

Use channels or the ProgressManager for updates:

```go
// Don't: Direct shared state mutation
// progressMap[id] = current  // Race condition!

// Do: Use ProgressManager
pm.Update(id, current, total)
```

## Related Documentation

- [Architecture Overview](overview) - High-level system design
- [Frontend Structure](frontend-structure) - Web UI architecture
- [Contributing Guide](../../contributing/backend) - Backend development guide
