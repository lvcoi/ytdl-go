---
title: "Playlist Downloads"
weight: 60
---

# Playlist Downloads

This guide covers downloading entire YouTube playlists, including organization, indexing, and batch processing.

## Basic Playlist Download

Download all videos in a playlist by providing the playlist URL:

```bash
ytdl-go https://www.youtube.com/playlist?list=PLxxxxxxxxxxxxxxxxxx
```

This downloads all videos in the playlist sequentially with default quality settings.

> **Playlist Detection**
>
> ytdl-go automatically detects playlist URLs and processes all videos in the list.

## Playlist Organization

### Using Playlist Placeholders

Organize downloaded playlist videos using special placeholders:

```bash
# Organize by playlist title
ytdl-go -o "{playlist-title}/{title}.{ext}" https://www.youtube.com/playlist?list=PLxxxxx

# Include video index in filename
ytdl-go -o "{playlist-title}/{index} - {title}.{ext}" https://www.youtube.com/playlist?list=PLxxxxx

# Full organization with quality indicator
ytdl-go -o "{playlist-title}/{index} - {title} [{quality}].{ext}" https://www.youtube.com/playlist?list=PLxxxxx
```

### Available Playlist Placeholders

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `{playlist-title}` or `{playlist_title}` | Playlist name | `My Awesome Playlist` |
| `{playlist-id}` or `{playlist_id}` | Playlist ID | `PLxxxxxxxxxxxxxxxxxxx` |
| `{index}` | Video position in playlist (1-based) | `1`, `2`, `3`, ... |
| `{count}` | Total videos in playlist | `25` |

### Example Directory Structures

**Simple organization:**
```
My Playlist/
├── Video Title 1.mp4
├── Video Title 2.mp4
└── Video Title 3.mp4
```

```bash
ytdl-go -o "{playlist-title}/{title}.{ext}" PLAYLIST_URL
```

**Numbered organization:**
```
My Playlist/
├── 1 - Video Title 1.mp4
├── 2 - Video Title 2.mp4
└── 3 - Video Title 3.mp4
```

```bash
ytdl-go -o "{playlist-title}/{index} - {title}.{ext}" PLAYLIST_URL
```

**With quality label:**
```
My Playlist/
├── 1 - Video Title 1 [1080p].mp4
├── 2 - Video Title 2 [720p].mp4
└── 3 - Video Title 3 [1080p].mp4
```

```bash
ytdl-go -o "{playlist-title}/{index} - {title} [{quality}].{ext}" PLAYLIST_URL
```

## Quality Control for Playlists

Apply quality and format preferences to all playlist videos:

```bash
# Download entire playlist in 720p
ytdl-go -quality 720p https://www.youtube.com/playlist?list=PLxxxxx

# Download playlist in MP4 format
ytdl-go -format mp4 https://www.youtube.com/playlist?list=PLxxxxx

# Combine quality, format, and organization
ytdl-go -quality 1080p -format mp4 \
        -o "{playlist-title}/{index} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

## Audio-Only Playlist Downloads

Download entire playlists as audio:

```bash
# Basic audio playlist download
ytdl-go -audio https://www.youtube.com/playlist?list=PLxxxxx

# Organized music library
ytdl-go -audio -o "Music/{playlist-title}/{index} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

> **Music Playlists**
>
> For music-focused organization, see the [Audio-Only Downloads](audio-only) guide.

## Playlist Progress Tracking

ytdl-go shows progress for each video in the playlist:

```
[1/25] Downloading: Video Title 1
████████████████████████████████████ 100% 15.2 MB/s

[2/25] Downloading: Video Title 2
████████████████░░░░░░░░░░░░░░░░░░░░  45% 12.8 MB/s
```

## Handling Playlist Entries

### Skipped Videos

ytdl-go automatically skips:
- Deleted videos
- Private videos
- Unavailable videos
- Empty entries

The download continues with the next available video.

> **Error Handling**
>
> Failed downloads don't stop the entire playlist. ytdl-go continues processing remaining videos.

### Resume Interrupted Playlists

If a playlist download is interrupted:

```bash
# Simply run the same command again
ytdl-go -o "{playlist-title}/{index} - {title}.{ext}" PLAYLIST_URL
```

By default, ytdl-go will prompt when files already exist. To avoid duplicates when resuming, use a unique output template that includes the video ID:

```bash
# Use unique identifier to avoid prompts
ytdl-go -o "{playlist-title}/{index} - {title}-{id}.{ext}" PLAYLIST_URL
```

## Concurrent Processing

Control how many playlist videos are downloaded in parallel:

```bash
# Download 4 videos at a time
ytdl-go -playlist-concurrency 4 PLAYLIST_URL

# Automatic (based on system resources)
ytdl-go -playlist-concurrency 0 PLAYLIST_URL
```

For now, to download multiple playlists concurrently:

```bash
# Download multiple playlists in parallel using -jobs
ytdl-go -jobs 3 PLAYLIST_URL1 PLAYLIST_URL2 PLAYLIST_URL3
```

## Common Playlist Workflows

### Archive Complete Playlist

```bash
ytdl-go -quality 1080p -format mp4 \
        -o "Archive/{playlist-title}/{index} - {title} [{id}].{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

### Music Playlist Collection

```bash
ytdl-go -audio \
        -o "Music/{playlist-title}/{index} - {artist} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

### Educational Series

