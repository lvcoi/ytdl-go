---
title: "Quick Start"
weight: 20
---

# Quick Start

Get up and running with ytdl-go in minutes!

## Your First Download

The simplest way to download a video:

```bash
ytdl-go https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

This downloads the video in the best available quality to your current directory with the filename pattern `{title}.{ext}`.

---

## Common Use Cases

### Download Audio Only

Perfect for music or podcasts:

```bash
ytdl-go -audio https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

Output: `Rick Astley - Never Gonna Give You Up.{ext}` (extension depends on audio format)

### Download a Playlist

Download all videos from a playlist:

```bash
ytdl-go https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf
```

### Choose Quality

Download in a specific quality:

```bash
# 1080p video
ytdl-go -quality 1080p https://www.youtube.com/watch?v=dQw4w9WgXcQ

# 720p video
ytdl-go -quality 720p https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Best quality (default)
ytdl-go -quality best https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

### Interactive Format Selection

Browse and select formats interactively:

```bash
ytdl-go -list-formats https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

This opens a beautiful TUI (Terminal User Interface) where you can:

- Browse all available formats
- See quality, codec, and file size information
- Select the exact format you want

### Custom Output Location

Save to a specific directory:

```bash
ytdl-go -o "downloads/{title}.{ext}" https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

Or with a custom filename:

```bash
ytdl-go -o "my-video.mp4" https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

---

## Understanding Output Templates

Output templates use placeholders that get replaced with video metadata:

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `{title}` | Video title | `Rick Astley - Never Gonna Give You Up` |
| `{id}` | Video ID | `dQw4w9WgXcQ` |
| `{ext}` | File extension | `mp4`, `webm`, `m4a` |
| `{artist}` | Artist/channel name | `Rick Astley` |
| `{playlist-title}` | Playlist name | `Best Music Videos` |

**Example patterns:**

```bash
# Organize by artist
ytdl-go -o "{artist}/{title}.{ext}" URL

# Add quality to filename
ytdl-go -o "{title} [{quality}].{ext}" URL

# Playlist organization
ytdl-go -o "{playlist-title}/{title}.{ext}" PLAYLIST_URL
```

See [Output Templates](../usage/output-templates) for complete documentation.

---

## Next Steps

Now that you know the basics, explore more features:

- **[Basic Downloads](../usage/basic-downloads)** - Detailed download options
- **[Playlists](../usage/playlists)** - Advanced playlist handling
- **[Audio-Only Mode](../usage/audio-only)** - Music and podcast downloads
- **[Format Selection](../usage/format-selection)** - Advanced format control
- **[Metadata & Sidecars](../usage/metadata-sidecars)** - Rich metadata handling

---

## Common Options Reference

| Flag | Description |
|------|-------------|
| `-audio` | Download audio only |
| `-quality VALUE` | Set quality (e.g., `1080p`, `720p`, `best`) |
| `-o TEMPLATE` | Output filename template |
| `-list-formats` | Interactive format selector |
| `-info` | Show video info without downloading |
| `-json` | Output JSON for scripting |
| `-quiet` | Suppress progress output |

See [CLI Options Reference](../../reference/cli-options) for the complete list.

---

## Getting Help

If you encounter issues:

- Check [Troubleshooting](../troubleshooting/common-issues)
- Review [FAQ](../troubleshooting/faq)
- Search [GitHub Issues](https://github.com/lvcoi/ytdl-go/issues)
- Open a [new issue](https://github.com/lvcoi/ytdl-go/issues/new)

---

## Web Interface

ytdl-go also includes an optional web interface:

```bash
# Build with web UI
./build.sh --web

# Or run the binary with -web flag
ytdl-go -web
```

Then open `http://localhost:8080` in your browser for a graphical interface.
