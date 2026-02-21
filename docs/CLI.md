# CLI Reference

This document covers everything you need to use ytdl-go from the command line. For the new browser-based Web UI, see the main [README](../README.md).

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Interactive Format Selector](#interactive-format-selector)
- [Output Templates](#output-templates)
- [Command-Line Options](#command-line-options)
- [Advanced Usage](#advanced-usage)
- [Examples by Use Case](#examples-by-use-case)
- [Troubleshooting](#troubleshooting)

---

## Installation

### Quick Install (Recommended)

Requires Go 1.24+.

```shell
go install github.com/lvcoi/ytdl-go@latest
```

Ensure `$GOPATH/bin` is in your `$PATH`.

### Build from Source

```shell
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go
go build -o ytdl-go .
```

### One-command build script (`build.sh`)

Use `build.sh` for an integrated build that compiles both the Go binary and the frontend assets.

```bash
# Build backend + frontend, then prompt to launch the Web UI
./build.sh

# Build and automatically launch the Web UI
./build.sh --web

# Build with a custom backend proxy port
VITE_API_PROXY_TARGET=http://127.0.0.1:9090 ./build.sh --web

# Show all options
./build.sh --help
```

**What it does:**
- Builds the Go binary to `./bin/yt`
- Builds frontend assets into `internal/web/assets/`
- Prompts to launch the Web UI unless `--web` is passed

**Options:**

| Option | Description |
| --- | --- |
| `-w`, `--web` | Automatically launch the Web UI after building |
| `-h`, `--help` | Show help message |

**Non-interactive (scripting):**

```bash
printf 'n\n' | ./build.sh
```

---

## Basic Usage

```bash
# Download best quality video
ytdl-go https://www.youtube.com/watch?v=BaW_jenozKc
```

![Video download](../screenshots/05-video-download.svg)

```bash
# Download audio-only
ytdl-go --audio https://www.youtube.com/watch?v=BaW_jenozKc
```

![Audio download](../screenshots/06-audio-download.svg)

```bash
# Get video metadata without downloading
ytdl-go --info https://www.youtube.com/watch?v=BaW_jenozKc
```

### The Essentials

| Goal | Command |
| --- | --- |
| **Download Best Video** | `ytdl-go "https://youtube.com/watch?v=..."` |
| **Download Audio Only** | `ytdl-go -audio "https://youtube.com/watch?v=..."` |
| **Interactive Mode** | `ytdl-go -list-formats "https://youtube.com/watch?v=..."` |
| **Download Playlist** | `ytdl-go "https://youtube.com/playlist?list=..."` |

---

## Interactive Format Selector

Use `-list-formats` to browse all available streams and pick a specific one.

```shell
ytdl-go -list-formats https://www.youtube.com/watch?v=BaW_jenozKc
```

![Interactive Format Selector](../screenshots/interactive-format-selector.svg)

**Keyboard controls:**

| Key | Action |
| --- | --- |
| `↑` / `↓` or `j` / `k` | Navigate the list |
| `Enter` | Download selected format |
| `1–9` | Quick-jump to itag starting with that digit |
| `Page Up` / `Page Down` | Jump 10 entries |
| `Home` / `End` or `g` / `G` | Jump to first / last |
| `b` | Go back without downloading |
| `q` / `Esc` / `Ctrl+C` | Quit |

---

## Output Templates

Customize filenames and directories with the `-o` flag.

**Supported placeholders:**

| Placeholder | Description | Example |
| --- | --- | --- |
| `{title}` | Video title (sanitized) | `My Video Title` |
| `{artist}` | Author/artist | `Artist Name` |
| `{album}` | Album (YouTube Music) | `Album Name` |
| `{id}` | YouTube video ID | `dQw4w9WgXcQ` |
| `{ext}` | File extension | `mp4`, `webm` |
| `{quality}` | Quality label | `1080p`, `128k` |
| `{playlist-title}` | Playlist name | `My Playlist` |
| `{playlist-id}` | Playlist ID | `PLxxx...` |
| `{index}` | Video position in playlist (1-based) | `1`, `2`, `3` |
| `{count}` | Total playlist video count | `25` |

**Examples:**

```bash
# Music organized by artist and album
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" URL

# Playlist with numbered files
ytdl-go -o "Archive/{playlist-title}/{index} - {title}.{ext}" URL

# Quality in filename
ytdl-go -o "Videos/{title} [{quality}].{ext}" URL
```

---

## Command-Line Options

### Core Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-o` | `{title}.{ext}` | Output filename/path template |
| `-output-dir` | (none) | Constrain all output to this directory (prevents path traversal) |
| `-audio` | `false` | Download best audio-only format |
| `-list-formats` | `false` | Launch interactive format selector |
| `-quality` | `best` | Target quality: `1080p`, `720p`, `best`, `worst`, etc. |
| `-format` | (any) | Preferred container: `mp4`, `webm`, `m4a`, etc. |
| `-itag` | `0` | Download a specific YouTube itag (overrides `-quality`/`-format`) |

### Network Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-timeout` | `3m` | Per-request timeout (e.g. `30s`, `5m`, `1h`) |

### Concurrency Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-jobs` | `1` | Number of URLs to download simultaneously |
| `-segment-concurrency` | `0` (auto) | Parallel segment downloads for HLS/DASH streams |
| `-playlist-concurrency` | `0` | Reserved for future parallel playlist downloads (currently ignored) |

### Output Control

| Flag | Default | Description |
| --- | --- | --- |
| `-quiet` | `false` | Suppress progress output (errors still shown) |
| `-json` | `false` | Emit newline-delimited JSON events instead of human-readable output |
| `-info` | `false` | Print metadata as JSON and exit without downloading |
| `-log-level` | `info` | Log verbosity: `debug`, `info`, `warn`, `error` |
| `-progress-layout` | (built-in) | Custom progress bar format (placeholders: `{label}`, `{percent}`, `{current}`, `{total}`, `{rate}`, `{eta}`) |

### Metadata Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-meta key=value` | (none) | Override a metadata field; repeatable (e.g. `-meta title="My Song" -meta artist="Me"`) |

### Web UI Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-web` | `false` | Start the Web UI server instead of downloading |
| `-web-addr` | `127.0.0.1:8080` | Address and port for the Web UI server |

---

## Advanced Usage

### Format Selection

```bash
# Specific resolution
ytdl-go -quality 1080p URL

# Specific container
ytdl-go -format mp4 URL

# Exact itag
ytdl-go -itag 137 URL
```

### Network & Performance

```bash
# Download 4 URLs concurrently
ytdl-go -jobs 4 URL1 URL2 URL3 URL4

# Extended timeout for slow connections
ytdl-go -timeout 10m URL
```

### Scripting & Automation

```bash
# JSON output (machine-readable)
ytdl-go -json URL

# Get metadata only
ytdl-go -info -json URL | jq .title

# Quiet mode for background scripts
ytdl-go -quiet URL
```

---

## Examples by Use Case

### Music Library Archiving

```bash
ytdl-go -audio \
  -o "Music/{artist}/{album}/{index} - {title}.{ext}" \
  -meta album="Album Name" \
  PLAYLIST_URL
```

### High-Quality Video Collection

```bash
ytdl-go -quality 1080p \
  -format mp4 \
  -o "Videos/{playlist-title}/{index} - {title} [{quality}].{ext}" \
  PLAYLIST_URL
```

### Batch Download with JSON Output

```bash
ytdl-go -json -jobs 4 -quiet \
  URL1 URL2 URL3 URL4 \
  > downloads.jsonl
```

### Debug Mode

```bash
ytdl-go -log-level debug -timeout 30m -list-formats URL
```

### Flag Precedence

1. `-itag` overrides `-quality` and `-format`
2. `-json` implies `-quiet`
3. `-info` exits after printing metadata (no download)
4. `-list-formats` prevents automatic format selection
5. `-json -list-formats` emits a machine-readable `{"type":"formats", ...}` payload instead of the interactive TUI

---

## Troubleshooting

**403 Forbidden Errors**
The tool automatically retries with different strategies. If errors persist, check your network or try `-timeout 10m`. Audio-only formats may fall back to FFmpeg extraction automatically.

**Restricted Content**
Private, age-gated, or member-only videos require authentication, which is not currently supported via the CLI.

**Playlists with missing videos**
Deleted or private playlist entries are logged as errors and skipped; the download continues for remaining entries.

**FFmpeg not found**
FFmpeg is optional but required for the audio-extraction fallback. Install it via your package manager (`brew install ffmpeg`, `apt install ffmpeg`, etc.) and ensure it is on your `$PATH`.

---

**See also:** [README](../README.md) • [Architecture](ARCHITECTURE.md) • [Flags Reference](FLAGS.md) • [Contributing](CONTRIBUTING.md)
