# Implementation Plan: Dashboard Evolution v2 P0

## Phase 1: Header Alignment and Direct Download Module
- [x] Task: Fix Header Alignment for YT Auth Status and Advanced Mode. d38f05f
    - [x] Write tests for Header component alignment.
    - [x] Adjust CSS/Layout in `Header.jsx` to ensure vertical alignment.
- [ ] Task: Implement Direct Download Component.
    - [ ] Write tests for `DirectDownload` component (URL input, button click, validation).
    - [ ] Create `DirectDownload.jsx` component with "Hot Input" styling.
    - [ ] Integrate `DirectDownload` into `DashboardView.jsx`.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Header Alignment and Direct Download Module' (Protocol in workflow.md)

## Phase 2: Modular Grid and Edit Mode
- [ ] Task: Refactor Dashboard into Modular Grid.
    - [ ] Write tests for `Grid` and `GridItem` modularity.
    - [ ] Update `DashboardView.jsx` to use a dynamic list of modules.
- [ ] Task: Implement Edit Mode Toggle and UI.
    - [ ] Write tests for Edit Mode state and UI changes (wobble, handles, remove buttons).
    - [ ] Add Edit toggle to Header/Dashboard.
    - [ ] Implement "wobble" animation and visibility of edit controls.
- [ ] Task: Implement Module Removal and Re-adding.
    - [ ] Write tests for removing and adding modules from the registry.
    - [ ] Implement module removal logic.
    - [ ] Implement "+" menu for adding available modules.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Modular Grid and Edit Mode' (Protocol in workflow.md)

## Phase 3: Layout Persistence and Drag-and-Drop
- [ ] Task: Implement Layout Persistence.
    - [ ] Write tests for `localStorage` saving/loading of layout.
    - [ ] Implement layout sync with `localStorage`.
- [ ] Task: Implement Drag-and-Drop Reordering (Optional/Bonus if time permits, or basic reorder).
    - [ ] Research and implement basic reordering mechanism.
    - [ ] Write tests for reordering logic.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Layout Persistence and Drag-and-Drop' (Protocol in workflow.md)
