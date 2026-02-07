---
title: "Audio-Only Downloads"
weight: 50
---

# Audio-Only Downloads

This guide covers downloading audio-only content from YouTube, perfect for music, podcasts, and other audio-focused content.

## Basic Audio Download

Download audio without video using the `-audio` flag:

```bash
ytdl-go -audio https://www.youtube.com/watch?v=VIDEO_ID
```

This downloads the highest quality audio-only stream available.

> **Automatic Best Quality**
>
> The `-audio` flag automatically selects the highest bitrate audio format available.

## How Audio Selection Works

ytdl-go uses a smart audio selection strategy:

1. **First Choice**: Highest bitrate audio-only format
2. **Codec Priority**: Opus > AAC > MP3
3. **Fallback**: Extract audio from progressive video if audio-only unavailable

### Audio Extraction Strategy

If direct audio download fails (403 error):

1. Retry with single request method
2. Use FFmpeg fallback (extracts from video)

> **No External Dependencies**
>
> ytdl-go handles audio extraction internally without requiring FFmpeg to be installed.

## Audio Quality Selection

### Specific Bitrate

Target a specific audio bitrate using the `-quality` flag:

```bash
# High quality (320kbps)
ytdl-go -audio -quality 320k https://www.youtube.com/watch?v=VIDEO_ID

# Standard quality (192kbps)
ytdl-go -audio -quality 192k https://www.youtube.com/watch?v=VIDEO_ID

# Medium quality (128kbps)
ytdl-go -audio -quality 128k https://www.youtube.com/watch?v=VIDEO_ID

# Lower quality (96kbps)
ytdl-go -audio -quality 96k https://www.youtube.com/watch?v=VIDEO_ID
```

### Available Audio Quality Options

| Quality Option | Bitrate | Description | File Size | Use Case |
|----------------|---------|-------------|-----------|----------|
| `best` | Variable | Highest available (default) | Largest | Archival, high-fidelity |
| `320k` | ~320 kbps | Near-lossless quality | Large | Music collection |
| `256k` | ~256 kbps | Very high quality | Medium-Large | Premium audio |
| `192k` | ~192 kbps | High quality | Medium | General music |
| `128k` | ~128 kbps | Standard quality | Small | Casual listening |
| `96k` | ~96 kbps | Acceptable quality | Very Small | Podcasts, speech |
| `64k` | ~64 kbps | Low quality | Minimal | Limited storage/bandwidth |
| `worst` | Variable | Lowest available | Smallest | Extreme space constraints |

> **Actual Bitrates**
>
> Actual bitrates may vary based on what's available. ytdl-go selects the closest match.

## Audio Format Selection

### Preferred Audio Format

Use `-format` to specify the audio container:

```bash
# M4A format (AAC codec)
ytdl-go -audio -format m4a https://www.youtube.com/watch?v=VIDEO_ID

# Opus format (best quality/size ratio)
ytdl-go -audio -format opus https://www.youtube.com/watch?v=VIDEO_ID

# MP3 format (maximum compatibility)
ytdl-go -audio -format mp3 https://www.youtube.com/watch?v=VIDEO_ID
```

### Audio Format Comparison

| Format | Codec | Quality/Size | Compatibility | Recommended For |
|--------|-------|--------------|---------------|-----------------|
| `m4a` | AAC | Excellent | Very Good | General use, Apple devices |
| `opus` | Opus | Best | Good (modern devices) | High quality, efficient |
| `mp3` | MP3 | Good | Excellent | Maximum compatibility |
| `webm` | Vorbis/Opus | Good | Good | Web-based use |

> **Format Availability**
>
> If your preferred format isn't available, use the [interactive format selector](format-selection) to see all options.

### Combining Quality and Format

```bash
# High-quality M4A
ytdl-go -audio -quality 256k -format m4a https://www.youtube.com/watch?v=VIDEO_ID

# Standard MP3 for compatibility
ytdl-go -audio -quality 192k -format mp3 https://www.youtube.com/watch?v=VIDEO_ID

# Best Opus for efficiency
ytdl-go -audio -quality best -format opus https://www.youtube.com/watch?v=VIDEO_ID
```

## Organizing Audio Downloads

### Music Library Organization

