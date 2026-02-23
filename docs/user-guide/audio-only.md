# Audio-Only Mode

## Basic Audio Download

```bash
ytdl-go -audio URL
```

Selects the best available audio-only format automatically.

## Selection Priority

1. Highest bitrate audio-only format
2. Best codec (Opus > AAC > MP3)
3. Falls back to progressive format if audio-only is unavailable

## Specific Audio Quality

```bash
# 128kbps audio
ytdl-go -audio -quality 128k URL

# Best quality audio
ytdl-go -audio -quality best URL

# Lowest quality (smallest file)
ytdl-go -audio -quality worst URL
```

## Specific Audio Format

```bash
# M4A (AAC)
ytdl-go -audio -format m4a URL

# Opus
ytdl-go -audio -format opus URL
```

## Download Strategy

Audio-only downloads use a multi-layered fallback strategy:

1. **Direct audio-only download** — Standard chunked download of the audio stream
2. **Single request retry** — If a 403 error occurs, retries without chunking
3. **FFmpeg fallback** — Downloads a progressive video and extracts the audio track

The FFmpeg fallback requires FFmpeg to be installed. See [Installation](installation.md) for setup instructions.

## Music Library Archiving

Combine audio mode with output templates and metadata overrides:

```bash
ytdl-go -audio \
  -o "Music/{artist}/{album}/{index} - {title}.{ext}" \
  -meta album="Album Name" \
  URL
```

## ID3 Tags

For MP3 files, ytdl-go automatically embeds ID3 tags with available metadata (title, artist, album). Use `-meta` to override:

```bash
ytdl-go -audio \
  -meta title="Custom Title" \
  -meta artist="Custom Artist" \
  -meta album="Custom Album" \
  URL
```
