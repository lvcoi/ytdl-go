---
title: "Roles and Responsibilities"
weight: 10
---

# Maintainer's Guide

This document provides guidance for maintainers on releases, dependency management, and handling recurring issues.

## Table of Contents

- [Release Process](#release-process)
- [Dependency Management](#dependency-management)
- [Troubleshooting Recurring Issues](#troubleshooting-recurring-issues)
- [Security](#security)
- [Communication](#communication)

## Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (x.0.0): Breaking changes to CLI flags, behavior, or public API
- **MINOR** (0.x.0): New features, backwards-compatible
- **PATCH** (0.0.x): Bug fixes, backwards-compatible

### Creating a Release

#### 1. Prepare the Release

Ensure the codebase is ready:

```bash
# Ensure you're on main branch and up-to-date
git checkout main
git pull origin main

# Run all quality checks
go fmt ./...
go vet ./...
golangci-lint run
go test ./...

# Test build
go build .
```

#### 2. Update CHANGELOG (if present)

If the project has a CHANGELOG.md, update it with:

- All notable changes since last release
- Bug fixes
- New features
- Breaking changes
- Security fixes

#### 3. Tag the Release

```bash
# Create an annotated tag
git tag -a v1.2.3 -m "Release v1.2.3: Brief description"

# Push the tag to GitHub
git push origin v1.2.3
```

This will trigger GitHub Actions (if configured) or prepare for manual release.

#### 4. Build Release Binaries

Build binaries for all major platforms:

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ytdl-go-linux-amd64 .

# Linux (arm64)
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ytdl-go-linux-arm64 .

# macOS (amd64 - Intel)
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ytdl-go-darwin-amd64 .

# macOS (arm64 - Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ytdl-go-darwin-arm64 .

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ytdl-go-windows-amd64.exe .

# Windows (arm64)
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o ytdl-go-windows-arm64.exe .
```

**Build Script** (for convenience):

```bash
#!/bin/bash
# build-release.sh

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

# Create output directory
mkdir -p dist

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    
    output="ytdl-go-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output="${output}.exe"
    fi
    
    echo "Building $output..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "dist/${output}" .
done

echo "Build complete! Binaries in dist/"
```

#### 5. Create GitHub Release

1. Go to GitHub repository → Releases → "Draft a new release"
2. Select the tag you just created (v1.2.3)
3. Add release title: "v1.2.3: Brief description"
4. Add release notes:
   - Summary of changes
   - Breaking changes (if any)
   - Notable bug fixes
   - New features
5. Attach the built binaries
6. Check "Set as latest release" (if appropriate)
7. Click "Publish release"

#### 6. Announce the Release

- Update README.md badge if needed
- Post in GitHub Discussions (if enabled)
- Share on relevant social channels (if applicable)

### Release Checklist

- [ ] Code is tested and working
- [ ] All tests pass (if tests exist)
- [ ] Linting passes
- [ ] CHANGELOG updated (if present)
- [ ] Version tag created (`git tag -a vX.Y.Z`)
- [ ] Tag pushed to GitHub
- [ ] Binaries built for all platforms
- [ ] GitHub Release created with binaries attached
- [ ] Release notes are clear and complete
- [ ] Documentation updated if needed

## Dependency Management

### Core Dependencies

ytdl-go relies on several critical dependencies:

| Dependency | Purpose | Update Frequency |
|------------|---------|------------------|
| `github.com/kkdai/youtube/v2` | YouTube API client | **Critical - Monitor closely** |
| `github.com/charmbracelet/bubbletea` | TUI framework | Stable, periodic updates |
| `github.com/charmbracelet/bubbles` | TUI components | Stable, periodic updates |
| `github.com/charmbracelet/lipgloss` | TUI styling | Stable, periodic updates |
| `github.com/bogem/id3v2/v2` | ID3 tag embedding | Stable, rare updates |
| `github.com/u2takey/ffmpeg-go` | FFmpeg bindings | Stable, rare updates |

### Critical: `kkdai/youtube` Updates

The `kkdai/youtube` library is the **most critical dependency** because YouTube frequently changes their API, which can break downloads.

#### When to Update

Update `kkdai/youtube` when:

1. **Users report widespread 403 errors** - YouTube likely changed their API
2. **The upstream library releases a fix** - Check https://github.com/kkdai/youtube/releases
3. **New YouTube features** - New formats, quality options, etc.
4. **Security vulnerabilities** - Rare, but important

#### How to Update

```bash
# Check current version
grep "github.com/kkdai/youtube/v2" go.mod

# Update to latest
go get github.com/kkdai/youtube/v2@latest

# Or update to specific version
go get github.com/kkdai/youtube/v2@v2.10.6

# Clean up dependencies
go mod tidy

# Test thoroughly!
go build .
./ytdl-go -info https://www.youtube.com/watch?v=dQw4w9WgXcQ
./ytdl-go -audio https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

#### Testing After Update

Always test these scenarios after updating `kkdai/youtube`:

- [ ] Regular video download
- [ ] Audio-only download
- [ ] Playlist download
- [ ] YouTube Music URL
- [ ] Format listing (`-list-formats`)
- [ ] Metadata extraction (`-info`)

### Other Dependencies

For other dependencies, follow this schedule:

- **Security updates:** Immediate
- **Bug fixes:** Within 1-2 weeks
- **Feature updates:** Evaluate need, test thoroughly
- **Major version bumps:** Careful evaluation, extensive testing

### Checking for Updates

```bash
# List outdated dependencies
go list -u -m all

# Update all dependencies to latest minor/patch
go get -u ./...
go mod tidy
```

### Dependency Security

Monitor for security vulnerabilities:

- Enable GitHub Dependabot alerts
- Check https://pkg.go.dev/vuln/ regularly
- Review `go.sum` changes in PRs

## Troubleshooting Recurring Issues

### Issue: Users Report 403 Forbidden Errors

**Symptoms:**
- Downloads fail with "403 Forbidden" error
- Affects multiple users simultaneously
- May affect specific formats (audio-only) more than others

**Root Causes:**
1. YouTube API changes (most common)
2. Rate limiting / IP blocks
3. Outdated `kkdai/youtube` library

**Diagnostic Steps:**

1. **Verify the issue:**
   ```bash
   # Test a known-working video
   ytdl-go -info https://www.youtube.com/watch?v=dQw4w9WgXcQ
   
   # Try audio-only download
   ytdl-go -audio https://www.youtube.com/watch?v=dQw4w9WgXcQ
   ```

2. **Check if it's widespread:**
   - Look for similar issues on GitHub
   - Check `kkdai/youtube` issues: https://github.com/kkdai/youtube/issues
   - Test from different networks/IPs

3. **Check YouTube library status:**
   - Visit https://github.com/kkdai/youtube/releases
   - Look for recent updates mentioning "403" or "API changes"
   - Check issue tracker for related reports

**Solutions:**

1. **Update `kkdai/youtube`:**
   ```bash
   go get github.com/kkdai/youtube/v2@latest
   go mod tidy
   go build .
   ```

2. **Try FFmpeg fallback:**
   - Ensure FFmpeg is installed
   - The tool automatically falls back to FFmpeg for audio downloads
   - If not working, check `youtube.go` FFmpeg fallback logic

3. **Temporary workaround (user-facing):**
   ```bash
   # If audio-only fails, try downloading progressive video format
   ytdl-go -itag 22 [URL]  # 720p progressive
   ytdl-go -itag 18 [URL]  # 360p progressive
   ```

**Prevention:**
- Monitor `kkdai/youtube` repository for updates
- Set up GitHub watch notifications for releases
- Keep a test suite of videos for quick verification

### Issue: FFmpeg Not Found

**Symptoms:**
- "ffmpeg not found" errors
- Audio extraction fallback fails

**Solution:**

1. **Verify FFmpeg installation:**
   ```bash
   which ffmpeg    # Linux/macOS
   where ffmpeg    # Windows
   ```

2. **Installation guide** (point users to):
   - macOS: `brew install ffmpeg`
   - Ubuntu/Debian: `sudo apt-get install ffmpeg`
   - Windows: Chocolatey or manual download from https://ffmpeg.org/download.html

3. **Check PATH:**
   ```bash
   echo $PATH  # Linux/macOS
   echo %PATH% # Windows (cmd)
   $env:Path   # Windows (PowerShell)
   ```

### Issue: Permission Denied / File Access Errors

**Symptoms:**
- "permission denied" when writing files
- "directory does not exist" errors

**Common Causes:**
1. Output directory doesn't exist
2. No write permissions
3. File is locked by another process
4. Invalid characters in filename

**Solutions:**

1. **Check directory permissions:**
   ```bash
   ls -ld /path/to/output  # Linux/macOS
   ```

2. **Sanitize filenames:**
   - The tool already sanitizes filenames in `path.go`
   - If issues persist, check `sanitizePathComponent()` function

3. **User guidance:**
   - Use `-o` with relative output templates (paths are resolved under the output directory)
   - Use `-output-dir` with an absolute path to control the base output directory and ensure it exists
   - Avoid special characters in custom paths

### Issue: Memory Usage / Performance Problems

**Symptoms:**
- High memory consumption
- Slow downloads
- System becomes unresponsive

**Diagnostic Steps:**

1. **Check concurrency settings:**
   ```bash
   # Too many concurrent downloads
   ytdl-go -jobs 50 [URLs...]  # Reduce this!
   ```

2. **Monitor resource usage:**
   ```bash
   # While ytdl-go is running
   top  # or htop on Linux/macOS
   ```

**Solutions:**

1. **Reduce concurrency:**
   ```bash
   # Lower -jobs flag
   ytdl-go -jobs 2 [URLs...]
   
   # Lower playlist/segment concurrency
   ytdl-go -playlist-concurrency 1 -segment-concurrency 1 [URL]
   ```

2. **Check for leaks:**
   - Review `progress_manager.go` and `unified_tui.go`
   - Ensure all goroutines are properly cleaned up
   - Check for unclosed file handles

### Issue: Playlist Downloads Incomplete

**Symptoms:**
- Some playlist videos skipped
- Early termination

**Common Causes:**
1. Private or deleted videos in playlist
2. Region-restricted content
3. Network timeouts

**Expected Behavior:**
- Tool should skip problematic videos and continue
- Summary shows skipped/failed count

**If not working:**
- Check error handling in playlist download logic (`youtube.go`, `runner.go`)
- Ensure errors don't cause early exit

## Security

### Reporting Security Issues

Security vulnerabilities should be reported privately:

1. Do NOT open a public GitHub issue
2. Use GitHub's "Report a vulnerability" feature
3. Or email maintainers directly (if specified in SECURITY.md)

### Handling Security Reports

When you receive a security report:

1. **Acknowledge receipt** within 24-48 hours
2. **Investigate** the issue thoroughly
3. **Develop a fix** in a private branch
4. **Test extensively** before releasing
5. **Coordinate disclosure** with reporter
6. **Release patch** as soon as ready
7. **Publish security advisory** on GitHub

### Security Best Practices

- Never commit credentials or API keys
- Validate all user inputs
- Sanitize file paths (already done in `path.go`)
- Keep dependencies updated
- Enable GitHub security features (Dependabot, Code Scanning)

## Communication

### Responding to Issues

**Response Time Goals:**
- Security issues: 24-48 hours
- Bug reports: Within 1 week
- Feature requests: Within 2 weeks
- General questions: Best effort

**Issue Triage:**

1. **Label appropriately**: bug, feature, question, etc.
2. **Ask for details**: Version, OS, command used, error output
3. **Reproduce**: Try to reproduce the issue locally
4. **Prioritize**: Security > Bugs > Features
5. **Close duplicates**: Link to original issue

### Pull Request Reviews

**Review Checklist:**
- [ ] Code is clean and follows Go conventions
- [ ] No obvious bugs or security issues
- [ ] Tests pass (if tests exist)
- [ ] Documentation updated if needed
- [ ] Commit messages are clear
- [ ] No merge conflicts

**Providing Feedback:**
- Be respectful and constructive
- Explain the "why" behind requested changes
- Acknowledge good work
- Help newcomers learn

### Maintaining Documentation

- Keep README.md up-to-date with features
- Update docs/ when architecture changes
- Add examples for new features
- Fix documentation bugs promptly

## Maintainer Roster

> **Note:** Add maintainer names, roles, and contact information here.

**Last Updated:** 2026-02-05

## Related Documentation

- [Release Process](#release-process) - Detailed release steps
- [Contributing Guide](../contributing/getting-started) - For contributors
- [Architecture Overview](../architecture/overview) - System design
