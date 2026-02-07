---
title: "FAQ"
weight: 30
---

# Frequently Asked Questions

Common questions and answers about ytdl-go.

---

## General Questions

### What is ytdl-go?

ytdl-go is a command-line tool and web application for downloading videos and audio from YouTube. It's written in Go for high performance and includes features like:

- Parallel downloads
- Interactive format selection
- Rich metadata embedding
- Resume capability
- Web UI

### Is ytdl-go free?

Yes! ytdl-go is open source and released under the MIT License. It's free to use, modify, and distribute.

### What platforms does ytdl-go support?

ytdl-go runs on:
- **Linux** (x86-64, ARM)
- **macOS** (Intel, Apple Silicon)
- **Windows** (x86-64)

See [Installation](../getting-started/installation) for platform-specific instructions.

### Do I need FFmpeg?

FFmpeg is **optional but recommended**. ytdl-go can download most content without FFmpeg, but it's helpful for:

- Audio extraction fallback
- Format conversion
- Certain legacy formats

---

## Downloads

### Can I download entire playlists?

Yes! Just provide the playlist URL:

```bash
ytdl-go https://www.youtube.com/playlist?list=PLAYLIST_ID
```

See [Playlists](../usage/playlists) for organization options.

### Can I download only audio?

Absolutely:

```bash
ytdl-go -audio URL
```

This downloads the best available audio-only format. See [Audio-Only Mode](../usage/audio-only) for details.

### How do I resume a failed download?

ytdl-go automatically resumes downloads! If a download is interrupted, just run the same command again. It will detect the `.part` file and continue from where it left off.

### Can I download multiple videos at once?

Yes, with the `-jobs` flag:

```bash
ytdl-go -jobs 4 URL1 URL2 URL3
```

Or for playlists:

```bash
ytdl-go -jobs 4 PLAYLIST_URL
```

> **Warning:** Higher job counts increase bandwidth usage and may trigger rate limits.

### What's the best quality I can download?

The best available quality from YouTube. Use:

```bash
ytdl-go -quality best URL
```

For specific resolutions:

```bash
ytdl-go -quality 1080p URL
ytdl-go -quality 720p URL
```

### Can I download 4K or 8K videos?

Yes, if available:

```bash
ytdl-go -quality 2160p URL   # 4K
ytdl-go -quality 4320p URL   # 8K (rare)
```

Use `-list-formats` to see what's available for a specific video.

---

## Format Selection

### How do I see all available formats?

Use the interactive format selector:

```bash
ytdl-go -list-formats URL
```

This shows a beautiful TUI where you can browse and select formats.

### What's the difference between progressive and adaptive formats?

- **Progressive**: Video + audio in single file (easy to play, larger files)
- **Adaptive**: Separate video and audio streams (better quality options, requires combining)

ytdl-go handles both automatically.

### Can I download specific formats by ID?

Yes, using the itag:

```bash
ytdl-go -itag 137 URL  # Example: 1080p video only
```

Find itags using `-list-formats`.

### What format should I choose?

For most users:

```bash
# Video (best quality, MP4)
ytdl-go -quality best -format mp4 URL

# Audio (high quality, M4A)
ytdl-go -audio -quality 320k URL
```

See [Format Selection](../usage/format-selection) for advanced options.

---

## Output and Organization

### How do I change the output filename?

Use the `-o` flag with placeholders:

```bash
ytdl-go -o "{title}.{ext}" URL
ytdl-go -o "downloads/{artist}/{title}.{ext}" URL
```

See [Output Templates](../usage/output-templates) for all placeholders.

### Can I organize downloads by artist or album?

Yes! Use output template placeholders:

```bash
# By artist
ytdl-go -audio -o "Music/{artist}/{title}.{ext}" URL

# By playlist
ytdl-go -o "{playlist-title}/{title}.{ext}" PLAYLIST_URL
```

### What placeholders are available?

Common placeholders:

| Placeholder | Description |
|-------------|-------------|
| `{title}` | Video title |
| `{artist}` | Channel/artist name |
| `{id}` | Video ID |
| `{ext}` | File extension |
| `{quality}` | Quality (e.g., 1080p) |
| `{upload-date}` | Upload date (YYYYMMDD) |
| `{playlist-title}` | Playlist name |
| `{index}` | Position in playlist |

See [Output Templates](../usage/output-templates) for the complete list.

### How do I avoid overwriting files?

Use the `-on-duplicate` flag:

```bash
# Skip if exists
ytdl-go -on-duplicate skip URL

# Overwrite existing
ytdl-go -on-duplicate overwrite URL

# Prompt (default)
ytdl-go -on-duplicate prompt URL
```

---

## Authentication and Restrictions

### Can I download age-restricted videos?

Yes, but you need authentication cookies:

```bash
ytdl-go -cookies cookies.txt URL
```

Export cookies from your browser using an extension like "Get cookies.txt".

### Can I download private videos?

Only if you're logged in. Use cookies from an authenticated browser session:

```bash
ytdl-go -cookies cookies.txt PRIVATE_URL
```

### Can I download members-only content?

Yes, with cookies from a logged-in YouTube Premium or channel member account:

```bash
ytdl-go -cookies cookies.txt MEMBERS_URL
```

### How do I get cookies from my browser?

1. Install a cookie export extension:
   - **Chrome/Edge**: "Get cookies.txt"
   - **Firefox**: "cookies.txt"
2. Visit YouTube while logged in
3. Export cookies in Netscape format
4. Save as `cookies.txt`
5. Use with `-cookies cookies.txt`

---

## Metadata

### Does ytdl-go embed metadata?

