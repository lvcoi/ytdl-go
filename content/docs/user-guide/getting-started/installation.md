---
title: "Installation"
weight: 10
---

# Installation

This guide covers different ways to install ytdl-go on your system.

## Requirements

- **Go 1.24+** (for building from source)
- **FFmpeg** (optional, recommended for audio extraction fallback)

> **Tip: FFmpeg Installation**
>
> While FFmpeg is optional, it's recommended for the best experience with audio extraction:
>
> **macOS:**
> ```bash
> brew install ffmpeg
> ```
>
> **Linux (Debian/Ubuntu):**
> ```bash
> sudo apt-get install ffmpeg
> ```
>
> **Linux (Fedora):**
> ```bash
> sudo dnf install ffmpeg
> ```
>
> **Windows:**
> Download from [ffmpeg.org](https://ffmpeg.org/download.html) and add to PATH.

---

## Installation Methods

### Pre-built Binaries (Recommended)

Download pre-compiled binaries from the [GitHub Releases](https://github.com/lvcoi/ytdl-go/releases) page.

**Linux:**

```bash
# Download latest release
curl -L https://github.com/lvcoi/ytdl-go/releases/latest/download/ytdl-go-linux-amd64 -o ytdl-go

# Make executable
chmod +x ytdl-go

# Move to system path (optional)
sudo mv ytdl-go /usr/local/bin/

# Verify installation
ytdl-go -version
```

**macOS:**

```bash
# Download latest release
curl -L https://github.com/lvcoi/ytdl-go/releases/latest/download/ytdl-go-darwin-amd64 -o ytdl-go

# Make executable
chmod +x ytdl-go

# Move to system path (optional)
sudo mv ytdl-go /usr/local/bin/

# Verify installation
ytdl-go -version
```

**Windows:**

1. Download `ytdl-go-windows-amd64.exe` from [releases](https://github.com/lvcoi/ytdl-go/releases/latest)
2. Rename to `ytdl-go.exe`
3. Add to your system PATH
4. Verify: `ytdl-go -version`

### Using `go install`

If you have Go 1.24+ installed:

```bash
go install github.com/lvcoi/ytdl-go@latest
```

> **Note:** Ensure your `$GOPATH/bin` is in your system `$PATH`.

### Build from Source

For the latest development version or if you want to contribute:

```bash
# Clone the repository
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go

# Build the binary
go build -o ytdl-go .

# (Optional) Install to system path
go install .
```

### One-Command Build Script

Use the integrated build script that builds both backend and frontend:

```bash
# Build everything
./build.sh

# Build and automatically launch the web UI
./build.sh --web

# Non-interactive build (for automation)
printf 'n\n' | ./build.sh
```

**What the script does:**

- Builds the Go binary to `./bin/yt`
- Builds frontend assets into `internal/web/assets/`
- Optionally launches the web UI

**Options:**

| Option | Description |
|--------|-------------|
| `-w`, `--web` | Automatically launch the UI after building |
| `-h`, `--help` | Show help message |

**Custom proxy target:**

```bash
VITE_API_PROXY_TARGET=http://127.0.0.1:9090 ./build.sh --web
```

---

## Verifying Installation

After installation, verify ytdl-go is working:

```bash
# Check version
ytdl-go -version

# Test with a simple download
ytdl-go -info https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

---

## Updating

### Pre-built Binaries

Download the latest release and replace your existing binary.

### `go install` Method

```bash
go install github.com/lvcoi/ytdl-go@latest
```

### From Source

```bash
cd ytdl-go
git pull origin main
go build -o ytdl-go .
```

---

## Uninstalling

### If installed to `/usr/local/bin`

```bash
sudo rm /usr/local/bin/ytdl-go
```

### If installed via `go install`

```bash
rm $(go env GOPATH)/bin/ytdl-go
```

---

## Next Steps

- [Quick Start Guide](quick-start) - Learn the basics
- [Configuration](configuration) - Configure ytdl-go for your needs
- [Basic Downloads](../usage/basic-downloads) - Start downloading videos
