# Installation

## Quick Install (Recommended)

Requires Go 1.24+ installed on your system.

```bash
go install github.com/lvcoi/ytdl-go@latest
```

Ensure your `$GOPATH/bin` is in your system `$PATH`.

## Build from Source

```bash
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go
go build -o ytdl-go .
```

Or install directly to `$GOPATH/bin`:

```bash
go install .
```

## Build Script

Use `build.sh` for an integrated build (Go binary + frontend assets):

```bash
# Build backend + frontend, then prompt to launch the UI
./build.sh

# Build and automatically launch backend + frontend UI
./build.sh --web

# Build and launch against a custom web address
YTDL_WEB_ADDR=127.0.0.1:9090 ./build.sh --web

# Show script options
./build.sh --help
```

| Option | Description |
| --- | --- |
| `-w`, `--web` | Automatically launch the UI after building |
| `-h`, `--help` | Show help message |

The build script:

- Builds the Go binary to `./bin/yt`
- Builds frontend assets into `internal/web/assets/`
- Launches backend via `yt -web --web-addr "${YTDL_WEB_ADDR:-127.0.0.1:8080}"`

## Optional Dependencies

**FFmpeg** enables the audio extraction fallback strategy when YouTube blocks direct audio-only downloads. The tool works without it, but some downloads may fail that would otherwise succeed.

=== "macOS"

    ```bash
    brew install ffmpeg
    ```

=== "Ubuntu/Debian"

    ```bash
    sudo apt-get install ffmpeg
    ```

=== "Windows"

    ```bash
    choco install ffmpeg
    ```
