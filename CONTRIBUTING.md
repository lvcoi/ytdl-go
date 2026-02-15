# Contributing to ytdl-go

Thank you for your interest in contributing to ytdl-go! This guide will help you get your development environment set up and understand the project structure.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Running Tests](#running-tests)
- [Code Quality](#code-quality)
- [Development Workflow](#development-workflow)
- [Submitting Changes](#submitting-changes)

## Prerequisites

Before you begin, ensure you have the following installed:

### Required

- **Go 1.24+** - [Download and install Go](https://golang.org/dl/)
  ```bash
  go version  # Should show 1.24 or higher
  ```

### Recommended

- **FFmpeg** (optional but recommended for audio extraction fallback)
  ```bash
  # macOS
  brew install ffmpeg
  
  # Ubuntu/Debian
  sudo apt-get install ffmpeg
  
  # Windows (via Chocolatey)
  choco install ffmpeg
  ```
  
  FFmpeg enables the audio extraction fallback strategy when YouTube blocks direct audio-only downloads. The tool works without it, but some downloads may fail that would otherwise succeed.

- **Make** (if using Makefile - currently not present but may be added)
  ```bash
  # macOS (included in Xcode Command Line Tools)
  xcode-select -install
  
  # Ubuntu/Debian
  sudo apt-get install build-essential
  
  # Windows (via Chocolatey)
  choco install make
  ```

### Development Tools

- **golangci-lint** (for code linting)
  ```bash
  # macOS
  brew install golangci-lint
  
  # Linux/Windows
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

- **Git** (for version control)
  ```bash
  git -version
  ```

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go
```

### 2. Download Dependencies

```bash
go mod download
```

This downloads all required Go modules listed in `go.mod`:
- `github.com/kkdai/youtube/v2` - YouTube API client
- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - TUI styling
- `github.com/bogem/id3v2/v2` - ID3 tag embedding
- `github.com/u2takey/ffmpeg-go` - FFmpeg Go bindings

### 3. Build the Project

```bash
# Build the binary
go build -o ytdl-go .

# Or build and install to $GOPATH/bin
go install .
```

### 4. Verify Installation

```bash
# Run the built binary
./ytdl-go -help

# Or if installed to $GOPATH/bin
ytdl-go -help
```

### Alternative: Custom Cache Locations

If your environment blocks writes to the default Go caches:

```bash
export GOCACHE=/tmp/gocache
export GOMODCACHE=/tmp/gomodcache
go mod tidy
go build .
```

## Project Structure

Understanding the project layout will help you navigate the codebase:

```
ytdl-go/
├── main.go                          # Entry point, CLI flag parsing
├── go.mod                           # Go module definition
├── go.sum                           # Dependency checksums
├── internal/                        # Private application code
│   ├── app/                         # Application orchestration layer
│   │   └── runner.go                # Main download workflow coordinator
│   ├── downloader/                  # Core download functionality
│   │   ├── downloader.go            # Core download logic
│   │   ├── youtube.go               # YouTube-specific extraction
│   │   ├── direct.go                # Direct URL downloads
│   │   ├── segment_downloader.go    # HLS/DASH segment handling
│   │   ├── unified_tui.go           # Terminal UI (format selector + progress)
│   │   ├── progress_manager.go      # Concurrent progress coordination
│   │   ├── output.go                # File writing and path management
│   │   ├── metadata.go              # Metadata extraction and sidecar JSON
│   │   ├── tags.go                  # ID3 tag embedding
│   │   ├── prompt.go                # Interactive user prompts
│   │   ├── path.go                  # Path sanitization and auto-rename
│   │   ├── http.go                  # HTTP client configuration
│   │   ├── errors.go                # Error categorization
│   │   ├── music.go                 # YouTube Music URL handling
│   │   └── ...                      # Other supporting files
│   └── web/                         # Web UI server (optional feature)
├── docs/                            # Documentation
│   ├── ARCHITECTURE.md              # Architecture overview and design
│   ├── FLAGS.md                     # Comprehensive flag reference
│   └── MAINTAINERS.md               # Release and maintenance guide
├── CONTRIBUTING.md                  # This file
├── README.md                        # User-facing documentation
├── LICENSE                          # MIT License
└── SECURITY.md                      # Security policy

```

### Key Directories

- **`internal/app/`** - High-level application logic
  - Orchestrates the download workflow
  - Manages concurrency (jobs, playlists)
  - Coordinates between downloader and UI

- **`internal/downloader/`** - Core download implementation
  - YouTube metadata extraction
  - Multiple download strategies (standard, retry, FFmpeg fallback)
  - Progress tracking and terminal UI
  - File I/O and metadata handling

- **`internal/web/`** - Web server (optional)
  - Web-based UI for downloads
  - API endpoints
  - Not required for CLI functionality

### Module Boundaries

- **`cmd` vs `internal`**: Currently, there's no `cmd/` directory. The main entry point is `main.go` at the root. All implementation code lives in `internal/` to prevent external imports.

- **`internal/downloader`**: Contains all download logic. If you're adding features related to downloading, metadata, or progress display, this is where you'll work.

- **`internal/app`**: Contains orchestration logic. If you're changing how multiple downloads are coordinated or how command-line arguments flow into the downloader, work here.

## Running Tests

### Execute All Tests

```bash
go test -v ./...
```

**Note:** As of now, ytdl-go has no test files (`.go` files are present, but no `*_test.go` files). This is an area where contributions would be valuable!

### Run Tests for a Specific Package

```bash
go test -v ./internal/downloader
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

### Run Specific Test

```bash
go test -v -run TestFunctionName ./internal/downloader
```

## Code Quality

### Linting

We use `golangci-lint` for comprehensive code linting. Run it before submitting changes:

```bash
golangci-lint run
```

To automatically fix issues where possible:

```bash
golangci-lint run -fix
```

### Formatting

Go code should be formatted with `gofmt`:

```bash
# Check formatting
gofmt -l .

# Format all files
gofmt -w .
```

Or use `go fmt`:

```bash
go fmt ./...
```

### Vetting

Use `go vet` to catch common mistakes:

```bash
go vet ./...
```

### Pre-Commit Checklist

Before committing, ensure:
- [ ] Code is formatted: `go fmt ./...`
- [ ] No vet warnings: `go vet ./...`
- [ ] Linter passes: `golangci-lint run`
- [ ] Tests pass: `go test ./...` (when tests exist)
- [ ] Binary builds: `go build .`

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Changes

- Write code following Go best practices
- Add comments for exported functions and types
- Keep functions focused and single-purpose
- Update documentation if behavior changes

### 3. Test Your Changes

```bash
# Build and test locally
go build .
./ytdl-go [test with your changes]

# Run formatting and linting
go fmt ./...
go vet ./...
golangci-lint run
```

### 4. Commit Changes

```bash
git add .
git commit -m "Brief description of your change"
```

**Commit Message Guidelines:**
- Use present tense ("Add feature" not "Added feature")
- Be concise but descriptive
- Reference issue numbers if applicable (#123)

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then open a pull request on GitHub with:
- Clear description of changes
- Motivation and context
- Any related issues
- Screenshots/examples if applicable

## Frontend Development

This section covers contributions to the web UI under `frontend/`.

### Frontend Workflow

1.  Install frontend dependencies:
    ```sh
    cd frontend
    npm install
    ```
2.  Start the development server:
    ```sh
    npm run dev
    ```
3.  Implement your change in `frontend/src/`.
4.  Validate the frontend build:
    ```sh
    npm run build
    ```

### Frontend Conventions

*   **Components**: keep components focused and colocate related logic where practical.
*   **Naming**: PascalCase for component files, camelCase for helpers/utilities.
*   **Styling**: prefer Tailwind utility classes, keeping global CSS in `frontend/index.css` minimal.
*   **Icons**: use `lucide-solid` through `frontend/src/components/Icon.jsx`; add icons by importing them there and extending `iconMap`.

### Dependency Policy

Keep dependencies lean.

*   **Preferred**: small focused libraries.
*   **Avoid by default**: large UI frameworks and heavy libraries that increase bundle size significantly.

## Submitting Changes

### Pull Request Guidelines

1. **Clear Title**: Summarize the change in one line
2. **Description**: Explain what changed and why
3. **Testing**: Describe how you tested the changes
4. **Breaking Changes**: Clearly mark any breaking changes
5. **Documentation**: Update relevant docs if needed

### Code Review Process

1. Maintainers will review your PR
2. Address any feedback or requested changes
3. Once approved, your PR will be merged

### What We Look For

- **Correctness**: Does it work as intended?
- **Simplicity**: Is it the simplest solution?
- **Performance**: Does it maintain/improve performance?
- **Error Handling**: Are errors handled gracefully?
- **Documentation**: Are changes documented?
- **Testing**: Are there tests (or plans to add them)?

## Debugging Tips

### Enable Debug Logging

```bash
ytdl-go -log-level debug [URL]
```

### Inspect Network Requests

The codebase uses standard Go `net/http`. You can add debug prints to `internal/downloader/http.go` to see requests/responses.

### Test with Specific Videos

For testing, use videos you control or well-known test videos:
- Public YouTube videos (avoid copyrighted content)
- Short videos to speed up testing
- Videos with various formats available

### Common Issues

1. **403 Errors**: YouTube frequently changes API. Check if `kkdai/youtube` needs updating.
2. **FFmpeg Not Found**: Ensure FFmpeg is in PATH: `which ffmpeg` (Linux/Mac) or `where ffmpeg` (Windows)
3. **Import Errors**: Run `go mod tidy` to clean up dependencies

## Architecture Resources

For a deep dive into the codebase architecture:
- Read [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- Study the Mermaid flowchart showing data flow
- Understand the download strategy chain
- Learn about the TUI seamless transition model

## Getting Help

- **Issues**: Open a GitHub issue for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions
- **Documentation**: Check README.md and docs/ folder

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help newcomers feel welcome

## License

By contributing, you agree that your contributions will be licensed under the MIT License (see [LICENSE](LICENSE) file).

--

Thank you for contributing to ytdl-go! Your efforts help make this tool better for everyone. 🎉
