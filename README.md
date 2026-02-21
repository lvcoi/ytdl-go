<div align="center">
<img src="img/ytdl-Gopher.png" alt="ytdl-gopher">

# ytdl-go

**A powerful, blazing fast YouTube downloader written in Go.** _Feature-rich, interactive, with a brand-new Web UI._

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/LICENSE)
[![Release](https://img.shields.io/github/release/lvcoi/ytdl-go.svg?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/releases)
[![GoDoc](https://img.shields.io/badge/reference-go.dev-007d9c?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/lvcoi/ytdl-go)

[Web UI](#web-ui) &bull; [Installation](#installation) &bull; [CLI](#cli-usage) &bull; [Docs](#documentation)

</div>

---

## Web UI

v0.2.0 ships a **brand-new browser-based Web UI** -- a sleek, dark-themed single-page app built with SolidJS and Tailwind CSS, embedded directly in the Go binary. Launch it with:

```bash
ytdl-go -web
# or build + launch in one step:
./build.sh --web
```

Then open **http://127.0.0.1:8080** in your browser. The original CLI is still fully supported.

### New Download

Paste one or more YouTube URLs, choose quality, format, and concurrency, then hit **Download**. Real-time progress bars and log output appear inline as the job runs.

![New Download view](https://github.com/user-attachments/assets/d8567e50-4a51-4e25-ab66-68edbbea372d)

> **Features:** multi-URL input, quality & format selectors, audio-only toggle, concurrent jobs, real-time SSE progress stream, inline log

### Active Download Progress

Watch downloads in real time -- per-file progress bars show bytes, speed, and ETA alongside a live log stream.

![Active download with real-time progress](https://github.com/user-attachments/assets/f7fcd737-1b72-40c2-9a69-71a9fd9232f2)

> **Features:** per-file progress bars, speed & ETA, inline log panel, auto-reconnect (up to 5 retries, exponential back-off)

### Library

All your downloaded media in one place. Filter by artist/creator, album/channel, or source playlist and hit play to watch or listen right in the browser.

![Library view](https://github.com/user-attachments/assets/9ba2320f-3160-451a-8d82-8e6c1d1e936a)

> **Features:** video/audio tabs, artist/channel/playlist filters, saved custom playlists, multi-column sort, built-in media player

### Configurations

Manage YouTube auth cookies, PO Token extension, default output template, and duplicate file policy.

![Configurations view](https://github.com/user-attachments/assets/7ddd0e9f-7770-4284-9086-b95315674aeb)

> **Features:** YouTube cookie re-sync, PO Token extension toggle, `{title}.{ext}` output template, duplicate file policy

---

## Features at a Glance

| Core Capabilities | Advanced Tools |
| --- | --- |
| **High Performance**<br>Parallel downloads, automatic retries, and resume capability. | **Interactive TUI**<br>Visual format selector to browse and pick specific quality streams. |
| **Broad Support**<br>Download Videos, Audio, Playlists, and YouTube Music URLs. | **Rich Metadata**<br>Embeds ID3 tags, fetches structured JSON metadata, and handles sidecars. |
| **Web UI**<br>Full-featured browser interface with real-time progress and media library. | **Automation Ready**<br>JSON output mode, custom output templates, and quiet modes for scripts. |

---

## Installation

### Quick Install (Recommended)

Requires Go 1.24+ installed on your system.

```shell
go install github.com/lvcoi/ytdl-go@latest
```

*Ensure your `$GOPATH/bin` is in your system `$PATH`.*

### Build from Source

<details>
<summary>Click to expand build instructions</summary>

```shell
# Clone the repository
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go

# Build the binary
go build -o ytdl-go .

# (Optional) Install to system path
go install .
```

</details>

### One-command build script (build.sh)

Use `build.sh` for an integrated repository build (Go binary + frontend assets).

```bash
# Build backend + frontend, then prompt to launch the UI
./build.sh

# Build and automatically launch backend + frontend UI
./build.sh --web

# Build and launch against a backend on a custom requested web address
YTDL_WEB_ADDR=127.0.0.1:9090 ./build.sh --web

# Build and launch UI against an explicit API proxy target
VITE_API_PROXY_TARGET=http://127.0.0.1:9090 ./build.sh --web

# Show script options
./build.sh --help
```

What it does:

- Builds the Go binary to `./bin/yt`
- Builds frontend assets into `internal/web/assets/`
- Prompts to launch backend + frontend dev UI unless `--web` is passed
- Launches backend via `yt -web --web-addr "${YTDL_WEB_ADDR:-127.0.0.1:8080}"`
- Uses `VITE_API_PROXY_TARGET` only when explicitly set; otherwise Vite auto-detects backend from `http://127.0.0.1:8080` + fallback ports
- `ytdl-go -web` auto-falls back to the next available port if the requested port is already in use, and logs the final URL

Options:

| Option | Description |
| --- | --- |
| `-w`, `--web` | Automatically launch the UI after building |
| `-h`, `--help` | Show help message |

Automation example (non-interactive "do not launch UI" path):

```bash
printf 'n\n' | ./build.sh
```

---

## CLI Usage

### Basic usage

```bash
# Download video with best quality
ytdl-go https://www.youtube.com/watch?v=BaW_jenozKc
```

![Video download](screenshots/05-video-download.svg)

```bash
# Download audio-only
ytdl-go --audio https://www.youtube.com/watch?v=BaW_jenozKc
```

![Audio download](screenshots/06-audio-download.svg)

### The Essentials

| Goal | Command |
| --- | --- |
| **Download Best Video** | `ytdl-go "https://youtube.com/watch?v=..."` |
| **Download Audio Only** | `ytdl-go -audio "https://youtube.com/watch?v=..."` |
| **Interactive Mode** | `ytdl-go -list-formats "https://youtube.com/watch?v=..."` |
| **Download Playlist** | `ytdl-go "https://youtube.com/playlist?list=..."` |
| **Launch Web UI** | `ytdl-go -web` |

### Interactive Mode (TUI)

Don't guess the quality code. Use `-list-formats` to browse streams visually.

```shell
ytdl-go -list-formats https://www.youtube.com/watch?v=BaW_jenozKc
```

> **Controls:** up/down to navigate, Enter to download, type digits for itag (e.g., `101`), repeat quickly to cycle.

![Interactive Format Selector](screenshots/interactive-format-selector.svg)

### File Organization & Templates

Customize where files go using the `-o` flag with placeholders.

**Organize Music by Artist:**

```shell
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" URL
```

**Archive Playlists with Index:**

```shell
ytdl-go -o "Archive/{playlist-title}/{index} - {title}.{ext}" URL
```

For the complete CLI reference -- all flags, output template placeholders, concurrency options, scripting examples, and troubleshooting -- see **[docs/CLI.md](docs/CLI.md)**.

---

## Advanced Usage

### Format Selection & Quality

```shell
# Specific resolution
ytdl-go -quality 1080p URL

# Specific container
ytdl-go -format mp4 URL

# Exact YouTube itag
ytdl-go -itag 137 URL
```

### Network & Performance

```shell
# Parallel downloads (4 files at once)
ytdl-go -jobs 4 URL1 URL2...

# Custom Timeout
ytdl-go -timeout 5m URL

# Resume downloads
ytdl-go URL  # (Automatic detection of .part files)
```

### Scripting & Metadata

```shell
# Get JSON metadata (no download)
ytdl-go -info -json URL

# Override metadata manually
ytdl-go -meta artist="My Artist" -meta title="Custom Title" URL

# Quiet mode (no progress bars)
ytdl-go -quiet URL
```

---

## Documentation

| Document | Description |
| --- | --- |
| [docs/CLI.md](docs/CLI.md) | Full CLI reference: all flags, templates, examples, troubleshooting |
| [docs/FLAGS.md](docs/FLAGS.md) | Detailed per-flag reference with types, defaults, and edge cases |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Internal architecture, data-flow diagrams, concurrency model |
| [frontend/docs/ARCHITECTURE.md](frontend/docs/ARCHITECTURE.md) | Frontend SPA architecture (SolidJS, Vite, Tailwind CSS) |
| [frontend/docs/API.md](frontend/docs/API.md) | Backend REST + SSE API contract used by the Web UI |
| [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) | How to contribute (workflow, tests, style) |
| [docs/MAINTAINERS.md](docs/MAINTAINERS.md) | Release process, dependency management, security |

---

## Troubleshooting

<details>
<summary><b>Common Issues & Fixes</b></summary>

- **403 Forbidden Errors:** The tool automatically retries with different methods. If persistent, check your IP reputation or try `-timeout 10m`.
- **Restricted Content:** Private, age-gated, or member-only videos require authentication which is currently **not supported**.
- **Playlists:** Empty videos or deleted entries in playlists are automatically skipped.
- **Library Metadata/Thumbnails:** New downloads write sidecar metadata (`<media-file>.json`) used by the web UI for artist/album/thumbnail grouping. Legacy files without sidecars still load, but may appear under Unknown buckets until re-downloaded.
- **Web Port Already in Use:** `ytdl-go -web` retries on higher ports automatically. Check startup logs for the selected URL and point Vite with `VITE_API_PROXY_TARGET=http://127.0.0.1:<port>`.

</details>

--

<div align="center">

Made with ❤️ by the ytdl-go team

[License](LICENSE) &bull; [Report Issue](https://github.com/lvcoi/ytdl-go/issues)

</div>
