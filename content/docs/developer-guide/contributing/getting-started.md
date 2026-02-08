---
title: "Getting Started"
weight: 10
---

# Getting Started

This guide will help you set up your development environment for contributing to ytdl-go.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Getting the Code](#getting-the-code)
- [Installing Dependencies](#installing-dependencies)
- [Building the Project](#building-the-project)
- [Verifying Installation](#verifying-installation)
- [Development Tools](#development-tools)
- [Project Structure](#project-structure)

## Prerequisites

Before you begin, ensure you have the following installed:

### Required

#### Go 1.24+

ytdl-go requires Go 1.24 or higher.

```bash
# Check your Go version
go version  # Should show 1.24 or higher
```

**Installation:**
- Download from [golang.org/dl/](https://golang.org/dl/)
- Follow installation instructions for your platform

### Recommended

#### FFmpeg

FFmpeg enables the audio extraction fallback strategy when YouTube blocks direct audio-only downloads. The tool works without it, but some downloads may fail that would otherwise succeed.

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt-get install ffmpeg

# Windows (via Chocolatey)
choco install ffmpeg
```

**Verify installation:**
```bash
ffmpeg -version
```

#### Node.js 18+ (for frontend)

Required only if contributing to the web UI.

```bash
# Check version
node --version  # Should show 18.x or higher

# macOS
brew install node

# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Windows (via Chocolatey)
choco install nodejs
```

### Development Tools

#### Git

```bash
# Check version
git --version

# macOS (Xcode Command Line Tools)
xcode-select --install

# Ubuntu/Debian
sudo apt-get install git

# Windows (via Chocolatey)
choco install git
```

#### golangci-lint

For code linting.

```bash
# macOS
brew install golangci-lint

# Linux/Windows
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Verify installation:**
```bash
golangci-lint --version
```

## Getting the Code

### Fork and Clone

1. **Fork the repository** on GitHub

2. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/ytdl-go.git
   cd ytdl-go
   ```

3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/lvcoi/ytdl-go.git
   git fetch upstream
   ```

## Installing Dependencies

### Backend Dependencies

Download all Go modules:

```bash
go mod download
```

This downloads:
- `github.com/kkdai/youtube/v2` - YouTube API client
- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - TUI styling
- `github.com/bogem/id3v2/v2` - ID3 tag embedding
- `github.com/u2takey/ffmpeg-go` - FFmpeg Go bindings

### Frontend Dependencies

If working on the web UI:

```bash
cd frontend
npm install
cd ..
```

## Building the Project

### Backend Only

```bash
# Build binary
go build -o ytdl-go .

# Or build and install to $GOPATH/bin
go install .
```

### Full Stack (Backend + Frontend)

```bash
# Using provided build script
./build.sh

# Or manually:
cd frontend && npm run build && cd ..
go build -o ytdl-go .
```

### Build Output

- **Binary:** `ytdl-go` (or `ytdl-go.exe` on Windows)
- **Frontend assets:** `internal/web/assets/` (embedded in binary)

## Verifying Installation

### Run the Binary

```bash
# Show help
./ytdl-go -help

# Test with info mode (no download)
./ytdl-go -info https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

Expected output: JSON metadata about the video

### Run Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

> **Note:** As of now, ytdl-go has minimal test coverage. Contributions to improve testing are welcome!

### Check Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter
golangci-lint run
```

All commands should complete without errors.

## Development Tools

### Custom Cache Locations

If your environment blocks writes to default Go caches:

```bash
export GOCACHE=/tmp/gocache
export GOMODCACHE=/tmp/gomodcache
go mod tidy
go build .
```

### IDE Setup

#### VS Code

Recommended extensions:
- **Go** (`golang.go`) - Official Go extension
- **Even Better TOML** (`tamasfe.even-better-toml`) - TOML support
- **Markdown All in One** (`yzhang.markdown-all-in-one`) - Markdown editing

Settings (`.vscode/settings.json`):
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  }
}
```

#### GoLand / IntelliJ IDEA

1. Open project directory
2. Enable Go modules support (should auto-detect)
3. Configure golangci-lint:
   - **Settings** → **Tools** → **Go Linter**
   - Select `golangci-lint`

## Project Structure

Understanding the project layout:

```
ytdl-go/
├── main.go                     # Entry point, CLI flag parsing
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── build.sh                    # Build script (backend + frontend)
│
├── internal/                   # Private application code
│   ├── app/                    # Application orchestration
│   │   └── runner.go           # Main workflow coordinator
│   ├── downloader/             # Core download functionality
│   │   ├── downloader.go       # Download logic and strategy
│   │   ├── youtube.go          # YouTube-specific extraction
│   │   ├── direct.go           # Direct URL downloads
│   │   ├── unified_tui.go      # Terminal UI
│   │   ├── progress_manager.go # Progress coordination
│   │   └── ...                 # Other modules
│   └── web/                    # Web UI server (optional)
│       ├── server.go           # HTTP server
│       └── assets/             # Frontend build output (embedded)
│
├── frontend/                   # Web UI (SolidJS)
│   ├── src/                    # Source code
│   │   ├── App.jsx             # Root component
│   │   ├── components/         # UI components
│   │   └── store/              # State management
│   ├── package.json            # Node dependencies
│   └── vite.config.js          # Build configuration
│
├── content/                    # Hugo documentation site
│   └── docs/                   # Documentation content
│
├── docs/                       # Additional documentation
│   ├── ARCHITECTURE.md         # Architecture overview
│   ├── FLAGS.md                # CLI flag reference
│   └── MAINTAINERS.md          # Maintainer's guide
│
├── CONTRIBUTING.md             # This guide (original)
├── README.md                   # User-facing documentation
├── LICENSE                     # MIT License
└── SECURITY.md                 # Security policy
```

### Key Directories

#### `internal/app/`

High-level application logic:
- Orchestrates the download workflow
- Manages concurrency (jobs, playlists)
- Coordinates between downloader and UI

#### `internal/downloader/`

Core download implementation:
- YouTube metadata extraction
- Multiple download strategies
- Progress tracking and terminal UI
- File I/O and metadata handling

#### `internal/web/`

Web server (optional feature):
- Web-based UI for downloads
- REST API endpoints
- SSE for real-time updates
- Not required for CLI functionality

#### `frontend/`

SolidJS web UI:
- Single Page Application
- Builds to static assets
- Embedded in Go binary via `//go:embed`

## Next Steps

Now that you have ytdl-go set up:

1. **Read the architecture docs:**
   - [Architecture Overview](../architecture/overview)
   - [Backend Structure](../architecture/backend-structure)
   - [Frontend Structure](../architecture/frontend-structure)

2. **Choose your contribution area:**
   - [Backend Development](backend) - Go development guide
   - [Frontend Development](frontend) - SolidJS development guide

3. **Review code style guides:**
   - [Code Style](code-style) - Coding standards for both Go and JavaScript

4. **Understand the PR process:**
   - [Pull Request Process](pull-request-process) - Workflow and best practices

## Getting Help

- **Issues:** Open a GitHub issue for bugs or feature requests
- **Discussions:** Use GitHub Discussions for questions
- **Documentation:** Check `docs/` folder and README.md

## Related Documentation

- [Backend Contributing Guide](backend) - Go-specific development
- [Frontend Contributing Guide](frontend) - SolidJS-specific development
- [Code Style Guide](code-style) - Coding standards
- [Architecture Overview](../architecture/overview) - System design
