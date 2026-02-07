---
title: "Output Placeholders"
weight: 20
---

# Output Placeholders

Complete reference for template placeholders used with the `-o` (output template) flag. These placeholders are dynamically replaced with video metadata to create organized file structures.

## Overview

Output templates allow you to organize downloaded files using video metadata. Placeholders are enclosed in curly braces `{placeholder}` and are replaced with actual values during download.

**Basic Example:**

```bash
ytdl-go -o "{artist} - {title}.{ext}" [URL]
# Results in: "Artist Name - Video Title.mp4"
```

## Complete Placeholder Reference

### Video Metadata Placeholders

| Placeholder | Description | Example Value | Notes |
|-------------|-------------|---------------|-------|
| `{title}` | Video title (sanitized) | `My Video Title` | Special characters removed/replaced |
| `{id}` | YouTube video ID | `dQw4w9WgXcQ` | Unique 11-character identifier |
| `{ext}` | File extension | `mp4`, `webm`, `m4a` | Based on selected format |
| `{quality}` | Quality label or bitrate | `1080p`, `720p`, `128k` | Video resolution or audio bitrate |

### Author/Artist Placeholders

| Placeholder | Description | Example Value | Notes |
|-------------|-------------|---------------|-------|
| `{artist}` | Video author/artist (sanitized) | `Artist Name` | From uploader or metadata |
| `{author}` | Channel/uploader name (sanitized) | `Channel Name` | Same as artist in most cases |

### Album Placeholders

| Placeholder | Description | Example Value | Notes |
|-------------|-------------|---------------|-------|
| `{album}` | Album name (sanitized) | `Album Name` | From YouTube Music metadata |
| `{track}` | Track number | `5` | From metadata (if available) |
| `{disc}` | Disc number | `1` | From metadata (if available) |

### Playlist Placeholders

| Placeholder | Description | Example Value | Notes |
|-------------|-------------|---------------|-------|
| `{playlist_title}` | Playlist name (sanitized) | `My Playlist` | Underscore or hyphen variants |
| `{playlist-title}` | Playlist name (sanitized) | `My Playlist` | Alternative syntax |
| `{playlist_id}` | Playlist ID | `PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf` | Unique playlist identifier |
| `{playlist-id}` | Playlist ID | `PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf` | Alternative syntax |
| `{index}` | Video position in playlist | `1`, `2`, `3` | 1-based indexing |
| `{count}` | Total videos in playlist | `25` | Total playlist size |

### Date/Time Placeholders

| Placeholder | Description | Example Value | Notes |
|-------------|-------------|---------------|-------|
| `{release_date}` | Release date | `2023-01-15` | Format: YYYY-MM-DD |
| `{release_year}` | Release year | `2023` | Format: YYYY |
| `{upload_date}` | Upload date | `20230115` | Format: YYYYMMDD |

### Technical Placeholders

| Placeholder | Description | Example Value | Notes |
|-------------|-------------|---------------|-------|
| `{itag}` | YouTube format tag | `251`, `140`, `22` | Internal format identifier |
| `{codec}` | Video/audio codec | `h264`, `vp9`, `opus` | Container codec |
| `{bitrate}` | Audio bitrate | `128000`, `256000` | In bits per second |
| `{fps}` | Video frame rate | `30`, `60` | Frames per second |
| `{width}` | Video width | `1920`, `1280` | In pixels |
| `{height}` | Video height | `1080`, `720` | In pixels |

## Placeholder Behavior

### Sanitization

All placeholders containing text are automatically sanitized for filesystem compatibility:

**Characters Removed/Replaced:**
- `/` → `-` (path separator)
- `\` → `-` (Windows path separator)
- `:` → `-` (Windows drive separator)
- `*` → `_` (wildcard)
- `?` → `_` (wildcard)
- `"` → `'` (quotes)
- `<` → `(` (redirects)
- `>` → `)` (redirects)
- `|` → `-` (pipe)

**Example:**

```bash
# Original title: "Video: Part 1/2 (HD)"
# Sanitized: "Video- Part 1-2 (HD)"
```

### Missing Values

If a placeholder's value is not available, it is replaced with a safe default:

- Text placeholders → `"unknown"` or empty string
- Numeric placeholders → `0`
- ID placeholders → video ID or `"unknown"`

