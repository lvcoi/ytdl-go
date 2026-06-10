<div align="center">
<img src="img/ytdl-Gopher.png" alt="ytdl-gopher" width="500">

# ytdl-go

**A powerful, blazing-fast YouTube downloader written in Go — now with a full Web UI.**

[![v0.2.0-beta](https://img.shields.io/badge/release-v0.2.0--beta-blueviolet?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/releases)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/LICENSE)
[![Docs](https://img.shields.io/badge/Wiki_&_Docs-blue?style=for-the-badge&logo=readthedocs&logoColor=white)](https://lvcoi.github.io/ytdl-go/)

</div>

---

## TL;DR

**ytdl-go** downloads YouTube videos, audio, and playlists — fast. Use it from the **command line** or launch the brand-new **Web UI** (`ytdl-go -web`). Parallel downloads, automatic retries, resume support, ID3 tagging, multi-account profiles, playlist management, and an interactive TUI format selector are all built in.

> **v0.2.0 Beta is out!** This release introduces the Web UI with a customizable dashboard, media library with built-in player, real-time download progress, multi-account support, queue & playlist management, and a responsive collapsible sidebar — all served from a single binary. The CLI works exactly as before.
>
> 📖 **Full documentation lives on the [docs site](https://lvcoi.github.io/ytdl-go/) and the [GitHub Wiki](https://github.com/lvcoi/ytdl-go/wiki)** — installation guides, CLI reference, architecture docs, and more.

---

## ✨ Features

| Feature | Description |
| --- | --- |
| 🚀 **Parallel Downloads** | Multiple concurrent downloads with retry & resume |
| 🌐 **Web UI** | Dashboard, download queue, media library & player |
| 🎮 **Interactive TUI** | Browse and pick formats visually with `-list-formats` |
| 🎵 **Audio & Video** | Videos, audio-only, playlists, YouTube Music |
| 🏷️ **Metadata** | Automatic ID3 tags, JSON sidecars, custom templates |
| 👤 **Multi-Account** | Per-account state isolation with profile switching |
| 🎶 **Playlists & Queue** | Create playlists, assign songs to multiple lists, reorder queue |
| 📱 **Responsive UI** | Collapsible sidebar, gallery/list view modes, mobile-friendly layout |
| ⚙️ **Automation** | JSON output, quiet mode, custom output paths |

---

## 🚀 Quick Start

```bash
# Install (requires Go 1.24+)
go install github.com/lvcoi/ytdl-go@latest

# Download a video
ytdl-go "https://youtube.com/watch?v=..."

# Download audio only
ytdl-go -audio "https://youtube.com/watch?v=..."

# Launch the Web UI
ytdl-go -web
```

<details>
<summary><b>Build from source (with Web UI)</b></summary>

```bash
git clone https://github.com/lvcoi/ytdl-go.git && cd ytdl-go

# Build everything and launch the UI
./build.sh --web
```

</details>

---

## 🌐 Web UI

The v0.2.0 Web UI is a SolidJS single-page app served directly by the Go backend — no separate process needed.

Launch it with:

```bash
ytdl-go -web
```

**What you get:**

- **Dashboard** — Customizable widget grid with drag-and-drop layout, recent activity, quick download, and system stats.
- **Download View** — Paste one or more URLs, pick format & quality, and watch real-time progress over a WebSocket connection.
- **Media Library** — Browse all downloaded media with thumbnail gallery or list view, search, filters (Music / Videos / Podcasts), sortable columns, and hover-reveal actions.
- **Built-in Player** — Floating audio/video player with queue support (add, remove, reorder, clear) and minimized mode. Queue persists across page refreshes.
- **Playlists** — Create custom playlists, assign songs to multiple playlists via checkbox UI, and play or queue entire playlists at once.
- **Multi-Account** — Switch between profiles (e.g. Personal / Work) with per-account state isolation and independent localStorage.
- **Collapsible Sidebar** — Responsive navigation that auto-collapses on mobile, with hover tooltips in collapsed mode.
- **Settings** — Configure concurrency, storage paths, network options (cookie usage, PO Token), with nested sub-navigation.
- **Toast Notifications** — Success/error feedback with 3-second auto-dismiss.

### Screenshots

> **Note:** Screenshots below show the UI with sample data. Your instance will look similar after downloading content.

| Dashboard | Download | Library |
| :---: | :---: | :---: |
| ![Dashboard](img/screenshots/dashboard.png) | ![Download](img/screenshots/download.png) | ![Library](img/screenshots/library.png) |

*Glassmorphic dark UI with cyan/emerald accents — sidebar navigation, drag-and-drop dashboard widgets, real-time download progress, and a thumbnail gallery with built-in player.*

### Interactive CLI (TUI)

The command-line interface includes a visual format selector:

```bash
ytdl-go -list-formats "https://youtube.com/watch?v=..."
```

![Interactive Format Selector](img/interactive-format-selector.svg)

---

## 🛠️ Tech Stack

| Layer | Technology |
| --- | --- |
| **Backend** | [Go](https://golang.org/) 1.24 |
| **Frontend** | [SolidJS](https://www.solidjs.com/) + [Tailwind CSS](https://tailwindcss.com/) v4 |
| **TUI** | [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| **Database** | [SQLite](https://www.sqlite.org/) via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go) |
| **Real-time** | WebSocket ([gorilla/websocket](https://github.com/gorilla/websocket)) |
| **Media** | [FFmpeg](https://ffmpeg.org/) (muxing & metadata), [ID3v2](https://github.com/bogem/id3v2) (audio tags) |
| **Icons** | [Lucide](https://lucide.dev/) |
| **Docs** | [Zensical](https://zensical.org/) (Material Design docs) |
| **Build** | [Vite](https://vitejs.dev/) (frontend), `go build` (backend) |

---

## 📖 Documentation

Detailed docs are available in two places — pick whichever you prefer:

| Source | Link |
| --- | --- |
| 📘 **Docs Site** | **[lvcoi.github.io/ytdl-go](https://lvcoi.github.io/ytdl-go/)** — full documentation with search, navigation, and Material Design theme |
| 📗 **GitHub Wiki** | **[github.com/lvcoi/ytdl-go/wiki](https://github.com/lvcoi/ytdl-go/wiki)** — same content, accessible directly from the repo |

Both cover:

| Section | What's There |
| --- | --- |
| **[User Guide](https://lvcoi.github.io/ytdl-go/user-guide/installation/)** | Installation, quick start, format selection, playlists, output templates, troubleshooting |
| **[Developer Guide](https://lvcoi.github.io/ytdl-go/developer-guide/architecture/)** | Architecture, API reference, contributing, best practices |
| **[CLI Reference](https://lvcoi.github.io/ytdl-go/reference/cli-options/)** | Full list of command-line flags |

---

## 🔧 Troubleshooting

<details>
<summary><b>Common issues</b></summary>

- **403 Forbidden** — Automatic retries handle most cases. Try `-timeout 10m` if persistent.
- **Restricted content** — Private / age-gated videos are not yet supported.
- **Port in use** — `ytdl-go -web` auto-falls back to the next available port.
- **Missing thumbnails** — Re-download to generate sidecar metadata used by the library.

</details>

---

## 🙏 Acknowledgments

ytdl-go stands on the shoulders of these projects:

- **[yt-dlp](https://github.com/yt-dlp/yt-dlp)** — The gold standard YouTube downloader. ytdl-go was heavily inspired by yt-dlp's feature set and approach.
- **[kkdai/youtube](https://github.com/kkdai/youtube)** — The original Go YouTube library that ytdl-go forked and extended as [ytdl-lib](https://github.com/lvcoi/ytdl-lib).
- **[Charm](https://charm.sh/)** — [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss) power the interactive terminal UI.
- **[SolidJS](https://www.solidjs.com/)** — The reactive framework behind the Web UI.
- **[Zensical](https://zensical.org/)** — Documentation site generator (Material Design).
- **[Gorilla WebSocket](https://github.com/gorilla/websocket)** — Real-time communication in the Web UI.

---

<div align="center">

Made with ❤️ by the ytdl-go team

[License](LICENSE) · [Report Issue](https://github.com/lvcoi/ytdl-go/issues) · [Docs Site](https://lvcoi.github.io/ytdl-go/) · [Wiki](https://github.com/lvcoi/ytdl-go/wiki)

</div>
