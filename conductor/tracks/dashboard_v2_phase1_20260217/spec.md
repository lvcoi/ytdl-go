# Track Specification: Dashboard Evolution v2 P0

## Overview
This track focuses on a significant overhaul of the Dashboard UI/UX, transforming it into a more modular, interactive, and user-friendly experience. The goal is to move away from static grids and hardcoded layouts towards a flexible widget-based system. Key improvements include a refactored "Quick Download" module, an enhanced "Active Downloads" widget with real-time progress for playlists/multi-URL inputs, a dynamic "Recent Activity" carousel, and a comprehensive onboarding experience for new users.

## Objectives
-   **Refactor Dashboard Layout:** Implement a modular grid system capable of hosting diverse widgets.
-   **Enhanced Quick Download:** Rename "Direct Download" to "Quick Download" and integrate it with the backend's robust CLI-like download routine.
-   **Advanced Active Downloads:** Support granular progress tracking for playlists and multi-URL inputs, displaying individual items and their download status.
-   **Interactive Recent Activity:** Implement a carousel for recent downloads with a "center-focused" design and direct media playback integration.
-   **Smart Welcome/Onboarding:** Create a "Welcome" widget that guides new users through the interface features and transitions to a "Welcome Back" status widget for returning users.
-   **Library Stats Refinement:** Update statistics to be media-centric (Music, Podcasts, Videos) rather than creator-centric.

## Functional Requirements

### 1. Quick Download Widget (formerly Direct Download)
-   **Input:** Single URL input field supporting individual videos, playlists, or multiple URLs (comma-separated).
-   **Action:** Trigger a backend Go routine identical to the CLI command `ytdl-go -o media/{artist}-{album}-{title}.{ext} ...`.
-   **Feedback:**
    -   Immediately add the download task(s) to the "Active Downloads" widget.
    -   Show a toast/modal notification confirming the download start.
    -   Provide a link/button in the toast to navigate directly to the full "Downloads" page.

### 2. Active Downloads Widget
-   **List View:** Display a list of currently active download tasks.
-   **Playlist/Multi-URL Handling:**
    -   When a playlist or multi-URL input is processed, *expand* the list to show individual media items (e.g., songs in a playlist) as they are identified.
    -   Do not just show a single "Playlist" entry; show the *contents* of the download job.
-   **Progress Tracking:**
    -   Display the title of the media being downloaded.
    -   Show real-time progress percentage (e.g., "45%") to the right of the title.
    -   **Completion:** Automatically remove the item from the list/widget once the download is 100% complete.

### 3. Recent Activity Carousel
-   **Layout:** A horizontal carousel displaying thumbnail cards of recently downloaded media.
-   **Visual Style:**
    -   Display approximately 3 items at a time.
    -   **Center Item:** Larger, more prominent, and centered.
    -   **Side Items:** Smaller, partially visible or faded, creating a 3D/depth effect.
    -   **Navigation:** Left/Right arrow buttons to manually scroll through the history (last 10-15 items).
    -   **Auto-Scroll:** Automatically cycle through items when idle (e.g., every 5 seconds).
-   **Content:**
    -   Thumbnail image.
    -   Media Title below the thumbnail.
-   **Interaction:** Clicking a card plays the media in the integrated player (existing component).

### 4. Welcome & Onboarding Widget
-   **New User State ("Welcome"):**
    -   Triggered on the very first visit (tracked via local storage/cookie).
    -   **Tutorial Mode:**
        -   A "Start Tour" button triggers a modal overlay that dims the rest of the screen.
        -   **Step-by-Step:** Highlights specific widgets (Quick Download, Active Downloads, Library, Settings) one by one.
        -   **Tooltips:** specific "Tooltip" components explain the function of each highlighted area.
        -   **Controls:** "Next", "Previous", and a clear "Skip Tutorial" button.
-   **Returning User State ("Welcome Back"):**
    -   Displayed after the tutorial is completed or skipped.
    -   **Visuals:** Features the "Super Gopher" image (from README) instead of a generic icon.
    -   **Content:** Personalized greeting (e.g., "Welcome back! You have X new downloads...").
-   **Manual Access:** A generic "?" circular icon/button in the lower-right corner to restart the tutorial at any time.

### 5. Library Stats Widget
-   **Metrics:** Display total counts for specific media types:
    -   Total Music Tracks
    -   Total Podcasts/Episodes
    -   Total Videos
    -   Total Recent Downloads (e.g., last 24h or 7 days)
-   **Removal:** Remove "Creator" or "Artist" specific counts as they are ambiguous across different media types.

## Non-Functional Requirements
-   **Performance:** Carousel animations must be smooth (60fps) using CSS transitions/transforms where possible, driven by SolidJS state.
-   **Responsiveness:** Layout must adapt to different screen sizes (though specific mobile layout is secondary to desktop for this phase).
-   **Modularity:** Code structure should separate widgets into distinct components (`QuickDownloadWidget`, `ActiveDownloadsWidget`, `CarouselWidget`, `WelcomeWidget`) for easier maintenance and future reordering capabilities.

## Out of Scope
-   Drag-and-drop widget reordering (Phase 2).
-   Resizing widgets (Phase 2).
-   Saving multiple dashboard layouts (Phase 2).
-   Major refactoring of the Media Player itself (focus is on integration).

## Acceptance Criteria
-   [ ] Quick Download triggers the correct backend routine and updates the Active Downloads list.
-   [ ] Active Downloads correctly parses and displays individual items from playlists/multi-URL inputs.
-   [ ] Active Downloads items show accurate progress and disappear upon completion.
-   [ ] Recent Activity Carousel displays 3 items with a prominent center, supports manual/auto navigation, and plays media on click.
-   [ ] New users see the "Welcome" widget and Tutorial overlay; returning users see the "Welcome Back" widget with the Gopher image.
-   [ ] Library Stats accurately reflect the counts of Music, Podcasts, and Videos.
