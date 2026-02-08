---
title: "Basic Downloads"
weight: 40
---

# Basic Downloads

This guide covers the fundamentals of downloading videos with ytdl-go, including quality selection and format options.

## Quick Start

The simplest way to download a video is to provide just the URL:

```bash
ytdl-go https://www.youtube.com/watch?v=BaW_jenozKc
```

This downloads the video in the best available quality with the default filename pattern: `{title}.{ext}`

> **Default Behavior**
>
> By default, ytdl-go automatically selects the highest quality video format available and saves it with the video's title as the filename.

## Quality Selection

### Specific Resolution

You can target a specific video resolution using the `-quality` flag:

```bash
# Download in 1080p
ytdl-go -quality 1080p https://www.youtube.com/watch?v=VIDEO_ID

# Download in 720p
ytdl-go -quality 720p https://www.youtube.com/watch?v=VIDEO_ID

# Download in 480p
ytdl-go -quality 480p https://www.youtube.com/watch?v=VIDEO_ID
```

### Available Quality Options

| Quality Option | Description | Use Case |
|----------------|-------------|----------|
| `best` | Highest available quality (default) | Maximum quality |
| `1080p` | Full HD resolution | High quality viewing |
| `720p` | HD resolution | Balanced quality/size |
| `480p` | Standard definition | Smaller file size |
| `360p` | Low resolution | Limited bandwidth |
| `240p` | Very low resolution | Minimal bandwidth |
| `144p` | Lowest resolution | Extremely limited bandwidth |
| `worst` | Lowest available quality | Smallest possible file |

> **Quality Matching**
>
> If the exact quality you specify isn't available, ytdl-go automatically selects the closest match.

### Example Quality Commands

```bash
# Get the absolute best quality available
ytdl-go -quality best https://www.youtube.com/watch?v=VIDEO_ID

# Get the smallest file size
ytdl-go -quality worst https://www.youtube.com/watch?v=VIDEO_ID

# Download in HD
ytdl-go -quality 720p https://www.youtube.com/watch?v=VIDEO_ID
```

## Format Selection

### Container Format Preference

Use the `-format` flag to specify your preferred container format:

```bash
# Prefer MP4 container (most compatible)
ytdl-go -format mp4 https://www.youtube.com/watch?v=VIDEO_ID

# Prefer WebM container (open source)
ytdl-go -format webm https://www.youtube.com/watch?v=VIDEO_ID
```

### Common Video Formats

| Format | Description | Compatibility |
|--------|-------------|---------------|
| `mp4` | MPEG-4 container | Excellent - plays on nearly all devices |
| `webm` | WebM container | Good - modern browsers and players |

> **Format Availability**
>
> If your requested format isn't available, the download may fail. Consider using the [interactive format selector](format-selection) to see available options.

### Combining Quality and Format

You can combine quality and format preferences for precise control:

```bash
# Download 720p MP4
ytdl-go -quality 720p -format mp4 https://www.youtube.com/watch?v=VIDEO_ID

# Download best quality WebM
ytdl-go -quality best -format webm https://www.youtube.com/watch?v=VIDEO_ID

# Download 1080p MP4 (most popular combination)
ytdl-go -quality 1080p -format mp4 https://www.youtube.com/watch?v=VIDEO_ID
```

## Basic Output Control

### Custom Filename

Use the `-o` flag to specify a custom filename:

```bash
# Simple filename
ytdl-go -o "myvideo.mp4" https://www.youtube.com/watch?v=VIDEO_ID

# Filename with quality indicator
ytdl-go -o "video-{quality}.{ext}" https://www.youtube.com/watch?v=VIDEO_ID

# Filename with video ID
ytdl-go -o "{title}-{id}.{ext}" https://www.youtube.com/watch?v=VIDEO_ID
```

> **Output Templates**
>
> For advanced filename customization, see the [Output Templates](output-templates) guide.

### Download to Specific Directory

Create a relative path to organize downloads:

```bash
# Save to Downloads folder
ytdl-go -o "Downloads/{title}.{ext}" https://www.youtube.com/watch?v=VIDEO_ID

# Organize by quality
ytdl-go -o "Videos/{quality}/{title}.{ext}" https://www.youtube.com/watch?v=VIDEO_ID
```

