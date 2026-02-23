# Basic Downloads

## Single Video

Download a video with the best available quality:

```bash
ytdl-go https://www.youtube.com/watch?v=BaW_jenozKc
```

## Specific Quality

```bash
# 720p video
ytdl-go -quality 720p URL

# Lowest quality (smallest file)
ytdl-go -quality worst URL
```

## Specific Container Format

```bash
# Prefer MP4
ytdl-go -format mp4 URL

# Prefer WebM
ytdl-go -format webm URL
```

## Specific Itag

Use `-list-formats` to find the itag number, then download it directly:

```bash
ytdl-go -itag 137 URL
```

## Multiple URLs

Download several videos at once:

```bash
ytdl-go URL1 URL2 URL3
```

Use `-jobs` to download in parallel:

```bash
ytdl-go -jobs 4 URL1 URL2 URL3 URL4
```

## Network Options

```bash
# Custom timeout
ytdl-go -timeout 5m URL

# Resume downloads (automatic .part file detection)
ytdl-go URL
```

## Scripting

```bash
# JSON output for machine parsing
ytdl-go -json URL

# Quiet mode (no progress bars, errors still shown)
ytdl-go -quiet URL

# Combine for pipelines
ytdl-go -json URL | jq -r 'select(.type=="item" and .status=="ok") | .output'
```
