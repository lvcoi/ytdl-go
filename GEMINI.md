# GEMINI.md - Project Context: ytdl-go

## Project Overview
`ytdl-go` is a powerful, high-performance YouTube downloader written in Go, featuring both a feature-rich Terminal User Interface (TUI) and a modern Web UI. It supports downloading videos, audio, playlists, and YouTube Music with advanced features like parallel downloads, automatic retries, and rich metadata handling.

### Key Technologies
- **Backend:** Go 1.24+ (using `ytdl-lib` fork, `bubbletea` for TUI, `id3v2` for tags, `ffmpeg-go` for processing).
- **Frontend:** SolidJS, Vite, Tailwind CSS (Single Page Application).
- **Architecture:** Modular Go backend with a REST API serving a SolidJS frontend.

---

## Building and Running

### Quick Start
- **Integrated Build:** Run `./build.sh` to build both the backend and frontend.
- **Launch Web UI:** Run `./build.sh --web` to build and automatically start the backend with the web interface.

### Backend (Go)
- **Build:** `go build -o ytdl-go .`
- **Install:** `go install .`
- **Run Tests:** `go test ./...`
- **Lint/Format:** `go fmt ./...` and `go vet ./...`

### Frontend (SolidJS)
- **Directory:** `frontend/`
- **Install Dependencies:** `npm install`
- **Development Server:** `npm run dev`
- **Build Assets:** `npm run build` (outputs to `internal/web/assets/` via Vite config).
- **Run Tests:** `npm test`

---

## Development Conventions

### Backend (Go) Best Practices
- **Concurrency:** Use fixed-size worker pools managed by the `ProgressManager`. The number of workers is controlled via the `-jobs` flag.
- **Cancellation:** Always pass `context.Context` to long-running operations and respect `ctx.Done()`.
- **Error Handling:** Use custom error types (see `internal/downloader/errors.go`) and wrap errors with context using `fmt.Errorf("...: %w", err)`.
- **Metadata:** The system generates sidecar JSON files (`.json`) for each download to store rich metadata and thumbnails.

### Frontend (SolidJS) Best Practices
- **Reactivity:** Use `<For>` for lists and `<Show>` for conditional rendering. Use `splitProps` to preserve reactivity when destructuring.
- **State Management:** Application state is organized in `frontend/src/store/`. Use `createStore` for complex state.
- **Components:** All form inputs must be controlled components.

### Engineering Process
- **TDD:** Write a failing test before implementing bug fixes or new features.
- **Branching:** Work in feature branches (`feature/xxx`) or bugfix branches (`bugfix/xxx`).
- **Commits:** Follow Conventional Commits specification.
- **Reviews:** All changes require a Pull Request and code review.

---

## AI Automation and Maintenance
The project features a sophisticated AI-integrated maintenance system located in the `.github` directory.

### Gemini Automation
- **Gemini Dispatch:** A custom GitHub Action (`.github/workflows/gemini-dispatch.yml`) that listens for specific triggers:
    - **Pull Requests:** Automatically triggers a `/review` when a PR is opened.
    - **Issues:** Automatically triggers a `/triage` when an issue is opened or reopened.
    - **Interactive Commands:** Responds to comments starting with `@gemini-cli` from authorized users (Owner/Member/Collaborator).
- **Command Configurations:** Detailed prompts and personas for AI actions are stored in `.github/commands/` (e.g., `gemini-review.toml`, `gemini-triage.toml`).
- **Review Criteria:** The AI reviewer focuses on Correctness, Security, Efficiency, Maintainability, and Testing, using a standardized severity scale (🔴 to 🟢).

### Standard CI/CD
- **Go CI:** `.github/workflows/go.yml` handles building and testing the Go backend on every push and PR to the `main` branch.
- **Dependency Review:** Automated scanning for dependency vulnerabilities.

---

## Project Structure and API
- `main.go`: CLI entry point and flag parsing.
- `internal/app/runner.go`: Main execution flow and job coordination via `app.Run`.
- `internal/downloader/`:
    - `downloader.go`: Defines the central `Options` struct used by CLI and Web UI.
    - `youtube.go`: YouTube-specific extraction and strategy selection.
    - `unified_tui.go`: Bubble Tea TUI implementation.
    - `progress_manager.go`: Multi-threaded progress tracking.
    - `output.go`: File writing and template expansion.
    - `errors.go`: Categorized error system (`CategoryRestricted`, `CategoryNetwork`, `CategoryInvalidURL`).
- `internal/web/`: 
    - `server.go`: Go web server implementing a REST API.
    - **Endpoints:**
        - `POST /api/download`: Starts a download job.
        - `GET /api/download/progress`: **SSE (Server-Sent Events)** endpoint for real-time progress updates (NOT WebSockets).
        - `GET /api/media/`: Lists or serves downloaded media files.
    - **Media Layout:** Files are organized into `audio/`, `video/`, `playlist/`, and `data/`.
    - **Communication Architecture:** Uses SSE for server→client streaming (progress updates) and standard HTTP POST for client→server commands. SSE is simpler than WebSockets for unidirectional data flow.
- `frontend/src/`: SolidJS application source.
- `docs/`: Detailed documentation on architecture and flags.

---

## Documentation Links
- [Architecture Overview](./docs/ARCHITECTURE.md)
- [Backend Deep Dive](./internal/docs/BACKEND_ARCHITECTURE.md)
- [Command-Line Flags](./docs/FLAGS.md)
- [Engineering Process](./ENGINEERING_PROCESS.md)
- [Best Practices](./BEST_PRACTICES.md)
