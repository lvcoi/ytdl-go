---
title: "Configuration"
weight: 30
---

# Configuration

Learn how to configure ytdl-go for your specific needs.

## Overview

ytdl-go is designed to work out of the box with sensible defaults, but offers extensive configuration options for power users.

---

## Cookie Files

Some video platforms require authentication (for example, for age-restricted, private, or members-only content). A common way for CLI tools to support this is by accepting a browser-exported cookie file.

At the moment, the `ytdl-go` CLI does **not** support passing a cookie file via a `-cookies` flag or any other command-line option. This means that:

- You can only download content that is accessible without logging in.
- Examples you may see elsewhere using `ytdl-go -cookies …` refer to a flag that is not implemented in the current CLI.

Support for cookie files may be added in a future release. Until then, there is no need to export cookies from your browser for use with `ytdl-go`.
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

> **Note: Future Enhancement**
>
> Configuration file support (e.g., `~/.config/ytdl-go/config.yaml`) may be added in future releases.

---

## Default Settings

Understanding the defaults helps you use ytdl-go effectively:

| Setting | Default Value | Description |
|---------|---------------|-------------|
| **Output Template** | `{title}.{ext}` | Filename pattern |
| **Quality** | `best` | Video quality selection |
| **Format** | Auto | Container format (mp4, webm, etc.) |
| **Jobs** | `1` | Concurrent downloads |
| **Timeout** | `3m` | Network timeout |
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

> **Warning:** Higher job counts increase bandwidth usage and may trigger rate limits.

### Network Timeouts

Adjust for slow connections:

```bash
# Increase timeout to 5 minutes
ytdl-go -timeout 5m URL
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

> **Info:** This prevents path traversal attacks when using user-provided templates.

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

- [Basic Downloads](../usage/basic-downloads) - Start downloading
- [Output Templates](../usage/output-templates) - Master template syntax
- [CLI Options Reference](../../reference/cli-options) - Complete flag list
