# Implementation Plan: Dashboard Evolution v2 P0

## Phase 1: Header Alignment and Direct Download Module
- [x] Task: Fix Header Alignment for YT Auth Status and Advanced Mode.
    - [x] Write tests for Header component alignment.
    - [x] Adjust CSS/Layout in `Header.jsx` to ensure vertical alignment.
- [x] Task: Implement Direct Download Component with Robust State Management.
    - [x] Write tests for `DirectDownload` component.
    - [x] Create `DirectDownload.jsx` component.
    - [x] Integrate into `DashboardView.jsx` using `batch()` for all status updates.
- [x] Task: Ensure Library Sync and Navigation Stability.
    - [x] Optimize `useDownloadManager.js` with `batch()` for SSE updates.
    - [x] Refine terminal status mapping ('done' -> 'complete').
    - [x] Verify library sync triggers on terminal job completion.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Header Alignment and Direct Download Module'


## Phase 2: Modular Grid and Edit Mode
- [ ] Task: Refactor Dashboard into Modular Grid.
- [ ] Task: Implement Edit Mode Toggle and UI.
- [ ] Task: Implement Module Removal and Re-adding.

## Phase 3: Layout Persistence and Drag-and-Drop
- [ ] Task: Implement Layout Persistence.
- [ ] Task: Implement Drag-and-Drop Reordering.
