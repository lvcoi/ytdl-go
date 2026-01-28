# Feature Review Assessment - ytdl-go

**Date**: 2026-01-28  
**Reviewer**: Automated Review  
**Repository**: lvcoi/ytdl-go  
**Branch**: copilot/review-public-video-download  

## Executive Summary

This document provides a comprehensive assessment of the ytdl-go project against the PR checklist requirements for a public video download tool with progress UI. The tool is **YouTube-specific** and built on the `kkdai/youtube` library, which inherently limits some of the checklist's more general requirements.

### Overall Status: **PARTIALLY COMPLIANT**

**Key Findings:**
- ✅ **Strong Progress UI**: Excellent implementation with multiple bars, resize handling, and logging
- ✅ **Solid Error Handling**: Good restricted access detection
- ⚠️ **Limited to YouTube**: Not a general-purpose tool (by design)
- ❌ **Missing DRM/Encryption Detection**: No explicit checks for encrypted streams
- ❌ **No Metadata Export**: Missing sidecar JSON files and metadata embedding
- ❌ **No Playlist Manifest**: Doesn't preserve playlist ordering in a structured file
- ⚠️ **Limited Testing**: Some critical scenarios untested

---

## A) Scope, Compliance, and Non-Goals (Hard Gates)

### Status: ✅ PARTIAL PASS (with documentation gaps)

| Item | Status | Notes |
|------|--------|-------|
| Only supports publicly accessible video URLs | ✅ YES | YouTube-only; uses public APIs |
| No DRM circumvention logic | ✅ YES | No DRM code present |
| No paywall/login bypass | ✅ YES | No authentication code |
| No token spoofing/credential harvesting | ✅ YES | Clean implementation |
| No browser automation for protected streams | ✅ YES | Direct API usage only |
| Detects restricted content | ✅ YES | `isRestrictedAccess()` function exists |
| Fails gracefully with clear errors | ✅ YES | `wrapAccessError()` provides context |
| Non-zero exit code on failure | ✅ YES | Errors propagate to main |

**Issues:**
- ⚠️ Documentation doesn't explicitly state "no DRM/paywall support" in non-goals
- ⚠️ No explicit statement about YouTube-only limitation

