---
title: "Output Templates"
weight: 20
---

# Output Templates

This guide covers ytdl-go's powerful output template system for customizing file names and directory structures.

## Overview

The `-o` flag accepts a template string with placeholders that are dynamically replaced with video metadata. This allows you to organize downloads exactly how you want.

**Default Template**: `{title}.{ext}`

```bash
# Basic usage
ytdl-go -o "template-string" URL
```

> **Relative Paths Only**
>
> Output paths must be relative for security reasons. Absolute paths are rejected.

## Available Placeholders

### Video Placeholders

| Placeholder | Description | Example | Notes |
|-------------|-------------|---------|-------|
| `{title}` | Video title (sanitized) | `My Video Title` | Special characters removed |
| `{id}` | YouTube video ID | `dQw4w9WgXcQ` | Unique identifier |
| `{ext}` | File extension | `mp4`, `webm`, `m4a` | Determined by format |
| `{quality}` | Quality label or bitrate | `1080p`, `720p`, `128k` | Video resolution or audio bitrate |

### Creator Placeholders

| Placeholder | Description | Example | Notes |
|-------------|-------------|---------|-------|
| `{artist}` | Video author/artist | `Artist Name` | Sanitized for filesystem |
| `{album}` | Album name | `Album Name` | From YouTube Music metadata |

### Playlist Placeholders

| Placeholder | Description | Example | Notes |
|-------------|-------------|---------|-------|
| `{playlist-title}` | Playlist name | `My Playlist` | Sanitized for filesystem |
| `{playlist_title}` | Playlist name (alt) | `My Playlist` | Same as playlist-title |
| `{playlist-id}` | Playlist ID | `PLxxxxxxxxxxxxxxxxxxx` | Unique identifier |
| `{playlist_id}` | Playlist ID (alt) | `PLxxxxxxxxxxxxxxxxxxx` | Same as playlist-id |
| `{index}` | Video position (1-based) | `1`, `2`, `3` | Position in playlist |
| `{count}` | Total videos in playlist | `25` | Total number of videos |

> **Playlist Context**
>
> Playlist placeholders are only available when downloading from a playlist URL. They remain empty for single video downloads.

## Path Behavior

### Directory Creation

Directories in the template are automatically created:

```bash
# Creates Music/Artist/Album/ if it doesn't exist
ytdl-go -o "Music/{artist}/{album}/{title}.{ext}" URL
```

### Relative Path Resolution

- Paths are resolved against `-output-dir` if set
- Otherwise resolved against current working directory
- Parent directory references (`..`) are validated for security

### Trailing Slash

A trailing slash forces directory interpretation:

```bash
# Creates directory named after title, file has default name
ytdl-go -o "{title}/" URL
```

## Template Examples

### Simple Filename Templates

```bash
# Just the title
ytdl-go -o "{title}.{ext}" URL

# Title with video ID
ytdl-go -o "{title}-{id}.{ext}" URL

# Title with quality indicator
ytdl-go -o "{title} [{quality}].{ext}" URL

# Artist and title
ytdl-go -o "{artist} - {title}.{ext}" URL
```

### Single Directory Organization

```bash
# Videos folder
ytdl-go -o "Videos/{title}.{ext}" URL

# Downloads with quality
ytdl-go -o "Downloads/{title} [{quality}].{ext}" URL

# By artist
ytdl-go -o "Videos/{artist} - {title}.{ext}" URL
```

### Multi-Level Directory Structure

```bash
# Organize by quality
ytdl-go -o "Videos/{quality}/{title}.{ext}" URL

# Organize by artist and album
ytdl-go -o "Music/{artist}/{album}/{title}.{ext}" URL

# Complex organization
ytdl-go -o "Content/{artist}/{quality}/{title}-{id}.{ext}" URL
```

### Playlist Templates

```bash
# Basic playlist organization
ytdl-go -o "{playlist-title}/{title}.{ext}" URL

# Numbered files
ytdl-go -o "{playlist-title}/{index} - {title}.{ext}" URL

# Complete playlist info
ytdl-go -o "{playlist-title}/{index} of {count} - {title}.{ext}" URL

# Playlist with quality
ytdl-go -o "{playlist-title}/{index} - {title} [{quality}].{ext}" URL
```

### Audio/Music Templates

