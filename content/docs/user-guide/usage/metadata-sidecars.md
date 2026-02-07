---
title: "Metadata and Sidecars"
weight: 30
---

# Metadata and Sidecars

This guide covers ytdl-go's metadata handling, including automatic embedding, sidecar files, JSON output, and metadata customization.

## Overview

ytdl-go automatically extracts video metadata from YouTube and provides multiple ways to use it:

1. **Embedded Metadata** - ID3 tags in audio files
2. **Sidecar JSON Files** - Complete metadata stored alongside downloads
3. **Template Placeholders** - Metadata used in filenames
4. **Custom Overrides** - Manual metadata specification

## Automatic Metadata Extraction

ytdl-go automatically extracts metadata from every video:

```bash
# Download with automatic metadata
ytdl-go URL
```

### Extracted Metadata Fields

| Field | Description | Example |
|-------|-------------|---------|
| `id` | Video ID | `dQw4w9WgXcQ` |
| `title` | Video title | `My Video Title` |
| `description` | Full description | `This video is about...` |
| `author` | Channel/uploader name | `Channel Name` |
| `artist` | Artist/creator name | `Artist Name` |
| `duration` | Length in seconds | `180` |
| `upload_date` | Upload date | `20240115` |
| `release_date` | Release date | `2024-01-15` |
| `release_year` | Release year | `2024` |
| `thumbnail` | Thumbnail URL | `https://i.ytimg.com/...` |
| `view_count` | Number of views | `1000000` |
| `like_count` | Number of likes | `50000` |

### Playlist-Specific Metadata

When downloading from playlists:

| Field | Description | Example |
|-------|-------------|---------|
| `playlist_title` | Playlist name | `My Playlist` |
| `playlist_id` | Playlist ID | `PLxxxxxxxxxxxxxxxxxxx` |
| `playlist_index` | Position in playlist | `5` |
| `playlist_count` | Total videos | `25` |

## Embedded Metadata (ID3 Tags)

For audio downloads, ytdl-go automatically embeds ID3 tags:

```bash
# Download with embedded metadata
ytdl-go -audio URL
```

### Embedded Fields

Audio files include:

- **Title** (`TIT2`) - Video title
- **Artist** (`TPE1`) - Channel/artist name
- **Album** (`TALB`) - Playlist name (if from playlist)
- **Track** (`TRCK`) - Position in playlist
- **Date** (`TYER`) - Release/upload year
- **URL** (`WXXX`) - Source URL

> **Audio Formats Only**
>
> ID3 tags are embedded in audio formats (M4A, MP3, Opus). Video files use sidecar JSON files instead.

### Verifying Embedded Metadata

Use media players or command-line tools to verify:

```bash
# Using ffprobe
ffprobe -show_format "audio.m4a" 2>&1 | grep TAG:

# Using exiftool (if installed)
exiftool "audio.m4a"

# Using mediainfo (if installed)
mediainfo "audio.m4a"
```

## Sidecar JSON Files

ytdl-go can save complete metadata in JSON sidecar files alongside downloads.

### Generating Sidecar Files

Sidecar files are automatically created during downloads:

```bash
# Download creates video.mp4 and video.info.json
ytdl-go URL
```

File structure:
```
downloads/
├── My Video.mp4
└── My Video.info.json
```

### Sidecar File Format

```json
{
  "id": "dQw4w9WgXcQ",
  "title": "My Video Title",
  "description": "Full video description...",
  "author": "Channel Name",
  "artist": "Artist Name",
  "duration_seconds": 180,
  "upload_date": "20240115",
  "release_date": "2024-01-15",
  "release_year": "2024",
  "thumbnail": "https://i.ytimg.com/...",
  "view_count": 1000000,
  "like_count": 50000,
  "formats": [...],
  "quality": "1080p",
  "container": "mp4"
}
```

### Using Sidecar Files

Parse sidecar files with `jq` or other JSON tools:

```bash
# Extract title
jq -r '.title' "My Video.info.json"

# Get duration
jq -r '.duration_seconds' "My Video.info.json"

# List all fields
jq 'keys' "My Video.info.json"

# Format upload date
jq -r '.upload_date' "My Video.info.json" | sed 's/\(....\)\(..\)\(..\)/\1-\2-\3/'
```

