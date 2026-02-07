# Configuration

Learn how to configure ytdl-go for your specific needs.

## Overview

ytdl-go is designed to work out of the box with sensible defaults, but offers extensive configuration options for power users.

---

## Cookie Files

Use cookie files to access age-restricted, private, or members-only content.

### Exporting Cookies

!!! warning "Browser Extensions"
    Use a browser extension to export cookies in Netscape format:
    
    - **Chrome/Edge**: [Get cookies.txt](https://chrome.google.com/webstore/detail/get-cookiestxt/bgaddhkoddajcdgocldbbfleckgcbcid)
    - **Firefox**: [cookies.txt](https://addons.mozilla.org/en-US/firefox/addon/cookies-txt/)

### Using Cookies

```bash
# Download age-restricted video
ytdl-go -cookies cookies.txt https://www.youtube.com/watch?v=VIDEO_ID

# Download members-only content
ytdl-go -cookies cookies.txt https://www.youtube.com/watch?v=MEMBER_VIDEO_ID
```

!!! tip
    Store your cookie file in a secure location and never commit it to version control.

---

## Environment Variables

Configure ytdl-go behavior via environment variables:

### Output Directory

```bash
# Set default output directory
export YTDL_OUTPUT_DIR="$HOME/Videos/YouTube"
ytdl-go URL
```

### Proxy Configuration

```bash
# Use HTTP proxy
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080

ytdl-go URL
```

---

## Configuration Files

Currently, ytdl-go does not use configuration files. All settings are passed via command-line flags or environment variables.

!!! note "Future Enhancement"
    Configuration file support (e.g., `~/.config/ytdl-go/config.yaml`) may be added in future releases.

---

## Default Settings

Understanding the defaults helps you use ytdl-go effectively:

| Setting | Default Value | Description |
|---------|---------------|-------------|
| **Output Template** | `{title}.{ext}` | Filename pattern |
| **Quality** | `best` | Video quality selection |
| **Format** | Auto | Container format (mp4, webm, etc.) |
| **Jobs** | `1` | Concurrent downloads |
| **Timeout** | `180` seconds | Network timeout |
| **Retries** | `3` | Download retry attempts |
| **On Duplicate** | `prompt` | Behavior when file exists |

---

## Common Configuration Patterns

### Organizing Downloads

Create a directory structure for different content types:

```bash
# Music downloads
ytdl-go -audio -o "Music/{artist}/{title}.{ext}" URL

# Video downloads by channel
ytdl-go -o "Videos/{artist}/{title}.{ext}" URL

# Playlists
ytdl-go -o "Playlists/{playlist-title}/{title}.{ext}" PLAYLIST_URL
```

### Batch Processing

Create a shell alias for your preferred settings:

```bash
# Add to ~/.bashrc or ~/.zshrc
alias ytdl-music='ytdl-go -audio -o "Music/{artist}/{title}.{ext}"'
alias ytdl-video='ytdl-go -quality 1080p -o "Videos/{title}.{ext}"'
alias ytdl-playlist='ytdl-go -o "Playlists/{playlist-title}/{title}.{ext}"'
```

Usage:

```bash
ytdl-music https://www.youtube.com/watch?v=MUSIC_VIDEO
ytdl-video https://www.youtube.com/watch?v=VIDEO_ID
ytdl-playlist https://www.youtube.com/playlist?list=PLAYLIST_ID
```

### Automation Scripts

For scripting, use JSON output mode:

```bash
#!/bin/bash
# download-playlist.sh

PLAYLIST_URL="$1"
OUTPUT_DIR="$2"

ytdl-go -json -o "${OUTPUT_DIR}/{title}.{ext}" "$PLAYLIST_URL" | \
while IFS= read -r line; do
    echo "$line" | jq -r 'select(.type=="item") | "\(.status): \(.title)"'
done
```

---

## Web Interface Configuration

When using the web interface, configuration is managed through the UI:

```bash
# Start web interface on default port (8080)
ytdl-go -web

# Use custom port
ytdl-go -web -port 3000

# Bind to specific address
ytdl-go -web -addr 0.0.0.0:8080
```

### Web UI Settings

The web interface persists settings in browser localStorage:

- Output template
- Quality preference
- Audio-only mode
- Concurrent job count
- Duplicate handling policy

---

## Performance Tuning

### Concurrent Downloads

Download multiple videos simultaneously:

```bash
# Download 4 videos in parallel
ytdl-go -jobs 4 PLAYLIST_URL
```

!!! warning
    Higher job counts increase bandwidth usage and may trigger rate limits.

### Network Timeouts

Adjust for slow connections:

```bash
# Increase timeout to 5 minutes
ytdl-go -timeout 300 URL
```

### Retry Logic

Configure retry attempts:

```bash
# Retry up to 5 times
ytdl-go -retries 5 URL
```

---

## Advanced Configuration

### Output Directory Constraints

Restrict output to a specific directory:

```bash
# Only allow output within this directory
ytdl-go -output-dir /safe/path -o "{title}.{ext}" URL
```

!!! info
    This prevents path traversal attacks when using user-provided templates.

### Metadata Embedding

Control metadata embedding:

```bash
# Write metadata to JSON sidecar
ytdl-go -write-info-json URL

# Write description to .description file
ytdl-go -write-description URL

# Write thumbnail to .jpg file
ytdl-go -write-thumbnail URL

# All metadata options
ytdl-go -write-info-json -write-description -write-thumbnail URL
```

---

## Next Steps

- [Basic Downloads](../usage/basic-downloads.md) - Start downloading
- [Output Templates](../usage/output-templates.md) - Master template syntax
- [CLI Options Reference](../../reference/cli-options.md) - Complete flag list
