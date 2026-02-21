<div align="center">

# ytdl-go

**A powerful, blazing-fast YouTube downloader written in Go.**
_Now featuring a brand-new browser-based Web UI — no CLI required._

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/blob/main/LICENSE)
[![Release](https://img.shields.io/github/release/lvcoi/ytdl-go.svg?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/releases)
[![GoDoc](https://img.shields.io/badge/reference-go.dev-007d9c?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/lvcoi/ytdl-go)

[Web UI](#-web-ui) • [Quick Start](#-quick-start) • [CLI](#-command-line-interface) • [Docs](#-documentation)

</div>

---

## ✨ What's New in v0.2.0

v0.2.0 ships a **completely new browser-based Web UI** — a sleek, dark-themed single-page app built with SolidJS and Tailwind CSS, embedded directly in the Go binary (no separate install needed). The original CLI is still fully supported for scripting and power users.

| | |
| --- | --- |
| 🖥️ **Web UI** | Point-and-click downloads, real-time progress, built-in media player |
| 📚 **Library** | Browse, filter, sort, and organize all your downloaded media |
| ⚙️ **Settings** | Auth cookies, PO token extension, output templates, duplicate handling |
| 🔒 **Auth** | YouTube cookie sync for age-restricted content |
| ⚡ **Same engine** | The same high-performance Go download core powers both interfaces |

---

## 🖥️ Web UI

Launch the Web UI with a single flag:

```bash
ytdl-go -web
# or build everything and launch automatically:
./build.sh --web
```

Then open **http://127.0.0.1:8080** in your browser.

### New Download

Paste one or more YouTube URLs, choose quality, format, and concurrency, then hit **Download**. Real-time progress bars and log output appear inline as the job runs.

![New Download view — paste URLs and configure options](https://github.com/user-attachments/assets/d8567e50-4a51-4e25-ab66-68edbbea372d)

> **Features:** multi-URL input · quality & format selectors · audio-only toggle · concurrent jobs · real-time SSE progress stream with per-file progress bars · inline download log

### Active Download Progress

Watch downloads in real time — per-file progress bars show bytes, speed, and ETA. A live log stream displays metadata fetch, format selection, and completion events.

![Active download — real-time progress bar and log stream](https://github.com/user-attachments/assets/f7fcd737-1b72-40c2-9a69-71a9fd9232f2)

> **Features:** per-file progress bars · download speed & ETA · inline log panel · automatic reconnect on network interruption (up to 5 retries with exponential back-off)

### Library

All your downloaded media in one place. Switch between Video and Audio tabs, filter by artist/creator, album/channel, or source playlist, and sort by date, creator, or collection. Hit ▶ to play right in the browser.

![Library view — browse, filter, and play downloaded media](https://github.com/user-attachments/assets/9ba2320f-3160-451a-8d82-8e6c1d1e936a)

> **Features:** video/audio tabs with counts · artist/channel/playlist filters · saved custom playlists · multi-column sort · built-in media player (video & audio) · persistent filter state across reloads

### Configurations

Manage YouTube auth cookies, enable the PO Token extension for restricted content, set a default output template, and choose how duplicate files are handled.

![Configurations view — auth cookies, PO token, output template, duplicate handling](https://github.com/user-attachments/assets/7ddd0e9f-7770-4284-9086-b95315674aeb)

> **Features:** YouTube cookie re-sync · PO Token extension toggle · default `{title}.{ext}` output template · duplicate file policy (prompt / overwrite / skip / auto-rename)

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+** — [download](https://golang.org/dl/)
- **Node.js / npm** — required only if building the frontend from source

### Option A — `go install` (CLI only)

```shell
go install github.com/lvcoi/ytdl-go@latest
ytdl-go -web   # launches the Web UI
```

### Option B — Build script (recommended for Web UI)

```bash
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go
./build.sh --web   # builds Go binary + frontend, then opens the UI
```

The binary is written to `./bin/yt`. The Web UI is embedded — no separate frontend server needed.

### Option C — Manual build

```bash
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go

# Build frontend assets (requires npm)
cd frontend && npm install && npm run build && cd ..

# Build Go binary
go build -o ytdl-go .

# Launch Web UI
./ytdl-go -web
```

---

## 💻 Command-Line Interface

The CLI remains fully functional for scripting, automation, and power users.

```bash
# Download best quality video
ytdl-go https://www.youtube.com/watch?v=BaW_jenozKc

# Download audio-only
ytdl-go -audio https://www.youtube.com/watch?v=BaW_jenozKc

# Interactive format selector (TUI)
ytdl-go -list-formats https://www.youtube.com/watch?v=BaW_jenozKc

# Organize a music playlist
ytdl-go -audio -o "Music/{artist}/{album}/{index} - {title}.{ext}" PLAYLIST_URL
```

For the complete CLI reference — all flags, output template placeholders, concurrency options, scripting examples, and troubleshooting — see **[docs/CLI.md](docs/CLI.md)**.

---

## 📚 Documentation

| Document | Description |
| --- | --- |
| [docs/CLI.md](docs/CLI.md) | Full command-line reference: all flags, templates, examples, troubleshooting |
| [docs/FLAGS.md](docs/FLAGS.md) | Detailed per-flag reference with types, defaults, and edge cases |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Internal architecture, data-flow diagrams, concurrency model |
| [frontend/docs/ARCHITECTURE.md](frontend/docs/ARCHITECTURE.md) | Frontend SPA architecture (SolidJS, Vite, Tailwind CSS) |
| [frontend/docs/API.md](frontend/docs/API.md) | Backend REST + SSE API contract used by the Web UI |
| [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) | How to contribute (workflow, tests, style) |
| [docs/MAINTAINERS.md](docs/MAINTAINERS.md) | Release process, dependency management, security |
| [SECURITY.md](SECURITY.md) | Security policy and vulnerability reporting |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Top-level contribution guidelines |

---

## 🛡️ Security

Please **do not** open public issues for security vulnerabilities. Follow the process in [SECURITY.md](SECURITY.md).

The repository uses [Gitleaks](https://github.com/gitleaks/gitleaks) for automated secret scanning and a pre-commit hook to prevent accidental credential commits. Run `./setup-hooks.sh` once after cloning to enable it.

---

<div align="center">

Made with ❤️ by the ytdl-go team

[License](LICENSE) • [Report Issue](https://github.com/lvcoi/ytdl-go/issues) • [Discussions](https://github.com/lvcoi/ytdl-go/discussions)

</div>