Yes! ytdl-go automatically embeds:

- Title
- Artist/channel
- Album (playlist name)
- Track number (playlist position)
- Upload date
- Thumbnail (for audio files)

### Can I override metadata?

Yes, use the `-meta` flag:

```bash
ytdl-go -meta artist="My Artist" -meta title="My Title" URL
```

### Can I save metadata to a separate file?

Yes:

```bash
# JSON sidecar file
ytdl-go -write-info-json URL

# Description file
ytdl-go -write-description URL

# Thumbnail
ytdl-go -write-thumbnail URL
```

See [Metadata & Sidecars](../usage/metadata-sidecars) for details.

---

## Performance

### Why is my download slow?

Common causes:

1. **Slow internet connection** - Test your connection speed
2. **YouTube rate limiting** - Try fewer concurrent jobs
3. **Network congestion** - Try downloading at different times
4. **Too many concurrent downloads** - Reduce `-jobs` count

### How many concurrent downloads should I use?

Depends on your connection:

- **Slow connection (< 10 Mbps)**: `-jobs 1`
- **Medium connection (10-50 Mbps)**: `-jobs 2-3`
- **Fast connection (> 50 Mbps)**: `-jobs 4-6`

Start conservative and increase if needed.

### Does ytdl-go use a lot of memory?

No. ytdl-go is designed to be memory-efficient. If you experience high memory usage:

1. Reduce `-jobs` count
2. Download playlists in batches
3. Report as a bug with details

---

## Web Interface

### How do I use the web UI?

Build and start the web server:

```bash
# Using build script
./build.sh --web

# Or with pre-built binary
ytdl-go -web
```

Then open `http://localhost:8080` in your browser.

### Can I access the web UI from other devices?

Yes, bind to all interfaces:

```bash
ytdl-go -web -addr 0.0.0.0:8080
```

Then access from other devices using your computer's IP: `http://YOUR_IP:8080`

> **Warning:** This exposes ytdl-go to your local network. Use firewall rules if needed.

### Can I use a different port?

Yes:

```bash
ytdl-go -web -port 3000
```

---

## Scripting and Automation

### Can I use ytdl-go in scripts?

Absolutely! ytdl-go is designed for automation:

```bash
#!/bin/bash
ytdl-go -quiet -on-duplicate skip URL
```

### How do I get machine-readable output?

Use JSON mode:

```bash
ytdl-go -json URL
```

This outputs NDJSON (newline-delimited JSON) with:
- Download progress
- Format information
- Errors

### What JSON event types are available?

```json
{"type": "item", "status": "ok", "title": "...", ...}
{"type": "formats", "formats": [...]}
{"type": "error", "category": "network", "message": "..."}
```

See [Metadata & Sidecars](../usage/metadata-sidecars#json-output-mode) for details.

### How do I handle errors in scripts?

Use exit codes:

```bash
ytdl-go URL
if [ $? -eq 0 ]; then
  echo "Success"
else
  echo "Failed with code: $?"
fi
```

See [Error Codes](error-codes) for all exit codes.

---

## Troubleshooting

### ytdl-go says "command not found"

Your `PATH` is not configured. Either:

1. **Use full path**: `/path/to/ytdl-go URL`
2. **Add to PATH**:
   ```bash
   export PATH="$PATH:/path/to/ytdl-go-directory"
   ```
3. **Install to system path**: `sudo mv ytdl-go /usr/local/bin/`

### I get 403 Forbidden errors

This is usually temporary. ytdl-go automatically retries. If persistent:

```bash
# Increase timeout
ytdl-go -timeout 600 URL

# Use cookies for authentication
ytdl-go -cookies cookies.txt URL
```

### Downloads hang or freeze

1. **Increase timeout**: `-timeout 600`
2. **Check network**: `ping youtube.com`
3. **Try one at a time**: `-jobs 1`
4. **Enable JSON mode for debug**: `-json`

### I found a bug. How do I report it?

1. Check [existing issues](https://github.com/lvcoi/ytdl-go/issues)
2. Open a [new issue](https://github.com/lvcoi/ytdl-go/issues/new) with:
   - ytdl-go version (`ytdl-go -version`)
   - Full command used
   - Complete error message
   - Operating system and Go version

---

## Development

### How do I build ytdl-go from source?

```bash
git clone https://github.com/lvcoi/ytdl-go.git
cd ytdl-go
./build.sh
```

See [Installation](../getting-started/installation) for details.

### What Go version do I need?

Go 1.24 or later:

```bash
go version  # Should show 1.24+
```

### Can I contribute to ytdl-go?

Yes! We welcome contributions. See:

- [Contributing Guide](../../developer-guide/contributing/getting-started) (coming soon)
- [Architecture Overview](../../developer-guide/architecture/overview) (coming soon)

---

## Legal and Privacy

### Is it legal to download YouTube videos?

This depends on:
- YouTube's Terms of Service
- Copyright law in your jurisdiction
- The specific video's license

ytdl-go is a tool. You are responsible for ensuring your use complies with applicable laws and terms of service.

### Does ytdl-go collect data?

No. ytdl-go:
- Doesn't track usage
- Doesn't send telemetry
- Doesn't collect personal information
- Runs entirely on your machine

### Where are downloads stored?

In your current directory by default, or wherever you specify with `-o`.

ytdl-go doesn't upload or sync files anywhere.

---

## Still Have Questions?

- [Common Issues](common-issues) - Solutions to problems
- [Error Codes](error-codes) - Exit code reference
- [GitHub Issues](https://github.com/lvcoi/ytdl-go/issues) - Search or ask
- [Documentation Home](../..) - Full documentation