> **Relative Paths Only**
>
> Output paths must be relative. Absolute paths are rejected for security reasons.

## Resume Capability

ytdl-go automatically resumes interrupted downloads:

```bash
# Start download
ytdl-go https://www.youtube.com/watch?v=VIDEO_ID

# If interrupted, just run the same command again
ytdl-go https://www.youtube.com/watch?v=VIDEO_ID
```

The tool detects `.part` files and resumes from where it left off.

## Multiple Downloads

Download multiple videos in sequence:

```bash
ytdl-go https://www.youtube.com/watch?v=VIDEO_ID1 \
        https://www.youtube.com/watch?v=VIDEO_ID2 \
        https://www.youtube.com/watch?v=VIDEO_ID3
```

For concurrent downloads, use the `-jobs` flag:

```bash
# Download 4 videos simultaneously
ytdl-go -jobs 4 \
        https://www.youtube.com/watch?v=VIDEO_ID1 \
        https://www.youtube.com/watch?v=VIDEO_ID2 \
        https://www.youtube.com/watch?v=VIDEO_ID3 \
        https://www.youtube.com/watch?v=VIDEO_ID4
```

> **Concurrency Recommendations**
>
> - Network-bound: 4-8 jobs
    - Disk I/O-bound: 1-2 jobs
    - Large files: Keep low (1-2) to avoid memory pressure

## Getting Video Information

Preview video metadata without downloading:

```bash
ytdl-go -info https://www.youtube.com/watch?v=VIDEO_ID
```

This outputs JSON with:
- Video ID
- Title
- Description
- Author/Channel
- Duration

```bash
# Pretty-print with jq
ytdl-go -info https://www.youtube.com/watch?v=VIDEO_ID | jq .

# Extract just the title
ytdl-go -info https://www.youtube.com/watch?v=VIDEO_ID | jq -r .title
```

## Quiet Mode

Suppress progress output for scripting:

```bash
ytdl-go -quiet https://www.youtube.com/watch?v=VIDEO_ID
```

Errors are still shown to stderr, but progress bars are hidden.

## Common Workflows

### High-Quality Video Download

```bash
ytdl-go -quality 1080p -format mp4 -o "Videos/{title}.{ext}" https://www.youtube.com/watch?v=VIDEO_ID
```

### Small File Size Download

```bash
ytdl-go -quality 480p -format mp4 -o "Videos/{title}.{ext}" https://www.youtube.com/watch?v=VIDEO_ID
```

### Batch Download with Organization

```bash
ytdl-go -quality 720p -format mp4 -o "Downloads/{title}-{id}.{ext}" \
        https://www.youtube.com/watch?v=VIDEO_ID1 \
        https://www.youtube.com/watch?v=VIDEO_ID2 \
        https://www.youtube.com/watch?v=VIDEO_ID3
```

### Check Information Before Downloading

```bash
# First, check the video info
ytdl-go -info https://www.youtube.com/watch?v=VIDEO_ID

# Then download if it looks good
ytdl-go -quality 1080p https://www.youtube.com/watch?v=VIDEO_ID
```

## Troubleshooting

### 403 Forbidden Errors

ytdl-go automatically retries with different methods. If issues persist:

```bash
# Increase timeout
ytdl-go -timeout 10m https://www.youtube.com/watch?v=VIDEO_ID
```

### Quality Not Available

If a specific quality isn't available:

1. Use the [interactive format selector](format-selection) to see all options
2. Try a different quality level
3. Check if the video has that quality available on YouTube

```bash
# See all available formats
ytdl-go -list-formats https://www.youtube.com/watch?v=VIDEO_ID
```

### Slow Downloads

For slow connections, increase the timeout:

```bash
ytdl-go -timeout 30m https://www.youtube.com/watch?v=VIDEO_ID
```

## Next Steps

- **Audio Downloads**: Learn about [audio-only downloads](audio-only)
- **Playlists**: Download [entire playlists](playlists)
- **Format Selection**: Use the [interactive format selector](format-selection)
- **Templates**: Master [output templates](output-templates) for file organization
- **Metadata**: Explore [metadata and sidecars](metadata-sidecars)

## Related References

- [Command-Line Options Reference](../../reference/cli-options)
- [Format Selection Guide](format-selection)
- [Troubleshooting Guide](../troubleshooting/common-issues)