```bash
# Artist - Title
ytdl-go -audio -o "Music/{artist} - {title}.{ext}" URL

# Artist/Album structure
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" URL

# With track numbers (if metadata available)
ytdl-go -audio -o "Music/{artist}/{album}/{title}.{ext}" URL

# Music playlist
ytdl-go -audio -o "Music/{artist}/{playlist-title}/{index} - {title}.{ext}" URL
```

## Number Formatting

### Index Numbers

For playlists, use the `{index}` placeholder:

```bash
# Simple index: 1, 2, 3, ...
ytdl-go -o "{playlist-title}/{index} - {title}.{ext}" URL
```

Indexes are not zero-padded. Most modern file managers use "natural sort" which correctly handles numeric sorting.

## Real-World Use Cases

### Video Course Archive

```bash
ytdl-go -quality 720p \
        -o "Courses/{playlist-title}/Lesson {index} - {title}.{ext}" \
        PLAYLIST_URL
```

Result:
```
Courses/
└── Python Programming/
    ├── Lesson 1 - Introduction.mp4
    ├── Lesson 2 - Variables.mp4
    └── Lesson 3 - Functions.mp4
```

### Music Album Collection

```bash
ytdl-go -audio \
        -o "Music/{artist}/{album}/{index} - {title}.{ext}" \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        PLAYLIST_URL
```

Result:
```
Music/
└── Artist Name/
    └── Album Name/
        ├── 1 - Song One.m4a
        ├── 2 - Song Two.m4a
        └── 3 - Song Three.m4a
```

### Podcast Episodes

```bash
ytdl-go -audio -quality 96k \
        -o "Podcasts/{playlist-title}/Episode {index} - {title}.{ext}" \
        PLAYLIST_URL
```

Result:
```
Podcasts/
└── My Podcast/
    ├── Episode 1 - Pilot Episode.m4a
    ├── Episode 2 - Guest Interview.m4a
    └── Episode 3 - Deep Dive.m4a
```

### Documentary Series

```bash
ytdl-go -quality 1080p -format mp4 \
        -o "Documentaries/{playlist-title}/Part {index} - {title}.{ext}" \
        PLAYLIST_URL
```

Result:
```
Documentaries/
└── Nature Series/
    ├── Part 1 - Oceans.mp4
    ├── Part 2 - Forests.mp4
    └── Part 3 - Mountains.mp4
```

### DJ Sets Archive

```bash
ytdl-go -audio -quality best \
        -o "DJ Sets/{artist}/{title}.{ext}" \
        URL
```

Result:
```
DJ Sets/
└── DJ Name/
    ├── Live Set Volume 1.opus
    └── Studio Mix.opus
```

### Video Quality Variants

```bash
ytdl-go -o "Videos/{quality}/{artist} - {title}.{ext}" URL
```

Result:
```
Videos/
├── 1080p/
│   ├── Artist - Video One.mp4
│   └── Artist - Video Two.mp4
└── 720p/
    └── Artist - Video Three.mp4
```

### Research Archive

```bash
ytdl-go -o "Research/{artist}/{title} [{id}].{ext}" URL
```

Result:
```
Research/
└── Channel Name/
    ├── Research Paper [abc123].mp4
    └── Follow-up Study [xyz789].mp4
```

## Advanced Templates

### Including Video ID for Uniqueness

```bash
ytdl-go -o "{artist}/{title}-{id}.{ext}" URL
```

This ensures unique filenames even if titles are identical.

### Quality-Based Organization

```bash
ytdl-go -o "{quality}/{playlist-title}/{index} - {title}.{ext}" URL
```

Separates downloads by quality level.

### Date-Based Archival

```bash
ytdl-go -o "Archive/{artist} - {title}.{ext}" URL
```

Organizes by artist.

### Combined Information

```bash
ytdl-go -o "{artist}/{album}/{index} - {title} [{quality}] {id}.{ext}" URL
```

Includes multiple metadata fields for maximum information.

## Working with Metadata Overrides

The `-meta` flag allows overriding metadata that gets embedded in the output file (e.g., ID3 tags), but it does NOT affect template placeholders. Placeholders are always populated from the video's original metadata.

```bash
# This overrides embedded metadata, but template uses original artist
ytdl-go -audio \
        -o "Music/{artist}/{album}/{title}.{ext}" \
        -meta artist="Custom Artist" \
        -meta album="Custom Album" \
        URL
```

