# ytdl-go 📺

![ytdl-go logo](screenshots/logo.svg)

## A powerful yt-dlp-style downloader written in Go

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](https://github.com/lvcoi/ytdl-go/LICENSE)
[![Release](https://img.shields.io/github/release/lvcoi/ytdl-go.svg?style=flat-square)](https://github.com/lvcoi/ytdl-go/releases)
[![GoDoc](https://img.shields.io/badge/go.dev-reference-007d9c?style=flat-square&logo=go)](https://pkg.go.dev/github.com/lvcoi/ytdl-go)

⚡ Blazing fast YouTube downloader with automatic retry, progress tracking, and YouTube Music support ⚡

---

## 📑 Table of Contents

- [✨ Features](#-features)
- [🚀 Installation](#-installation)
- [📖 Quick Start](#-quick-start)
- [💡 Usage Examples](#-usage-examples)
- [📊 Command Line Options](#-command-line-options)
- [🏷️ Output Templates](#️-output-templates)
- [🔧 Troubleshooting](#-troubleshooting)
- [🙏 Acknowledgments](#-acknowledgments)
- [📜 License](#-license)

---

## ✨ Features

| Core Capabilities | Advanced Tools |
| :--- | :--- |
| **🚀 High Performance**<br>Parallel downloads, automatic retries, and resume capability | **🎮 Interactive TUI**<br>Visual format selector to browse and pick quality streams |
| **📺 Broad Support**<br>Videos, Audio, Playlists, and YouTube Music URLs | **🏷️ Rich Metadata**<br>ID3 tags, JSON metadata, and sidecar files |
| **🎨 Format Control**<br>Select by quality (`1080p`, `best`), container (`mp4`), or itag | **⚙️ Automation Ready**<br>JSON output, custom templates, and quiet modes |

---

## 🚀 Installation

### Quick Install (Recommended)

```bash
go install github.com/lvcoi/ytdl-go@latest
```

*Requires Go 1.23+. Ensure `$GOPATH/bin` is in your `$PATH`.*

### Build from Source

```bash
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go
go build -o ytdl-go .
```

---

## 📖 Quick Start

### Basic Downloads

Download a video with best quality:
```bash
ytdl-go "https://youtube.com/watch?v=BaW_jenozKc"
```

![Video Download](screenshots/05-video-download.svg)

Download audio only:
```bash
ytdl-go -audio "https://youtube.com/watch?v=BaW_jenozKc"
```

![Audio Download](screenshots/06-audio-download.svg)

### Interactive Format Selection

Browse available formats and select interactively:
```bash
ytdl-go -list-formats "https://youtube.com/watch?v=BaW_jenozKc"
```

**Controls:** `↑/↓` navigate • `Enter` download • `1-9` filter by itag • `q` quit

### Playlist Downloads

Download entire playlists:
```bash
ytdl-go "https://youtube.com/playlist?list=PL59FEE129ADFF2B12"
```

![Playlist Download](screenshots/11-playlist-download.svg)

---

## 💡 Usage Examples

### Music Library Organization

**By Artist & Album:**
```bash
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" URL
```

![Audio with Folders](screenshots/13-audio-playlist-folders.svg)

**Playlist with Track Numbers:**
```bash
ytdl-go -audio -o "Music/{playlist-title}/{index:02d} - {title}.{ext}" URL
```

![Custom Playlist](screenshots/12-custom-playlist.svg)

### Video Collection

**With Quality Labels:**
```bash
ytdl-go -o "Videos/{title} [{quality}].{ext}" URL
```

![Quality Indicator](screenshots/10-quality-indicator.svg)

### Advanced Usage

**Get Metadata Only:**
```bash
ytdl-go -info URL
```

![Metadata Info](screenshots/07-metadata-info.svg)

**Specific Quality & Format:**
```bash
ytdl-go -quality 720p -format mp4 URL
```

**Parallel Downloads:**
```bash
ytdl-go -jobs 4 URL1 URL2 URL3 URL4
```

![Multiple URLs](screenshots/17-multiple-urls.svg)

**Quiet Mode (for scripts):**
```bash
ytdl-go -quiet -audio URL
```

![Quiet Mode](screenshots/15-quiet-mode.svg)

---

## 📊 Command Line Options

| Flag | Default | Description |
| :--- | :--- | :--- |
| `-o` | `{title}.{ext}` | Output template (see [Output Templates](#️-output-templates)) |
| `-audio` | `false` | Download best audio-only format |
| `-info` | `false` | Print video metadata as JSON without downloading |
| `-list-formats` | `false` | Launch interactive format selector |
| `-quality` | `best` | Target quality (`1080p`, `720p`, `128k`, `best`, `worst`) |
| `-format` | `` | Preferred container (`mp4`, `webm`, `m4a`) |
| `-itag` | `0` | Download specific format by itag number |
| `-meta` | `` | Override metadata (`key=value`, repeatable) |
| `-jobs` | `1` | Number of concurrent downloads |
| `-json` | `false` | Output as JSON (suppresses progress) |
| `-quiet` | `false` | Suppress progress output |
| `-timeout` | `3m` | Per-request timeout (e.g., `30s`, `5m`, `1h`) |
| `-log-level` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `-progress-layout` | `` | Custom progress template |
| `-segment-concurrency` | `auto` | Parallel segment downloads for HLS/DASH |
| `-playlist-concurrency` | `auto` | Parallel playlist entry downloads |

---

## 🏷️ Output Templates

Use the `-o` flag with placeholders to customize output paths:

| Placeholder | Description | Example |
| :--- | :--- | :--- |
| `{title}` | Video title (sanitized) | `My Video Title` |
| `{artist}` | Video author/artist | `Artist Name` |
| `{album}` | Album name (YouTube Music) | `Album Name` |
| `{id}` | Video ID | `dQw4w9WgXcQ` |
| `{ext}` | File extension | `mp4`, `webm`, `m4a` |
| `{quality}` | Quality label or bitrate | `1080p`, `128k` |
| `{playlist-title}` | Playlist name | `My Playlist` |
| `{playlist-id}` | Playlist ID | `PL59FEE129ADFF2B12` |
| `{index}` | Video index in playlist | `1`, `2`, `3` |
| `{count}` | Total videos in playlist | `25` |

**Template Examples:**
```bash
# Organize by artist and album
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" URL

# Playlist with index
ytdl-go -o "{playlist-title}/{index:02d} - {title}.{ext}" URL

# Quality in filename
ytdl-go -o "Videos/{title} [{quality}].{ext}" URL
```

---

## 🧾 Metadata & Sidecars

Each successful download emits a sidecar JSON file alongside the media (`<output>.json`). Playlist downloads also emit a `playlist.json` manifest in the output folder.

Metadata sources (in order of preference):

- Platform/structured metadata from the extractor (e.g., YouTube API)
- oEmbed/OG tags when downloading direct URLs that point to HTML pages
- Manifest hints (when available)
- `-meta` overrides

Sidecar schema (fields omitted when unknown):

```json
{
  "id": "dQw4w9WgXcQ",
  "title": "Example Title",
  "artist": "Example Artist",
  "author": "Example Author",
  "album": "Example Album",
  "track": 1,
  "disc": 1,
  "release_date": "2009-10-25",
  "release_year": 2009,
  "duration_seconds": 212,
  "thumbnail_url": "https://...",
  "source_url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "extractor": "kkdai/youtube",
  "extractor_version": "v2.10.5",
  "output": "Downloads/Example Title.mp4",
  "format": "mp4",
  "quality": "720p",
  "status": "ok",
  "warnings": ["missing artist"],
  "playlist": {
    "id": "PL123...",
    "title": "Example Playlist",
    "url": "https://www.youtube.com/playlist?list=PL123...",
    "index": 1,
    "count": 25
  }
}
```

Metadata overrides:

```bash
ytdl-go -meta title="Custom Title" -meta artist="Custom Artist" -meta album="Custom Album" URL
```

Supported keys: `title`, `artist`, `author`, `album`, `track`, `disc`, `release_date`, `release_year`, `source_url`.

Audio tag embedding: ID3 tags are embedded for `.mp3` outputs. Other containers are currently logged as “unsupported” and skipped.

## 🎮 Interactive Features

### Format Selection

When using `-list-formats`, you can interactively browse and select formats:

- **↑/↓ or j/k** - Navigate between formats
- **Enter** - Download the selected format
- **1-9** - Quick jump to formats with itags starting with that digit
- **Page Up/Page Down** - Jump 10 formats at a time
- **Home/End or g/G** - Go to first/last format
- **b** - Go back without downloading
- **q/esc/ctrl+c** - Quit

The selected format is highlighted and the itag number is displayed in real-time.

### File Handling

When downloading to an existing file and running in a terminal:

- **[o]verwrite** - Replace the existing file
- **[s]kip** - Skip downloading this file
- **[r]ename** - Automatically rename to `filename (1).ext`
- **[q]uit** - Abort the entire download process

If stdin is not a TTY (e.g., when piping), existing files will be overwritten automatically with a warning.

## 🛡️ Error Handling

The downloader includes robust error handling:

- **403 Forbidden errors** - Automatically retries with a different download method
- **Network timeouts** - Respects the configured timeout per request
- **Playlist errors** - Continues downloading remaining videos if one fails
- **Detailed reporting** - Each error is reported with context about which video failed
- **Categorized exit codes** - Stable exit codes for invalid URLs, unsupported formats, restricted/DRM responses, network failures, and filesystem errors

## 💡 Examples by Use Case

### 🎵 Music archiving

```bash
# Download playlist as organized music library by artist/album/song
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" "https://music.youtube.com/playlist?list=..."

# Download with track numbers
ytdl-go -audio -o "Music/{artist}/{album}/{index:02d} - {title}.{ext}" "https://music.youtube.com/playlist?list=..."

# Download playlist songs under playlist folder
ytdl-go -audio -o "Music/{playlist-title}/{title}.{ext}" "https://www.youtube.com/playlist?list=..."

# Download with playlist and track number
ytdl-go -audio -o "Music/{playlist-title}/{index:02d} - {title}.{ext}" "https://www.youtube.com/playlist?list=..."
```

### 🎬 Video collection

```bash
# Download with quality and date
ytdl-go -o "Videos/{playlist-title}/{title} [{quality}].{ext}" "https://www.youtube.com/playlist?list=..."
```

### 📄 Metadata extraction

```bash
# Export playlist metadata to JSON
ytdl-go -info "https://www.youtube.com/playlist?list=..." > playlist.json
```

### 📦 Bulk downloads

```bash
# Download multiple playlists
for url in $(cat playlist_urls.txt); do
    ytdl-go -audio -o "music/{artist}/{title}.{ext}" "$url"
done
```

## 📝 Notes / Limitations

### Supported Formats

- **YouTube-only**: This tool is specifically designed for YouTube and YouTube Music
- **Progressive formats**: Only downloads videos with audio+video combined (no DASH muxing)
- **Audio formats**: Supports best available audio-only formats via `-audio` flag
- **HLS/DASH parsing supported**: Unencrypted manifests are parsed and segments are concatenated; adaptive A/V muxing is not implemented

### Technical Limitations

- Only progressive formats are pulled for video; DASH-only video+audio muxing is not implemented
- Only public, non-DRM content is supported; login-required/paywalled/DRM manifests are rejected
- Adaptive HLS/DASH downloads concatenate segments; if a manifest uses unsupported segment templates the download will fail fast
- Direct URL downloads rely on the server exposing a video/audio content type or a recognizable file extension
- YouTube Music URLs (`music.youtube.com`) are converted to regular YouTube URLs automatically
- Authentication, cookies, proxies, and subtitle downloads are not yet supported
- No credential storage or persistent cookie jars are used; no browser automation is performed
- Output directories are created as needed; trailing slash on `-o` forces treating it as a directory
- Maximum 9999 automatic renames when using the rename option to prevent infinite loops
- Resume support for direct/HLS/DASH downloads via `.resume.json` state; progressive YouTube streams do not resume

### Non-Goals / Explicitly Not Supported

This tool is designed exclusively for **publicly accessible YouTube content**. The following are explicitly **not supported** and will not be added:

❌ **DRM/Encrypted Content**

- Widevine, PlayReady, or other DRM-protected streams
- AES-128 encrypted HLS streams
- DASH manifests with CENC encryption

❌ **Access Control Bypass**

- Login-required or members-only videos
- Paywall-protected content
- Age-restricted content requiring sign-in
- Private or unlisted videos requiring authentication

❌ **Platform Limitations**

- Non-YouTube platforms
- Browser automation or cookie extraction
- Credential harvesting or token spoofing

### ⚖️ Legal Notice

**Important**: Users are solely responsible for ensuring their use of this tool complies with:

- YouTube's Terms of Service
- Applicable copyright laws in their jurisdiction
- Content creator rights and licenses

This tool:

- ✅ Only accesses publicly available content through standard APIs
- ✅ Does not circumvent any technical protection measures or DRM
- ✅ Does not bypass authentication or access controls
- ❌ Should not be used to download copyrighted content without permission

**When restricted content is detected** (private videos, login-required, etc.), the tool will refuse to proceed and exit with a clear error message.

## 🚫 Non-goals

- Bypassing DRM, paywalls, or login restrictions
- Automating browser sessions or extracting authenticated cookies
- Re-encoding/transcoding media (output is the source container)
- Executing downloaded content

## ⚖️ Legal / Copyright Notice

This tool is intended for downloading publicly accessible content that you have the right to access and use. Respect copyright laws, license terms, and the Terms of Service of the source site.

## 🔧 Troubleshooting

<details>
<summary><b>Common Issues & Fixes</b></summary>

- **403 Forbidden Errors:** The tool automatically retries with different methods. If persistent, check your IP reputation or try `-timeout 10m`.
- **Restricted Content:** Private, age-gated, or member-only videos require authentication which is currently **not supported**.
- **Playlists:** Empty videos or deleted entries in playlists are automatically skipped.

</details>

## ⚡ Performance

The downloader is optimized for:

- Concurrent metadata fetching for playlists
- Parallel segment downloads for adaptive streams
- Minimal memory usage with streaming downloads
- Fast resumption on network errors
- Efficient progress tracking without impacting download speed

## 🙏 Acknowledgments

### 🛠️ Tech Stack & Dependencies

This project wouldn't be possible without the amazing open-source community:

- **[Go](https://golang.org/)** - The powerful programming language that makes ytdl-go fast and efficient
- **[github.com/kkdai/youtube/v2](https://github.com/kkdai/youtube)** - The core YouTube API library that handles all the heavy lifting
- **[YouTube](https://youtube.com)** - The platform we all love (and sometimes need to download from)

### 🌟 Special Thanks

- The Go community for creating such an amazing ecosystem
- The maintainers of `kkdai/youtube` for their excellent library
- All the contributors and users who help improve this project
- The yt-dlp project for inspiration and setting the standard for YouTube downloaders

### 🤝 Contributing

We welcome contributions! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### 📬 Contact

Have questions or feedback? Feel free to open an issue on GitHub.

---

## Made with ❤️ by the ytdl-go team (aka ...me)

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
