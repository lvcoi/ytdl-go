---
title: "Type Definitions"
weight: 30
---

# Type Definitions

This document provides comprehensive type definitions for the ytdl-go API, including request payloads, response structures, and event schemas.

## Table of Contents

- [Request Types](#request-types)
- [Response Types](#response-types)
- [SSE Event Types](#sse-event-types)
- [Common Types](#common-types)

## Request Types

### DownloadRequest

Payload for `POST /api/download`.

```typescript
interface DownloadRequest {
  urls: string[];
  options?: DownloadOptions;
}
```

#### Example

```json
{
  "urls": [
    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "https://www.youtube.com/watch?v=9bZkp7q19f0"
  ],
  "options": {
    "output": "{uploader} - {title}.{ext}",
    "audio": false,
    "quality": "best",
    "format": "mp4",
    "jobs": 2,
    "timeout": 180,
    "on-duplicate": "prompt"
  }
}
```

### DownloadOptions

Download configuration options.

```typescript
interface DownloadOptions {
  output?: string;           // Output filename template (default: "{title}.{ext}")
  audio?: boolean;           // Audio-only mode (default: false)
  quality?: string;          // Video/audio quality (default: "best")
  format?: string;           // Preferred container format (default: "")
  jobs?: number;             // Concurrent jobs (default: 1)
  timeout?: number;          // Timeout in seconds (default: 180)
  "on-duplicate"?: string;   // Duplicate file policy (default: "prompt")
}
```

#### Field Details

| Field | Type | Default | Constraints | Description |
|-------|------|---------|-------------|-------------|
| `output` | `string` | `{title}.{ext}` | No absolute paths, no `..` | Output filename template with variables |
| `audio` | `boolean` | `false` | - | Extract audio only (no video) |
| `quality` | `string` | `best` | See [Quality Values](#quality-values) | Video/audio quality preference |
| `format` | `string` | `""` | See [Format Values](#format-values) | Preferred container format |
| `jobs` | `number` | `1` | Min: 1 | Number of concurrent downloads |
| `timeout` | `number` | `180` | Min: 1 | Per-download timeout in seconds |
| `on-duplicate` | `string` | `prompt` | See [Duplicate Policies](#duplicate-policies) | How to handle existing files |

##### Quality Values

| Value | Description |
|-------|-------------|
| `best` | Best available quality |
| `worst` | Lowest available quality |
| `720p` | 720p video resolution |
| `1080p` | 1080p video resolution |
| `1440p` | 1440p video resolution |
| `2160p` | 4K video resolution |
| `128k` | 128 kbps audio bitrate |
| `192k` | 192 kbps audio bitrate |
| `256k` | 256 kbps audio bitrate |
| `320k` | 320 kbps audio bitrate |

##### Format Values

| Value | Description |
|-------|-------------|
| `mp4` | MP4 container (H.264 video) |
| `webm` | WebM container (VP9 video) |
| `m4a` | M4A container (AAC audio) |
| `mp3` | MP3 audio |
| `opus` | Opus audio |

##### Duplicate Policies

| Policy | Description |
|--------|-------------|
| `prompt` | Ask user for each duplicate file |
| `overwrite` | Overwrite existing files without asking |
| `skip` | Skip downloads if file exists |
| `rename` | Auto-rename with numeric suffix (e.g., `file(1).mp4`) |
| `prompt_all` | Ask once, apply to all duplicates |
| `overwrite_all` | Overwrite all duplicates without asking |
| `skip_all` | Skip all duplicates |
| `rename_all` | Auto-rename all duplicates |

### DuplicateResponse

Payload for `POST /api/download/duplicate-response`.

```typescript
interface DuplicateResponse {
  jobId: string;      // Job ID from download start
  promptId: string;   // Prompt ID from SSE duplicate event
  choice: string;     // User's decision
}
```

#### Choice Values

| Choice | Description |
|--------|-------------|
| `overwrite` | Overwrite this file |
| `skip` | Skip this download |
| `rename` | Auto-rename this file |
| `overwrite_all` | Overwrite this and all future duplicates |
| `skip_all` | Skip this and all future duplicates |
| `rename_all` | Rename this and all future duplicates |

#### Example

```json
{
  "jobId": "job_1",
  "promptId": "dup_2",
  "choice": "overwrite"
}
```

## Response Types

### DownloadStartResponse

Response from `POST /api/download`.

```typescript
interface DownloadStartResponse {
  status: string;   // Job status: "queued", "running", "complete", "error"
  jobId: string;    // Unique job identifier
  message: string;  // Human-readable status message
}
```

#### Example

```json
{
  "status": "queued",
  "jobId": "job_1",
  "message": "Download started for 2 item(s)."
}
```

### ErrorResponse

Generic error response.

```typescript
interface ErrorResponse {
  type: "error";
  status: "error";
  error: string;    // Error message
}
```

#### Example

```json
{
  "type": "error",
  "status": "error",
  "error": "invalid JSON payload"
}
```

### ServerStatusResponse

Response from `GET /api/status`.

```typescript
interface ServerStatusResponse {
  active_downloads: number;  // Number of currently running jobs
  uptime: string;            // Server uptime duration (e.g., "2m31s")
}
```

#### Example

```json
{
  "active_downloads": 1,
  "uptime": "2m31s"
}
```

### MediaListResponse

Response from `GET /api/media/`.

```typescript
interface MediaListResponse {
  items: MediaItem[];          // Array of media items
  next_offset: number | null;  // Next pagination offset, or null if no more
}
```

#### Example

```json
{
  "items": [
    {
      "id": "file.mp4",
      "title": "Rick Astley - Never Gonna Give You Up",
      "artist": "RickAstleyVEVO",
      "size": "12.5 MB",
      "date": "2026-02-07",
      "type": "video",
      "filename": "file.mp4"
    }
  ],
  "next_offset": null
}
```

### MediaItem

Individual media item in library.

```typescript
interface MediaItem {
  id: string;        // Unique identifier (filename)
  title: string;     // Media title
  artist: string;    // Artist/uploader name
  size: string;      // Human-readable file size (e.g., "12.5 MB")
  date: string;      // Download date (YYYY-MM-DD)
  type: string;      // Media type: "video" or "audio"
  filename: string;  // Actual filename on disk
}
```

### DuplicateResponseSuccess

Response from `POST /api/download/duplicate-response`.

```typescript
interface DuplicateResponseSuccess {
  status: "ok";
}
```

#### Example

```json
{
  "status": "ok"
}
```

## SSE Event Types

All SSE events are sent as JSON in the `data:` field.

### Common Envelope

Most events include these fields:

```typescript
interface EventEnvelope {
  type: string;    // Event type
  jobId: string;   // Job identifier
  seq: number;     // Sequence number (monotonically increasing)
  at: string;      // RFC3339 timestamp
}
```

### SnapshotEvent

Initial state snapshot sent on connection.

```typescript
interface SnapshotEvent extends EventEnvelope {
  type: "snapshot";
  status: JobStatus;
  tasks: TaskState[];
  logs: string[];
}
```

#### Example

```json
{
  "type": "snapshot",
  "jobId": "job_1",
  "seq": 0,
  "at": "2026-02-07T10:30:00Z",
  "status": "running",
  "tasks": [
    {
      "taskId": "task_1",
      "description": "Downloading video.mp4",
      "current": 1024000,
      "total": 5120000,
      "finished": false
    }
  ],
  "logs": [
    "Starting download...",
    "Extracting metadata..."
  ]
}
```

### StatusEvent

Job status change.

```typescript
interface StatusEvent extends EventEnvelope {
  type: "status";
  status: JobStatus;
  message: string;
}
```

#### Example

```json
{
  "type": "status",
  "jobId": "job_1",
  "seq": 1,
  "at": "2026-02-07T10:30:01Z",
  "status": "running",
  "message": "Download started"
}
```

### RegisterEvent

New task registered.

```typescript
interface RegisterEvent extends EventEnvelope {
  type: "register";
  taskId: string;
  description: string;
  total: number;  // Total bytes (0 if unknown)
}
```

#### Example

```json
{
  "type": "register",
  "jobId": "job_1",
  "seq": 2,
  "at": "2026-02-07T10:30:02Z",
  "taskId": "task_1",
  "description": "Downloading video.mp4",
  "total": 5120000
}
```

### ProgressEvent

Download progress update.

```typescript
interface ProgressEvent extends EventEnvelope {
  type: "progress";
  taskId: string;
  current: number;   // Bytes downloaded
  total: number;     // Total bytes
  percent: number;   // Percentage (0-100)
  speed: string;     // Human-readable speed (e.g., "512 KB/s")
  eta: string;       // Human-readable ETA (e.g., "8s")
}
```

#### Example

```json
{
  "type": "progress",
  "jobId": "job_1",
  "seq": 3,
  "at": "2026-02-07T10:30:03Z",
  "taskId": "task_1",
  "current": 1024000,
  "total": 5120000,
  "percent": 20.0,
  "speed": "512 KB/s",
  "eta": "8s"
}
```

### FinishEvent

Task completed.

```typescript
interface FinishEvent extends EventEnvelope {
  type: "finish";
  taskId: string;
  success: boolean;
  error: string | null;
}
```

#### Example

```json
{
  "type": "finish",
  "jobId": "job_1",
  "seq": 4,
  "at": "2026-02-07T10:30:10Z",
  "taskId": "task_1",
  "success": true,
  "error": null
}
```

### LogEvent

Log message.

```typescript
interface LogEvent extends EventEnvelope {
  type: "log";
  level: LogLevel;
  message: string;
}
```

#### Example

```json
{
  "type": "log",
  "jobId": "job_1",
  "seq": 5,
  "at": "2026-02-07T10:30:11Z",
  "level": "info",
  "message": "Download complete: video.mp4"
}
```

### DuplicateEvent

Duplicate file prompt.

```typescript
interface DuplicateEvent extends EventEnvelope {
  type: "duplicate";
  promptId: string;
  filename: string;
  message: string;
}
```

#### Example

```json
{
  "type": "duplicate",
  "jobId": "job_1",
  "seq": 6,
  "at": "2026-02-07T10:30:12Z",
  "promptId": "dup_1",
  "filename": "video.mp4",
  "message": "File already exists: video.mp4"
}
```

### DuplicateResolvedEvent

Duplicate prompt resolved.

```typescript
interface DuplicateResolvedEvent extends EventEnvelope {
  type: "duplicate-resolved";
  promptId: string;
  choice: string;
}
```

#### Example

```json
{
  "type": "duplicate-resolved",
  "jobId": "job_1",
  "seq": 7,
  "at": "2026-02-07T10:30:13Z",
  "promptId": "dup_1",
  "choice": "overwrite"
}
```

### DoneEvent

Job complete (terminal event).

```typescript
interface DoneEvent extends EventEnvelope {
  type: "done";
  status: JobStatus;
  message: string;
  exitCode: number;
  error: string | null;
  stats: JobStats;
}
```

#### Example

```json
{
  "type": "done",
  "jobId": "job_1",
  "seq": 8,
  "at": "2026-02-07T10:30:15Z",
  "status": "complete",
  "message": "All downloads completed successfully",
  "exitCode": 0,
  "error": null,
  "stats": {
    "total": 3,
    "success": 3,
    "failed": 0,
    "skipped": 0
  }
}
```

## Common Types

### JobStatus

Job execution status.

```typescript
type JobStatus = "queued" | "running" | "complete" | "error";
```

| Value | Description |
|-------|-------------|
| `queued` | Job is queued for execution |
| `running` | Job is currently executing |
| `complete` | Job completed successfully |
| `error` | Job failed with error |

### TaskState

State of a download task.

```typescript
interface TaskState {
  taskId: string;
  description: string;
  current: number;
  total: number;
  finished: boolean;
}
```

### JobStats

Job completion statistics.

```typescript
interface JobStats {
  total: number;     // Total items processed
  success: number;   // Successfully downloaded
  failed: number;    // Failed downloads
  skipped: number;   // Skipped downloads
}
```

### LogLevel

Log message severity.

```typescript
type LogLevel = "debug" | "info" | "warn" | "error";
```

| Value | Description |
|-------|-------------|
| `debug` | Debug information (verbose) |
| `info` | Informational message |
| `warn` | Warning message |
| `error` | Error message |

## Template Variables

Variables available in `output` template strings:

```typescript
interface TemplateVariables {
  title: string;      // Video title
  id: string;         // Video ID
  ext: string;        // File extension (mp4, m4a, etc.)
  uploader: string;   // Channel/uploader name
  date: string;       // Upload date (YYYYMMDD)
}
```

### Example Usage

```json
{
  "output": "{uploader} - {title} ({date}).{ext}"
}
```

**Result:** `RickAstleyVEVO - Never Gonna Give You Up (20091024).mp4`

## Validation Rules

### URL Validation

- Must be valid HTTP(S) URL
- Supports YouTube, direct video links, HLS/DASH manifests
- No file:// or other protocols

### Output Template Validation

- **Forbidden:** Absolute paths (e.g., `/tmp/file.mp4`)
- **Forbidden:** Path traversal (e.g., `../../../file.mp4`)
- **Allowed:** Relative paths with variables (e.g., `{uploader}/{title}.{ext}`)

### Field Constraints

| Field | Min | Max | Default |
|-------|-----|-----|---------|
| `jobs` | 1 | - | 1 |
| `timeout` | 1 | - | 180 |
| `offset` (pagination) | 0 | - | 0 |
| `limit` (pagination) | 1 | 500 | 200 |

## Related Documentation

- [Endpoints](endpoints) - REST API endpoints
- [Events](events) - Server-Sent Events details
- [Frontend Structure](../architecture/frontend-structure) - Frontend API integration
