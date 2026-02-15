# GEMINI.md - Context for ytdl-go

This file provides instructional context and project standards for AI interactions within the `ytdl-go` repository.

## Project Overview

`ytdl-go` is a high-performance, feature-rich YouTube downloader written in Go. It provides both a command-line interface (CLI) with an interactive terminal UI (TUI) and a modern web-based user interface.

- **Primary Technologies:** Go (Backend), SolidJS (Frontend), Vite (Build Tool), Tailwind CSS (Styling).
- **Core Capabilities:** Parallel downloads, interactive format selection, playlist support, YouTube Music integration, and rich metadata (ID3) embedding.
- **Key Libraries:** 
  - `github.com/kkdai/youtube/v2` (YouTube interaction)
  - `github.com/charmbracelet/bubbletea` (TUI framework)
  - `github.com/u2takey/ffmpeg-go` (Media processing)

## Architecture

- `main.go`: Entry point, CLI flag parsing, and execution dispatch.
- `internal/downloader/`: Core logic for fetching metadata, selecting formats, and managing stream downloads.
- `internal/app/`: Runner logic for managing concurrent jobs and worker pools.
- `internal/web/`: Go web server that hosts the API and serves the frontend assets.
- `frontend/`: SolidJS-based web application.

## Building and Running

### Development
- **Build All:** `./build.sh` (builds backend to `bin/yt` and frontend to `internal/web/assets`).
- **Build & Launch UI:** `./build.sh --web` (builds and starts both backend and frontend dev server).
- **Manual Launch:** `go run main.go -web` (after frontend build).

### Installation
- **CLI Only:** `go install github.com/lvcoi/ytdl-go@latest`

### Testing
- **Backend:** `go test ./...`
- **Frontend:** `cd frontend && npm test`

## Development Conventions

### Backend (Go)
- **Concurrency:** Use fixed-size worker pools for downloads (configured via `-jobs`).
- **Cancellation:** Every long-running operation must accept and respect `context.Context`.
- **Error Handling:** 
  - Use custom error types for distinct failure modes.
  - Always wrap errors with context using `fmt.Errorf("...: %w", err)`.
  - Check `downloader.IsReported(err)` to avoid duplicate error logging.
- **Style:** Follow standard `go fmt` and `go vet` practices. Prefer interfaces to decouple components.

### Frontend (SolidJS)
- **Components:** Use controlled components for all form inputs.
- **State:** Use SolidJS stores (`createStore`) located in `frontend/src/store/`.
- **Reactivity:** Use `splitProps` when destructuring to preserve reactivity.
- **Conditional Rendering:** Prefer `<Show>` and `<For>` components over logical `&&` or `.map()`.

### Process
- **TDD:** Write failing tests before implementing features or bug fixes.
- **Commits:** Use [Conventional Commits](https://www.conventionalcommits.org/).
- **CI/CD:** Ensure all tests pass and linters are clean (`go vet`, `npm run lint`) before merging.

## Key Files
- `README.md`: General overview and usage instructions.
- `docs/FLAGS.md`: Comprehensive reference for CLI options.
- `BEST_PRACTICES.md`: Detailed coding standards.
- `ENGINEERING_PROCESS.md`: Standardized development lifecycle.
- `internal/downloader/downloader.go`: Main download processing entry point.
- `frontend/src/App.jsx`: Main UI component.