### Batch Processing Sidecars

```bash
# Find all sidecar files
find . -name "*.info.json"

# Extract all titles
find . -name "*.info.json" -exec jq -r '.title' {} \;

# Create CSV of metadata
find . -name "*.info.json" -exec jq -r '[.title, .author, .duration_seconds] | @csv' {} \;
```

## Metadata-Only Mode

Get metadata without downloading files:

```bash
ytdl-go -info URL
```

This outputs JSON metadata to stdout without creating any files.

### Metadata-Only Output

```json
{
  "id": "dQw4w9WgXcQ",
  "title": "My Video Title",
  "description": "Full description...",
  "author": "Channel Name",
  "duration_seconds": 180,
  "upload_date": "20240115"
}
```

### Metadata-Only Use Cases

**Check video information:**
```bash
ytdl-go -info URL | jq .
```

**Extract specific field:**
```bash
# Get title
ytdl-go -info URL | jq -r '.title'

# Get duration
ytdl-go -info URL | jq -r '.duration_seconds'

# Get author
ytdl-go -info URL | jq -r '.author'
```

**Batch metadata collection:**
```bash
# Get metadata for multiple videos
for url in $(cat urls.txt); do
  ytdl-go -info "$url" | jq -r '{title, author, duration_seconds}'
done
```

**Decision making:**
```bash
# Check duration before downloading
duration=$(ytdl-go -info URL | jq -r '.duration_seconds')
if [ "$duration" -lt 600 ]; then
  ytdl-go URL
fi
```

## Custom Metadata Overrides

Override metadata fields with the `-meta` flag:

```bash
ytdl-go -meta artist="Artist Name" -meta title="Song Title" URL
```

### Available Override Fields

| Field | Description | Used In |
|-------|-------------|---------|
| `title` | Track/video title | Templates, ID3, JSON |
| `artist` | Artist/creator name | Templates, ID3, JSON |
| `author` | Uploader/channel | Templates, ID3, JSON |
| `album` | Album name | Templates, ID3, JSON |
| `track` | Track number | ID3, JSON |
| `disc` | Disc number | ID3, JSON |
| `release_date` | Release date (YYYY-MM-DD) | Templates, ID3, JSON |
| `release_year` | Release year (YYYY) | Templates, ID3, JSON |
| `source_url` | Source URL | ID3, JSON |

### Multiple Overrides

Specify multiple metadata fields:

```bash
ytdl-go -audio \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        -meta title="Track Title" \
        -meta track=5 \
        -meta release_year=2024 \
        URL
```

### Metadata Override Examples

**Music download with full metadata:**
```bash
ytdl-go -audio \
        -o "Music/{artist}/{album}/{track:02d} - {title}.{ext}" \
        -meta artist="Pink Floyd" \
        -meta album="The Dark Side of the Moon" \
        -meta title="Time" \
        -meta track=4 \
        -meta release_year=1973 \
        URL
```

**Podcast episode:**
```bash
ytdl-go -audio \
        -meta artist="Podcast Name" \
        -meta album="Season 1" \
        -meta title="Episode 5: Topic" \
        -meta track=5 \
        -meta release_date="2024-01-15" \
        URL
```

**Audiobook chapter:**
```bash
ytdl-go -audio \
        -meta artist="Author Name" \
        -meta album="Book Title" \
        -meta title="Chapter 3: The Beginning" \
        -meta track=3 \
        URL
```

## Metadata in Output Templates

Use metadata in filename templates:

```bash
# Artist and title
ytdl-go -o "{artist} - {title}.{ext}" URL

# With album
ytdl-go -o "{artist}/{album}/{title}.{ext}" URL

# With track number
ytdl-go -o "{artist}/{album}/{track:02d} - {title}.{ext}" URL

# With release year
ytdl-go -o "{artist}/{release_year} - {album}/{title}.{ext}" URL
```

> **Template Placeholders**
>
> For complete template documentation, see [Output Templates](output-templates).

## Playlist Metadata Workflows

### Basic Playlist Metadata

