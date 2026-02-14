---
title: "Format Selection"
weight: 10
---

# Format Selection

This guide covers ytdl-go's format selection capabilities, including the interactive format selector, quality options, and direct itag selection.

## Format Selection Methods

ytdl-go offers three ways to select formats:

1. **Automatic Selection** - Let ytdl-go pick the best format (default)
2. **Preference-Based** - Specify quality/format preferences with flags
3. **Interactive Selection** - Browse and choose from all available formats
4. **Direct Selection** - Specify exact format by itag number

## Automatic Format Selection

By default, ytdl-go automatically selects the best available format:

```bash
# Best video quality
ytdl-go URL

# Best audio quality
ytdl-go -audio URL
```

### Selection Logic

**For Video:**
1. Highest resolution progressive format (audio+video in one stream)
2. Selects based on resolution and bitrate
3. No codec-priority ranking implemented

**For Audio:**
1. Highest bitrate audio-only format
2. Selects based on bitrate only
3. No codec-priority ranking implemented (Opus/AAC/MP3 selected by bitrate)

## Preference-Based Selection

### Quality Flags

Specify desired quality with `-quality`:

```bash
# Specific video resolution
ytdl-go -quality 1080p URL
ytdl-go -quality 720p URL
ytdl-go -quality 480p URL

# Specific audio bitrate
ytdl-go -audio -quality 256k URL
ytdl-go -audio -quality 192k URL
ytdl-go -audio -quality 128k URL

# Best or worst quality
ytdl-go -quality best URL
ytdl-go -quality worst URL
```

### Format Flags

Specify preferred container with `-format`:

```bash
# Video containers
ytdl-go -format mp4 URL
ytdl-go -format webm URL

# Audio containers
ytdl-go -audio -format m4a URL
ytdl-go -audio -format opus URL
ytdl-go -audio -format mp3 URL
```

### Combined Preferences

Combine quality and format for precise selection:

```bash
# 720p MP4 video
ytdl-go -quality 720p -format mp4 URL

# 256k M4A audio
ytdl-go -audio -quality 256k -format m4a URL

# Best quality WebM
ytdl-go -quality best -format webm URL
```

> **Info: Closest Match**
    If exact preferences aren't available, ytdl-go selects the closest available match.

## Interactive Format Selector

The interactive format selector provides a visual interface to browse and select formats.

### Launching the Selector

```bash
ytdl-go -list-formats URL
```

This opens an interactive terminal UI showing all available formats.

![Interactive Format Selector](../../screenshots/interactive-format-selector.svg)

### Format Display

The selector shows:
- **Itag**: YouTube format identifier
- **Type**: Video, Audio, or Video+Audio (progressive)
- **Quality**: Resolution (e.g., 1080p) or bitrate (e.g., 128k)
- **Codec**: Video codec (H.264, VP9) and audio codec (Opus, AAC)
- **Container**: File format (MP4, WebM, M4A)
- **Size**: Estimated file size

### Keyboard Controls

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Page Up` | Jump up 10 formats |
| `Page Down` | Jump down 10 formats |
| `Home` / `g` | Jump to first format |
| `End` / `G` | Jump to last format |
| `Enter` | Download selected format |
| `0-9` | Quick jump to itag (type digits) |
| `b` | Go back without downloading |
| `q` / `Esc` / `Ctrl+C` | Quit |

### Quick Itag Jump

Type digits to quickly jump to formats starting with those numbers:

```
Type: 1    → Jumps to formats starting with 1 (18, 137, 140, etc.)
Type: 14   → Jumps to formats starting with 14 (140, 141, etc.)
Type: 251  → Jumps to format 251 (if it exists)
```

Press the same digits again to cycle through matching formats.

> **Tip: Fast Navigation**
    If you know the itag number, type it directly instead of scrolling through all formats.

### Format Categories

Formats are typically organized as:

1. **Progressive Formats** (Video + Audio combined)
   - Lower quality but simpler
   - Single file download
   - Best compatibility

2. **Video-Only Formats**
   - Higher quality options
   - Requires separate audio download
   - Best quality available

3. **Audio-Only Formats**
   - Various bitrates and codecs
   - Smaller file sizes
   - Perfect for music/podcasts

### Example Session

```bash
# Launch interactive selector
ytdl-go -list-formats https://www.youtube.com/watch?v=VIDEO_ID

# Navigate with arrow keys to desired format
# Press Enter to download

