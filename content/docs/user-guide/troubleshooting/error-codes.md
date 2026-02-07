---
title: "Error Codes"
weight: 20
---

# Error Codes

ytdl-go uses standardized exit codes to indicate different types of failures. This helps with scripting and automation.

---

## Exit Code Reference

| Exit Code | Category | Meaning | Common Causes |
|-----------|----------|---------|---------------|
| **0** | Success | Command completed successfully | - |
| **1** | Unknown | General error or unknown failure | Unexpected errors, bugs |
| **2** | Invalid URL | URL is malformed or not supported | Wrong URL format, unsupported site |
| **3** | Unsupported | Feature or format not supported | Format unavailable, unsupported feature |
| **4** | Restricted | Content is restricted | Age-restricted, private, members-only |
| **5** | Network | Network connectivity issues | Timeout, connection refused, DNS failure |
| **6** | Filesystem | File system errors | Permission denied, disk full, invalid path |

---

## Exit Code 0: Success

**Meaning:** The command completed successfully.

```bash
ytdl-go URL
echo $?  # Prints: 0
```

---

## Exit Code 1: Unknown Error

**Meaning:** A general error occurred that doesn't fit into other categories.

**Common Causes:**
- Unexpected program errors
- Internal bugs
- Uncategorized failures

**Example:**

```bash
ytdl-go URL
echo $?  # Prints: 1 (if unknown error)
```

**What to do:**
1. Enable JSON mode for details: `ytdl-go -json URL`
2. Check the error message
3. Report as bug if unexpected

---

## Exit Code 2: Invalid URL

**Meaning:** The provided URL is malformed or points to an unsupported site.

**Common Causes:**
- Typo in URL
- URL doesn't point to YouTube
- Missing URL parameter

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
- Double-check the URL
- Ensure URL starts with `https://youtube.com` or `https://www.youtube.com`
- Use full YouTube URLs, not shortened links (unless supported)

---

## Exit Code 3: Unsupported

**Meaning:** The requested feature or format is not supported.

**Common Causes:**
- Requested format not available
- Feature not implemented
- Progressive format missing when required

**Examples:**

```bash
# Format not available
ytdl-go -format flv URL
# Exit code: 3 (if FLV not available)

# Quality not available
ytdl-go -quality 4320p URL
# Exit code: 3 (if 8K not available)
```

**Solutions:**
- Use `-list-formats` to see available formats
- Try different quality or format
- Use `best` quality to let ytdl-go choose

```bash
# Check available formats
ytdl-go -list-formats URL

# Use best available
ytdl-go -quality best URL
```

---

## Exit Code 4: Restricted

**Meaning:** Content is restricted and cannot be accessed without authentication.

**Common Causes:**
- Age-restricted videos
- Private videos
- Members-only content
- Region-locked content

**Examples:**

```bash
# Age-restricted video
ytdl-go URL
# Exit code: 4 (if age-restricted)
```

**Solutions:**

> **Note:** Age-restricted and private content requires authentication cookies.

```bash
# Export cookies from browser (use browser extension)
# Then use with ytdl-go
ytdl-go -cookies cookies.txt URL
```

**Cookie Export:**
1. Install browser extension:
   - Chrome/Edge: "Get cookies.txt"
   - Firefox: "cookies.txt"
2. Export cookies for YouTube
3. Save as `cookies.txt`
4. Use with `-cookies` flag

---

## Exit Code 5: Network

**Meaning:** Network connectivity issues prevented the download.

**Common Causes:**
- Connection timeout
- DNS resolution failure
- Server unreachable
- Network interruption
- Rate limiting

**Examples:**

```bash
# Timeout
ytdl-go URL
# Exit code: 5 (if timeout)

# Connection refused
ytdl-go URL
# Exit code: 5 (if cannot connect)
```

**Solutions:**

```bash
# Increase timeout
ytdl-go -timeout 10m URL

# Retry with exponential backoff (manual)
for i in 1 2 3; do
  ytdl-go URL && break
  sleep $((i * 5))
done

# Check network connectivity
ping youtube.com
curl -I https://www.youtube.com
```

