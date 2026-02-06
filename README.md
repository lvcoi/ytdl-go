<div align="center">

# ytdl-go

**A powerful, blazing fast YouTube downloader written in Go.** _Feature-rich, interactive, and dependency-free._

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/LICENSE)
[![Release](https://img.shields.io/github/release/lvcoi/ytdl-go.svg?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/releases)
[![GoDoc](https://img.shields.io/badge/reference-go.dev-007d9c?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/lvcoi/ytdl-go)

[Features](#-features-at-a-glance) • [Installation](#-installation) • [Usage](#-usage-guide) • [Options](#-command-line-options)

</div>

---

## ✨ Features at a Glance

| Core Capabilities | Advanced Tools |
| --- | --- |
| **🚀 High Performance**<br>Parallel downloads, automatic retries, and resume capability. | **🎮 Interactive TUI**<br>Visual format selector to browse and pick specific quality streams. |
| **📺 Broad Support**<br>Download Videos, Audio, Playlists, and YouTube Music URLs. | **🏷️ Rich Metadata**<br>Embeds ID3 tags, fetches structured JSON metadata, and handles sidecars. |
| **🎨 Format Control**<br>Select by quality (`1080p`, `best`), container (`mp4`), or exact `itag`. | **⚙️ Automation Ready**<br>JSON output mode, custom output templates, and quiet modes for scripts. |

---

## 🚀 Installation

### Quick Install (Recommended)

Requires Go 1.23+ installed on your system.

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

---

## 📖 Usage Guide

### 🎯 The Essentials

| Goal | Command |
| --- | --- |
| **Download Best Video** | `ytdl-go "https://youtube.com/watch?v=..."` |
| **Download Audio Only** | `ytdl-go -audio "https://youtube.com/watch?v=..."` |
| **Interactive Mode** | `ytdl-go -list-formats "https://youtube.com/watch?v=..."` |
| **Download Playlist** | `ytdl-go "https://youtube.com/playlist?list=..."` |

### 🎮 Interactive Mode (TUI)

Don't guess the quality code. Use `-list-formats` to browse streams visually.

```shell
ytdl-go -list-formats https://www.youtube.com/watch?v=BaW_jenozKc
```

> **Controls:** `↑/↓` to navigate, `Enter` to download, `1-9` to filter by itag.

![Interactive Format Selector](screenshots/interactive-format-selector.svg)

### 📂 File Organization & Templates

Customize where files go using the `-o` flag with placeholders.

**Organize Music by Artist:**

```shell
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" URL
```

**Archive Playlists with Index:**

```shell
ytdl-go -o "Archive/{playlist-title}/{index} - {title}.{ext}" URL
```

**Supported Placeholders:**
`{title}`, `{artist}`, `{album}`, `{id}`, `{ext}`, `{quality}`, `{playlist-title}`, `{playlist-id}`, `{index}`, `{count}`

---

## ⚙️ Advanced Usage

### 📺 Format Selection & Quality

```shell
# Specific resolution
ytdl-go -quality 1080p URL

# Specific container
ytdl-go -format mp4 URL

# Exact YouTube itag
ytdl-go -itag 137 URL
```

### 🛠️ Network & Performance

```shell
# Parallel downloads (4 files at once)
ytdl-go -jobs 4 URL1 URL2...

# Custom Timeout
ytdl-go -timeout 5m URL

# Resume downloads
ytdl-go URL  # (Automatic detection of .part files)
```

### 🤖 Scripting & Metadata

```shell
# Get JSON metadata (no download)
ytdl-go -info -json URL

# Override metadata manually
ytdl-go -meta artist="My Artist" -meta title="Custom Title" URL

# Quiet mode (no progress bars)
ytdl-go -quiet URL
```

---

## 📊 Command Line Options

| Flag | Default | Description |
| --- | --- | --- |
| `-o` | `{title}.{ext}` | Output template. |
| `-audio` | `false` | Download best audio-only format. |
| `-list-formats` | `false` | Launch interactive format selector. |
| `-quality` | `best` | Target quality (`1080p`, `720p`, `worst`). |
| `-format` | `` | Preferred container (`mp4`, `webm`, `m4a`). |
| `-jobs` | `1` | Concurrent download jobs. |
| `-json` | `false` | Output logs/status as JSON lines. |
| `-quiet` | `false` | Suppress standard output. |
| `-meta` | `` | Override metadata field (`key=value`). |

---

## 🔧 Troubleshooting

<details>
<summary><b>Common Issues & Fixes</b></summary>

- **403 Forbidden Errors:** The tool automatically retries with different methods. If persistent, check your IP reputation or try `-timeout 10m`.
- **Restricted Content:** Private, age-gated, or member-only videos require authentication which is currently **not supported**.
- **Playlists:** Empty videos or deleted entries in playlists are automatically skipped.

</details>

---

<div align="center">

Made with ❤️ by the ytdl-go team

[License](LICENSE) • [Report Issue](https://github.com/lvcoi/ytdl-go/issues)

</div>
