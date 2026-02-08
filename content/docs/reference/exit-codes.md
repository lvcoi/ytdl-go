---
title: "Exit Codes"
weight: 30
---

# Exit Codes

Quick reference for ytdl-go exit codes used in scripting and automation. Exit codes indicate the type of error or success condition.

## Quick Reference Table

| Exit Code | Category | Meaning | Common Causes |
|-----------|----------|---------|---------------|
| **0** | Success | Command completed successfully | - |
| **1** | Unknown | General error or unknown failure | Unexpected errors, bugs |
| **2** | Invalid URL | URL is malformed or not supported | Wrong URL format, unsupported site |
| **3** | Unsupported | Feature or format not supported | Format unavailable, unsupported feature |
| **4** | Restricted | Content is restricted | Age-restricted, private, members-only |
| **5** | Network | Network connectivity issues | Timeout, connection refused, DNS failure |
| **6** | Filesystem | File system errors | Permission denied, disk full, invalid path |

## Checking Exit Codes

### Unix/Linux/macOS

```bash
ytdl-go [URL]
echo $?  # Prints exit code
```

### Windows (PowerShell)

```powershell
ytdl-go [URL]
echo $LASTEXITCODE  # Prints exit code
```

### Windows (CMD)

```cmd
ytdl-go [URL]
echo %ERRORLEVEL%  REM Prints exit code
```

## Using Exit Codes in Scripts

### Basic Error Handling

```bash
#!/bin/bash

ytdl-go "$1"
exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo "✓ Download successful"
else
    echo "✗ Download failed with exit code: $exit_code"
    exit $exit_code
fi
```

### Conditional Actions

```bash
#!/bin/bash

ytdl-go "$1"

case $? in
    0)
        echo "✓ Success"
        ;;
    2)
        echo "✗ Invalid URL"
        ;;
    3)
        echo "✗ Format not supported - try different quality"
        ytdl-go -quality best "$1"
        ;;
    4)
        echo "✗ Restricted content - authentication required"
        ;;
    5)
        echo "✗ Network error - retrying..."
        sleep 5
        ytdl-go "$1"
        ;;
    6)
        echo "✗ Filesystem error - check permissions"
        ;;
    *)
        echo "✗ Unknown error"
        ;;
esac
```

### Retry Logic

```bash
#!/bin/bash

max_retries=3
url="$1"

for i in $(seq 1 $max_retries); do
    ytdl-go "$url"
    exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        echo "✓ Success!"
        exit 0
    elif [ $exit_code -eq 5 ]; then
        echo "Network error. Retry $i/$max_retries..."
        sleep $((i * 5))
    else
        echo "Non-retryable error: $exit_code"
        exit $exit_code
    fi
done

echo "Failed after $max_retries retries"
exit 5
```

## Exit Code Details

### Exit Code 0: Success

Command completed successfully. All downloads finished without errors.

```bash
ytdl-go [URL]
echo $?  # Prints: 0
```

### Exit Code 1: Unknown Error

General error that doesn't fit other categories. May indicate internal bugs or unexpected conditions.

**What to do:**
- Enable debug logging: `ytdl-go -log-level debug [URL]`
- Use JSON mode for details: `ytdl-go -json [URL]`
- Report as bug if unexpected

### Exit Code 2: Invalid URL

The provided URL is malformed or points to an unsupported site.

**Examples:**
```bash
# No URL provided
ytdl-go
# Exit code: 2

# Invalid URL format
ytdl-go "not-a-url"
# Exit code: 2

# Unsupported website
ytdl-go "https://example.com/video"
# Exit code: 2
```

**Solutions:**
- Verify URL format
- Ensure URL is a valid YouTube link
- Use full URLs, not shortened links

### Exit Code 3: Unsupported

The requested feature or format is not supported or available.

**Examples:**
```bash
# Format not available
ytdl-go -format flv [URL]
# Exit code: 3 (if FLV unavailable)

# Quality not available
ytdl-go -quality 4320p [URL]
# Exit code: 3 (if 8K unavailable)
```