**Debugging:**

```bash
# JSON mode shows detailed network errors
ytdl-go -json URL

# Test without downloading
ytdl-go -info URL
```

---

## Exit Code 6: Filesystem

**Meaning:** File system operations failed.

**Common Causes:**
- Permission denied
- Disk full
- Invalid path
- Directory doesn't exist
- File already exists (with `on-duplicate` setting)

**Examples:**

```bash
# Permission denied
ytdl-go -o "/root/video.mp4" URL
# Exit code: 6 (if not root)

# Disk full
ytdl-go URL
# Exit code: 6 (if no space)

# Invalid path
ytdl-go -o "/invalid/../path.mp4" URL
# Exit code: 6 (path traversal blocked)
```

**Solutions:**

```bash
# Check permissions
ls -la /path/to/output

# Check disk space
df -h

# Use valid output path
ytdl-go -o "$HOME/Downloads/{title}.{ext}" URL

# Handle duplicates
ytdl-go -on-duplicate skip URL
```

---

## Using Exit Codes in Scripts

Exit codes are particularly useful for automation and error handling in scripts.

### Bash Script Example

```bash
#!/bin/bash

url="$1"
ytdl-go "$url"
exit_code=$?

case $exit_code in
  0)
    echo "✓ Download successful"
    ;;
  2)
    echo "✗ Invalid URL: $url"
    ;;
  3)
    echo "✗ Format not supported"
    ;;
  4)
    echo "✗ Content restricted - try with cookies"
    ;;
  5)
    echo "✗ Network error - retrying..."
    sleep 5
    ytdl-go "$url"
    ;;
  6)
    echo "✗ File system error - check permissions"
    ;;
  *)
    echo "✗ Unknown error (code: $exit_code)"
    ;;
esac

exit $exit_code
```

### Retry Logic Example

```bash
#!/bin/bash

max_retries=3
retry_count=0
url="$1"

while [ $retry_count -lt $max_retries ]; do
  ytdl-go "$url"
  exit_code=$?
  
  if [ $exit_code -eq 0 ]; then
    echo "Success!"
    exit 0
  elif [ $exit_code -eq 5 ]; then
    retry_count=$((retry_count + 1))
    echo "Network error. Retry $retry_count/$max_retries..."
    sleep $((retry_count * 5))
  else
    echo "Non-retryable error: $exit_code"
    exit $exit_code
  fi
done

echo "Failed after $max_retries retries"
exit 5
```

### Conditional Downloads

```bash
#!/bin/bash

# Download only if it doesn't exist
ytdl-go -on-duplicate skip "$1"

if [ $? -eq 6 ]; then
  echo "File already exists, skipping"
  exit 0
fi
```

---

## Checking Exit Codes

### Unix/Linux/macOS

```bash
ytdl-go URL
echo $?  # Prints exit code
```

### Windows (PowerShell)

```powershell
ytdl-go URL
echo $LASTEXITCODE  # Prints exit code
```

### Windows (CMD)

```cmd
ytdl-go URL
echo %ERRORLEVEL%  REM Prints exit code
```

---

## Error Categories in JSON Mode

When using `-json` flag, errors include category information:

```json
{
  "type": "error",
  "category": "network",
  "message": "connection timeout",
  "url": "https://www.youtube.com/watch?v=..."
}
```

**Category Field Values:**
- `invalid_url` → Exit code 2
- `unsupported` → Exit code 3
- `restricted` → Exit code 4
- `network` → Exit code 5
- `filesystem` → Exit code 6
- `unknown` → Exit code 1

---

## Related Documentation

- [Common Issues](common-issues) - Solutions to frequent problems
- [FAQ](faq) - Frequently asked questions
- [CLI Options](../../reference/cli-options) - Complete flag reference
- [Basic Downloads](../usage/basic-downloads) - Download guide
