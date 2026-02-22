# Metadata & Sidecars

## Sidecar JSON

Each download produces a sidecar JSON file (`<media-file>.json`) containing structured metadata. This file is used by the web UI for artist/album/thumbnail grouping.

Fields include: `id`, `title`, `artist`, `album`, `thumbnail_url`, `source_url`, `release_date`, `duration_seconds`, `output`, `status`.

Legacy files without sidecars still load in the library but may appear under "Unknown" buckets until re-downloaded.

## Metadata Overrides

Use the `-meta` flag to override metadata fields. It can be specified multiple times:

```bash
ytdl-go -meta title="Custom Title" -meta artist="Custom Artist" URL
```

### Supported Keys

| Key | Description |
| --- | --- |
| `title` | Video title |
| `artist` | Artist/author name |
| `author` | Uploader/channel name |
| `album` | Album name |
| `track` | Track number |
| `disc` | Disc number |
| `release_date` | Release date (YYYY-MM-DD) |
| `release_year` | Release year (YYYY) |
| `source_url` | Source URL |

Overrides affect:

- Output template placeholders (`{title}`, `{artist}`, etc.)
- Sidecar JSON content
- ID3 tags embedded in audio files

## ID3 Tag Embedding

For MP3 files, ytdl-go automatically embeds ID3 tags with available metadata. Tags include title, artist, and album when available.

```bash
ytdl-go -audio \
  -meta album="Greatest Hits" \
  -meta track=5 \
  -meta artist="The Band" \
  URL
```

!!! note
    ID3 tag embedding is currently supported for MP3 files only. Other audio formats (Opus, AAC) receive sidecar metadata but no embedded tags.

## Info Mode

Extract metadata as JSON without downloading:

```bash
ytdl-go -info URL
```

Output includes `id`, `title`, `description`, `author`, and `duration_seconds`.

Combine with `jq` for field extraction:

```bash
ytdl-go -info URL | jq -r .title
```