```bash
# Playlist with default metadata
ytdl-go -audio -o "{playlist-title}/{index} - {title}.{ext}" PLAYLIST_URL
```

Result includes:
- Playlist title in directory name
- Track numbers from playlist index
- Individual video titles

### Playlist with Album Metadata

```bash
# Set album name for entire playlist
ytdl-go -audio \
        -o "Music/{artist}/{album}/{index} - {title}.{ext}" \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        PLAYLIST_URL
```

This sets consistent artist and album for all tracks.

### Individual Track Overrides

For individual tracks in a playlist, download separately with metadata:

```bash
# Track 1
ytdl-go -audio \
        -meta artist="Artist" \
        -meta album="Album" \
        -meta track=1 \
        -meta title="Track 1" \
        URL1

# Track 2
ytdl-go -audio \
        -meta artist="Artist" \
        -meta album="Album" \
        -meta track=2 \
        -meta title="Track 2" \
        URL2
```

## JSON Output Mode

For machine-readable output, use JSON mode:

```bash
ytdl-go -json URL
```

### JSON Output Format

Each line is a JSON object:

```json
{"type":"item","status":"ok","url":"https://...","output":"video.mp4","bytes":4194304,"retries":false}
{"type":"item","status":"error","url":"https://...","error":"connection timeout"}
```

### Event Types

| Type | Description | Fields |
|------|-------------|--------|
| `item` | Download completion | `status`, `url`, `output`, `bytes`, `retries` |
| `error` | Error occurrence | `url`, `category`, `error` |
| `formats` | Available formats | `formats` (array) |

### Parsing JSON Output

```bash
# Get successful downloads
ytdl-go -json URL1 URL2 URL3 | jq -r 'select(.type=="item" and .status=="ok") | .output'

# Get failed downloads
ytdl-go -json URL1 URL2 URL3 | jq -r 'select(.type=="item" and .status=="error") | .url'

# Calculate total bytes
ytdl-go -json URL1 URL2 URL3 | jq -r 'select(.type=="item") | .bytes' | awk '{sum+=$1} END {print sum}'

# Extract all output filenames
ytdl-go -json URL1 URL2 URL3 | jq -r 'select(.type=="item") | .output'
```

### JSON Mode for Monitoring

Monitor batch downloads:

```bash
# Log to file
ytdl-go -json -o "{title}.{ext}" URL1 URL2 URL3 > download.log

# Real-time monitoring
ytdl-go -json -o "{title}.{ext}" URL1 URL2 URL3 | \
  jq -r 'select(.type=="item") | "\(.status): \(.output)"'

# Progress tracking
ytdl-go -json PLAYLIST_URL | \
  jq -r 'select(.type=="item") | "Downloaded: \(.output) (\(.bytes) bytes)"'
```

## Metadata Best Practices

1. **Use automatic metadata** when possible for accuracy
2. **Override metadata** for music collections and albums
3. **Include metadata in templates** for organized file naming
4. **Save sidecar files** for future reference
5. **Verify embedded tags** in audio files after download
6. **Use JSON mode** for automated workflows
7. **Document metadata schemes** for consistency
8. **Test metadata** on single file before batch operations

## Common Metadata Workflows

### Music Library with Full Metadata

```bash
ytdl-go -audio -quality 256k \
        -o "Music/{artist}/{album}/{track:02d} - {title}.{ext}" \
        -meta artist="Artist Name" \
        -meta album="Album Name" \
        -meta release_year=2024 \
        PLAYLIST_URL
```

### Podcast Archive

```bash
ytdl-go -audio -quality 96k \
        -o "Podcasts/{artist}/Season {season}/{episode:02d} - {title}.{ext}" \
        -meta artist="Podcast Name" \
        -meta album="Season 1" \
        PLAYLIST_URL
```

### Video Archive with Metadata

```bash
ytdl-go -quality 1080p \
        -o "Archive/{author}/{release_date} - {title} [{id}].{ext}" \
        URL
```

Each file gets:
- Author/channel in directory
- Release date in filename
- Video ID for uniqueness
- Sidecar JSON with full metadata

### Research Collection

