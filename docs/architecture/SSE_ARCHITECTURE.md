# Real-Time Progress Communication Architecture

## Technology Choice: Server-Sent Events (SSE), Not WebSockets

The ytdl-go application uses **Server-Sent Events (SSE)** for real-time progress updates from the backend to the frontend. This is often confused with WebSockets, but they are different technologies with different use cases.

## Why SSE Instead of WebSockets?

| Feature | SSE | WebSockets |
|---------|-----|------------|
| **Direction** | Server → Client only | Bidirectional |
| **Protocol** | HTTP (GET request) | WebSocket protocol (upgrade from HTTP) |
| **Reconnection** | Built-in automatic reconnection | Must implement manually |
| **Browser API** | EventSource (simple) | WebSocket (more complex) |
| **Use Case** | Server pushing updates to client | Real-time bidirectional chat, gaming |

For ytdl-go, we only need **server → client** communication (progress updates, logs, status). The client only sends occasional requests (download start, duplicate file decisions) via standard HTTP POST requests. SSE is the simpler, more appropriate choice.

## Architecture Overview

```
┌─────────────────┐                                    ┌──────────────────┐
│  SolidJS        │                                    │   Go Backend     │
│  Frontend       │                                    │                  │
└─────────────────┘                                    └──────────────────┘
        │                                                       │
        │  1. POST /api/download                               │
        │     { urls: [...], options: {...} }                  │
        ├──────────────────────────────────────────────────────>│
        │                                                       │
        │  2. Response: { jobId: "job_123", status: "queued" } │
        │<──────────────────────────────────────────────────────┤
        │                                                       │
        │  3. EventSource connection                           │
        │     GET /api/download/progress?id=job_123            │
        ├──────────────────────────────────────────────────────>│
        │                                                       │
        │  4. SSE Stream (long-lived connection)               │
        │     id: 0                                            │
        │     data: {"type":"snapshot",...}                    │
        │                                                       │
        │     id: 1                                            │
        │     data: {"type":"status","status":"running"}       │
        │                                                       │
        │     id: 2                                            │
        │     data: {"type":"register","id":"dl_123"}          │
        │                                                       │
        │     id: 3                                            │
        │     data: {"type":"progress","current":50,"total":100}│
        │<──────────────────────────────────────────────────────┤
        │                  [connection stays open]              │
        │                                                       │
        │  5. POST /api/download/duplicate-response            │
        │     (optional, only when duplicate file prompt)      │
        ├──────────────────────────────────────────────────────>│
        │                                                       │
        │     id: 45                                           │
        │     data: {"type":"done","status":"complete"}        │
        │<──────────────────────────────────────────────────────┤
        │     [connection closes]                               │
```

## Implementation Details

### Backend (Go)

**File:** `internal/web/server.go`

The progress endpoint handler:
```go
mux.HandleFunc("/api/download/progress", func(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // Support reconnection with event replay
    afterSeq := int64(0)
    if seq, ok := parseProgressSeq(r.URL.Query().Get("since")); ok {
        afterSeq = seq
    } else if seq, ok := parseProgressSeq(r.Header.Get("Last-Event-ID")); ok {
        afterSeq = seq
    }
    
    stream, cancel := job.Subscribe(afterSeq)
    defer cancel()
    
    // Stream events in SSE format
    for evt := range stream {
        writeSSEEvent(w, flusher, enc, evt)
    }
})
```

**SSE Format Function:**
```go
func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, enc *json.Encoder, evt ProgressEvent) {
    if evt.Seq > 0 || evt.Type == "snapshot" {
        fmt.Fprintf(w, "id: %d\n", evt.Seq)
    }
    fmt.Fprintf(w, "data: ")
    _ = enc.Encode(evt)  // Adds newline after JSON
    fmt.Fprintf(w, "\n")  // Blank line terminates the event
    flusher.Flush()
}
```

### Frontend (SolidJS)

**File:** `frontend/src/hooks/useDownloadManager.js`