**Solutions:**
```bash
# List available formats
ytdl-go -list-formats [URL]

# Use best available quality
ytdl-go -quality best [URL]
```

### Exit Code 4: Restricted

Content is restricted and cannot be accessed. May require authentication.

**Common causes:**
- Age-restricted videos
- Private videos
- Members-only content
- Region-locked content

**Solutions:**

> **Note:** ytdl-go does not currently support authentication for restricted content. Only publicly accessible videos can be downloaded.

### Exit Code 5: Network

Network connectivity issues prevented the download.

**Common causes:**
- Connection timeout
- DNS resolution failure
- Server unreachable
- Rate limiting

**Solutions:**
```bash
# Increase timeout
ytdl-go -timeout 10m [URL]

# Retry manually
ytdl-go [URL] || ytdl-go [URL]

# Debug network issues
ytdl-go -log-level debug [URL]
```

### Exit Code 6: Filesystem

File system operations failed.

**Common causes:**
- Permission denied
- Disk full
- Invalid path
- Directory doesn't exist
- Path traversal blocked

**Examples:**
```bash
# Permission denied
ytdl-go -o "/root/video.mp4" [URL]
# Exit code: 6 (if not root)

# Disk full
ytdl-go [URL]
# Exit code: 6 (if no space)

# Invalid path
ytdl-go -o "/invalid/../path.mp4" [URL]
# Exit code: 6 (path traversal blocked)
```

**Solutions:**
```bash
# Check permissions
ls -la /path/to/output

# Check disk space
df -h

# Use valid output path
ytdl-go -o "$HOME/Downloads/{title}.{ext}" [URL]
```

## JSON Mode Error Categories

When using `-json` flag, errors include category information that maps to exit codes:

```json
{
  "type": "error",
  "category": "network",
  "message": "connection timeout",
  "url": "https://www.youtube.com/watch?v=..."
}
```

**Category Mapping:**
- `invalid_url` → Exit code 2
- `unsupported` → Exit code 3
- `restricted` → Exit code 4
- `network` → Exit code 5
- `filesystem` → Exit code 6
- `unknown` → Exit code 1

## Advanced Scripting Examples

### Parallel Downloads with Error Tracking

```bash
#!/bin/bash

success=0
failed=0

for url in "$@"; do
    ytdl-go "$url" &
done

wait

for job in $(jobs -p); do
    wait $job
    if [ $? -eq 0 ]; then
        ((success++))
    else
        ((failed++))
    fi
done

echo "Success: $success, Failed: $failed"
[ $failed -eq 0 ]  # Exit 0 only if all succeeded
```

### Error-Specific Notifications

```bash
#!/bin/bash

ytdl-go "$1"
exit_code=$?

case $exit_code in
    0)
        notify-send "Download Complete" "Successfully downloaded video"
        ;;
    5)
        notify-send "Network Error" "Check your internet connection"
        ;;
    6)
        notify-send "Storage Error" "Check disk space and permissions"
        ;;
    *)
        notify-send "Download Failed" "Exit code: $exit_code"
        ;;
esac

exit $exit_code
```

### Logging with Exit Codes

```bash
#!/bin/bash

log_file="downloads.log"
url="$1"

ytdl-go "$url" 2>&1 | tee -a "$log_file"
exit_code=${PIPESTATUS[0]}

echo "[$(date)] URL: $url, Exit Code: $exit_code" >> "$log_file"
exit $exit_code
```

---

## Related Documentation

- [Error Codes (Detailed)](../user-guide/troubleshooting/error-codes) - Comprehensive error documentation with solutions
- [Common Issues](../user-guide/troubleshooting/common-issues) - Troubleshooting guide
- [CLI Options](cli-options) - Command-line flag reference
- [Advanced Usage](../user-guide/usage/advanced-usage) - Complex scenarios and scripting
