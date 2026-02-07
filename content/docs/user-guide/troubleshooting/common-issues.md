---
title: "Common Issues"
weight: 10
---

# Common Issues

Solutions to the most frequently encountered problems with ytdl-go.

---

## Download Failures

### 403 Forbidden Errors

**Symptom:** Downloads fail with HTTP 403 errors.

**Solutions:**

1. **Automatic retry** - ytdl-go automatically retries with different methods
2. **Increase timeout** - Some videos need more time to respond:
   ```bash
   ytdl-go -timeout 5m URL
   ```
3. **Check IP reputation** - Your IP might be rate-limited by YouTube
4. **Use cookies** - For age-restricted or members-only content:
   ```bash
   ytdl-go -cookies cookies.txt URL
   ```

### Network Timeouts

**Symptom:** Downloads fail with timeout errors.

**Solutions:**

```bash
# Increase timeout to 10 minutes
ytdl-go -timeout 10m URL

# Retry with fewer concurrent jobs
ytdl-go -jobs 1 URL
```

### Download Incomplete or Corrupted

**Symptom:** Downloaded files are incomplete or won't play.

**Solutions:**

1. **Resume failed downloads** - ytdl-go automatically resumes `.part` files
2. **Check disk space** - Ensure sufficient free space
3. **Try different format**:
   ```bash
   ytdl-go -format mp4 URL
   ```

---

## Format Selection Issues

### "No Progressive Format Available"

**Symptom:** Error about missing progressive formats.

**Solution:** This happens when FFmpeg fallback is required but the format isn't available. Try:

```bash
# Let ytdl-go pick best available format
ytdl-go URL

# Or specify a different quality
ytdl-go -quality 720p URL
```

### Interactive Format Selector Not Working

**Symptom:** `-list-formats` doesn't show interactive UI.

**Solutions:**

1. **Terminal compatibility** - Ensure your terminal supports TUI
2. **JSON mode conflict** - Remove `-json` flag if present
3. **Use direct format selection**:
   ```bash
   ytdl-go -itag 137 URL
   ```

---

## Restricted Content

### Age-Restricted Videos

**Symptom:** "Content is restricted" or exit code 4.

**Solution:**

> **Note:** Age-restricted content requires authentication cookies.

```bash
# Export cookies from browser (use extension)
# Then use with ytdl-go
ytdl-go -cookies cookies.txt URL
```

### Private or Members-Only Videos

**Symptom:** Cannot access private or members-only content.

**Current Limitation:** Authentication is required. Use cookies:

```bash
ytdl-go -cookies cookies.txt URL
```

### Region-Locked Content

**Symptom:** Video not available in your region.

**Workaround:** Use a VPN or proxy to access content from allowed regions.

---

## Playlist Issues

### Empty or Deleted Videos

**Symptom:** Playlist contains unavailable videos.

**Behavior:** ytdl-go automatically skips deleted or private videos and continues with the rest.

```bash
# Download playlist, skipping unavailable videos
ytdl-go https://www.youtube.com/playlist?list=PLAYLIST_ID
```

### Partial Playlist Download

**Symptom:** Only some videos download from playlist.

**Solutions:**

1. **Check network stability**
2. **Increase timeout**:
   ```bash
   ytdl-go -timeout 10m -jobs 2 PLAYLIST_URL
   ```
3. **Review skipped videos** - Some may be restricted or deleted

---

## File System Issues

### Permission Denied

**Symptom:** Cannot write files, permission errors.

**Solutions:**

```bash
# Check output directory permissions
ls -la /path/to/output

# Use directory you own
ytdl-go -o "$HOME/Downloads/{title}.{ext}" URL

# Or fix permissions
chmod u+w /path/to/output
```

### File Already Exists

**Symptom:** "File already exists" error.

**Solutions:**

```bash
# Skip duplicates
ytdl-go -on-duplicate skip URL

# Overwrite existing files
ytdl-go -on-duplicate overwrite URL

# Prompt for each duplicate (default)
ytdl-go -on-duplicate prompt URL
```

### Path Too Long

**Symptom:** Filename or path exceeds system limits.

**Solution:** Use shorter output template:

```bash
# Shorter template
ytdl-go -o "{id}.{ext}" URL

# Or limit title length manually
ytdl-go -o "short/{title}.{ext}" URL
```

---

## Audio Extraction Issues

### No Audio Stream Found

**Symptom:** Audio-only mode fails with "no audio stream" error.