**Example:**

```bash
# No album metadata available
ytdl-go -o "{artist}/{album}/{title}.{ext}" [URL]
# Results in: "Artist Name/unknown/Video Title.mp4"
```

### Case Sensitivity

Placeholders are case-sensitive. Use the exact casing shown in this reference.

**Incorrect:**
```bash
ytdl-go -o "{Title}.{ext}" [URL]  # Won't work
```

**Correct:**
```bash
ytdl-go -o "{title}.{ext}" [URL]  # Works
```

## Common Use Cases

### Simple File Naming

```bash
# Just the title
ytdl-go -o "{title}.{ext}" [URL]
# Output: "Video Title.mp4"

# Include quality
ytdl-go -o "{title} [{quality}].{ext}" [URL]
# Output: "Video Title [1080p].mp4"

# Include video ID
ytdl-go -o "{title} ({id}).{ext}" [URL]
# Output: "Video Title (dQw4w9WgXcQ).mp4"
```

### Music Library Organization

```bash
# Artist/Album/Track structure
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" [URL]
# Output: "Music/Artist Name/Album Name/Song Title.m4a"

# Include track numbers
ytdl-go -audio -o "Music/{artist}/{album}/{track} - {title}.{ext}" [URL]
# Output: "Music/Artist Name/Album Name/05 - Song Title.m4a"

# Flat structure with artist
ytdl-go -audio -o "Music/{artist} - {title}.{ext}" [URL]
# Output: "Music/Artist Name - Song Title.m4a"
```

### Playlist Downloads

```bash
# Numbered playlist items
ytdl-go -o "{playlist_title}/{index} - {title}.{ext}" [PLAYLIST_URL]
# Output: "My Playlist/01 - Video Title.mp4"

# Include total count
ytdl-go -o "{playlist_title}/{index} of {count} - {title}.{ext}" [PLAYLIST_URL]
# Output: "My Playlist/01 of 25 - Video Title.mp4"

# Organize by playlist
ytdl-go -o "Playlists/{playlist_title}/{title}.{ext}" [PLAYLIST_URL]
# Output: "Playlists/My Playlist/Video Title.mp4"
```

### Date-Based Organization

```bash
# By year
ytdl-go -o "Videos/{release_year}/{title}.{ext}" [URL]
# Output: "Videos/2023/Video Title.mp4"

# By full date
ytdl-go -o "Videos/{release_date}/{title}.{ext}" [URL]
# Output: "Videos/2023-01-15/Video Title.mp4"

# Year/Month folders (requires manual extraction)
ytdl-go -o "Videos/{release_year}/{title}.{ext}" [URL]
```

### Quality-Based Organization

```bash
# Separate by quality
ytdl-go -o "Videos/{quality}/{title}.{ext}" [URL]
# Output: "Videos/1080p/Video Title.mp4"

# Quality in filename
ytdl-go -o "Videos/{title} [{quality}].{ext}" [URL]
# Output: "Videos/Video Title [720p].mp4"
```

### Advanced Templates

```bash
# Complete organization
ytdl-go -o "{artist}/{album}/{index} - {title} [{quality}].{ext}" [URL]
# Output: "Artist/Album/01 - Song [1080p].mp4"

# Podcast structure
ytdl-go -audio -o "Podcasts/{author}/{release_date} - {title}.{ext}" [URL]
# Output: "Podcasts/Podcast Name/2023-01-15 - Episode Title.m4a"

# Technical details included
ytdl-go -o "Videos/{title} [{quality}] [{itag}].{ext}" [URL]
# Output: "Videos/Video Title [720p] [22].mp4"
```

## Best Practices

### Use Relative Paths

Always use relative paths for portability and security:

**Good:**
```bash
ytdl-go -o "downloads/{title}.{ext}" [URL]
ytdl-go -o "{artist}/{title}.{ext}" [URL]
```

**Bad:**
```bash
ytdl-go -o "/home/user/downloads/{title}.{ext}" [URL]  # Absolute path
```

### Handle Missing Data

Plan for missing metadata by using safe defaults:

