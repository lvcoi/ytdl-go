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
├── App.jsx              # Root component and top-level layout
├── index.jsx            # Entry point, mounts App + store provider
├── store/
│   └── appStore.jsx     # Central app state + localStorage persistence
└── components/
    ├── DownloadView.jsx  # URL input, download options, result display
    ├── LibraryView.jsx   # Downloaded media list, search/filter
    ├── SettingsView.jsx  # Configuration (cookies, extensions)
    └── Player.jsx        # Media playback
```

## State Management

State is centralized in `store/appStore.jsx` using Solid's context + store primitives. The app persists key UI fields in localStorage so layout and form choices survive reloads:

- **`ui.activeTab`** — Current view (download, library, settings)
- **`ui.isAdvanced`** — Power-user toggle state
- **`settings`** — Download options (output template, quality, jobs, audio-only, duplicate policy)
- **`download.urlInput`** — Current URL draft in the download textarea

Runtime download progress state is also centralized so tab navigation does not reset active download progress UI.
Only durable state is persisted to localStorage. Transient runtime state (active job status/progress/logs/prompts) is intentionally not persisted.

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