```javascript
const listenForProgress = (jobId) => {
    const connect = () => {
        const query = new URLSearchParams({ id: jobId });
        if (lastEventSeq > 0) {
            query.set('since', String(lastEventSeq));
        }
        
        eventSource = new EventSource(`/api/download/progress?${query.toString()}`);
        
        eventSource.onmessage = (e) => {
            const evt = JSON.parse(e.data);
            // Handle event based on type
            switch (evt.type) {
                case 'snapshot': applySnapshot(evt.snapshot, jobId); break;
                case 'status': handleStatusEvent(evt, jobId); break;
                case 'progress': updateProgress(evt); break;
                // ... other event types
            }
        };
        
        eventSource.onerror = () => {
            // Automatic reconnection with exponential backoff
            reconnectAttempts += 1;
            const delay = reconnectDelaysMs[Math.min(reconnectAttempts - 1, reconnectDelaysMs.length - 1)];
            reconnectTimer = setTimeout(connect, delay);
        };
    };
    
    connect();
};
```

## Event Types

| Event Type | Description | Critical? |
|------------|-------------|-----------|
| `snapshot` | Full state snapshot (sent on connect/reconnect) | Yes |
| `status` | Job status change (queued, running, complete, error) | Yes |
| `register` | New progress task registered | No |
| `progress` | Progress update for a task | No |
| `finish` | Task completion | No |
| `log` | Log message | No |
| `duplicate` | Duplicate file prompt | Yes |
| `duplicate-resolved` | Duplicate prompt resolved | Yes |
| `done` | Job completed (closes connection) | Yes |

Critical events have a 200ms timeout for delivery to ensure they're not dropped even if subscriber buffers are full.

## Reconnection & Resume

### Backend Support

1. **Event Sequencing:** Every event (except snapshots) has a monotonically increasing `seq` number
2. **Event History:** The backend retains the last 4096 events per job
3. **Snapshot + Replay:** On reconnection, the backend sends:
   - A `snapshot` event with current state
   - Historical events with `seq > since` parameter
   - Live events as they occur

### Frontend Support

1. **Last Event Tracking:** Frontend tracks `lastEventSeq` from received events
2. **Reconnection Query:** On reconnect, sends `?since={lastEventSeq}` to resume
3. **Exponential Backoff:** Reconnection delays: 1s, 2s, 4s, 8s, 10s (max 5 attempts)
4. **EventSource Auto-Reconnect:** Browser's EventSource API also has built-in reconnection using `Last-Event-ID` header

## Common Misconceptions

### "Why not use WebSockets?"

WebSockets are bidirectional and more complex. For ytdl-go:
- **Server → Client:** Progress updates (high frequency) → **Use SSE**
- **Client → Server:** Control messages (low frequency) → **Use standard HTTP requests**

This separation is simpler, more reliable, and easier to test than managing bidirectional WebSocket messages.

### "SSE doesn't work with proxies/load balancers"

This is outdated. Modern proxies (nginx, Apache, Caddy) support SSE correctly. The key is setting proper headers:
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Accel-Buffering: no  # For nginx
```

### "SSE has connection limits"

Browsers limit concurrent EventSource connections per domain (typically 6). For ytdl-go, this is not an issue because:
1. Only one download job is active at a time in the UI
2. Each job uses one SSE connection
3. Connections close when jobs complete

## Testing

The SSE implementation is tested in `internal/web/server_test.go`:

```go
func TestJobSubscribeSnapshotReplayAndLiveEvent(t *testing.T) {
    // Test verifies:
    // 1. Snapshot event is sent first
    // 2. Historical events are replayed based on 'since' parameter
    // 3. Live events are streamed
    // 4. Event sequencing is monotonic
    // 5. SSE format is correct (id: and data: lines)
}
```

## Performance Characteristics

- **Latency:** ~10-50ms from backend event emission to frontend receipt
- **Throughput:** Can handle 100+ events/second per connection
- **Memory:** ~128KB buffer per SSE subscriber (base + event history)
- **Reconnection:** Sub-second for temporary disconnects (with event replay)

## Security

1. **Origin Validation:** Web server enforces same-origin policy
2. **Job ID Validation:** Only events for requested job are sent
3. **No Credentials in Stream:** Job IDs are returned from authenticated download requests
4. **Auto-Expiry:** Completed jobs expire after 30 minutes, error jobs after 2 hours

## References

- [SSE Specification (HTML5)](https://html.spec.whatwg.org/multipage/server-sent-events.html)
- [EventSource API (MDN)](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)
- [SSE vs WebSockets](https://www.ably.io/topic/websockets-vs-sse)