```bash
# Album might be missing - provide context
ytdl-go -o "Music/{artist}/{album}/{title}.{ext}" [URL]
# Falls back to: "Music/Artist Name/unknown/Song Title.m4a"

# Or use flat structure when metadata uncertain
ytdl-go -o "Music/{artist} - {title}.{ext}" [URL]
```

### Avoid Special Characters

Let sanitization handle special characters automatically:

```bash
# Template with special chars in placeholders is OK
ytdl-go -o "{artist}/{title}.{ext}" [URL]

# Don't add problematic chars in template literal parts
# Bad: ytdl-go -o "{artist}: {title}.{ext}" [URL]  # Colon problematic on Windows
# Good: ytdl-go -o "{artist} - {title}.{ext}" [URL]
```

### Use Meaningful Structures

Create intuitive folder hierarchies:

```bash
# Good: Clear hierarchy
ytdl-go -o "Music/{artist}/{album}/{title}.{ext}" [URL]

# Bad: Unclear structure
ytdl-go -o "{album}/{ext}/{artist}/{title}.{ext}" [URL]
```

### Pad Numbers for Sorting

Use index for proper sorting in file managers:

```bash
# Good: Natural sorting
ytdl-go -o "{playlist_title}/{index} - {title}.{ext}" [PLAYLIST_URL]
# Results in: 1, 2, 3... 10, 11, 12...

# Note: Index is not zero-padded automatically
# Without padding, files sort as: 1, 10, 11, 2, 20, 3...
# Most file managers use "natural sort" which handles this correctly
# However, tools like `ls` may sort incorrectly without additional flags
```

## Template Testing

Test templates with `-info` flag to preview filenames without downloading:

```bash
# Check what filename will be generated
ytdl-go -info [URL] | jq -r .title
# Then use in template:
ytdl-go -o "{title}.{ext}" [URL]
```

## Metadata Override

You can override placeholder values using the `-meta` flag:

```bash
# Override title
ytdl-go -meta title="Custom Title" -o "{title}.{ext}" [URL]
# Output: "Custom Title.mp4"

# Override multiple fields
ytdl-go \
  -meta artist="Custom Artist" \
  -meta album="Custom Album" \
  -o "Music/{artist}/{album}/{title}.{ext}" \
  [URL]
# Output: "Music/Custom Artist/Custom Album/Video Title.mp4"
```

See the [CLI Options - Metadata Flags](cli-options#metadata-flags) section for complete metadata override documentation.

## Directory Creation

ytdl-go automatically creates directories specified in templates:

```bash
# Creates nested folders as needed
ytdl-go -o "a/b/c/d/{title}.{ext}" [URL]
# Creates: ./a/b/c/d/Video Title.mp4
```

### Output Directory Constraint

Use `-output-dir` to constrain all outputs to a base directory:

```bash
# All paths relative to /safe/downloads
ytdl-go -output-dir /safe/downloads -o "{artist}/{title}.{ext}" [URL]
# Output: /safe/downloads/Artist/Video Title.mp4

# Prevents directory traversal
ytdl-go -output-dir /safe/downloads -o "../../etc/passwd" [URL]
# Error: Path escapes output directory
```

## Troubleshooting

### Long Filenames

Some filesystems have filename length limits (typically 255 characters):

**Solution:** Use shorter templates or fewer placeholders:

```bash
# Too long
ytdl-go -o "{playlist_title}/{artist}/{album}/{index} - {title} [{quality}] ({id}).{ext}"

# Better
ytdl-go -o "{artist}/{album}/{index} - {title}.{ext}"
```

### Invalid Characters in Output

If you encounter filesystem errors:

1. Check for problematic characters in template literal parts
2. Rely on automatic sanitization for placeholders
3. Avoid colons, slashes, and other special characters in literal text

### Placeholder Not Replaced

If a placeholder appears literally in the filename:

1. Check spelling and case sensitivity
2. Verify placeholder is supported (see reference table)
3. Use `-info` to check available metadata

```bash
# Debug: Check available metadata
ytdl-go -info [URL] | jq .
```

---

## Related Documentation

- [CLI Options](cli-options) - Complete flag reference
- [Basic Downloads](../user-guide/usage/basic-downloads) - Getting started guide
- [Advanced Usage](../user-guide/usage/advanced-usage) - Complex download scenarios
- [Troubleshooting](../user-guide/troubleshooting) - Common issues and solutions
