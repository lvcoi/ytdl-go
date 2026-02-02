# Refactor `internal/downloader/` for Maintainability

Split the 2182-line `downloader.go` monolith into focused modules while keeping it as a thin orchestrator.

---

## Completed Refactor Checklist

### Phase 1: Low-Risk Extractions ✅

- [x] Extract `url.go` - pure functions, no deps on other downloader code
- [x] Extract `path.go` - pure functions, only uses `youtube.Format`
- [x] Extract `http.go` - transport/client setup, no internal deps

### Phase 2: State-Carrying Extractions ✅

- [x] Extract `prompt.go` - has global state, careful with mutex
- [x] Extract `output.go` - JSON output, depends on structs

### Phase 3: Core Logic Extractions ✅

- [x] Extract `format.go` - format selection logic
- [x] Extract `music.go` - YouTube Music API (largest single-concern block)
- [x] Extract `youtube.go` - client, download core

### Phase 4: Playlist ✅

- [x] Extract `playlist.go` - playlist processing

### Phase 5: Cleanup ✅

- [x] Slim `downloader.go` to orchestrator (~265 lines)
- [x] Update imports, verify no cycles
- [x] All existing functionality preserved

---

## Current State (Post-Refactor)

| File                    | Lines | Concern                                |
|-------------------------|-------|----------------------------------------|
| `downloader.go`         | 265   | ✅ Thin orchestrator                   |
| `youtube.go`            | 314   | ✅ YouTube client & download core      |
| `music.go`              | 460   | ✅ YouTube Music API                   |
| `playlist.go`           | 334   | ✅ Playlist processing                 |
| `format.go`             | 238   | ✅ Format selection logic              |
| `format_selector.go`    | 468   | ✅ Interactive format picker (TUI)     |
| `path.go`               | 136   | ✅ Output path resolution              |
| `http.go`               | 75    | ✅ HTTP transport setup + user agents  |
| `output.go`             | 254   | ✅ JSON output + file validation       |
| `prompt.go`             | 178   | ✅ Duplicate file handling             |
| `url.go`                | 126   | ✅ URL validation & normalization      |
| `errors.go`             | 79    | ✅ Error categories & exit codes       |
| `printer.go`            | 373   | ✅ Console output formatting           |
| `progress.go`           | 160   | ✅ Progress writer                     |
| `progress_manager.go`   | 720   | ✅ TUI progress manager                |
| `adaptive.go`           | 391   | ✅ HLS/DASH streaming                  |
| `dash.go`               | 469   | ✅ DASH manifest parsing               |
| `direct.go`             | 317   | ✅ Non-YouTube URL downloads           |
| `segments.go`           | 189   | ✅ HLS/DASH segment parsing            |
| `segment_downloader.go` | 188   | ✅ Parallel segment downloads          |
| `metadata.go`           | 224   | ✅ Item metadata & sidecar             |
| `metadata_web.go`       | 128   | ✅ Web page metadata extraction        |
| `tags.go`               | 53    | ✅ ID3 tag embedding                   |
| `utils.go`              | 25    | ✅ Duration formatting                 |

| `unified_tui.go`        | 469   | ⏳ Seamless format→download TUI (pending integration) |

**Total: 25 files, well-organized by concern**

---

## Cleanup TODO ✅

### Dead Code Removal ✅

- [x] **Remove unused error sentinels in `segments.go`**
  - `ErrEncryptedHLS` and `ErrEncryptedDASH` removed

### Potential Consolidation ✅

- [x] **Merged `output.go` + `output_validate.go`**
  - Combined into single `output.go` (~254 lines)

- [ ] **Consider merging `progress.go` + `progress_manager.go`**
  - Skipped: `progress_manager.go` is already 720 lines

### Code Quality ✅

- [x] **Consolidated `stringsOrFallback` and `firstNonEmpty`**
  - Kept `stringsOrFallback` in `metadata.go`
  - Updated `metadata_web.go` to use it

- [x] **Moved `musicUserAgent` to `http.go`**
  - All user agent constants now in `http.go`

### Pending Integration

- [ ] **Integrate `unified_tui.go`** — SeamlessTUI for format selection → download transition
  - Provides seamless viewport transition without terminal flash
  - Not yet wired into main flow

---

## Risks & Mitigations (Original)

| Risk                                  | Mitigation                                                                         |
|---------------------------------------|------------------------------------------------------------------------------------|
| Circular imports                      | Extract pure functions first; `playlist.go` imports `youtube.go`, not vice versa   |
| Global state (`globalDuplicateAction`)| Keep in `prompt.go`, accessed via exported functions                               |
| Breaking tests                        | Each phase: extract → compile → test before next                                   |
| Merge conflicts                       | Small, focused PRs per phase if preferred                                          |

---

## Summary

**Refactor Status: ✅ COMPLETE**

The original 2182-line monolith has been successfully split into 24 focused modules.

**Cleanup completed:**
- ✅ Removed unused error sentinels (`ErrEncryptedHLS`, `ErrEncryptedDASH`)
- ✅ Merged `output.go` + `output_validate.go`
- ✅ Consolidated `stringsOrFallback`/`firstNonEmpty` into single function
- ✅ Moved `musicUserAgent` to `http.go`

**Pending:**
- ⏳ Integrate `unified_tui.go` (SeamlessTUI) into `-list-formats` flow
