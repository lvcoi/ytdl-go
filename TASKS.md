# Task List

## Comments

- Execute one PR at a time.
- After each PR, stop for review before starting the next PR.
- Keep this file as the source of truth for scope and status.

## Status Legend

- ✅ Completed and committed
- 🟡 Implemented locally (not committed)
- ⏳ Pending

## Current Branch Snapshot

- Branch: `web-ui`
- Last committed PR work: `PR10` (`921d5f3`)
- Latest commit on branch: Update TASKS.md to reflect PR10 completion and add feature implementation planning to PR12 scope (`3c66346`)
- Local in-progress changes: PR11 backend + docs updates (uncommitted)

## Recent Non-PR Work

- Documentation structure + consistency correction completed and committed (`33d6bb0`):
  - Restored frontend-owned docs under `frontend/docs/`
  - Rewrote root `docs/` files as project-wide
  - Updated frontend architecture API endpoint docs to current backend routes

## PR Roadmap

1. **PR0: Dev API Port Fix (Immediate blocker)**
   - Covers: `9` (dev config part)
   - Scope: make Vite proxy target configurable via env (`VITE_API_PROXY_TARGET`), default to `http://127.0.0.1:8080`, update docs/scripts.
   - Status: ✅ Completed and committed (`3071431`)

2. **PR1: App State Foundation + Persistence**
   - Covers: `3`
   - Scope: central app store, persist key UI state across tab/page switches, keep download state alive when navigating.
   - Status: ✅ Completed and committed (`e9cfbbc`)

3. **PR2: Backend Progress/Event Model Upgrade**
   - Covers: `1a`, foundation for `1b` and `6`
   - Scope: SSE includes real playlist totals and active-item metadata (title/artist/thumbnail/source/index/count).
   - Status: ✅ Completed and committed (`7b28aab`)

4. **PR3: Download UI Upgrade**
   - Covers: `1b`, part of `2`
   - Scope: rich progress bar (thumbnail/title/artist), accurate playlist progress, expandable active-download list UI.
   - Status: ✅ Completed and committed (`6f69f8a`)

5. **PR4: Download-to-Library Navigation + Auto Sync**
   - Covers: rest of `2`, `4`
   - Scope: click active download item to jump to matching Library entry; auto refresh Library after completed downloads.
   - Status: ✅ Completed and committed (`191b97b`)

6. **PR5: Unified Metadata API + Media Folder Layout Prep**
   - Covers: `6`, groundwork for `5` and comment `#4`
   - Scope: backend reads sidecar metadata and returns consistent fields; establish/normalize `media/{audio,video,playlist,data}` usage.
   - Status: ✅ Completed and committed (`b8c2182`)

7. **PR6: Library Tabs + Filtering/Sorting**
   - Covers: core of `5`
   - Scope: Video/Audio sub-tabs; filter/sort by Artist/Creator, Album/Channel, Playlist.
   - Status: ✅ Completed and committed (`f048529`)

8. **PR7: Saved Playlists (Fast Frontend Phase)**
   - Covers: `5` (phase A)
   - Scope: localStorage playlist CRUD and assignment in Library.
   - Status: ✅ Completed and committed

9. **PR8: Saved Playlists Backend Persistence + Migration**
   - Covers: `5` (phase B)
   - Scope: backend playlist storage in `media/data`, API endpoints, one-time migration from localStorage, and follow-up robustness fixes from review.
   - Status: ✅ Completed and committed (`6d068fa`)

10. **PR9: Player Visual Refresh**
    - Covers: `7`
    - Scope: replace generic icon with thumbnail on left, unify metadata presentation.
    - Status: ✅ Completed and committed (`e7d0009`)

11. **PR10: Player Window Controls + Minimized Mode**
    - Covers: `8`
    - Scope: close/minimize buttons, docked mini bar with Play/Next/Queue, draggable full player (disabled when minimized).
    - Status: ✅ Completed and committed (`921d5f3`)

12. **PR11: Port Bind Failure Handling**
    - Covers: `10`, final part of `9`
    - Scope: backend auto-tries alternate ports on bind failure, clear startup/error messaging, dev flow guidance.
    - Status: 🟡 Implemented locally (not committed)

13. **PR12: Inactive Feature Feedback**
    - Covers: `12`
    - Scope: disabled visual states + "Coming Soon" tooltip for non-wired controls. Make a plan for implementing the disconnected features in the same task list format as this TASKS.md document.
    - Status: ⏳ Pending
