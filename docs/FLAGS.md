# Command-Line Flags Reference

This comprehensive reference documents all command-line flags available in ytdl-go. For quick usage examples, see the main [README.md](../README.md).

## Table of Contents

- [Input/Output Flags](#inputoutput-flags)
- [Format Selection Flags](#format-selection-flags)
- [Network Flags](#network-flags)
- [Concurrency Flags](#concurrency-flags)
- [Output Control Flags](#output-control-flags)
- [Metadata Flags](#metadata-flags)
- [Advanced Flags](#advanced-flags)
- [Flag Combinations](#flag-combinations)

## Input/Output Flags

### `-o` / `--output` (Output Template)

**Default:** `{title}.{ext}`  
**Type:** String  
**Example:** `ytdl-go -o "downloads/{title}.{ext}" [URL]`

Specifies the output path or template for downloaded files. Supports placeholders that are dynamically replaced with video metadata.

**Available Placeholders:**

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `{title}` | Video title (sanitized for filesystem) | `My Video Title` |
| `{artist}` | Video author/artist (sanitized) | `Artist Name` |
| `{album}` | Album name (YouTube Music metadata) | `Album Name` |
| `{id}` | YouTube video ID | `dQw4w9WgXcQ` |
| `{ext}` | File extension from format | `mp4`, `webm`, `m4a` |
| `{quality}` | Quality label or bitrate | `1080p`, `720p`, `128k` |
| `{playlist_title}` or `{playlist-title}` | Playlist name | `My Playlist` |
| `{playlist_id}` or `{playlist-id}` | Playlist ID | `PLxxx...` |
| `{index}` | Video index in playlist (1-based) | `1`, `2`, `3` |
| `{count}` | Total videos in playlist | `25` |

**Path Behavior:**
- Relative paths are relative to current working directory
- Absolute paths are used as-is
- Directories are created automatically if they don't exist
- Trailing slash forces directory interpretation

**Examples:**

```bash
# Simple filename
ytdl-go -o "video.mp4" [URL]

# Include artist in filename
ytdl-go -o "{artist} - {title}.{ext}" [URL]

# Organize by artist/album
ytdl-go -o "Music/{artist}/{album}/{title}.{ext}" [URL]

# Playlist with index
ytdl-go -o "Playlist/{playlist_title}/{index} - {title}.{ext}" [URL]

# Quality indicator
ytdl-go -o "Videos/{title} [{quality}].{ext}" [URL]
```

### `--output-dir` (Output Directory Constraint)

**Default:** (none)  
**Type:** String  
**Example:** `ytdl-go --output-dir /safe/downloads -o "../../etc/passwd" [URL]`

Security feature that constrains all output paths to be within a specified base directory. Prevents directory traversal attacks.

**Behavior:**
- All output paths are validated against this directory
- Paths attempting to escape are rejected
- Useful in server/script environments where user input is involved

## Format Selection Flags

### `--audio` (Audio-Only Mode)

**Default:** `false`  
**Type:** Boolean  
**Example:** `ytdl-go --audio [URL]`

Downloads the best available audio-only format. Automatically selects the highest bitrate audio stream.

**Selection Priority:**
1. Highest bitrate audio-only format
2. Best codec (Opus > AAC > MP3)
3. Falls back to progressive format if audio-only unavailable

**Strategy Chain:**
1. Try direct audio-only download
2. If 403 error, retry with single request
3. If still fails, use FFmpeg fallback (extract from progressive video)

### `--list-formats` (Interactive Format Browser)

**Default:** `false`  
**Type:** Boolean  
**Example:** `ytdl-go --list-formats [URL]`

Launches an interactive terminal UI to browse and select from all available formats.

**Features:**
- View all video and audio formats
- See quality, codec, bitrate, file size
- Quick jump by itag number

**Keyboard Controls:**
- `↑/↓` or `j/k` - Navigate
- `Enter` - Download selected format
- `1-9` - Jump to itags starting with digit
- `Page Up/Down` - Jump 10 formats
- `Home/End` or `g/G` - First/Last format
- `b` - Go back without downloading
- `q/Esc/Ctrl+C` - Quit

### `--quality` (Quality Preference)

**Default:** (best available)  
**Type:** String  
**Example:** `ytdl-go --quality 720p [URL]`

Specifies preferred quality for video or audio downloads.

**Video Quality Options:**
- `1080p`, `720p`, `480p`, `360p`, `240p`, `144p` - Specific resolution
- `best` - Highest quality available (default)
- `worst` - Lowest quality available

**Audio Quality Options:**
- `320k`, `256k`, `192k`, `128k`, `96k`, `64k` - Specific bitrate
- `best` - Highest bitrate (default)
- `worst` - Lowest bitrate

**Behavior:**
- Selects closest available quality if exact match not found
- For video, only progressive formats are selected
- For audio (with `--audio`), selects audio-only formats

**Examples:**

```bash
# Download 720p video
ytdl-go --quality 720p [URL]

# Download 128k audio
ytdl-go --audio --quality 128k [URL]

# Download worst quality (smallest file)
ytdl-go --quality worst [URL]
```

### `--format` (Container/Extension Preference)

**Default:** (any)  
**Type:** String  
**Example:** `ytdl-go --format mp4 [URL]`

Specifies preferred container format or file extension.

**Common Formats:**
- `mp4` - MPEG-4 container (most compatible)
- `webm` - WebM container (open source)
- `m4a` - MPEG-4 Audio (AAC)
- `opus` - Opus audio
- `mp3` - MP3 audio

**Behavior:**
- Filters available formats by container type
- Combined with `--quality` for precise selection
- If format unavailable, download may fail

**Examples:**

```bash
# Prefer MP4 container
ytdl-go --format mp4 [URL]

# Download audio as M4A
ytdl-go --audio --format m4a [URL]

# Specific quality and format
ytdl-go --quality 720p --format mp4 [URL]
```

### `--itag` (Direct Format Selection)

**Default:** `0` (auto-select)  
**Type:** Integer  
**Example:** `ytdl-go --itag 251 [URL]`

Downloads a specific format by its YouTube itag number. Bypasses all automatic format selection.

**Common YouTube Itags:**

| Itag | Type | Description |
|------|------|-------------|
| 18 | Video | 360p MP4 (progressive) |
| 22 | Video | 720p MP4 (progressive) |
| 251 | Audio | Opus ~160kbps |
| 140 | Audio | AAC 128kbps |
| 141 | Audio | AAC 256kbps |

**How to Find Itags:**

```bash
# List all available formats with itags
ytdl-go --list-formats [URL]
```

**Use Cases:**
- Download specific format for compatibility
- Testing different formats
- Scripting with known format requirements

## Network Flags

### `--timeout` (Request Timeout)

**Default:** `3m` (3 minutes)  
**Type:** Duration  
**Example:** `ytdl-go --timeout 10m [URL]`

Sets the timeout for individual network requests. Does not apply to entire download, only individual HTTP requests.

**Duration Format:**
- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `90s` - 90 seconds (1.5 minutes)

**When to Adjust:**
- **Increase**: Slow network, large files, congested servers
- **Decrease**: Want faster failure detection

**Examples:**

```bash
# Long timeout for slow connections
ytdl-go --timeout 10m [URL]

# Short timeout for quick failure detection
ytdl-go --timeout 30s [URL]
```

**Note:** For HLS/DASH streams, this applies to each segment download, not the entire stream.

## Concurrency Flags

### `--jobs` (Concurrent Downloads)

**Default:** `1`  
**Type:** Integer  
**Example:** `ytdl-go --jobs 4 [URL1] [URL2] [URL3] [URL4]`

Number of URLs to download simultaneously when multiple URLs are provided.

**Behavior:**
- Downloads N URLs in parallel
- Each job shows its own progress bar
- All jobs managed by ProgressManager

**Recommendations:**
- **Network-bound**: Can increase (4-8)
- **Disk I/O-bound**: Keep low (1-2)
- **Large files**: Keep low to avoid memory pressure
- **Small files**: Can increase safely

**Examples:**

```bash
# Download 4 videos concurrently
ytdl-go --jobs 4 video1.com video2.com video3.com video4.com

# Single job (sequential)
ytdl-go --jobs 1 video1.com video2.com
```

### `--playlist-concurrency` (Playlist Entry Concurrency)

**Default:** `0` (auto - based on CPU count)  
**Type:** Integer  
**Example:** `ytdl-go --playlist-concurrency 3 [PLAYLIST_URL]`

Number of playlist entries to download in parallel.

**Special Values:**
- `0` - Auto (runtime.NumCPU() / 2)
- `1` - Sequential (one at a time)
- `N` - Specific concurrency level

**Recommendations:**
- **Default (auto)**: Good for most cases
- **Sequential (1)**: When order matters or debugging
- **High (4-8)**: Fast network, many small videos

**Examples:**

```bash
# Auto-detect concurrency
ytdl-go --playlist-concurrency 0 [PLAYLIST_URL]

# Sequential downloads
ytdl-go --playlist-concurrency 1 [PLAYLIST_URL]

# High concurrency
ytdl-go --playlist-concurrency 8 [PLAYLIST_URL]
```

### `--segment-concurrency` (Segment Download Concurrency)

**Default:** `0` (auto - based on CPU count)  
**Type:** Integer  
**Example:** `ytdl-go --segment-concurrency 4 [HLS_URL]`

For HLS/DASH streams, number of segments to download in parallel.

**Special Values:**
- `0` - Auto (runtime.NumCPU())
- `1` - Sequential (one segment at a time)
- `N` - Specific concurrency level

**When It Matters:**
- HLS streams (`.m3u8`)
- DASH streams (`.mpd`)
- Not used for progressive downloads

**Recommendations:**
- **Default (auto)**: Optimal for most cases
- **Sequential (1)**: Debugging or server restrictions
- **High (8-16)**: Many small segments, fast network

## Output Control Flags

### `--quiet` (Suppress Progress)

**Default:** `false`  
**Type:** Boolean  
**Example:** `ytdl-go --quiet [URL]`

Suppresses all progress output. Errors are still shown to stderr.

**Use Cases:**
- Scripting
- Logging to file
- Avoiding terminal clutter
- Background processing

**Examples:**

```bash
# Quiet download
ytdl-go --quiet [URL]

# Quiet with output file capture
ytdl-go --quiet [URL] 2>errors.log
```

### `--json` (JSON Output Mode)

**Default:** `false`  
**Type:** Boolean  
**Example:** `ytdl-go --json [URL]`

Emits newline-delimited JSON output for machine parsing. Suppresses all human-readable progress output.

**Output Format:**

Each line is a JSON object representing an event:

```json
{"type":"start","url":"...","title":"..."}
{"type":"progress","url":"...","percent":25.5,"current":1048576,"total":4194304}
{"type":"complete","url":"...","output":"video.mp4","size":4194304}
{"type":"error","url":"...","category":"network","error":"connection timeout"}
```

**Event Types:**
- `start` - Download started
- `progress` - Progress update
- `complete` - Download successful
- `error` - Download failed

**Use Cases:**
- Integration with other tools
- Progress monitoring from scripts
- Automated processing pipelines
- Web UI backends

**Examples:**

```bash
# JSON output
ytdl-go --json [URL]

# JSON output piped to jq
ytdl-go --json [URL] | jq -r 'select(.type=="complete") | .output'
```

### `--info` (Metadata Only)

**Default:** `false`  
**Type:** Boolean  
**Example:** `ytdl-go --info [URL]`

Extracts and prints metadata as JSON without downloading. Does not write any files.

**Output Includes:**
- Video ID, title, author
- Duration, views, upload date
- Available formats with quality/codec/bitrate
- Thumbnail URLs
- Playlist information (if applicable)

**Use Cases:**
- Checking video information
- Scripting (selecting format programmatically)
- Cataloging content
- Verifying URLs

**Examples:**

```bash
# Get metadata as JSON
ytdl-go --info [URL]

# Pretty-print with jq
ytdl-go --info [URL] | jq .

# Extract just the title
ytdl-go --info [URL] | jq -r .title
```

## Metadata Flags

### `--meta` (Metadata Override)

**Default:** (none)  
**Type:** Key=Value (repeatable)  
**Example:** `ytdl-go --meta title="Custom Title" --meta artist="Custom Artist" [URL]`

Overrides specific metadata fields. Can be specified multiple times for different fields.

**Supported Keys:**
- `title` - Video title
- `artist` - Artist/author name
- `author` - Uploader/channel name
- `album` - Album name
- `track` - Track number
- `disc` - Disc number
- `release_date` - Release date (YYYY-MM-DD)
- `release_year` - Release year (YYYY)
- `source_url` - Source URL

**Effects:**
- Overrides template placeholders (`{title}`, `{artist}`, etc.)
- Written to sidecar JSON file
- Embedded in audio tags (MP3 ID3)

**Examples:**

```bash
# Override title and artist
ytdl-go --meta title="My Song" --meta artist="My Artist" [URL]

# Set album and track info
ytdl-go --audio \
  --meta album="Greatest Hits" \
  --meta track=5 \
  --meta artist="The Band" \
  [URL]

# Multiple overrides
ytdl-go \
  --meta title="Episode 1" \
  --meta release_date="2023-01-15" \
  --meta author="Podcast Name" \
  [URL]
```

### `--progress-layout` (Custom Progress Format)

**Default:** (built-in format)  
**Type:** String  
**Example:** `ytdl-go --progress-layout "{label} {percent}% {rate}" [URL]`

Customizes the progress bar display format.

**Available Placeholders:**
- `{label}` - Task label/filename
- `{percent}` - Completion percentage (0-100)
- `{current}` - Downloaded bytes (formatted)
- `{total}` - Total bytes (formatted)
- `{rate}` - Download speed (formatted)
- `{eta}` - Estimated time remaining

**Examples:**

```bash
# Minimal progress
ytdl-go --progress-layout "{percent}% {rate}" [URL]

# Detailed progress
ytdl-go --progress-layout "{label} | {current}/{total} | {rate} | ETA: {eta}" [URL]
```

## Advanced Flags

### `--log-level` (Logging Verbosity)

**Default:** `info`  
**Type:** String (debug, info, warn, error)  
**Example:** `ytdl-go --log-level debug [URL]`

Controls logging verbosity.

**Levels:**
- `debug` - All messages including detailed diagnostics
- `info` - Informational messages and above (default)
- `warn` - Warnings and errors only
- `error` - Errors only

**Use Cases:**
- `debug`: Troubleshooting, development
- `info`: Normal usage
- `warn`: Quiet operation with error reporting
- `error`: Minimal output, errors only

**Examples:**

```bash
# Debug mode for troubleshooting
ytdl-go --log-level debug [URL]

# Warnings and errors only
ytdl-go --log-level warn [URL]
```

### `--web` (Web UI Server)

**Default:** `false`  
**Type:** Boolean  
**Example:** `ytdl-go --web`

Launches the web-based UI server instead of downloading from CLI.

**Usage:**

```bash
# Start web server on default address (127.0.0.1:8080)
ytdl-go --web

# Custom address
ytdl-go --web --web-addr "0.0.0.0:3000"
```

**Web UI Features:**
- Browser-based download interface
- Multiple concurrent downloads
- Real-time progress
- Format selection

### `--web-addr` (Web Server Address)

**Default:** `127.0.0.1:8080`  
**Type:** String  
**Example:** `ytdl-go --web --web-addr "0.0.0.0:3000"`

Specifies the address and port for the web server.  
For security reasons, you should normally bind the web server only to localhost (`127.0.0.1`) unless you fully understand the risks and have additional protections (such as host allowlisting and a reverse proxy) in place.

**Format:** `host:port`

**Examples:**

```bash
# Localhost only (default and recommended)
ytdl-go --web --web-addr "127.0.0.1:8080"

# Custom port on localhost
ytdl-go --web --web-addr "127.0.0.1:3000"
```

## Flag Combinations

### Common Workflows

#### Music Library Archiving
```bash
ytdl-go --audio \
  -o "Music/{artist}/{album}/{index} - {title}.{ext}" \
  --meta album="Album Name" \
  [PLAYLIST_URL]
```

#### High-Quality Video Collection
```bash
ytdl-go --quality 1080p \
  --format mp4 \
  -o "Videos/{playlist_title}/{index} - {title} [{quality}].{ext}" \
  [PLAYLIST_URL]
```

#### Batch Download with JSON Output
```bash
ytdl-go --json \
  --jobs 4 \
  --quiet \
  [URL1] [URL2] [URL3] [URL4] \
  > downloads.jsonl
```

#### Fast Playlist Download
```bash
ytdl-go --audio \
  --playlist-concurrency 8 \
  --quiet \
  [PLAYLIST_URL]
```

#### Debug Mode with Detailed Logging
```bash
ytdl-go --log-level debug \
  --timeout 30m \
  --list-formats \
  [URL]
```

### Flag Precedence

When flags conflict:
1. `--itag` overrides `--quality` and `--format`
2. `--json` implies `--quiet`
3. `--info` prevents downloading (metadata only)
4. `--list-formats` prevents automatic downloading

### Invalid Combinations

These combinations don't make sense:
- `--info` + `-o` (metadata only, no file written)
- `--list-formats` + `--quiet` (interactive UI needs output)
- `--json` + `--list-formats` (JSON mode is non-interactive)

---

**Last Updated:** 2026-02-05  
**Version:** 1.0  
**See Also:** [README.md](../README.md), [ARCHITECTURE.md](ARCHITECTURE.md)