**Solutions:**

```bash
# Let ytdl-go select best audio
ytdl-go -audio URL

# Try specific quality
ytdl-go -audio -quality 128k URL

# Check available formats first
ytdl-go -list-formats URL
```

### Poor Audio Quality

**Symptom:** Downloaded audio quality is lower than expected.

**Solutions:**

```bash
# Request highest quality audio
ytdl-go -audio -quality best URL

# Or specific bitrate
ytdl-go -audio -quality 320k URL
```

---

## Output Template Issues

### Invalid Path or Characters

**Symptom:** Template produces invalid filenames.

**Solution:** ytdl-go automatically sanitizes filenames, but avoid:

- Absolute paths (security risk)
- `..` path segments (security risk)
- Special characters in manual filenames

```bash
# Good templates
ytdl-go -o "{title}.{ext}" URL
ytdl-go -o "downloads/{artist}/{title}.{ext}" URL

# Bad templates (rejected)
ytdl-go -o "/absolute/path/{title}.{ext}" URL  # Rejected
ytdl-go -o "../../../{title}.{ext}" URL        # Rejected
```

### Placeholders Not Replaced

**Symptom:** Filename contains literal `{title}` instead of video title.

**Solution:** Ensure placeholders are spelled correctly:

```bash
# Correct placeholders (see docs)
ytdl-go -o "{title}.{ext}" URL          # ✓
ytdl-go -o "{artist}/{title}.{ext}" URL # ✓

# Wrong placeholders
ytdl-go -o "{name}.{ext}" URL           # ✗ No {name} placeholder
ytdl-go -o "{author}.{ext}" URL         # ✗ Use {artist} instead
```

---

## Performance Issues

### Slow Downloads

**Solutions:**

1. **Check your connection** - Run speed test
2. **Reduce concurrent jobs** - Default is 1, don't exceed your bandwidth:
   ```bash
   ytdl-go -jobs 1 URL
   ```
3. **Network issues** - May be YouTube rate limiting

### High Memory Usage

**Symptom:** ytdl-go consumes excessive memory.

**Solutions:**

```bash
# Download one at a time
ytdl-go -jobs 1 URL

# Avoid very long playlists in one command
# Split into smaller batches
```

---

## Metadata Issues

### Missing ID3 Tags

**Symptom:** Downloaded audio files lack metadata.

**Solution:** ytdl-go automatically embeds metadata. If missing:

1. **Check audio format** - ID3 tags work best with m4a
2. **Use `-write-info-json`** for sidecar metadata:
   ```bash
   ytdl-go -audio -write-info-json URL
   ```

### Incorrect Metadata

**Symptom:** Wrong artist, title, or other metadata.

**Solution:** Override manually:

```bash
ytdl-go -meta artist="Correct Artist" -meta title="Correct Title" URL
```

---

## Web UI Issues

### Cannot Access Web Interface

**Symptom:** `http://localhost:8080` not accessible.

**Solutions:**

1. **Ensure web UI is enabled**:
   ```bash
   ytdl-go -web
   ```
2. **Check port availability**:
   ```bash
   # Use different port if 8080 is taken
   ytdl-go -web -port 3000
   ```
3. **Firewall blocking** - Allow the port through firewall

### Web UI Not Updating

**Symptom:** Downloads don't show in web interface.

**Solution:** Refresh the browser page. The UI uses real-time updates via Server-Sent Events.

---

## Build Issues

### Build Fails

**Symptom:** `go build` or `./build.sh` fails.

**Solutions:**

1. **Check Go version** - Requires Go 1.24+:
   ```bash
   go version  # Should show 1.24 or higher
   ```
2. **Clean build**:
   ```bash
   go clean -cache
   go build -o ytdl-go .
   ```
3. **Dependencies**:
   ```bash
   go mod download
   go mod verify
   ```

---

## Still Having Issues?

If your problem isn't listed here:

1. **Check [FAQ](faq)** for more questions
2. **Review [Error Codes](error-codes)** for specific exit codes
3. **Search [GitHub Issues](https://github.com/lvcoi/ytdl-go/issues)**
4. **Open a new issue** with:
   - Command used
   - Complete error message
   - ytdl-go version
   - Operating system

---

## Related Documentation

- [Error Codes](error-codes) - Exit code reference
- [FAQ](faq) - Frequently asked questions
- [Configuration](../getting-started/configuration) - Setup and configuration
- [CLI Options](../../reference/cli-options) - Complete flag reference
