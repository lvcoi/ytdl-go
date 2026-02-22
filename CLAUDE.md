# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ytdl-go** is a high-performance YouTube downloader in Go with both a CLI/TUI and a modern Web UI. The binary is self-contained: the SolidJS frontend compiles into `internal/web/assets/` and is embedded into the Go binary via `//go:embed`.

Tech stack:
- **Backend:** Go 1.24+, Bubble Tea (TUI), SQLite (`modernc.org/sqlite`), `gorilla/websocket`, `ffmpeg-go`
- **Frontend:** SolidJS, Tailwind CSS v4, Vite, Lucide Icons, Vitest

## Commands

### Build

```bash
# Integrated build (Go binary + frontend assets) — preferred
./build.sh

# Build and immediately launch the Web UI
./build.sh --web

# Backend only
go build -o ./bin/yt .

# Frontend only (dev with hot reload)
cd frontend && npm install && npm run dev

# Frontend production build → outputs to internal/web/assets/
cd frontend && npm run build
```

### Test

```bash
# Backend
go test ./...

# Frontend (run from frontend/)
cd frontend && npm test
cd frontend && npm run test:coverage
```

### Lint / Format

```bash
# Backend
go fmt ./...
go vet ./...

# Frontend (from frontend/)
cd frontend && npm run lint   # if configured
```

### Run a single Go test

```bash
go test ./internal/downloader/... -run TestName
```

## Architecture

### Backend (`internal/`)

```
main.go                         # CLI flag parsing; entry point
internal/app/runner.go          # Orchestrates download workflow and concurrency
internal/downloader/
  downloader.go                 # Core download logic and strategy selection
  youtube.go                    # YouTube metadata extraction + download strategies
  direct.go                     # Direct URL downloads (MP4, WebM, HLS, DASH)
  segment_downloader.go         # HLS/DASH segment parallel downloading
  unified_tui.go                # Bubble Tea TUI (format selector + progress views)
  progress_manager.go           # Thread-safe progress bar coordinator
  output.go                     # File writing, path resolution, template expansion
  metadata.go                   # Sidecar JSON generation
  tags.go                       # ID3 tag embedding for audio
  prompt.go                     # Interactive file-conflict resolution
internal/web/server.go          # REST API + WebSocket hub; serves embedded frontend
internal/db/db.go               # SQLite schema and data access
internal/ws/                    # WebSocket messaging types
```

Download strategy chain (auto-selected):
1. Standard chunked download → 2. Single-request retry on 403 → 3. FFmpeg fallback (audio-only)

The TUI uses a single Bubble Tea program that transitions seamlessly between `SeamlessViewFormatSelector` and `SeamlessViewProgress` without restarting, via a `selectionChan`.

Progress updates from download goroutines are routed through `ProgressManager` → `tea.Program.Send()` (thread-safe). Each goroutine is identified by a unique task ID.

### Frontend (`frontend/src/`)

```
App.jsx                         # Root component and top-level layout
index.jsx                       # Entry point, mounts App + store provider
store/appStore.jsx              # Central app state + localStorage persistence
components/
  DashboardView.jsx             # Dashboard with drag/resize widget grid
  DownloadView.jsx              # URL input, options, result display
  LibraryView.jsx               # Explorer-style media library (gallery/list/detail)
  SettingsView.jsx              # Configuration
  Player.jsx                    # Media playback
  dashboard/widgetRegistry.js  # Central registry for all dashboard widgets
routes/                         # SolidJS Router route definitions
hooks/                          # Custom SolidJS hooks
```

State is centralized in `store/appStore.jsx` using SolidJS context + store. Only durable UI state persists to localStorage; transient download progress state is intentionally not persisted.

API communication:
- `POST /api/download` — start async download job, returns `jobId`
- `GET /api/download/progress?id={jobId}` — SSE stream for progress events
- `POST /api/download/duplicate-response` — resolve file-conflict prompts
- `GET /api/media/` — paginated media list
- `GET /api/status` — server health

### Web UI port handling

`ytdl-go -web` auto-falls back to the next available port if the requested port is busy. If `npm run dev` proxying stops working, set `VITE_API_PROXY_TARGET=http://127.0.0.1:<port>` to match the logged backend URL.

## Development Conventions

### Go
- Pass `context.Context` as first argument to all long-running functions.
- Use custom error types for distinct failure modes; always wrap with `fmt.Errorf("...: %w", err)`.
- Use interfaces to decouple components and enable mocking. Avoid global state.
- Code must pass `go fmt` and `go vet`.

### SolidJS
- Use `createSignal` for local state, `createStore` for shared/complex state.
- Use `<Show>` (not `&&`) for conditional rendering; `<For>` for dynamic lists.
- Use `splitProps` when destructuring props to preserve reactivity.
- Treat store state as immutable.
- All interactive elements must have ARIA attributes.

### Testing
- Follow TDD: write a failing test before fixing a bug or implementing a feature.
- Use `testify` (`assert`, `require`) for Go test assertions.
- **Guard-pair testing:** every boolean guard needs both a block test *and* a pass-through test.
- Each test must set up its own preconditions — do not rely on test ordering.
- Mock state (`localStorage`, `fetch`) must be reset in `beforeEach`.

### Commits and Branching
- Follow Conventional Commits: `feat:`, `fix:`, `docs:`, `chore:`, etc.
- Branch names: `feature/<issue-number>-description` or `bugfix/<issue-number>-description`.

## Adding a Dashboard Widget

Adding a widget requires exactly 3 steps (see `frontend/docs/WIDGET_AUTHORING.md` for the full guide):

1. Create `frontend/src/components/dashboard/YourWidget.jsx` using the standard card template (`rounded-[2rem] border border-white/10 bg-black/20 backdrop-blur-sm p-6 h-full`).
2. Register it in `frontend/src/components/dashboard/widgetRegistry.js` with `id`, `label`, `icon`, `defaultW/H`, and `minW/H`.
3. Add a `case 'your-widget':` to the `renderWidget()` switch in `DashboardView.jsx`.

## Key Reference Docs

- `docs/BEST_PRACTICES.md` — authoritative coding, state management, and testing patterns
- `docs/FLAGS.md` — comprehensive CLI flag reference
- `docs/ENGINEERING_PROCESS.md` — full development lifecycle
- `internal/docs/BACKEND_ARCHITECTURE.md` — detailed backend flow diagrams
- `frontend/docs/FRONTEND_ARCHITECTURE.md` — frontend state and build pipeline
- `frontend/docs/API.md` — full backend API contract