**Recommendation:**
- Add explicit non-goals section to README stating no support for:
  - DRM/encrypted content
  - Login/paywall bypass
  - Non-YouTube platforms (or clarify it's YouTube-specific)

---

## B) Source & Format Support

### Status: ❌ DOES NOT MEET (YouTube-only, no DRM detection)

| Item | Status | Notes |
|------|--------|-------|
| Accepts any public video URL | ❌ NO | **YouTube-only** via kkdai/youtube library |
| Direct file downloads (.mp4, .webm, .mov) | ⚠️ LIMITED | Only what YouTube provides as progressive formats |
| HLS .m3u8 (unencrypted) | ❌ NO | Not implemented |
| DASH .mpd (unencrypted) | ❌ NO | Not implemented |
| Detects encrypted HLS (AES-128/key URIs) | ❌ NO | No detection code |
| Detects encrypted DASH (Widevine/PlayReady/CENC) | ❌ NO | No detection code |
| Source handling is modular/extensible | ⚠️ LIMITED | Tightly coupled to kkdai/youtube |

**Critical Gaps:**
1. **No multi-platform support**: Tool only works with YouTube
2. **No DRM/encryption detection**: Could attempt to download encrypted content and fail silently
3. **No HLS/DASH parsing**: Relies entirely on kkdai/youtube's format selection

**Recommendation:**
- **If YouTube-only is acceptable**: Document this clearly and remove multi-platform requirements
- **If general-purpose is required**: Major refactoring needed to add:
  - HLS manifest parser with AES-128 key detection
  - DASH manifest parser with DRM signaling detection
  - Pluggable extractor architecture

---

## C) URL Analysis & Validation

### Status: ✅ PASS

| Item | Status | Notes |
|------|--------|-------|
| Validates input URL format | ✅ YES | `validateInputURL()` checks scheme |
| Invalid URLs produce explicit errors | ✅ YES | Clear error messages |
| Detects downloadability prior to download | ✅ YES | Restricted access detection |
| Implements `--list-formats` | ✅ YES | Implemented with tabular output |

**Code Evidence:**
```go
// validation.go
func validateInputURL(raw string) error {
    parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
    if err != nil {
        return fmt.Errorf("invalid url %q: %w", raw, err)
    }
    switch parsed.Scheme {
    case "http", "https":
        return nil
    default:
        return fmt.Errorf("invalid url %q: scheme must be http or https", raw)
    }
}
```

---

## D) Download Behavior & Output Correctness

### Status: ⚠️ PARTIAL (missing resume support)

| Item | Status | Notes |
|------|--------|-------|
| Defaults to best available quality | ✅ YES | `selectFormat()` picks best |
| Supports `--quality` | ❌ NO | Not implemented |
| Supports `--format` | ❌ NO | Not implemented |
| Supports `--output` | ✅ YES | `-o` flag with templates |
| Downloads segments | ⚠️ N/A | Handled by kkdai/youtube library |
| Assembles segments in correct order | ⚠️ N/A | Library responsibility |
| Supports resume without corruption | ❌ NO | No resume logic |
| Output is playable and correctly muxed | ✅ YES | Progressive formats only |

**Missing Features:**
- No `--quality` flag (always selects best)
- No `--format` flag (automatic selection)
- No resume capability (would require range requests and state tracking)

**Recommendation:**
- Add `--quality` flag with options like `720p`, `1080p`, `best`, `worst`
- Add `--format` flag to select by itag or codec preference
- Implement resume with partial file detection and HTTP range requests

---

## E) CLI Interface Requirements

### Status: ✅ PASS (mostly complete)

| Item | Status | Notes |
|------|--------|-------|
| Terminal-only operation | ✅ YES | CLI tool |
| Required input: `url` | ✅ YES | Args via flag.Args() |
| Optional `--quality` | ❌ NO | Not implemented |
| Optional `--format` | ❌ NO | Not implemented |
| Optional `--output` | ✅ YES | `-o` flag |
| Optional `--list-formats` | ✅ YES | Implemented |
| Optional `--audio-only` | ✅ YES | `--audio` flag |
| Optional `--json` | ⚠️ PARTIAL | `--info` outputs JSON, but not for all operations |
| `--json` mode no stdout noise | ⚠️ NEEDS VERIFY | Not fully tested |

**Code Evidence:**
```go
// main.go
flag.StringVar(&opts.OutputTemplate, "o", "{title}.{ext}", ...)
flag.BoolVar(&opts.AudioOnly, "audio", false, ...)
flag.BoolVar(&opts.InfoOnly, "info", false, ...)
flag.BoolVar(&opts.ListFormats, "list-formats", false, ...)
flag.BoolVar(&opts.Quiet, "quiet", false, ...)
```

**Recommendation:**
- Implement `--quality` and `--format` flags
- Add `--json` flag that routes all logs to stderr and outputs only JSON to stdout
- Current `--info` could be merged into `--json` behavior

---

## F) Error Handling & Messaging

### Status: ✅ GOOD

| Item | Status | Notes |
|------|--------|-------|
| Invalid URL errors | ✅ YES | Clear validation messages |
| Unsupported format/source | ⚠️ IMPLICIT | No explicit check, but library handles |
| Restricted access errors | ✅ YES | Comprehensive detection |
| Network failure/timeout | ✅ YES | Timeout support, retry on 403 |
| File system errors | ✅ YES | Permission checks, disk errors |
| Actionable messages | ✅ YES | Good error context |
| Consistent exit codes | ✅ YES | Non-zero on failure |

**Strong Points:**
- Comprehensive restricted access detection (private, login, paywall, age-restricted, etc.)
- Automatic retry logic on 403 errors
- Interactive file overwrite prompts

---

## G) Performance & Robustness

### Status: ✅ GOOD

| Item | Status | Notes |
|------|--------|-------|
| Streaming I/O | ✅ YES | Uses `io.Copy`, no buffering of full files |
| Minimal memory usage | ✅ YES | Streams directly to disk |
| Parallel segment downloading | ⚠️ N/A | Handled by library |
| No busy-looping | ✅ YES | Event-driven progress manager |
| Progress updates throttled | ✅ YES | 200ms ticker in progress manager |
| Predictable concurrency scaling | ⚠️ LIMITED | Sequential download of playlist items |

**Code Evidence:**
```go
// progress_manager.go line 119
m.ticker = time.NewTicker(200 * time.Millisecond)
```

**Recommendation:**
- Consider parallel playlist downloads with configurable concurrency limit

---

## H) Security Requirements

### Status: ✅ PASS

| Item | Status | Notes |
|------|--------|-------|
| Sanitizes output paths | ✅ YES | `sanitize()` removes invalid chars |
| Prevents traversal/injection | ✅ YES | Uses `filepath.Join`, sanitizes names |
| Does not execute downloaded content | ✅ YES | Only writes files |
| Does not store credentials | ✅ YES | No auth code |
| Network requests use timeouts | ✅ YES | Configurable timeout |
| Safe retry strategy | ✅ YES | Single retry on 403 |

**Code Evidence:**
```go
// downloader.go line 601
func sanitize(name string) string {
    invalid := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
    clean := invalid.ReplaceAllString(name, "-")
    clean = strings.TrimSpace(clean)
    if clean == "" {
        return "video"
    }
    return clean
}
```

---

## I) Progress UI Integration (Hard Requirement)

### Status: ✅ EXCELLENT

| Item | Status | Notes |
|------|--------|-------|
| User-defined layouts supported | ✅ YES | `--progress-layout` flag |
| Configurable via CLI/config | ✅ YES | CLI flag present |
| Supports label, %, rate, ETA, bytes | ✅ YES | All fields implemented |
| Multiple progress bars simultaneously | ✅ YES | Map-based task tracking |
| Stable ordering | ✅ YES | Order slice maintains insertion order |
| No flicker/overlap/corruption | ✅ YES | ANSI escape codes for clean updates |
| Dynamic terminal resizing | ✅ YES | Signal handling + width updates |
| Detect width changes | ✅ YES | `terminalWidth()` on resize |
| Reflow without artifacts | ✅ YES | Bar width recalculation |
| Preserve alignment | ✅ YES | Padding and truncation logic |
| Interleaved logging | ✅ YES | Clears bars, logs, re-renders |
| Logs appear above bars | ✅ YES | Clear → log → render pattern |
| Rendering resumes after logs | ✅ YES | Event-driven architecture |
| Controlled refresh rate | ✅ YES | 200ms ticker |
| Minimal overhead | ✅ YES | Efficient string building |
| Graceful fallback non-TTY | ✅ YES | Checks `isTerminal()` and ANSI support |
| Completed bars persist | ✅ YES | Finished flag maintained |
| Renderer decoupled | ✅ YES | Event channel + progress manager |
| Renderer doesn't block I/O | ✅ YES | Separate goroutine |
| Compatible with parallel downloads | ✅ YES | Concurrent-safe task map |
| Compatible with resume | ⚠️ N/A | Resume not implemented |

**Outstanding Implementation:**
This is one of the strongest parts of the codebase. The progress manager is well-architected with:
- Event-driven design via channels
- Separate goroutine for rendering
- Terminal resize signal handling (Unix + Windows)
- User-customizable layouts
- Proper ANSI escape code usage
- Graceful degradation for non-TTY

**Test Coverage:**
```go
// Existing tests:
- TestProgressManagerMultipleBars ✅
- TestProgressManagerResizeEvent ✅
- TestProgressManagerLogging ✅
- TestProgressWriterNonTTYOutput ✅
```

---

## J) Video/Audio Metadata Collection & Playlist Metadata (Hard Requirement)

### Status: ❌ CRITICAL GAPS

| Item | Status | Notes |
|------|--------|-------|
| Collect `title` | ✅ YES | From YouTube API |
| Collect `artists[]` | ⚠️ PARTIAL | Uses `Author`, not array |
| Collect `album` | ⚠️ PARTIAL | YouTube Music only, not exported |
| Collect `track_number` / `disc_number` | ❌ NO | Not collected |
| Collect `release_date` / `year` | ❌ NO | Not collected |
| Collect `duration_seconds` | ✅ YES | Available in API response |
| Collect `thumbnail_url` | ❌ NO | Not collected |
| Collect `source_url` | ❌ NO | Not collected |
| Collect `source_id` | ⚠️ PARTIAL | Video ID available but not exported |
| Collect `extractor_name` + version | ❌ NO | Not tracked |
| **Playlist Metadata** | | |
| Collect `playlist_title` | ✅ YES | Available in code |
| Collect `playlist_id` | ✅ YES | Available in code |
| Collect `playlist_url` | ❌ NO | Not preserved |
| Collect `position` (1..N) | ⚠️ IMPLICIT | Index exists but not exported |
| Save ordering to manifest | ❌ NO | No `playlist.json` output |
| Robust when metadata scattered | ⚠️ PARTIAL | Some fallbacks exist |
| **Metadata Sources** | | |
| Platform APIs | ✅ YES | kkdai/youtube library |
| oEmbed | ❌ NO | Not used |
| Open Graph tags | ❌ NO | Not used |
| Manifest/container hints | ❌ NO | Not parsed |
| User overrides (--meta) | ❌ NO | Not implemented |
| **Graceful Failure** | | |
| Missing metadata doesn't crash | ✅ YES | Defaults to video ID |
| Null/empty for unavailable fields | ⚠️ PARTIAL | No structured output |
| Structured warnings | ❌ NO | No logging for missing fields |
| **Metadata Output** | | |
| Emit sidecar JSON per item | ❌ NO | **CRITICAL GAP** |
| Embed tags in audio files | ❌ NO | **CRITICAL GAP** |
| Sidecar JSON when embedding unsupported | ❌ NO | No sidecar at all |

**Critical Missing Features:**

### 1. No Sidecar Metadata Files
The tool does not generate `.info.json` files alongside downloads. This is required for:
- Reproducibility
- External tools integration
- Metadata preservation

### 2. No Metadata Embedding
Audio files (m4a, mp3) are downloaded without ID3/MP4 tags:
- No artist tags
- No album tags
- No track numbers
- No cover art

### 3. No Playlist Manifests
When downloading playlists, no `playlist.json` is created with:
- Playlist metadata
- Item ordering
- Per-item metadata

### 4. No User Metadata Overrides
Cannot specify custom metadata via CLI:
- `--meta title="Custom Title"`
- `--meta artist="Artist Name"`

**Code Evidence:**
```go
// downloader.go lines 649-702: printInfo() outputs to stdout
// But no persistent .info.json file is written during downloads
```

**Recommendation - High Priority:**

1. **Add sidecar JSON output**:
   ```go
   func writeSidecarMetadata(outputPath string, video *youtube.Video, meta musicEntryMeta) error {
       infoPath := outputPath + ".info.json"
       // Write comprehensive JSON
   }
   ```

2. **Add metadata embedding for audio**:
   ```go
   func embedAudioMetadata(filePath string, meta Metadata) error {
       // Use ID3/MP4 tagging library
   }
   ```

3. **Add playlist manifest**:
   ```go
   func writePlaylistManifest(playlist *youtube.Playlist, items []DownloadedItem) error {
       manifestPath := sanitize(playlist.Title) + ".playlist.json"
       // Write ordered list with metadata
   }
   ```

4. **Add `--meta` flag support**:
   ```go
   flag.Var(&metaOverrides, "meta", "metadata override (key=value)")
   ```

---

## K) Testing Requirements

### Status: ⚠️ PARTIAL

| Item | Status | Notes |
|------|--------|-------|
| Unit tests for URL parsing/validation | ✅ YES | `validation_test.go` |
| Unit tests for format selection | ❌ NO | Missing `TestSelectFormat` |
| Unit tests for DRM/encryption detection | ❌ NO | Not applicable (no detection code) |
| Unit tests for metadata parsing/normalization | ⚠️ PARTIAL | YouTube Music parsing untested |
| Unit tests for playlist ordering/manifest | ❌ NO | No manifest generation |
| Integration tests with public sources | ❌ NO | No integration tests |
| Progress UI tests for multiple bars | ✅ YES | Comprehensive |
| Progress UI tests for resize handling | ✅ YES | Good coverage |
| Progress UI tests for logging during progress | ✅ YES | Working |
| Progress UI tests for non-TTY behavior | ✅ YES | Tested |
| Tests validate stable output | ⚠️ PARTIAL | Some tests check strings |

**Missing Test Cases:**
1. `TestSelectFormat` - Format selection logic
2. `TestResolveOutputPath` - Template replacement
3. `TestSanitize` - Filename sanitization edge cases
4. `TestMusicPlaylistParsing` - YouTube Music metadata extraction
5. Integration tests with real YouTube URLs (mocked or fixture-based)

**Recommendation:**
```go
// Add to downloader_test.go
func TestSelectFormat(t *testing.T) { /* ... */ }
func TestResolveOutputPath(t *testing.T) { /* ... */ }
func TestSanitizeEdgeCases(t *testing.T) { /* ... */ }
```

---

## L) Documentation Requirements

### Status: ⚠️ NEEDS IMPROVEMENT

| Item | Status | Notes |
|------|--------|-------|
| Documents supported formats | ⚠️ IMPLICIT | Says "progressive formats" but not explicit list |
| Documents limitations | ⚠️ PARTIAL | Mentions no DASH muxing, but incomplete |
| Documents all CLI flags | ✅ YES | Good table in README |
| Includes examples | ✅ YES | Comprehensive examples |
| Includes legal/copyright notice | ⚠️ PARTIAL | MIT license present, but no usage guidelines |
| States non-goals explicitly | ❌ NO | **Missing** |
| Documents metadata behavior | ❌ NO | No metadata section |
| Which fields collected | ❌ NO | Not documented |
| Where they come from | ❌ NO | Not documented |
| When they may be missing | ❌ NO | Not documented |
| How overrides work | ❌ NO | No overrides implemented |
| Sidecar JSON schema | ❌ NO | No sidecar files |

**Required Documentation Additions:**

### 1. Non-Goals Section
```markdown
## Non-Goals / Limitations

This tool is designed for downloading **publicly accessible YouTube videos only**.

**Not Supported:**
- ❌ DRM-protected content (Widevine, PlayReady)
- ❌ Encrypted streams (HLS with AES-128, DASH with CENC)
- ❌ Login-required / members-only content
- ❌ Paywall-protected videos
- ❌ Non-YouTube platforms
- ❌ DASH muxing (audio+video combining)
- ❌ Subtitle downloads

**Legal Notice:**
Users are responsible for ensuring their use complies with YouTube's Terms of Service
and applicable copyright laws. This tool does not circumvent DRM or access controls.
```

### 2. Metadata Documentation
Currently missing entirely. Should document:
- Which metadata fields are collected
- Limitations of YouTube API metadata
- YouTube Music integration for album info
- Future plans for sidecar JSON files

---

## Acceptance Criteria Assessment

### AC-1: Public URL Download (Direct File) ✅ PASS
**Status**: WORKS for YouTube progressive formats

### AC-2: Public URL Download (HLS Unencrypted) ❌ FAIL
**Status**: NOT IMPLEMENTED - No HLS parsing

### AC-3: Public URL Download (DASH Unencrypted) ❌ FAIL
**Status**: NOT IMPLEMENTED - No DASH parsing

### AC-4: Restricted Content Detection ✅ PASS
**Status**: Good detection of login/paywall markers

### AC-5: DRM/Encrypted Stream Refusal ❌ FAIL
**Status**: NO DETECTION - Could attempt to download and fail

### AC-6: Format Enumeration ✅ PASS
**Status**: `--list-formats` works well

### AC-7: Quality Selection ❌ FAIL
**Status**: NO `--quality` FLAG - Always picks best

### AC-8: Resume Support ❌ FAIL
**Status**: NOT IMPLEMENTED

### AC-9: Multiple Concurrent Downloads ⚠️ PARTIAL
**Status**: Multiple URLs supported but downloaded sequentially

### AC-10: Progress Bars Render Correctly ✅ PASS
**Status**: Excellent implementation

### AC-11: User-Defined Progress Layout ✅ PASS
**Status**: `--progress-layout` flag works

### AC-12: Terminal Resize Handling ✅ PASS
**Status**: Signal-based resize detection works

### AC-13: Interleaved Logging With Progress ✅ PASS
**Status**: Clean clear-log-render pattern

### AC-14: Non-TTY Behavior ✅ PASS
**Status**: Graceful fallback implemented

### AC-15: JSON Output Mode Cleanliness ⚠️ NEEDS VERIFY
**Status**: `--info` outputs JSON but not tested for stderr cleanliness

### AC-16: Path Safety ✅ PASS
**Status**: Good sanitization

### AC-17: Tests & Docs Exist ⚠️ PARTIAL
**Status**: Some tests exist, docs incomplete

### AC-18: Metadata Collected When Available ❌ FAIL
**Status**: NO SIDECAR FILES

### AC-19: Playlist Ordering and Manifest ❌ FAIL
**Status**: NO PLAYLIST.JSON

### AC-20: Metadata Missing → Graceful Degradation ⚠️ PARTIAL
**Status**: Doesn't crash, but no structured warnings

### AC-21: User Metadata Overrides ❌ FAIL
**Status**: NOT IMPLEMENTED

### AC-22: Metadata Embedding Behavior ❌ FAIL
**Status**: NO TAG EMBEDDING

---

## Summary of Critical Gaps

### 🔴 High Priority (Blocking Issues)

1. **No Metadata Export** (J, AC-18, AC-22)
   - Missing sidecar `.info.json` files
   - No ID3/MP4 tag embedding for audio
   - No way to preserve metadata for downloaded content

2. **No Playlist Manifest** (J, AC-19)
   - Playlist ordering not preserved in machine-readable format
   - No `playlist.json` generation

3. **No DRM/Encryption Detection** (B, AC-5)
   - Could attempt downloads that will fail
   - No explicit error for encrypted content

4. **No HLS/DASH Support** (B, AC-2, AC-3)
   - Limited to YouTube progressive formats only
   - Cannot download many modern videos

### 🟡 Medium Priority (Feature Gaps)

5. **No Quality Selection** (D, AC-7)
   - `--quality` flag missing
   - Always downloads best (no user choice)

6. **No Resume Support** (D, AC-8)
   - Interrupted downloads must restart from zero

7. **No User Metadata Overrides** (J, AC-21)
   - `--meta` flag not implemented
   - Cannot correct or add metadata

8. **Documentation Gaps** (L)
   - Missing non-goals section
   - No metadata documentation
   - Missing legal/copyright usage guidelines

### 🟢 Low Priority (Nice to Have)

9. **Parallel Playlist Downloads** (G)
   - Currently sequential, could be faster

10. **Additional Test Coverage** (K)
    - Format selection tests
    - Metadata parsing tests
    - Integration tests

---

## Recommended Roadmap

### Phase 1: Metadata & Compliance (Critical)
**Estimated Effort**: 2-3 days

1. ✅ Fix test duplication (DONE)
2. ⚠️ Add sidecar JSON output (`.info.json`)
3. ⚠️ Add metadata embedding for audio files
4. ⚠️ Add playlist manifest generation
5. ⚠️ Update documentation with non-goals and legal notice
6. ⚠️ Add unit tests for metadata functions

### Phase 2: Enhanced Features (Important)
**Estimated Effort**: 3-4 days

1. Add `--quality` flag with format selection
2. Add `--meta` flag for metadata overrides
3. Add DRM/encryption detection with clear errors
4. Improve test coverage (format selection, sanitization)
5. Add `--json` mode that's stderr-clean

### Phase 3: Advanced Features (Optional)
**Estimated Effort**: 5-7 days

1. Implement resume support with partial file detection
2. Add parallel playlist downloads with concurrency limit
3. Add HLS/DASH parsing (major refactor)
4. Add integration tests with mocked YouTube responses

---

## Conclusion

The **ytdl-go** project has an **excellent foundation** with:
- ✅ Outstanding progress UI implementation
- ✅ Good error handling and security practices
- ✅ Clean, maintainable code structure

However, it has **critical gaps** in:
- ❌ Metadata collection and export (sidecar JSON, tag embedding)
- ❌ Playlist manifest generation
- ❌ DRM/encryption detection
- ❌ Format/quality selection options

**Recommendation**: 
- **If this is a YouTube-specific tool**: Update documentation to clarify scope, complete metadata features (Phase 1), and consider this READY with caveats
- **If this is a general-purpose tool**: Significant additional work needed (all phases) to meet the full checklist requirements

**Next Steps**:
1. Prioritize Phase 1 (metadata) as it's a hard requirement
2. Update documentation to clarify YouTube-only scope
3. Add missing tests for metadata parsing
4. Consider whether general-purpose support is actually needed

---

**Assessment Complete**: 2026-01-28