# Or type itag number (e.g., "140") and press Enter
```

## Direct Itag Selection

If you know the specific format you want, use the `-itag` flag:

```bash
ytdl-go -itag 137 URL
```

### Finding Itag Numbers

Use the interactive selector or `-list-formats` to find itags:

```bash
ytdl-go -list-formats URL
```

### Common YouTube Itags

#### Video Formats (Progressive - includes audio)

| Itag | Resolution | Container | Video Codec | Audio Codec |
|------|------------|-----------|-------------|-------------|
| 18 | 360p | MP4 | H.264 | AAC |
| 22 | 720p | MP4 | H.264 | AAC |

#### Video-Only Formats

| Itag | Resolution | Container | Codec | Notes |
|------|------------|-----------|-------|-------|
| 137 | 1080p | MP4 | H.264 | High quality |
| 136 | 720p | MP4 | H.264 | HD quality |
| 135 | 480p | MP4 | H.264 | Standard |
| 134 | 360p | MP4 | H.264 | Mobile |
| 247 | 720p | WebM | VP9 | Efficient |
| 248 | 1080p | WebM | VP9 | Very efficient |

#### Audio-Only Formats

| Itag | Bitrate | Container | Codec | Quality |
|------|---------|-----------|-------|---------|
| 251 | ~160kbps | WebM | Opus | High |
| 250 | ~70kbps | WebM | Opus | Medium |
| 249 | ~50kbps | WebM | Opus | Low |
| 140 | 128kbps | M4A | AAC | Standard |
| 141 | 256kbps | M4A | AAC | High |
| 139 | 48kbps | M4A | AAC | Low |

> **Warning: Itag Availability**
    Not all itags are available for every video. Use `-list-formats` to see what's available for your specific video.

### Itag Examples

```bash
# High-quality 1080p video-only (requires audio merge)
ytdl-go -itag 137 URL

# 720p progressive (video + audio)
ytdl-go -itag 22 URL

# High-quality Opus audio
ytdl-go -itag 251 URL

# Standard AAC audio
ytdl-go -itag 140 URL

# High-quality AAC audio
ytdl-go -itag 141 URL
```

## Understanding Format Types

### Progressive Formats

**Combined video and audio in single file**

Advantages:
- Single download
- No merging required
- Maximum compatibility
- Simpler processing

Disadvantages:
- Limited quality options
- Usually lower maximum resolution
- Larger file sizes for equivalent quality

```bash
# Download progressive format
ytdl-go -itag 22 URL  # 720p MP4 with audio
```

### Adaptive Formats

**Separate video and audio streams**

Advantages:
- Highest quality options
- More format choices
- Better compression efficiency
- Flexible audio/video combinations

Disadvantages:
- Requires merging (handled automatically)
- Two downloads required
- Slightly more complex

```bash
# Download adaptive format (auto-merges)
ytdl-go -itag 137 URL  # 1080p video + best audio
```

> **Info: Automatic Merging**
    ytdl-go automatically handles merging when necessary. You don't need to worry about combining audio and video manually.

## Format Selection Strategies

### Maximum Quality

```bash
# Best available video
ytdl-go -quality best URL

# Specific high quality
ytdl-go -quality 1080p -format mp4 URL

# Interactive selection of highest quality
ytdl-go -list-formats URL
# Select the highest resolution format
```

### Maximum Compatibility

```bash
# MP4 format (most compatible)
ytdl-go -format mp4 URL

# Progressive format (single file)
ytdl-go -itag 22 URL  # 720p MP4

# Standard quality MP4
ytdl-go -quality 720p -format mp4 URL
```

### Minimum File Size

```bash
# Lowest quality
ytdl-go -quality worst URL

# Specific low resolution
ytdl-go -quality 360p URL

# Low bitrate audio
ytdl-go -audio -quality 64k URL
```

### Balanced Quality/Size

```bash
# 720p is often the sweet spot
ytdl-go -quality 720p URL

# WebM offers good compression
ytdl-go -quality 720p -format webm URL

# Medium audio quality
ytdl-go -audio -quality 128k URL
```

### Specific Codec Requirements

```bash
# H.264 video codec (via itag)
ytdl-go -itag 137 URL  # 1080p H.264

# Opus audio codec
ytdl-go -itag 251 URL  # ~160kbps Opus

