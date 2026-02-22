# ytdl-go

A YouTube downloader written in Go with a terminal UI and web interface.

## Features

| Core Capabilities | Advanced Tools |
| --- | --- |
| **High Performance** — Parallel downloads, automatic retries, and resume capability. | **Interactive TUI** — Visual format selector to browse and pick specific quality streams. |
| **Broad Support** — Download videos, audio, playlists, and YouTube Music URLs. | **Rich Metadata** — Embeds ID3 tags, fetches structured JSON metadata, and handles sidecars. |
| **Format Control** — Select by quality (`1080p`, `best`), container (`mp4`), or exact `itag`. | **Automation Ready** — JSON output mode, custom output templates, and quiet modes for scripts. |

## Quick Start

```bash
# Install
go install github.com/lvcoi/ytdl-go@latest

# Download a video
ytdl-go https://www.youtube.com/watch?v=BaW_jenozKc

# Download audio only
ytdl-go -audio https://www.youtube.com/watch?v=BaW_jenozKc

# Launch the web UI
ytdl-go -web
```

See the [Installation](user-guide/installation.md) and [Quick Start](user-guide/quick-start.md) guides for details.

## Documentation Sections

- **[User Guide](user-guide/installation.md)** — Installation, usage, output templates, troubleshooting.
- **[Developer Guide](developer-guide/architecture.md)** — Architecture, API reference, contributing.
- **[Reference](reference/cli-options.md)** — CLI flags, roadmap, security policy.