Organize downloads by artist and album:

```bash
# Basic artist/album structure
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" URL

# With track numbers
ytdl-go -audio -o "Music/{artist}/{album}/{track} - {title}.{ext}" URL

# Artist - Title format
ytdl-go -audio -o "Music/{artist} - {title}.{ext}" URL
```

Example directory structure:
```
Music/
└── Artist Name/
    └── Album Name/
        ├── 01 - Song Title 1.m4a
        ├── 02 - Song Title 2.m4a
        └── 03 - Song Title 3.m4a
```

### Podcast Organization

```bash
ytdl-go -audio -quality 96k \
        -o "Podcasts/{artist}/{title}.{ext}" \
        URL
```

### Single Folder Collection

```bash
# Artist - Title format
ytdl-go -audio -o "Music/{artist} - {title}.{ext}" URL

# With quality indicator
ytdl-go -audio -o "Music/{artist} - {title} [{quality}].{ext}" URL
```

## Music Playlist Downloads

Download entire music playlists:

```bash
# Basic playlist download
ytdl-go -audio https://www.youtube.com/playlist?list=PLxxxxx

# Organized with track numbers
ytdl-go -audio \
        -o "Music/{playlist-title}/{index:02d} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx

# With artist/album structure
ytdl-go -audio \
        -o "Music/{artist}/{playlist-title}/{index:02d} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

> **Music Playlists**
>
> For more on playlist organization, see the [Playlist Downloads](playlists) guide.

## Audio Metadata

### Automatic Metadata

ytdl-go automatically extracts and embeds metadata:

- **Title**: Track title
- **Artist**: Channel/artist name
- **Album**: Playlist name (if from playlist)
- **Track Number**: Position in playlist
- **Release Date**: Upload date

### Custom Metadata

Override metadata fields:

```bash
# Set artist and album
ytdl-go -audio \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        URL

# Complete metadata override
ytdl-go -audio \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        -meta title="Track Title" \
        -meta track=1 \
        -meta release_year=2024 \
        URL
```

### Playlist Metadata

For playlists, set consistent metadata:

```bash
ytdl-go -audio \
        -o "Music/{artist}/{album}/{index:02d} - {title}.{ext}" \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

> **Metadata Details**
>
> For complete metadata options, see [Metadata and Sidecars](metadata-sidecars).

## YouTube Music Downloads

ytdl-go works with YouTube Music URLs:

```bash
# Single track
ytdl-go -audio https://music.youtube.com/watch?v=VIDEO_ID

# Music playlist
ytdl-go -audio https://music.youtube.com/playlist?list=PLxxxxx

# Album with metadata
ytdl-go -audio \
        -o "Music/{artist}/{album}/{index:02d} - {title}.{ext}" \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        https://music.youtube.com/playlist?list=PLxxxxx
```

## Common Audio Workflows

### High-Quality Music Collection

```bash
ytdl-go -audio -quality 256k -format m4a \
        -o "Music/{artist}/{album}/{index:02d} - {title}.{ext}" \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

### Podcast Episodes

```bash
ytdl-go -audio -quality 96k \
        -o "Podcasts/{playlist-title}/Episode {index:03d} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

### Audiobook Chapters

```bash
ytdl-go -audio -quality 128k \
        -o "Audiobooks/{playlist-title}/Chapter {index:02d} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

### Speech/Lecture Content

```bash
ytdl-go -audio -quality 96k -format m4a \
        -o "Lectures/{playlist-title}/{index:03d} - {title}.{ext}" \
        https://www.youtube.com/playlist?list=PLxxxxx
```

### DJ Mix/Set Archive

```bash
ytdl-go -audio -quality best -format opus \
        -o "DJ Sets/{artist}/{title} [{id}].{ext}" \
        URL
```

### Batch Music Downloads

```bash
# Multiple singles
ytdl-go -audio -quality 256k \
        -o "Music/{artist} - {title}.{ext}" \
        URL1 URL2 URL3

# With parallel downloads
ytdl-go -audio -quality 256k -jobs 4 \
        -o "Music/{artist} - {title}.{ext}" \
        URL1 URL2 URL3 URL4
