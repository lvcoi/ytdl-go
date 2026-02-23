# AGENTS.md

## Overview

YouTube downloader in Go with a SolidJS web frontend. Provides CLI (TUI) and web UI modes.

## Prerequisites

- Go 1.24+
- Node.js (for frontend)
- ffmpeg (runtime dependency for media processing)

## Setup & Build

```bash
# Full build (Go backend + frontend assets) — single source of truth
./build.sh

# Build and auto-launch web UI
./build.sh --web

# Backend only
go build -o ./bin/yt .

# Frontend only (outputs to internal/web/assets/)
cd frontend && npm install && npm run build
```

## Dev Server

```bash
# Frontend dev server with Vite (proxies /api and /ws to backend)
cd frontend && npm run dev

# Run backend in web mode (default 0.0.0.0:8888)
./bin/yt -web

# Override host/port
./bin/yt -web --host 127.0.0.1 --port 8080
```

The Vite config auto-detects the backend port. Set `VITE_API_PROXY_TARGET` to override.

## Testing

```bash
# Go — all packages
go test ./...

# Go — single package
go test ./internal/downloader/...
go test ./internal/web/...
go test ./internal/db/...

# Go — verbose
go test -v ./internal/downloader/...

# Frontend — all tests (vitest)
cd frontend && npm test

# Frontend — with coverage
cd frontend && npm run test:coverage
```

## Lint & Format

```bash
# Go
go fmt ./...
go vet ./...
```

No frontend linter is configured in package.json.

## CI Checks (GitHub Actions)

Triggered on push/PR to `main`:

| Workflow | What it does |
|---|---|
| `go.yml` | `go build ./...` and `go test ./...` |
| `dependency-review.yml` | Scans dependency changes in PRs |
| `ossar.yml` | Static analysis (runs on Windows) |

## PR Requirements

- **Branch naming:** `feature/<issue>-description` or `bugfix/<issue>-description`
- **Commit format:** Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`)
- **CI must pass:** Go build + tests (`go.yml`), dependency review
- **Target branch:** `main`

## Key Directories

```
main.go                  Entry point, CLI flag parsing
build.sh                 Unified build script (Go + frontend)
internal/
  app/                   Application runner, lifecycle
  downloader/            Core download engine, format selection, metadata
  web/                   REST API, WebSocket server, embedded frontend assets
  db/                    SQLite persistence (downloads, library)
  ws/                    WebSocket hub
frontend/
  src/
    App.jsx              Main SolidJS component
    components/          UI components
    hooks/               SolidJS hooks
    store/               State management (createStore)
    services/            API client layer
    utils/               Shared utilities
    test/                Test setup and helpers
  vite.config.js         Vite config with API proxy auto-detection
  vitest.config.js       Test config (jsdom, SolidJS)
docs/                    Architecture, flags, contributing guides
```

## Conventions

- **Go errors:** Use custom error types, always wrap with `fmt.Errorf("...: %w", err)`
- **Go concurrency:** Fixed-size worker pools, pass `context.Context` for cancellation
- **Frontend reactivity:** `createSignal` for local state, `createStore` for shared state
- **Frontend rendering:** Prefer `<Show>` over `&&`, use `<For>` for lists
- **Styling:** Tailwind CSS v4 utility classes
