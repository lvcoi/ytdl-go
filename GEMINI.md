# GEMINI.md - ytdl-go Context & Instructions

This file provides the foundational context, architecture, and engineering standards for the `ytdl-go` project. Adhere to these guidelines for all development tasks.

## Project Overview

**ytdl-go** is a high-performance, feature-rich YouTube downloader written in Go. It provides both a powerful Terminal User Interface (TUI) and a modern Web UI.

- **Purpose:** Fast, parallelized downloading of YouTube videos, audio, and playlists with rich metadata embedding.
- **Main Technologies:**
    - **Backend:** Go 1.24+, `bubbletea` (TUI), `sqlite` (Storage), `ffmpeg` (Processing), `gorilla/websocket` (Real-time updates).
    - **Frontend:** SolidJS, Tailwind CSS v4, Vite, Lucide Icons.
- **Architecture:** 
    - `main.go`: Entry point and flag parsing.
    - `internal/app`: Application runner and lifecycle management.
    - `internal/downloader`: Core downloading, format selection, and metadata logic.
    - `internal/web`: REST API and WebSocket server, serves embedded frontend assets.
    - `internal/db`: SQLite-based persistence for downloads and library state.
    - `frontend/`: Standalone SolidJS application.

---

## Building and Running

### Integrated Build (Recommended)
The `build.sh` script is the single source of truth for building the entire application (Go binary + frontend assets).

- **Build all:** `./build.sh`
- **Build and Launch Web UI:** `./build.sh --web`
- **Options:** Use `-p` for port, `-H` for host.

### Individual Components
- **Backend Only:** `go build -o ./bin/yt .`
- **Frontend Development:** `cd frontend && npm install && npm run dev`
- **Frontend Build:** `cd frontend && npm run build` (outputs to `internal/web/assets/`)

---

## Testing Strategy

The project follows a **Test-Driven Development (TDD)** approach. Always write a reproduction test before fixing a bug or implementing a feature.

- **Backend (Go):** 
    - Run all tests: `go test ./...`
    - Use `testify` (`assert`, `require`) for assertions.
    - Ensure new functions have corresponding unit tests.
- **Frontend (SolidJS):**
    - Run all tests: `npm test` (inside `frontend/`).
    - Uses `vitest` and `@solidjs/testing-library`.
- **Quality Gate:** `build.sh` performs basic verification. CI runs full tests and linting.

---

## Development Conventions

### Backend (Go)
- **Concurrency:** Use fixed-size worker pools for downloads. Always pass `context.Context` for cancellation.
- **Error Handling:** Use custom error types for distinct failure modes. Always wrap errors with context: `fmt.Errorf("...: %w", err)`.
- **Interfaces:** Favor interfaces to decouple components and enable mocking in tests.
- **Formatting:** Code must be formatted with `go fmt` and pass `go vet`.

### Frontend (SolidJS)
- **Reactivity:** Use `createSignal` for local state and `createStore` for shared/complex state.
- **Performance:** Prefer `<Show>` over `&&` for conditional rendering. Use `<For>` for dynamic lists.
- **State Management:** Keep stores in `frontend/src/store/`. Treat state as immutable.
- **Styling:** Use Tailwind CSS v4 utility classes.
- **Components:** All interactive elements must have appropriate ARIA attributes. Form inputs must be controlled components.

### General
- **Commits:** Follow **Conventional Commits** (e.g., `feat:`, `fix:`, `docs:`, `chore:`).
- **Branching:** Work in `feature/<issue-number>-description` or `bugfix/<issue-number>-description`.
- **Documentation:** Maintain `docs/FLAGS.md` and `docs/ARCHITECTURE.md` as the implementation evolves.

---

## Key Files & Directories

- `main.go`: Application entry point.
- `build.sh`: Unified build and execution script.
- `internal/downloader/`: Core engine for YouTube interaction.
- `internal/web/server.go`: API and WebSocket hub implementation.
- `internal/db/db.go`: SQLite schema and data access.
- `frontend/src/App.jsx`: Main frontend component.
- `ENGINEERING_PROCESS.md`: Detailed development lifecycle.
- `BEST_PRACTICES.md`: Detailed coding standards and patterns.
- `docs/FLAGS.md`: Comprehensive CLI flag documentation.
