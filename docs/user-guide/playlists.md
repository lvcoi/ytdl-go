# Playlists

## Download a Playlist

```bash
ytdl-go "https://youtube.com/playlist?list=..."
```

All videos in the playlist are downloaded sequentially. Empty or deleted entries are automatically skipped.

## Organize Playlist Downloads

Use output templates with playlist placeholders:

```bash
ytdl-go -o "Archive/{playlist-title}/{index} - {title}.{ext}" URL
```

Available playlist placeholders:

| Placeholder | Description | Example |
| --- | --- | --- |
| `{playlist_title}` or `{playlist-title}` | Playlist name | `My Playlist` |
| `{playlist_id}` or `{playlist-id}` | Playlist ID | `PLxxx...` |
| `{index}` | Video index (1-based) | `1`, `2`, `3` |
| `{count}` | Total videos in playlist | `25` |

## Audio-Only Playlist

```bash
ytdl-go -audio -o "Music/{playlist-title}/{index} - {title}.{ext}" URL
```

## Playlist Concurrency

!!! note
    The `-playlist-concurrency` flag is reserved for future versions. Playlist entries are currently always downloaded sequentially.

```bash
# Currently has no effect
ytdl-go -playlist-concurrency 8 URL
```

## Handling Failures

- Private or deleted videos are skipped automatically
- Region-restricted content is skipped with a warning
- Network timeouts trigger retries
- A summary at the end shows success/failure counts
