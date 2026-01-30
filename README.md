# ytdl-go 📺

![ytdl-go logo](screenshots/logo.svg)

## A powerful yt-dlp-style downloader written in Go

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](https://github.com/lvcoi/ytdl-go/LICENSE)
[![Release](https://img.shields.io/github/release/lvcoi/ytdl-go.svg?style=flat-square)](https://github.com/lvcoi/ytdl-go/releases)
[![GoDoc](https://img.shields.io/badge/go.dev-reference-007d9c?style=flat-square&logo=go)](https://pkg.go.dev/github.com/lvcoi/ytdl-go)

⚡ Blazing fast YouTube downloader with automatic retry, progress tracking, and YouTube Music support ⚡

---

## Table of Contents

- [✨ Features](#-features)
- [🚀 Installation](#-installation)
- [📖 Usage](#-usage)
  - [🎯 Basic usage](#-basic-usage)
  - [🎨 Output customization](#-output-customization)
  - [📚 Playlist downloads](#-playlist-downloads)
  - [⚙️ Advanced options](#️-advanced-options)
- [📊 Command Line Options](#-command-line-options)
- [🏷️ Output Template Placeholders](#️-output-template-placeholders)
- [🧾 Metadata & Sidecars](#-metadata--sidecars)
- [🎮 Interactive Features](#-interactive-features)
- [🛡️ Error Handling](#️-error-handling)
- [💡 Examples by Use Case](#-examples-by-use-case)
- [📝 Notes / Limitations](#-notes--limitations)
- [🔧 Troubleshooting](#-troubleshooting)
- [⚡ Performance](#-performance)
- [🙏 Acknowledgments](#-acknowledgments)
- [📜 License](#-license)

---

## ✨ Features

- **Metadata extraction** - `--info` prints detailed metadata (formats included) as pretty JSON
- **Format/quality control** - `--quality`, `--format`, and `--list-formats` let you pick the variant you want before downloading
- **Direct URL support** - Downloads public `.mp4`, `.webm`, `.mov`, `.m3u8`, and `.mpd` URLs (HLS/DASH unencrypted only)
- **Parallel downloads** - `--jobs` runs multiple URLs concurrently with stable progress output
- **Audio-only downloads** - `--audio` grabs the best audio-only format; otherwise downloads best progressive
- **Playlist support** - Playlist URLs are expanded and downloaded entry by entry with progress tracking
- **YouTube Music compatibility** - Automatically converts `music.youtube.com` URLs to regular YouTube URLs
- **Flexible output templating** - `-o` supports multiple placeholders:
  - `{title}` - Video title (sanitized)
  - `{artist}` - Video author/artist (sanitized)
  - `{album}` - Album name from YouTube Music metadata (when available)
  - `{id}` - Video ID
  - `{ext}` - File extension
  - `{quality}` - Quality label or bitrate
  - `{playlist_title}` - Playlist title (for playlists)
  - `{playlist_id}` - Playlist ID (for playlists)
  - `{index}` - Current index in playlist
  - `{count}` - Total number of videos in playlist
- **Progress tracking** - Real-time progress bar with speed indicators, or `--quiet` to suppress
- **Error resilience** - Automatic retry on 403 errors with fallback download method
- **Interactive file handling** - Prompts for overwrite/skip/rename when files exist (when TTY is available)
- **Configurable timeout** - Per-request timeout via `--timeout`
- **Comprehensive summaries** - Detailed download summary with success/failure/skip counts and total size

## 🚀 Installation

### 📋 Prerequisites

- ![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go) Go 1.23+

### ⚡ Quick Install (Recommended)

```bash
# Install directly from GitHub
go install github.com/lvcoi/ytdl-go@latest
```

This will download, build, and install `ytdl-go` to your `$GOPATH/bin` directory (usually `~/go/bin`). Make sure this directory is in your `PATH`.

### 🔧 Build from source

```bash
# Clone the repository
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go
```

![Clone repository](screenshots/01-git-clone.svg)

```bash
# Build the binary
go build -o ytdl-go .
```

![Build output](screenshots/02-go-build.svg)

```bash
# Optional: Install to system path
go install .
```

![Install output](screenshots/03-go-install.svg)

### 🏗️ Build with custom cache locations

If your environment blocks writes to the default Go caches, point them at a writable directory:

```bash
export GOCACHE=/tmp/gocache
export GOMODCACHE=/tmp/gomodcache
go mod tidy   # downloads dependencies
go build .
```

![Custom cache build](screenshots/04-custom-cache.svg)

## 📖 Usage

### 🎯 Basic usage

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

```bash
# Get video metadata without downloading
ytdl-go --info https://www.youtube.com/watch?v=BaW_jenozKc
```

![Metadata output example](screenshots/07-metadata-info.svg)

```bash
# List available formats without downloading
ytdl-go --list-formats https://www.youtube.com/watch?v=BaW_jenozKc
```

### 🎨 Output customization

```bash
# Custom output path
ytdl-go -o "downloads/{title}.{ext}" https://www.youtube.com/watch?v=BaW_jenozKc
```

![Custom output download](screenshots/08-custom-output.svg)

```bash
# Include artist in filename
ytdl-go -o "{artist} - {title}.{ext}" https://www.youtube.com/watch?v=BaW_jenozKc
```

![Artist in filename](screenshots/09-artist-filename.svg)

```bash
# Download with quality indicator
ytdl-go -o "{title} [{quality}].{ext}" https://www.youtube.com/watch?v=BaW_jenozKc
```

![Quality indicator example](screenshots/10-quality-indicator.svg)

### 📚 Playlist downloads

```bash
# Download entire playlist
ytdl-go "https://www.youtube.com/playlist?list=PL59FEE129ADFF2B12"
```

![Playlist download progress](screenshots/11-playlist-download.svg)

```bash
# Download playlist with custom structure
ytdl-go -o "downloads/{playlist_title}/{index:02d} - {title}.{ext}" "https://www.youtube.com/playlist?list=PL59FEE129ADFF2B12"
```

![Custom playlist structure](screenshots/12-custom-playlist.svg)

```bash
# Download playlist as audio-only with artist folders
ytdl-go --audio -o "music/{artist}/{playlist_title}/{index:02d} - {title}.{ext}" "https://www.youtube.com/playlist?list=PL59FEE129ADFF2B12"
```

![Audio playlist with folders](screenshots/13-audio-playlist-folders.svg)

```bash
# YouTube Music playlists (automatically converted)
ytdl-go --audio -o "music/{artist}/{title}.{ext}" "https://music.youtube.com/playlist?list=PLxUALHb15RSAPuTLY-05OageBIuHAOwJm"
```

![YouTube Music conversion](screenshots/14-youtube-music.svg)

### ⚙️ Advanced options

```bash
# Quiet mode (no progress output)
ytdl-go --quiet https://www.youtube.com/watch?v=BaW_jenozKc
```

![Quiet mode output](screenshots/15-quiet-mode.svg)

```bash
# Custom timeout
ytdl-go --timeout 10m https://www.youtube.com/watch?v=BaW_jenozKc
```

![Custom timeout example](screenshots/16-custom-timeout.svg)

```bash
# Multiple URLs
ytdl-go https://www.youtube.com/watch?v=video1 https://www.youtube.com/watch?v=video2
```

![Multiple URLs download](screenshots/17-multiple-urls.svg)

```bash
# List all available formats without downloading
ytdl-go --list-formats https://www.youtube.com/watch?v=BaW_jenozKc
```

```bash
# Pick a specific quality/container
ytdl-go --quality 720p --format mp4 https://www.youtube.com/watch?v=BaW_jenozKc
```

```bash
# Download a specific format by itag number
ytdl-go --itag 251 https://www.youtube.com/watch?v=BaW_jenozKc
```

```bash
# JSON-only output (no human progress noise)
ytdl-go --json --quality 1080p https://www.youtube.com/watch?v=BaW_jenozKc
```

```bash
# Direct file download (public URL)
ytdl-go https://example.com/video.mp4
```

```bash
# Custom progress layout
ytdl-go --progress-layout "{label} {percent} {current}/{total} {rate} {eta}" URL
```

```bash
# Parallel downloads
ytdl-go --jobs 4 URL1 URL2 URL3 URL4
```

## 📊 Command Line Options

| Option | Default | Description |
| -------- | --------- | ------------- |
| `-o` | `{title}.{ext}` | Output path or template with supported placeholders |
| `-audio` | `false` | Download best available audio-only format |
| `-info` | `false` | Print video metadata as JSON without downloading |
| `-list-formats` | `false` | List available formats for a URL and exit |
| `-quality` | `best` | Preferred quality (e.g., `1080p`, `720p`, `128k`, `worst`) |
| `-format` | `` | Preferred container/extension (e.g., `mp4`, `webm`, `m4a`) |
| `-itag` | `0` | Download specific format by itag number (use `-list-formats` to see available itags) |
| `-meta` | `` | Metadata override (`key=value`, repeatable) |
| `-progress-layout` | `` | Progress layout template (use `{label}`, `{percent}`, `{rate}`, `{eta}`, `{current}`, `{total}`) |
| `-segment-concurrency` | `auto` | Parallel segment downloads for HLS/DASH (0=auto, 1=disabled) |
| `-playlist-concurrency` | `auto` | Parallel playlist entry downloads (0=auto, 1=disabled) |
| `-jobs` | `1` | Number of concurrent downloads |
| `-json` | `false` | Emit newline-delimited JSON output (suppresses progress text) |
| `-quiet` | `false` | Suppress progress output (errors still shown) |
| `-timeout` | `3m` | Per-request timeout (e.g., 30s, 5m, 1h) |
| `-log-level` | `info` | Log level: `debug`, `info`, `warn`, `error` |

## 🏷️ Output Template Placeholders

| Placeholder | Description | Example |
| ------------- | ------------- | --------- |
| `{title}` | Video title (sanitized for filesystem) | `My Video Title` |
| `{artist}` | Video author/artist (sanitized) | `Artist Name` |
| `{album}` | Album name from YouTube Music metadata (when available) | `Album Name` |
| `{id}` | YouTube video ID | `dQw4w9WgXcQ` |
| `{ext}` | File extension from format | `mp4`, `webm`, `m4a` |
| `{quality}` | Quality label or bitrate | `1080p`, `128k` |
| `{playlist_title}` | Playlist name | `My Awesome Playlist` |
| `{playlist_id}` | Playlist ID | `PL59FEE129ADFF2B12` |
| `{index}` | Current video index in playlist | `1`, `2`, `3` |
| `{count}` | Total videos in playlist | `25` |

## 🧾 Metadata & Sidecars

Each successful download emits a sidecar JSON file alongside the media (`<output>.json`). Playlist downloads also emit a `playlist.json` manifest in the output folder.

Metadata sources (in order of preference):

- Platform/structured metadata from the extractor (e.g., YouTube API)
- oEmbed/OG tags when downloading direct URLs that point to HTML pages
- Manifest hints (when available)
- `--meta` overrides

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
ytdl-go --meta title="Custom Title" --meta artist="Custom Artist" --meta album="Custom Album" URL
```

Supported keys: `title`, `artist`, `author`, `album`, `track`, `disc`, `release_date`, `release_year`, `source_url`.

Audio tag embedding: ID3 tags are embedded for `.mp3` outputs. Other containers are currently logged as “unsupported” and skipped.

## 🎮 Interactive Features

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
ytdl-go --audio -o "Music/{artist}/{album}/{title}.{ext}" "https://music.youtube.com/playlist?list=..."

# Download with track numbers
ytdl-go --audio -o "Music/{artist}/{album}/{index:02d} - {title}.{ext}" "https://music.youtube.com/playlist?list=..."

# Download playlist songs under playlist folder
ytdl-go --audio -o "Music/{playlist_title}/{title}.{ext}" "https://www.youtube.com/playlist?list=..."

# Download with playlist and track number
ytdl-go --audio -o "Music/{playlist_title}/{index:02d} - {title}.{ext}" "https://www.youtube.com/playlist?list=..."
```

### 🎬 Video collection

```bash
# Download with quality and date
ytdl-go -o "Videos/{playlist_title}/{title} [{quality}].{ext}" "https://www.youtube.com/playlist?list=..."
```

### 📄 Metadata extraction

```bash
# Export playlist metadata to JSON
ytdl-go --info "https://www.youtube.com/playlist?list=..." > playlist.json
```

### 📦 Bulk downloads

```bash
# Download multiple playlists
for url in $(cat playlist_urls.txt); do
    ytdl-go --audio -o "music/{artist}/{title}.{ext}" "$url"
done
```

## 📝 Notes / Limitations

### Supported Formats

- **YouTube-only**: This tool is specifically designed for YouTube and YouTube Music
- **Progressive formats**: Only downloads videos with audio+video combined (no DASH muxing)
- **Audio formats**: Supports best available audio-only formats via `--audio` flag
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

### 403 Forbidden errors

The downloader automatically handles most 403 errors by retrying with a different method. If you continue to see issues:

1. Check your network connection
2. Try increasing the timeout: `--timeout 10m`
3. The video might be region-restricted or private

### Playlist download issues

- YouTube Music playlists are automatically converted
- Some private or unlisted playlists may require authentication (not yet supported)
- Empty playlist entries are automatically skipped

### File naming

- Invalid filesystem characters are replaced with hyphens
- Very long titles may be truncated
- Use the `{artist}` placeholder for better organization of music content

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