```bash
ytdl-go -quality 720p \
        -o "Courses/{playlist-title}/Lesson {index} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

### Podcast Episodes

```bash
ytdl-go -audio -quality 128k \
        -o "Podcasts/{playlist-title}/Episode {index} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

### Documentary Series with Metadata

```bash
ytdl-go -quality 1080p -format mp4 \
        -o "Documentaries/{playlist-title}/{index} - {title}.{ext}" \
        -meta album="{playlist-title}" \
        -meta track="{index}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

## Playlist with Custom Metadata

Override metadata for the entire playlist:

```bash
ytdl-go -audio \
        -o "Music/{artist}/{album}/{index} - {title}.{ext}" \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

This sets the same artist and album for all tracks in the playlist.

## Selective Playlist Downloads

### Download Specific Range

While ytdl-go doesn't have built-in range selection, you can use the interactive mode:

```bash
# View all formats and select specific videos
ytdl-go -list-formats https://www.youtube.com/playlist?list=PLxxxxx
```

### Using External Tools

For more control, extract URLs first:

```bash
# Get playlist info as JSON
ytdl-go -info https://www.youtube.com/playlist?list=PLxxxxx

# Then download specific videos
ytdl-go URL1 URL2 URL3
```

## Playlist Information

Get playlist metadata without downloading:

```bash
ytdl-go -info https://www.youtube.com/playlist?list=PLxxxxx
```

This returns JSON with:
- Playlist title
- Playlist ID
- Video count
- Video list with metadata

```bash
# Pretty print playlist info
ytdl-go -info https://www.youtube.com/playlist?list=PLxxxxx | jq .

# Count videos in playlist
ytdl-go -info https://www.youtube.com/playlist?list=PLxxxxx | jq '.entries | length'

# List all video titles
ytdl-go -info https://www.youtube.com/playlist?list=PLxxxxx | jq -r '.entries[].title'
```

## Quiet Playlist Downloads

For automated/scripted playlist downloads:

```bash
# Quiet mode (no progress bars)
ytdl-go -quiet -o "{playlist-title}/{index} - {title}.{ext}" PLAYLIST_URL

# JSON output for parsing
ytdl-go -json -o "{playlist-title}/{index} - {title}.{ext}" PLAYLIST_URL
```

## Large Playlists

For very large playlists (100+ videos):

```bash
# Increase timeout for slow connections
ytdl-go -timeout 30m \
        -o "{playlist-title}/{index} - {title}.{ext}" \
        PLAYLIST_URL

# Use quiet mode to reduce terminal clutter
ytdl-go -quiet \
        -o "{playlist-title}/{index} - {title}.{ext}" \
        PLAYLIST_URL

# JSON logging for monitoring
ytdl-go -json \
        -o "{playlist-title}/{index} - {title}.{ext}" \
        PLAYLIST_URL > playlist-download.log
```

## Troubleshooting Playlists

### Incomplete Playlist Downloads

**Problem**: Not all videos downloaded

**Solutions**:
```bash
# Run the same command again - ytdl-go will prompt for each existing file
# To avoid prompts, use {id} in the template for unique filenames
ytdl-go -o "{playlist-title}/{index} - {title}-{id}.{ext}" PLAYLIST_URL

# Check for errors with debug logging
ytdl-go -log-level debug -o "{playlist-title}/{index} - {title}.{ext}" PLAYLIST_URL
```

### Private or Unavailable Videos

**Problem**: Some videos in playlist are unavailable

**Behavior**: ytdl-go automatically skips unavailable videos and continues with the rest.

> **Authentication Not Supported**
>
> Private, age-gated, or member-only videos cannot be downloaded as authentication is currently not supported.

### Slow Playlist Processing

**Solutions**:
```bash
# Increase timeout
ytdl-go -timeout 30m PLAYLIST_URL

# Lower quality for faster downloads
ytdl-go -quality 480p PLAYLIST_URL
```

## Best Practices

1. **Use numbered filenames** for playlists to maintain order
2. **Include playlist title** in the path for organization
3. **Set appropriate quality** to balance size and download time
4. **Use quiet or JSON mode** for large automated downloads
5. **Monitor first few downloads** to ensure correct organization

## Example: Complete Workflow

Here's a complete workflow for archiving a video course:

```bash
# 1. First, check the playlist info
ytdl-go -info https://www.youtube.com/playlist?list=PLxxxxx | jq .

# 2. Download with proper organization
ytdl-go -quality 720p -format mp4 \
        -o "Courses/{playlist-title}/Lesson {index} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx

# 3. If interrupted, resume with the same command
ytdl-go -quality 720p -format mp4 \
        -o "Courses/{playlist-title}/Lesson {index} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

Result:
```
Courses/
└── Python Programming Course/
    ├── Lesson 1 - Introduction.mp4
    ├── Lesson 2 - Variables.mp4
    ├── Lesson 3 - Functions.mp4
    └── ...
```

## Next Steps

- **Audio-Only**: Download music playlists with [audio-only mode](audio-only)
- **Templates**: Learn more about [output templates](output-templates)
- **Metadata**: Add custom metadata with [metadata sidecars](metadata-sidecars)
- **Format Selection**: Fine-tune quality with [format selection](format-selection)

## Related References

- [Basic Downloads Guide](basic-downloads)
- [Output Templates Reference](output-templates)
- [Command-Line Flags Reference](../../reference/cli-options)
