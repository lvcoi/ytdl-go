# Frontend Architecture

This document describes the high-level architecture of the `ytdl-go` web frontend.

## Overview

The frontend is a **Single Page Application (SPA)** built with [SolidJS](https://www.solidjs.com/). It compiles to static assets (`index.html`, `app.js`, `styles.css`) that are embedded directly into the Go binary at build time via `//go:embed`. This means the final distributed artifact is a single, self-contained executable.

## Tech Stack

| Layer | Technology |
| ----- | ---------- |
| UI Framework | SolidJS |
| Build Tool | Vite |
| Styling | Tailwind CSS |
| Icons | Lucide (vanilla) |
| Language | JavaScript (JSX) |

## Component Structure

```text
src/
├── App.jsx              # Root component, routing, global state
├── index.jsx            # Entry point, mounts App into DOM
└── components/
    ├── DownloadView.jsx  # URL input, download options, result display
    ├── LibraryView.jsx   # Downloaded media list, search/filter
    ├── SettingsView.jsx  # Configuration (cookies, extensions)
    └── Player.jsx        # Media playback
```

## State Management

SolidJS signals and props are used for all state management — no external state library. Key signals live in `App.jsx` and are passed down as props:

- **`settings`** — Download options (output template, quality, jobs, audio-only, etc.)
- **`activeTab`** — Current view (download, library, settings)
- **`downloads`** — Library items (currently mock data)
- **`isAdvanced`** — Toggle for power-user options

## Build & Integration Pipeline

1. **Development:** `npm run dev` starts Vite dev server with hot reload. API calls proxy to the Go backend via `vite.config.js` proxy (`/api` → `VITE_API_PROXY_TARGET`, default `http://127.0.0.1:8080`).
2. **Production build:** `npm run build` compiles assets into `../internal/web/assets/`.
3. **Go embed:** The Go server uses `//go:embed assets/*` to bundle the compiled frontend into the binary.
4. **Runtime:** The Go server serves the SPA at `/` and API endpoints at `/api/*`.

## API Communication

The frontend communicates with the Go backend via JSON over HTTP:

- **`POST /api/download`** — Submit URLs for download (synchronous, returns results).
- **`GET /api/library`** — Fetch downloaded media list (planned).
- **`GET /api/status`** — Server health and active job count (planned).

See [API.md](API.md) for the full contract.
