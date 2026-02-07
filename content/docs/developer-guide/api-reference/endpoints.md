---
title: "REST API Endpoints"
weight: 10
---

# REST API Endpoints

This document defines the REST API endpoints provided by the ytdl-go backend server.

**Base URL:** `/api`

## Table of Contents

- [Start Download](#start-download)
- [Duplicate Prompt Response](#duplicate-prompt-response)
- [Server Status](#server-status)
- [Media Listing](#media-listing)
- [Media File Serve](#media-file-serve)

## Start Download

Starts an async download job.

### Endpoint

```
POST /api/download
```

### Request Headers

- `Content-Type: application/json`

### Request Body Limits

- **Max Body Size:** 1 MiB

### Request Schema

```json
{
  "urls": ["https://www.youtube.com/watch?v=dQw4w9WgXcQ"],
  "options": {
    "output": "{title}.{ext}",
    "audio": false,
    "quality": "best",
    "format": "",
    "jobs": 1,
    "timeout": 180,
    "on-duplicate": "prompt"
  }
}
```

### Request Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `urls` | `string[]` | **Yes** | - | List of URLs to download |
| `options` | `object` | No | See below | Download configuration options |

#### Options Object

| Field | Type | Default | Constraints | Description |
|-------|------|---------|-------------|-------------|
| `output` | `string` | `{title}.{ext}` | No absolute paths or `..` | Output filename template |
| `audio` | `boolean` | `false` | - | Audio-only mode |
| `quality` | `string` | `best` | `best`, `worst`, `720p`, `1080p`, `128k`, etc. | Video/audio quality |
| `format` | `string` | `""` | `mp4`, `webm`, `m4a`, etc. | Preferred container format |
| `jobs` | `number` | `1` | Min: 1 | Concurrent download jobs |
| `timeout` | `number` | `180` | Min: 1 | Timeout in seconds per download |
| `on-duplicate` | `string` | `prompt` | See [Duplicate Policies](#duplicate-policies) | Duplicate file handling |

##### Duplicate Policies

| Policy | Description |
|--------|-------------|
| `prompt` | Ask user for each duplicate |
| `overwrite` | Overwrite existing files |
| `skip` | Skip downloading duplicates |
| `rename` | Auto-rename with suffix |
| `prompt_all` | Apply choice to all duplicates |
| `overwrite_all` | Overwrite all duplicates |
| `skip_all` | Skip all duplicates |
| `rename_all` | Auto-rename all duplicates |

##### Output Template Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `{title}` | Video title | `Rick Astley - Never Gonna Give You Up` |
| `{id}` | Video ID | `dQw4w9WgXcQ` |
| `{ext}` | File extension | `mp4`, `m4a` |

### Success Response

**Status Code:** `200 OK`

```json
{
  "status": "queued",
  "jobId": "job_1",
  "message": "Download started for 1 item(s)."
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `status` | `string` | Job status: `queued`, `running`, `complete`, `error` |
| `jobId` | `string` | Unique job identifier for progress tracking |
| `message` | `string` | Human-readable status message |

### Error Responses

#### 400 Bad Request

Invalid request payload:

```json
{
  "type": "error",
  "status": "error",
  "error": "invalid JSON payload"
}
```

**Common Causes:**
- Malformed JSON
- Missing required fields
- Invalid field types
- Absolute paths in output template
- Path traversal attempts (`..` in output)

#### 413 Payload Too Large

Request body exceeds 1 MiB:

```json
{
  "type": "error",
  "status": "error",
  "error": "request body too large"
}
```

#### 500 Internal Server Error

Server-side error:

```json
{
  "type": "error",
  "status": "error",
  "error": "internal server error"
}
```

### Example Requests

#### Basic Video Download

```bash
curl -X POST http://localhost:8080/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["https://www.youtube.com/watch?v=dQw4w9WgXcQ"]
  }'
```

#### Audio-Only Download

```bash
curl -X POST http://localhost:8080/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["https://www.youtube.com/watch?v=dQw4w9WgXcQ"],
    "options": {
      "audio": true,
      "quality": "128k",
      "output": "{uploader} - {title}.{ext}"
    }
  }'
```

#### Multiple URLs with Custom Options

```bash
curl -X POST http://localhost:8080/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
      "https://www.youtube.com/watch?v=9bZkp7q19f0"
    ],
    "options": {
      "quality": "720p",
      "format": "mp4",
      "jobs": 2,
      "on-duplicate": "rename"
    }
  }'
```

---

## Duplicate Prompt Response

Submits a duplicate-file decision for a pending prompt.

### Endpoint

```
POST /api/download/duplicate-response
```

### Request Headers

- `Content-Type: application/json`

### Request Schema

```json
{
  "jobId": "job_1",
  "promptId": "dup_2",
  "choice": "overwrite"
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `jobId` | `string` | **Yes** | Job identifier from download start response |
| `promptId` | `string` | **Yes** | Prompt identifier from SSE `duplicate` event |
| `choice` | `string` | **Yes** | User's decision (see [Choices](#choices)) |

#### Choices

| Choice | Description |
|--------|-------------|
| `overwrite` | Overwrite the existing file |
| `skip` | Skip this download |
| `rename` | Auto-rename with suffix |
| `overwrite_all` | Overwrite this and all future duplicates |
| `skip_all` | Skip this and all future duplicates |
| `rename_all` | Rename this and all future duplicates |

### Success Response

**Status Code:** `200 OK`

```json
{
  "status": "ok"
}
```

### Error Responses

#### 400 Bad Request

```json
{
  "error": "invalid choice"
}
```

#### 404 Not Found

```json
{
  "error": "prompt not found"
}
```

### Example Request

```bash
curl -X POST http://localhost:8080/api/download/duplicate-response \
  -H "Content-Type: application/json" \
  -d '{
    "jobId": "job_1",
    "promptId": "dup_2",
    "choice": "overwrite"
  }'
```

---

## Server Status

Returns server health and active job information.

### Endpoint

```
GET /api/status
```

### Request Parameters

None

### Success Response

**Status Code:** `200 OK`

```json
{
  "active_downloads": 1,
  "uptime": "2m31s"
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `active_downloads` | `number` | Number of currently running jobs |
| `uptime` | `string` | Server uptime duration |

### Example Request

```bash
curl http://localhost:8080/api/status
```

---

## Media Listing

Lists downloaded media files with pagination.

### Endpoint

```
GET /api/media/
```

### Query Parameters

| Parameter | Type | Default | Max | Description |
|-----------|------|---------|-----|-------------|
| `offset` | `number` | `0` | - | Pagination offset |
| `limit` | `number` | `200` | `500` | Number of items to return |

### Success Response

**Status Code:** `200 OK`

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
    },
    {
      "id": "audio.m4a",
      "title": "Sample Audio",
      "artist": "Unknown Artist",
      "size": "3.2 MB",
      "date": "2026-02-06",
      "type": "audio",
      "filename": "audio.m4a"
    }
  ],
  "next_offset": null
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `items` | `object[]` | Array of media items |
| `next_offset` | `number` \| `null` | Next pagination offset, or `null` if no more results |

#### Media Item Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Unique item identifier (filename) |
| `title` | `string` | Media title |
| `artist` | `string` | Artist/uploader name |
| `size` | `string` | Human-readable file size |
| `date` | `string` | Download date (YYYY-MM-DD) |
| `type` | `string` | Media type: `video` or `audio` |
| `filename` | `string` | Actual filename on disk |

### Example Requests

#### Get First Page

```bash
curl http://localhost:8080/api/media/
```

#### Get Next Page

```bash
curl http://localhost:8080/api/media/?offset=200&limit=200
```

#### Get Small Batch

```bash
curl http://localhost:8080/api/media/?limit=50
```

---

## Media File Serve

Serves a media file from the media directory.

### Endpoint

```
GET /api/media/{filename}
```

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `filename` | `string` | Name of the media file to serve |

### Security

- **Path traversal prevention** - `..` sequences rejected
- **Symlink protection** - Symlinks outside media directory blocked
- **Directory listing disabled** - Only files can be served

### Success Response

**Status Code:** `200 OK`

**Content-Type:** Varies by file type
- Video: `video/mp4`, `video/webm`, etc.
- Audio: `audio/mp4`, `audio/mpeg`, etc.

**Response Body:** File contents (streamed)

### Error Responses

#### 400 Bad Request

Invalid filename (path traversal attempt):

```json
{
  "error": "invalid filename"
}
```

#### 404 Not Found

File not found:

```json
{
  "error": "file not found"
}
```

#### 403 Forbidden

Symlink escape attempt:

```json
{
  "error": "access denied"
}
```

### Example Request

```bash
# Stream video
curl http://localhost:8080/api/media/video.mp4 -o video.mp4

# Serve in HTML5 player
<video src="/api/media/video.mp4" controls></video>

# Serve in audio player
<audio src="/api/media/audio.m4a" controls></audio>
```

---

## Related Documentation

- [Events](events) - Server-Sent Events for real-time updates
- [Types](types) - Complete type definitions
- [Frontend Structure](../architecture/frontend-structure) - Frontend API integration
