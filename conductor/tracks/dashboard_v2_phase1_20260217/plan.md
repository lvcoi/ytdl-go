# Implementation Plan: Dashboard Evolution v2 P0

## Phase 1: Header Alignment and Quick Download Module
- [x] Task: Fix Header Alignment for YT Auth Status and Advanced Mode. [commit: ff325dd]
    - [x] Write tests for Header component alignment.
    - [x] Adjust CSS/Layout in `Header.jsx` to ensure vertical alignment.
- [x] Task: Implement Quick Download Component with Robust State Management. [commit: 4208ded]
    - [x] Write tests for `QuickDownload` component (simulating inputs and backend calls).
    - [x] Create `QuickDownload.jsx` component using `solid-component-builder` principles.
    - [x] Integrate into `DashboardView.jsx` using `batch()` for all status updates.
- [x] Task: Ensure Library Sync and Navigation Stability. [commit: c50e22d]
    - [x] Optimize `useDownloadManager.js` with `batch()` for SSE updates.
    - [x] Refine terminal status mapping ('done' -> 'complete').
    - [x] Verify library sync triggers on terminal job completion.
- [~] Task: Conductor - User Manual Verification 'Phase 1: Header Alignment and Quick Download Module' (Protocol in workflow.md)

## Phase 2: Modular Grid and Edit Mode
- [ ] Task: Refactor Dashboard into Modular Grid.
    - [ ] Write tests for `DashboardView` grid layout structure.
    - [ ] Refactor `DashboardView.jsx` to use a flexible grid container.
- [ ] Task: Implement Edit Mode Toggle and UI.
    - [ ] Write tests for "Edit Mode" state toggling.
    - [ ] Create "Edit Mode" toggle button in Header/Settings.
    - [ ] Implement visual indicators for "Edit Mode" (e.g., dashed borders, handles).
- [ ] Task: Implement Module Removal and Re-adding.
    - [ ] Write tests for removing and adding widgets to the grid.
    - [ ] Implement removal logic (removing component from grid state).
    - [ ] Create "Add Widget" drawer/modal.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Modular Grid and Edit Mode' (Protocol in workflow.md)

## Phase 3: Layout Persistence and Drag-and-Drop
- [ ] Task: Implement Layout Persistence.
    - [ ] Write tests for saving/loading layout from LocalStorage.
    - [ ] Integrate `createStore` for layout state management.
- [ ] Task: Implement Drag-and-Drop Reordering.
    - [ ] Write tests for drag interactions (using `solid-dnd` or similar if available/approved, or custom logic).
    - [ ] Implement drag handlers for grid items in Edit Mode.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Layout Persistence and Drag-and-Drop' (Protocol in workflow.md)

## Phase 4: Recent Activity Carousel
- [ ] Task: Implement Custom Carousel Component.
    - [ ] Write tests for Carousel logic (index navigation, circular buffer).
    - [ ] Implement `CarouselWidget.jsx` using SolidJS signals for active index and animation state.
    - [ ] Implement "Center Focus" scaling logic.
- [ ] Task: Integrate Media Player with Carousel.
    - [ ] Write tests for click-to-play interaction.
    - [ ] Connect Carousel card click events to the global Player store.
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Recent Activity Carousel' (Protocol in workflow.md)

## Phase 5: Welcome Widget & Onboarding
- [ ] Task: Implement Welcome Widget Logic.
    - [ ] Write tests for "New User" vs "Returning User" state detection.
    - [ ] Implement `WelcomeWidget.jsx` with conditional rendering.
- [ ] Task: Implement Tutorial Overlay.
    - [ ] Write tests for Tutorial flow (next, prev, skip).
    - [ ] Create `TutorialOverlay.jsx` with dimming backdrop and highlight zones.
- [ ] Task: Conductor - User Manual Verification 'Phase 5: Welcome Widget & Onboarding' (Protocol in workflow.md)

## Phase 6: Library Stats Refinement
- [ ] Task: Refactor Library Stats.
    - [ ] Write tests for new stats aggregation (Music vs Podcast vs Video).
    - [ ] Update `StatsWidget.jsx` to display new metrics.
- [ ] Task: Conductor - User Manual Verification 'Phase 6: Library Stats Refinement' (Protocol in workflow.md)