```

## Advanced Audio Options

### Interactive Audio Format Selection

Browse all available audio formats:

```bash
ytdl-go -audio -list-formats https://www.youtube.com/watch?v=VIDEO_ID
```

This shows:
- All audio-only formats
- Bitrates
- Codecs
- File sizes

Use arrow keys to select the specific format you want.

### Specific Audio Codec (by itag)

If you need a specific codec, use the itag number:

```bash
# Opus ~160kbps (itag 251)
ytdl-go -itag 251 URL

# AAC 128kbps (itag 140)
ytdl-go -itag 140 URL

# AAC 256kbps (itag 141)
ytdl-go -itag 141 URL
```

Common audio itags:

| Itag | Format | Codec | Bitrate | Quality |
|------|--------|-------|---------|---------|
| 251 | webm | Opus | ~160kbps | High |
| 250 | webm | Opus | ~70kbps | Medium |
| 249 | webm | Opus | ~50kbps | Low |
| 140 | m4a | AAC | 128kbps | Standard |
| 141 | m4a | AAC | 256kbps | High |

> **Find Itags**
>
> Use `-list-formats` to see all available itags for a specific video.

### Resume Interrupted Audio Downloads

```bash
# Start download
ytdl-go -audio -o "Music/{title}.{ext}" URL

# If interrupted, run same command to resume
ytdl-go -audio -o "Music/{title}.{ext}" URL
```

## Scripting and Automation

### JSON Output for Monitoring

```bash
ytdl-go -audio -json -o "Music/{artist} - {title}.{ext}" URL > download.log
```

### Quiet Mode for Cron Jobs

```bash
ytdl-go -audio -quiet -o "Music/{artist}/{title}.{ext}" URL
```

### Batch Processing with URLs from File

```bash
# Read URLs from file and download
while read url; do
    ytdl-go -audio -o "Music/{artist} - {title}.{ext}" "$url"
done < urls.txt
```

## Troubleshooting Audio Downloads

### No Audio-Only Format Available

**Problem**: Some videos don't have audio-only streams

**Solution**: ytdl-go automatically extracts audio from video
```bash
# This is handled automatically
ytdl-go -audio URL
```

### Audio Quality Lower Than Expected

**Problem**: Requested quality not available

**Solutions**:
```bash
# Check available formats
ytdl-go -list-formats URL

# Use best available
ytdl-go -audio -quality best URL
```

### Format Not Available

**Problem**: Requested audio format doesn't exist

**Solutions**:
```bash
# Check available formats
ytdl-go -audio -list-formats URL

# Try different format
ytdl-go -audio -format m4a URL
```

### Metadata Not Embedded

**Problem**: Audio file has no metadata tags

**Solutions**:
```bash
# Explicitly set metadata
ytdl-go -audio \
        -meta artist="Artist Name" \
        -meta title="Song Title" \
        -meta album="Album Name" \
        URL

# Check sidecar JSON file for metadata
cat "song-title.info.json"
```

## Performance Tips

1. **Use appropriate quality** - Lower bitrate for speech (96k), higher for music (256k)
2. **Choose efficient formats** - Opus provides best quality/size ratio
3. **Batch downloads** - Use `-jobs` for multiple files
4. **Resume interrupted downloads** - Re-run same command
5. **Use quiet mode** - For large automated downloads

## Best Practices

1. **Music**: Use 256k or 320k M4A/Opus for best quality
2. **Podcasts/Speech**: Use 96k or 128k for good quality and small size
3. **Audiobooks**: Use 96k or 128k for clarity with reasonable file size
4. **DJ Sets/Mixes**: Use best quality Opus for fidelity
5. **Organization**: Use consistent templates for easy library management
6. **Metadata**: Set proper metadata for music library integration
7. **Batch Processing**: Use quiet or JSON mode for monitoring

## Next Steps

- **Playlists**: Download music playlists with [playlist guides](playlists)
- **Metadata**: Learn about [metadata and sidecars](metadata-sidecars)
- **Templates**: Master [output templates](output-templates)
- **Format Selection**: Fine-tune with [interactive format selector](format-selection)

## Related References

- [Basic Downloads Guide](basic-downloads)
- [Metadata and Sidecars Guide](metadata-sidecars)
- [Command-Line Flags Reference](../../reference/flags)
