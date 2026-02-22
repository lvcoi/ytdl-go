# Format Selection

## Interactive Format Browser

Use `-list-formats` to browse all available formats in an interactive TUI:

```bash
ytdl-go -list-formats URL
```

### Keyboard Controls

| Key | Action |
| --- | --- |
| `↑/↓` or `j/k` | Navigate |
| `Enter` | Download selected format |
| `1-9` | Jump to itags starting with digit |
| `Page Up/Down` | Jump 10 formats |
| `Home/End` or `g/G` | First/Last format |
| `b` | Go back without downloading |
| `q/Esc/Ctrl+C` | Quit |

## Quality Selection

```bash
# Specific resolution
ytdl-go -quality 1080p URL
ytdl-go -quality 720p URL

# Best or worst
ytdl-go -quality best URL
ytdl-go -quality worst URL

# Audio bitrate (with -audio)
ytdl-go -audio -quality 128k URL
```

If the exact quality is unavailable, the closest match is selected.

## Container Format

```bash
ytdl-go -format mp4 URL
ytdl-go -format webm URL
ytdl-go -audio -format m4a URL
```

## Direct Itag Selection

Bypass automatic selection by specifying a YouTube itag number:

```bash
ytdl-go -itag 251 URL
```

Common itags:

| Itag | Type | Description |
| --- | --- | --- |
| 18 | Video | 360p MP4 (progressive) |
| 22 | Video | 720p MP4 (progressive) |
| 140 | Audio | AAC 128kbps |
| 251 | Audio | Opus ~160kbps |

## JSON Format Listing

For scripting, combine `-json` with `-list-formats` to get machine-readable format data:

```bash
ytdl-go -json -list-formats URL
```

## Flag Precedence

When flags conflict:

1. `-itag` overrides `-quality` and `-format`
2. `-json` implies `-quiet`
3. `-info` prevents downloading (metadata only)
4. `-list-formats` prevents automatic downloading
