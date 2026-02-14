---
title: "Server-Sent Events"
weight: 20
---

# Server-Sent Events

This document details the Server-Sent Events (SSE) API for real-time download progress updates in ytdl-go.

## Table of Contents

- [Overview](#overview)
- [Connection](#connection)
- [Event Types](#event-types)
- [Replay Support](#replay-support)
- [Error Handling](#error-handling)

## Overview

ytdl-go uses Server-Sent Events (SSE) to stream real-time progress updates from the backend to the frontend. SSE is a server-push technology that enables the server to send updates to the client over a single HTTP connection.

### Why SSE?

- **Simpler than WebSockets** - Standard HTTP, no protocol upgrade
- **Automatic reconnection** - Built-in browser retry mechanism
- **One-way communication** - Perfect for progress updates
- **Event-based** - Structured message types
- **Firewall-friendly** - Standard HTTP(S) ports

## Connection

### Endpoint

```
GET /api/download/progress?id={jobId}
```

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | `string` | **Yes** | Job ID from download start response |
| `since` | `number` | No | Event sequence number to replay from (default: 0) |

### Request Headers

```
Accept: text/event-stream
```

### Response Headers

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

### Connection Lifecycle

```
1. Client connects → GET /api/download/progress?id=job_1
2. Server sends snapshot event (current state)
3. Server sends progress updates as events occur
4. Job completes → Server sends done event
5. Connection closes
```

### JavaScript Example

```js
const jobId = 'job_1';
const eventSource = new EventSource(`/api/download/progress?id=${jobId}`);

eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data);
  
  switch (data.type) {
    case 'snapshot':
      // Initial state
      initializeUI(data);
      break;
    case 'progress':
      // Progress update
      updateProgressBar(data);
      break;
    case 'log':
      // Log message
      appendLog(data);
      break;
    case 'done':
      // Job complete
      showCompletion(data);
      eventSource.close();
      break;
  }
};

eventSource.onerror = (error) => {
  console.error('SSE Error:', error);
  eventSource.close();
};

// Cleanup on component unmount
function cleanup() {
  eventSource.close();
}
```

## Event Types

All SSE events are sent as JSON in the `data:` field. Each event includes common envelope fields for replay safety.

### Common Envelope Fields

Most events include these fields:

| Field | Type | Description |
|-------|------|-------------|
| `type` | `string` | Event type identifier |
| `jobId` | `string` | Owning job ID |
| `seq` | `number` | Monotonically increasing sequence number |
| `at` | `string` | RFC3339 timestamp |

### Event Type: `snapshot`

Sent immediately upon connection. Describes the job's **current** state, even when replaying historical events via `since` parameter.

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

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `status` | `string` | Current job status: `queued`, `running`, `complete`, `error` |
| `tasks` | `object[]` | Array of active download tasks |
| `logs` | `string[]` | Recent log messages |

### Event Type: `status`

Job status change event.

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

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `status` | `string` | New status: `queued`, `running`, `complete`, `error` |
| `message` | `string` | Human-readable status description |

### Event Type: `register`

New download task registered.

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

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `taskId` | `string` | Unique task identifier |
| `description` | `string` | Human-readable task description |
| `total` | `number` | Total bytes to download (0 if unknown) |

### Event Type: `progress`

Download progress update.

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

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `taskId` | `string` | Task identifier |
| `current` | `number` | Bytes downloaded so far |
| `total` | `number` | Total bytes to download |
| `percent` | `number` | Percentage complete (0-100) |
| `speed` | `string` | Current download speed (human-readable) |
| `eta` | `string` | Estimated time remaining (human-readable) |

### Event Type: `finish`

Task completed.

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

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `taskId` | `string` | Task identifier |
| `success` | `boolean` | Whether task succeeded |
| `error` | `string` \| `null` | Error message if failed, null if succeeded |

### Event Type: `log`

Log message.

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

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `level` | `string` | Log level: `debug`, `info`, `warn`, `error` |
| `message` | `string` | Log message text |

### Event Type: `duplicate`

Duplicate file prompt.

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

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `promptId` | `string` | Unique prompt identifier (for response) |
| `filename` | `string` | Name of the conflicting file |
| `message` | `string` | Human-readable prompt message |

> **Note:** The client should respond via [POST /api/download/duplicate-response](endpoints#duplicate-prompt-response).

### Event Type: `duplicate-resolved`

Duplicate prompt resolved.

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

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `promptId` | `string` | Prompt identifier |
| `choice` | `string` | User's choice: `overwrite`, `skip`, `rename`, etc. |

### Event Type: `done`

Job complete (terminal event).

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

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `status` | `string` | Final status: `complete` or `error` |
| `message` | `string` | Human-readable completion message |
| `exitCode` | `number` | Exit code (0 = success, non-zero = error) |
| `error` | `string` \| `null` | Error message if failed |
| `stats` | `object` | Download statistics |

#### Stats Object

| Field | Type | Description |
|-------|------|-------------|
| `total` | `number` | Total number of items processed |
| `success` | `number` | Number of successful downloads |
| `failed` | `number` | Number of failed downloads |
| `skipped` | `number` | Number of skipped downloads |

## Replay Support

The `since` query parameter enables event replay for reconnection scenarios.

### Usage

```
GET /api/download/progress?id=job_1&since=5
```

### Behavior

- **`since=0`** - Replay all retained events for the job
- **`since=N`** - Replay events with `seq > N`
- **No `since`** - Only new events (no replay)

### Replay Guarantees

1. **Snapshot always sent** - Regardless of `since` value, snapshot reflects **current** state
2. **Event ordering** - Events replayed in sequence order
3. **No duplicates** - Events already received (seq ≤ since) are not resent
4. **Retention window** - Server retains events for a limited time (typically until job completion)

### Reconnection Strategy

```js
let lastSeq = 0;

function connect() {
  const url = `/api/download/progress?id=${jobId}&since=${lastSeq}`;
  const eventSource = new EventSource(url);
  
  eventSource.onmessage = (event) => {
    const data = JSON.parse(event.data);
    
    // Track sequence number
    if (data.seq !== undefined) {
      lastSeq = Math.max(lastSeq, data.seq);
    }
    
    handleEvent(data);
  };
  
  eventSource.onerror = (error) => {
    eventSource.close();
    
    // Reconnect after delay
    setTimeout(() => {
      console.log('Reconnecting...');
      connect();
    }, 1000);
  };
}

connect();
```

## Error Handling

### Connection Errors

```js
eventSource.onerror = (error) => {
  console.error('SSE Error:', error);
  
  // Check readyState
  switch (eventSource.readyState) {
    case EventSource.CONNECTING:
      console.log('Reconnecting...');
      break;
    case EventSource.CLOSED:
      console.log('Connection closed');
      break;
  }
  
  // Implement exponential backoff
  setTimeout(() => {
    connect();
  }, backoffDelay);
};
```

### HTTP Error Responses

#### 400 Bad Request

Missing or invalid job ID:

```
HTTP/1.1 400 Bad Request
Content-Type: text/plain

Invalid job ID
```

#### 404 Not Found

Job not found:

```
HTTP/1.1 404 Not Found
Content-Type: text/plain

Job not found
```

### Best Practices

1. **Always close connections** - Call `eventSource.close()` when done
2. **Handle reconnection** - Implement retry logic with backoff
3. **Track sequence numbers** - Use `since` for replay on reconnect
4. **Set reasonable timeouts** - Don't keep connections open indefinitely
5. **Clean up on unmount** - Close EventSource in cleanup handlers

### Complete Example

```js
function useDownloadProgress(jobId) {
  let eventSource = null;
  let lastSeq = 0;
  let reconnectAttempts = 0;
  const maxReconnectAttempts = 5;
  
  function connect() {
    const url = `/api/download/progress?id=${jobId}&since=${lastSeq}`;
    eventSource = new EventSource(url);
    
    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      // Reset reconnect counter on successful message
      reconnectAttempts = 0;
      
      // Track sequence
      if (data.seq !== undefined) {
        lastSeq = Math.max(lastSeq, data.seq);
      }
      
      // Handle event
      switch (data.type) {
        case 'snapshot':
          console.log('Initial state:', data);
          break;
        case 'progress':
          updateProgress(data);
          break;
        case 'log':
          appendLog(data);
          break;
        case 'done':
          console.log('Job complete:', data);
          eventSource.close();
          break;
      }
    };
    
    eventSource.onerror = (error) => {
      console.error('SSE Error:', error);
      eventSource.close();
      
      // Exponential backoff reconnection
      if (reconnectAttempts < maxReconnectAttempts) {
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
        reconnectAttempts++;
        
        console.log(`Reconnecting in ${delay}ms (attempt ${reconnectAttempts})...`);
        setTimeout(connect, delay);
      } else {
        console.error('Max reconnect attempts reached');
      }
    };
  }
  
  // Start connection
  connect();
  
  // Cleanup function
  return () => {
    if (eventSource) {
      eventSource.close();
    }
  };
}
```

## Related Documentation

- [Endpoints](endpoints) - REST API endpoints
- [Types](types) - Complete type definitions
- [Frontend Structure](../architecture/frontend-structure) - Frontend SSE integration
