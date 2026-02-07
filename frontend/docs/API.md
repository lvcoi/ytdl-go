# Backend API Reference

This document defines the JSON contract between the `ytdl-go` frontend and the Go backend server.

**Base URL:** `/api`

## 1. Start Download

Starts an async download job.

- **URL:** `/download`
- **Method:** `POST`
- **Content-Type:** `application/json`
- **Max Body:** 1 MiB

### Request Body

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

Key option fields:

| Field | Type | Default | Notes |
| ----- | ---- | ------- | ----- |
| `options.output` | `string` | `{title}.{ext}` | No absolute paths or `..`. |
| `options.audio` | `boolean` | `false` | Audio-only mode. |
| `options.quality` | `string` | `best` | `best`, `worst`, `720p`, `128k`, etc. |
| `options.format` | `string` | `""` | Preferred container (`mp4`, `webm`, etc). |
| `options.jobs` | `number` | `1` | Concurrent jobs. |
| `options.timeout` | `number` | `180` | Timeout in seconds. |
| `options.on-duplicate` | `string` | `prompt` | `prompt`, `overwrite`, `skip`, `rename`, `*_all`. |

### Success Response

```json
{
  "status": "queued",
  "jobId": "job_1",
  "message": "Download started for 1 item(s)."
}
```

### Error Response

```json
{
  "type": "error",
  "status": "error",
  "error": "invalid JSON payload"
}
```

## 2. Download Progress Stream (SSE)

Streams progress and state events for a job.

- **URL:** `/download/progress?id={jobId}`
- **Method:** `GET`
- **Content-Type:** `text/event-stream`

Each SSE `data:` line is JSON. Event types include:

- `register`
- `progress`
- `finish`
- `log`
- `duplicate`
- `done`

`done` includes `message` with `complete` or `error`.

## 3. Duplicate Prompt Response

Submits a duplicate-file decision for a pending prompt.

- **URL:** `/download/duplicate-response`
- **Method:** `POST`
- **Content-Type:** `application/json`

### Request Body

```json
{
  "jobId": "job_1",
  "promptId": "dup_2",
  "choice": "overwrite"
}
```

## 4. Server Status

- **URL:** `/status`
- **Method:** `GET`

### Success Response

```json
{
  "active_downloads": 1,
  "uptime": "2m31s"
}
```

## 5. Media Listing

Lists downloaded media with pagination.

- **URL:** `/media/`
- **Method:** `GET`
- **Query Params:**
  - `offset` (default `0`)
  - `limit` (default `200`, max `500`)

### Success Response

```json
{
  "items": [
    {
      "id": "file.mp4",
      "title": "file",
      "artist": "Unknown Artist",
      "size": "12.5 MB",
      "date": "2026-02-07",
      "type": "video",
      "filename": "file.mp4"
    }
  ],
  "next_offset": null
}
```

- `next_offset` is `null` when there are no more results.

## 6. Media File Serve

Serves a media file from the media directory.

- **URL:** `/media/{filename}`
- **Method:** `GET`

Path traversal and symlink escape are rejected.