The template will expand using the video's original artist/album, while the embedded ID3 tags will contain your custom values.

## Security Considerations

### Output Directory Constraint

Use `-output-dir` to restrict all downloads to a base directory:

```bash
ytdl-go -output-dir /safe/downloads \
        -o "{artist}/{album}/{title}.{ext}" \
        URL
```

This prevents directory traversal attacks.

### Path Validation

ytdl-go automatically:
- Sanitizes special characters in placeholders
- Rejects absolute paths
- Validates paths against output directory constraints
- Prevents directory traversal (`../`)

## Troubleshooting Templates

### Missing Directories

**Problem**: Directories not created

**Solution**: ytdl-go creates directories automatically. Ensure you have write permissions.

```bash
# Check permissions
ls -la parent-directory/
```

### Empty Placeholders

**Problem**: Placeholder shows as empty or `[unknown]`

**Causes**:
- Metadata not available (e.g., `{album}` for non-music content)
- Playlist placeholder used for single video
- Information not in video metadata

**Solutions**:
```bash
# Check available metadata
ytdl-go -info URL | jq .

# Use different placeholders or templates
ytdl-go -o "{artist}/{title}.{ext}" URL
```

### Special Characters in Filenames

**Problem**: Unexpected characters in output files

**Behavior**: ytdl-go automatically sanitizes:
- Removes: `/`, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|`
- Replaces: spaces, special characters

**Solution**: Use the sanitized output or manually override with `-meta`.

### Path Too Long

**Problem**: Path exceeds filesystem limits

**Solutions**:
```bash
# Use shorter templates
ytdl-go -o "{artist}/{title}.{ext}" URL

# Use ID instead of full title
ytdl-go -o "{artist}/{id}.{ext}" URL
```

## Best Practices

1. **Keep templates readable**: Balance detail with simplicity
2. **Use consistent structure**: Maintain same pattern across downloads
3. **Include unique identifiers**: Add `{id}` to prevent overwrites
4. **Test templates**: Try on single video before batch downloads
5. **Consider metadata**: Not all videos have all metadata fields
6. **Plan for scale**: Design templates that work for large playlists
7. **Document your templates**: Keep notes on organizational schemes

## Template Cheat Sheet

### Video Downloads
```bash
# Basic
"{title}.{ext}"

# With quality
"{title} [{quality}].{ext}"

# Organized
"Videos/{artist}/{title}.{ext}"
```

### Audio Downloads
```bash
# Simple
"{artist} - {title}.{ext}"

# Album structure
"Music/{artist}/{album}/{title}.{ext}"
```

### Playlists
```bash
# Numbered
"{playlist-title}/{index} - {title}.{ext}"

# Complete info
"{playlist-title}/{index} of {count} - {title}.{ext}"

# With quality
"{playlist-title}/{index} - {title} [{quality}].{ext}"
```

### Archives
```bash
# By artist
"Archive/{artist}/{title}.{ext}"

# Complete organization
"Archive/{artist}/{title}-{id}.{ext}"

# Quality variants
"{quality}/{playlist-title}/{index} - {title}.{ext}"
```

## Examples by Scenario

### Student/Learning
```bash
ytdl-go -o "Courses/{playlist-title}/Lesson {index} - {title}.{ext}" URL
```

### Music Collector
```bash
ytdl-go -audio -o "Music/{artist}/{album}/{index} - {title}.{ext}" URL
```

### Podcast Listener
```bash
ytdl-go -audio -o "Podcasts/{playlist-title}/Episode {index} - {title}.{ext}" URL
```

### Content Creator
```bash
ytdl-go -quality 1080p -o "Reference/{artist}/{title} [{id}].{ext}" URL
```

### Archivist
```bash
ytdl-go -o "Archive/{artist} - {title} [{quality}].{ext}" URL
```

## Next Steps

- **Metadata**: Learn about [metadata and sidecars](metadata-sidecars)
- **Playlists**: Apply templates to [playlist downloads](playlists)
- **Audio**: Organize [audio downloads](audio-only)
- **Format Selection**: Choose formats with [format selection](format-selection)

## Related References

- [Basic Downloads Guide](basic-downloads)
- [Playlist Downloads Guide](playlists)
- [Metadata and Sidecars Guide](metadata-sidecars)
- [Command-Line Flags Reference](../../reference/cli-options)