# AAC audio codec
ytdl-go -itag 140 URL  # 128kbps AAC
```

## Machine-Readable Format List

Get formats as JSON for scripting:

```bash
ytdl-go -json -list-formats URL
```

Output format:
```json
{
  "type": "formats",
  "formats": [
    {
      "itag": 137,
      "container": "mp4",
      "quality": "1080p",
      "qualityLabel": "1080p",
      "codec": "avc1.640028",
      "bitrate": 2500000,
      "size": 52428800
    },
    ...
  ]
}
```

### Parsing with jq

```bash
# List all itags
ytdl-go -json -list-formats URL | jq -r '.formats[].itag'

# Find 1080p formats
ytdl-go -json -list-formats URL | jq '.formats[] | select(.quality == "1080p")'

# Get best audio bitrate
ytdl-go -json -list-formats URL | jq '[.formats[] | select(.type == "audio")] | sort_by(.bitrate) | last'

# List video-only formats
ytdl-go -json -list-formats URL | jq '.formats[] | select(.type == "video")'
```

## Advanced Format Selection

### Testing Formats

Test different formats to find the best balance:

```bash
# Get video info first
ytdl-go -info URL | jq .

# List available formats
ytdl-go -list-formats URL

# Download a few test formats
ytdl-go -itag 137 -o "test-1080p.{ext}" URL
ytdl-go -itag 136 -o "test-720p.{ext}" URL
ytdl-go -itag 135 -o "test-480p.{ext}" URL

# Compare file sizes and quality
ls -lh test-*.mp4
```

### Format Filtering Scripts

Example script to select format based on criteria:

```bash
#!/bin/bash
URL="$1"

# Get formats as JSON
formats=$(ytdl-go -json -list-formats "$URL")

# Find best MP4 format under 100MB
itag=$(echo "$formats" | jq -r '
  .formats[] 
  | select(.container == "mp4" and .size < 104857600) 
  | .itag' 
  | head -n1)

# Download selected format
ytdl-go -itag "$itag" "$URL"
```

### Fallback Strategy

```bash
#!/bin/bash
URL="$1"

# Try 1080p first
ytdl-go -quality 1080p -format mp4 "$URL" 2>/dev/null || \
# Fall back to 720p
ytdl-go -quality 720p -format mp4 "$URL" 2>/dev/null || \
# Fall back to best available
ytdl-go -quality best "$URL"
```

## Format Selection Best Practices

1. **Use interactive mode** when unsure about available formats
2. **Test formats** before batch downloads
3. **Consider bandwidth** when selecting quality
4. **Check file sizes** in the format list before downloading
5. **Use progressive formats** for maximum compatibility
6. **Choose adaptive formats** for highest quality
7. **Know your itags** for consistent automated downloads
8. **Document your choices** for reproducible workflows

## Troubleshooting Format Selection

### Format Not Available

**Problem**: Requested format doesn't exist

**Solutions**:
```bash
# Check available formats
ytdl-go -list-formats URL

# Use best available instead
ytdl-go -quality best URL

# Try different format
ytdl-go -format webm URL
```

### Quality Lower Than Expected

**Problem**: Downloaded quality doesn't match request

**Causes**:
- Video doesn't have that quality
- Closest match selected instead
- Format filtering too restrictive

**Solutions**:
```bash
# Check what's actually available
ytdl-go -list-formats URL

# Remove format restriction
ytdl-go -quality 1080p URL  # Without -format

# Use specific itag
ytdl-go -itag 137 URL
```

### Interactive Selector Not Showing

**Problem**: `-list-formats` doesn't launch UI

**Causes**:
- JSON mode active (`-json`)
- Quiet mode active (`-quiet`)
- Not a TTY (piped output)

**Solutions**:
```bash
# Remove conflicting flags
ytdl-go -list-formats URL  # No -json or -quiet

# Use in terminal (not piped)
ytdl-go -list-formats URL
```

### Large File Size

**Problem**: Downloaded file is too large

**Solutions**:
```bash
# Check sizes before downloading
ytdl-go -list-formats URL  # Review size column

# Select lower quality
ytdl-go -quality 720p URL

# Choose specific smaller format
ytdl-go -itag 135 URL  # 480p instead of 1080p
```

## Next Steps

- **Basic Downloads**: Start with [basic downloads](basic-downloads)
- **Audio-Only**: Learn about [audio downloads](audio-only)
- **Playlists**: Apply format selection to [playlists](playlists)
- **Templates**: Organize with [output templates](output-templates)

## Related References

- [Basic Downloads Guide](basic-downloads)
- [Audio-Only Downloads Guide](audio-only)
- [Command-Line Flags Reference](../../reference/cli-options)
