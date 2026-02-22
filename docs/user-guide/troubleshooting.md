# Troubleshooting

## 403 Forbidden Errors

**Symptoms:** Downloads fail with "403 Forbidden", may affect multiple users simultaneously.

**Causes:**

- YouTube API changes (most common)
- Rate limiting / IP blocks
- Outdated YouTube library dependency

**Solutions:**

1. The tool automatically retries with different download strategies
2. Try a longer timeout: `ytdl-go -timeout 10m URL`
3. For audio, ensure FFmpeg is installed for the fallback strategy
4. Try a progressive format directly: `ytdl-go -itag 22 URL` (720p) or `ytdl-go -itag 18 URL` (360p)

## Restricted Content

Private, age-gated, or member-only videos require authentication which is **not supported**. These videos are skipped automatically in playlists.

## FFmpeg Not Found

**Symptoms:** Audio extraction fallback fails.

**Solution:** Install FFmpeg and ensure it's in your PATH:

=== "macOS"

    ```bash
    brew install ffmpeg
    ```

=== "Ubuntu/Debian"

    ```bash
    sudo apt-get install ffmpeg
    ```

=== "Windows"

    ```bash
    choco install ffmpeg
    ```

Verify: `which ffmpeg` (Linux/macOS) or `where ffmpeg` (Windows).

## Permission Denied / File Access Errors

- Ensure the output directory exists and is writable
- Use `-o` with relative paths
- Avoid special characters in custom paths
- The `-output-dir` flag can constrain paths to a safe directory

## Playlist Downloads Incomplete

- Private or deleted videos are skipped automatically
- Region-restricted content is skipped with a warning
- A summary at the end shows success/failure counts

## Web Port Already in Use

`ytdl-go -web` retries on higher ports automatically. Check startup logs for the selected URL. When running the frontend dev server, align the proxy target:

```bash
VITE_API_PROXY_TARGET=http://127.0.0.1:<port> npm run dev
```

## Library Metadata/Thumbnails Missing

New downloads write sidecar metadata used by the web UI. Legacy files without sidecars still load but may appear under "Unknown" buckets until re-downloaded.

## Debug Mode

For detailed diagnostics:

```bash
ytdl-go -log-level debug URL
```