```bash
ytdl-go -quality 720p \
        -o "Research/{category}/{author} - {title}.{ext}" \
        -meta category="Topic Area" \
        URL
```

## Troubleshooting Metadata

### Missing Metadata Fields

**Problem**: Expected metadata field is empty

**Causes**:
- Field not available in video metadata
- Video doesn't have that information
- Playlist-specific field used for single video

**Solutions**:
```bash
# Check available metadata
ytdl-go -info URL | jq .

# Override missing fields
ytdl-go -meta artist="Artist Name" URL

# Use different placeholder in template
ytdl-go -o "{author} - {title}.{ext}" URL  # Use author instead of artist
```

### Incorrect Embedded Tags

**Problem**: Audio file has wrong ID3 tags

**Solutions**:
```bash
# Override metadata during download
ytdl-go -audio \
        -meta artist="Correct Artist" \
        -meta title="Correct Title" \
        URL

# Or manually edit tags after download with tools like:
# - kid3 (GUI)
# - id3v2 (CLI)
# - ffmpeg
```

### Sidecar Files Not Created

**Problem**: No .info.json files generated

**Current Behavior**: Sidecar files are created automatically when supported.

**Verify**:
```bash
# Check if file exists
ls -la *.info.json

# If missing, metadata is in embedded tags (audio) or needs to be extracted
```

### Special Characters in Metadata

**Problem**: Metadata contains characters that break filenames

**Solution**: ytdl-go automatically sanitizes metadata used in templates:
```bash
# Automatically sanitized
ytdl-go -o "{title}.{ext}" URL
```

Special characters (`/`, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|`) are removed or replaced.

## Advanced Metadata Techniques

### Metadata Extraction Script

```bash
#!/bin/bash
# Extract metadata from URLs

urls_file="$1"
output_csv="metadata.csv"

echo "url,title,author,duration,upload_date" > "$output_csv"

while read -r url; do
  meta=$(ytdl-go -info "$url")
  title=$(echo "$meta" | jq -r '.title')
  author=$(echo "$meta" | jq -r '.author')
  duration=$(echo "$meta" | jq -r '.duration_seconds')
  date=$(echo "$meta" | jq -r '.upload_date')
  
  echo "$url,\"$title\",\"$author\",$duration,$date" >> "$output_csv"
done < "$urls_file"
```

### Conditional Downloads Based on Metadata

```bash
#!/bin/bash
# Download only short videos

url="$1"
max_duration=600  # 10 minutes

duration=$(ytdl-go -info "$url" | jq -r '.duration_seconds')

if [ "$duration" -lt "$max_duration" ]; then
  echo "Downloading (${duration}s)"
  ytdl-go "$url"
else
  echo "Skipping (${duration}s exceeds ${max_duration}s)"
fi
```

### Metadata Database Creation

```bash
#!/bin/bash
# Build SQLite database of video metadata

db="videos.db"

sqlite3 "$db" << EOF
CREATE TABLE IF NOT EXISTS videos (
  id TEXT PRIMARY KEY,
  title TEXT,
  author TEXT,
  duration INTEGER,
  upload_date TEXT,
  url TEXT
);
EOF

while read -r url; do
  meta=$(ytdl-go -info "$url")
  
  id=$(echo "$meta" | jq -r '.id')
  title=$(echo "$meta" | jq -r '.title')
  author=$(echo "$meta" | jq -r '.author')
  duration=$(echo "$meta" | jq -r '.duration_seconds')
  date=$(echo "$meta" | jq -r '.upload_date')
  
  sqlite3 "$db" << EOF
INSERT OR REPLACE INTO videos VALUES (
  '$id', '$title', '$author', $duration, '$date', '$url'
);
EOF
done < urls.txt
```

## Next Steps

- **Templates**: Use metadata in [output templates](output-templates)
- **Audio**: Apply metadata to [audio downloads](audio-only)
- **Playlists**: Manage [playlist metadata](playlists)
- **Basics**: Start with [basic downloads](basic-downloads)

## Related References

- [Output Templates Guide](output-templates)
- [Audio-Only Downloads Guide](audio-only)
- [Playlist Downloads Guide](playlists)
- [Command-Line Flags Reference](../../reference/flags)
