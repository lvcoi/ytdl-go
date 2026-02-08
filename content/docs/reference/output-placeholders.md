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

### Album Placeholders

| Placeholder | Description | Example Value | Notes |
|-------------|-------------|---------------|-------|
| `{album}` | Album name (sanitized) | `Album Name` | From YouTube Music metadata |

### Playlist Placeholders

| Placeholder | Description | Example Value | Notes |
|-------------|-------------|---------------|-------|
| `{playlist_title}` | Playlist name (sanitized) | `My Playlist` | Underscore or hyphen variants |
| `{playlist-title}` | Playlist name (sanitized) | `My Playlist` | Alternative syntax |
| `{playlist_id}` | Playlist ID | `PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf` | Unique playlist identifier |
| `{playlist-id}` | Playlist ID | `PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf` | Alternative syntax |
| `{index}` | Video position in playlist | `1`, `2`, `3` | 1-based indexing |
| `{count}` | Total videos in playlist | `25` | Total playlist size |

## Placeholder Behavior

### Sanitization

All placeholders containing text are automatically sanitized for filesystem compatibility. Invalid filename characters are replaced with a dash (`-`) to ensure the resulting path is safe on common filesystems.

**Characters Removed/Replaced:**
- `/` → `-` (path separator)
- `\` → `-` (Windows path separator)
- `:` → `-` (Windows drive separator)
- `*` → `-` (wildcard)
- `?` → `-` (wildcard)
- `"` → `-` (quotes)
- `<` → `-` (redirects)
- `>` → `-` (redirects)
- `|` → `-` (pipe)

**Example:**

```bash
# Original title: "Video: Part 1/2 (HD)"
# Sanitized: "Video- Part 1-2 (HD)"
```

### Missing Values

If a placeholder's value is not available, it is replaced with a safe default:

- Text placeholders → empty string (for optional metadata like `{album}`)
- Numeric placeholders → empty string
- Required placeholders like `{title}`, `{id}`, `{ext}` → always available

**Example:**

```bash
# No album metadata available
ytdl-go -o "{artist}/{album}/{title}.{ext}" [URL]
# Results in: "Artist Name/Video Title.mp4"
# (empty album segment is removed during path cleaning)
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

# Include track numbers (from metadata if available)
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" [URL]
# Output: "Music/Artist Name/Album Name/Song Title.m4a"

# Flat structure with artist
ytdl-go -audio -o "Music/{artist} - {title}.{ext}" [URL]
# Output: "Music/Artist Name - Song Title.m4a"
```

### Playlist Downloads

```bash
# Numbered playlist items
ytdl-go -o "{playlist_title}/{index} - {title}.{ext}" [PLAYLIST_URL]
# Output: "My Playlist/1 - Video Title.mp4"

# Include total count
ytdl-go -o "{playlist_title}/{index} of {count} - {title}.{ext}" [PLAYLIST_URL]
# Output: "My Playlist/1 of 25 - Video Title.mp4"

# Organize by playlist
ytdl-go -o "Playlists/{playlist_title}/{title}.{ext}" [PLAYLIST_URL]
# Output: "Playlists/My Playlist/Video Title.mp4"
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

**Note:** Release date and release year placeholders like `{release_date}` and `{release_year}` are not currently supported.

### Advanced Templates

```bash
# Complete organization
ytdl-go -o "{artist}/{album}/{index} - {title} [{quality}].{ext}" [URL]
# Output: "Artist/Album/1 - Song [1080p].mp4"

# Playlist with quality
ytdl-go -o "Playlists/{playlist-title}/{index} - {title}.{ext}" [URL]
# Output: "Playlists/My Playlist/1 - Video Title.mp4"
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

Understand how missing metadata is handled so you can choose appropriate templates:

```bash
# Album might be missing - the {album} placeholder becomes empty
ytdl-go -o "Music/{artist}/{album}/{title}.{ext}" [URL]
# If album metadata is missing:
#   Template expands to: "Music/Artist Name//Song Title.m4a"
#   After filepath.Clean(): "Music/Artist Name/Song Title.m4a"

# Or use a flat structure when metadata is uncertain
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
# Index values generated: 1, 2, 3... 10, 11, 12...
ytdl-go -o "{playlist_title}/{index} - {title}.{ext}" [PLAYLIST_URL]

# Note: Index is not zero-padded automatically
# Without padding, files sort lexicographically as: 1, 10, 11, 2, 20, 3...
# Most file managers use "natural sort" which handles this correctly
# However, tools like `ls` may sort incorrectly without additional flags
```

Test templates with `-info` flag to preview filenames without downloading:

```bash
# Check what filename will be generated
ytdl-go -info [URL] | jq -r .title
# Then use in template:
ytdl-go -o "{title}.{ext}" [URL]
```

## Metadata Override

The `-meta` flag overrides metadata embedded in the output file (e.g., ID3 tags), but it does **not** affect template placeholder expansion. Template placeholders always use the video's original metadata from YouTube.

```bash
# The -meta flag sets the ID3 tags in the file, NOT the template placeholders
ytdl-go -audio \
  -meta artist="Custom Artist" \
  -meta album="Custom Album" \
  -o "Music/{artist}/{album}/{title}.{ext}" \
  [URL]
# Template uses original YouTube metadata for the path
# ID3 tags in the file will contain "Custom Artist" and "Custom Album"
```

See the [Output Templates](../user-guide/usage/output-templates#working-with-metadata-overrides) guide for details.

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
- [Troubleshooting](../user-guide/troubleshooting) - Common issues and solutions
