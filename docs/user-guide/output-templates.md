# Output Templates

## Overview

The `-o` flag controls where downloaded files are saved. It supports placeholders that are replaced with video metadata at download time.

**Default:** `{title}.{ext}`

## Available Placeholders

| Placeholder | Description | Example |
| --- | --- | --- |
| `{title}` | Video title (sanitized) | `My Video Title` |
| `{artist}` | Video author/artist | `Artist Name` |
| `{album}` | Album name (YouTube Music) | `Album Name` |
| `{id}` | YouTube video ID | `dQw4w9WgXcQ` |
| `{ext}` | File extension | `mp4`, `webm`, `m4a` |
| `{quality}` | Quality label or bitrate | `1080p`, `128k` |
| `{playlist_title}` or `{playlist-title}` | Playlist name | `My Playlist` |
| `{playlist_id}` or `{playlist-id}` | Playlist ID | `PLxxx...` |
| `{index}` | Video index in playlist (1-based) | `1`, `2`, `3` |
| `{count}` | Total videos in playlist | `25` |

## Examples

```bash
# Simple filename
ytdl-go -o "video.mp4" URL

# Include artist
ytdl-go -o "{artist} - {title}.{ext}" URL

# Organize by artist/album
ytdl-go -o "Music/{artist}/{album}/{title}.{ext}" URL

# Playlist with index
ytdl-go -o "Playlist/{playlist_title}/{index} - {title}.{ext}" URL

# Quality indicator
ytdl-go -o "Videos/{title} [{quality}].{ext}" URL
```

## Path Behavior

- Output paths must be **relative** (absolute paths are rejected)
- Relative paths resolve against `-output-dir` if set, otherwise the current working directory
- Directories are created automatically if they don't exist
- Trailing slash forces directory interpretation
- Filenames are sanitized for filesystem safety

## Output Directory Constraint

The `-output-dir` flag constrains all output paths to a base directory, preventing directory traversal:

```bash
ytdl-go -output-dir /safe/downloads -o "../../etc/passwd" URL
# Rejected — path escapes the output directory
```
