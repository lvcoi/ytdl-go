---
title: "Configuration"
weight: 30
---

# Configuration

Learn how to configure ytdl-go for your specific needs.

## Overview

ytdl-go is designed to work out of the box with sensible defaults, but offers extensive configuration options for power users.

---

## Configuration Files

Currently, ytdl-go does not use configuration files. All settings are passed via command-line flags.

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

---

## Advanced Configuration

### Output Directory Constraints

Restrict output to a specific directory:

```bash
# Only allow output within this directory
ytdl-go -output-dir /safe/path -o "{title}.{ext}" URL
```

> **Info:** This prevents path traversal attacks when using user-provided templates.

---

## Next Steps

- [Basic Downloads](../usage/basic-downloads) - Start downloading
- [Output Templates](../usage/output-templates) - Master template syntax
- [CLI Options Reference](../../reference/cli-options) - Complete flag list
