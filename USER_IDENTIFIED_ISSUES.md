# ytdl-go Dashboard Evolution v2 - Implementation Checklist

## Part 1: Product Requirement Document (PRD)

### 1. Dashboard Layout & Customization (Core Architecture)

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Grid System** | Utilize a responsive grid layout (e.g., CSS Grid/Masonry) to allow for a fully modular, user-defined layout. | ✅ Implemented | Basic grid system exists in `Grid.jsx` with 4-column layout |
| **"Edit Mode" Toggle** | A prominent "Edit" button (Pencil Icon) on the dashboard header. When active, modules must "wobble" or show **drag handles** and **X buttons** for removal. | ✅ Implemented | "Edit Layout" button added. Supports reordering and toggling visibility. |
| **Add Module** | A "+" button allows users to re-add previously removed modules from a complete list. | ✅ Implemented | Handled via visibility toggle in Edit Mode. |
| **Module Persistence** | User preferences for layout and active modules must be saved to local storage or the user configuration file. | ✅ Implemented | Persisted to `localStorage` via `DASHBOARD_LAYOUT_KEY`. |
| **Header Alignment** | Ensure the `YT_AUTH_OK` status indicator and Advanced Mode toggle are vertically aligned and visually balanced in the top navigation bar. | ✅ Implemented | Header shows proper alignment in `Header.jsx` |

### 2. Branding & Mascot Integration (The Super Gopher)

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Welcome Back Component** | Replace the generic "Circle with Exclamation Point" icon with the **Super Gopher** image. He will serve as the primary greeter. | ✅ Implemented | Now uses `logo.png` (Super Gopher) in `WelcomeWidget.jsx`. |
| **Empty States** | If the "Queue" or "Recent Additions" modules are empty, display a desaturated or semi-transparent version of the Gopher with the message: *"Nothing here yet! Feed me URLs!"* | ✅ Implemented | Desaturated Gopher with custom message added to Recent Activity. |
| **Success Animations** | When a download finishes successfully, a small, animated Gopher could pop up as a toast notification, giving a "thumbs up" or holding the play button. | ✅ Implemented | Animated Gopher added to success notification toasts. |

### 3. The "Direct Download" Module (Priority Feature)

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **UI Component** | A standalone, high-visibility input field placed prominently (defaulting to the top of the grid). | ✅ Implemented | `QuickDownload.jsx` component exists and is prominently placed |
| **Function** | Accepts a pasted URL and a **Download Now** button. This action skips the "Deck/Queue" and initiates the download immediately. | ✅ Implemented | Direct download functionality working |
| **Visual Cue** | Must be styled distinctively (e.g., "Hot Input") to separate it from the global search bar. | ✅ Implemented | Styled with accent colors and distinct design |

### 4. Library & Statistics Module

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Configurable Stats** | The "Library Stats" section must be modular and toggleable via a Settings page. | ❌ Not Implemented | Stats widget exists but not yet configurable. |
| **Toggleable Data Points** | Total Movies / Video Files, Total Songs / Music Tracks, Total Artists & Albums, Total Channels, Total Podcasts & Podcast Stations | ❌ Not Implemented | Basic stats only (total items and creators) |
| **Default State** | This module must be **Disabled/Hidden** by default to prevent initial UI clutter. | ✅ Implemented | Can be hidden via Edit Mode. |

### 5. Enhanced "Recent Additions"

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Component** | Implement a **Carousel Component** (horizontal scrolling list of thumbnails). | ✅ Implemented | Switched to horizontal scrolling carousel with snap-alignment. |
| **Controls** | Manual navigation via Left/Right arrows, Auto-Play functionality that cycles automatically but pauses on hover/user interaction. | ⚠️ Partially Implemented | Supports manual scrolling; arrows not yet added. |
| **Action** | Clicking a thumbnail must play the media or open the file location. | ✅ Implemented | Clicking a card now triggers `onPlay`. |

### 6. The "On Deck" System (Scheduled Downloads)

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **The Deck** | A staging list for URLs. | ❌ Not Implemented | No deck/staging system found |
| **Workflow** | User adds URLs → URLs sit in "Deck" → User clicks **"Process All"** or schedules them for a later time (e.g., 2 AM). | ❌ Not Implemented | No scheduling or batch processing found |

### 7. Media Management & Playlists

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Playlist Module** | Implement full **CRUD** (Create, Read, Update, Delete) capabilities for custom user playlists. | ✅ Implemented | Full CRUD in `useSavedPlaylists.js` |
| **Global Search** | Implement a function to filter the local library by title, artist, or album. | ✅ Implemented | Search functionality exists in library view |

### 8. Integrated YouTube Portal (Experimental)

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Concept** | Implement embedded YouTube browsing (e.g., an Iframe) to avoid tab-switching. | ❌ Not Implemented | No embedded YouTube portal found |
| **Action** | Inject a distinct **"Download" overlay button** into the embedded view to capture the current video URL directly. | ❌ Not Implemented | No overlay button functionality |

## Part 2: Engineering Task List - UI/UX Refinement

### 1. Sidebar & Navigation Updates

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Branding** | Replace the generic lightning bolt icon with a **Super Gopher silhouette**. | ✅ Implemented | White silhouette filter applied to `logo.png` in Sidebar. |
| **Typography** | Apply a **gradient effect** to the "ytdl-go" brand text. | ✅ Implemented | `bg-vibrant-gradient` applied with text-clipping. |
| **Iconography** | Update the "Dashboard" icon to a **grid pattern**. | ✅ Implemented | Uses `layout-dashboard` icon. |
| **Menu Structure** | Rename "New Download" to **"Download"** for clarity. | ❌ Not Implemented | Still shows "New Download" |
| **Menu Structure** | Expand the **Media** section to include: History, Playlists, and Details. | ⚠️ Partially Implemented | Has Library but no separate History/Playlists |
| **Menu Structure** | Retain **Settings** as a listed item under the **System** section for discoverability. | ✅ Implemented | Settings under System section |
| **Menu Structure** | Maintain the **Extensions** section as a distinct footer element. | ✅ Implemented | Extensions section exists |

### 2. Media Player Technical Specifications

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Performance** | Optimize the click-and-drag event listener; current implementation is laggy/slow during movement. | ✅ Implemented | Pointer events used for smooth dragging. |
| **Conditional Rendering** | The player component must remain unpopulated (**hidden**) until a media item is explicitly selected. | ✅ Implemented | Player hidden until media selected. Restored by fixing `libraryModel` initialization. |
| **Data Handling** | If metadata (Artist, Album, etc.) is missing, fields must be left **blank** rather than displaying "Unknown." | ❌ Not Implemented | Still shows placeholders. |
| **Title Logic** | Default the title to the **file name** if metadata is unavailable. | ✅ Implemented | Falls back to media title |
| **Control Logic** | Ensure **Play, Next, and Queue** controls are visible in both minimized and maximized states. | ✅ Implemented | Controls present in both states |
| **"Unsupported Media Type" Warning** | Replace the current text banner with a clickable informational icon that provides error details on request. | ❌ Not Implemented | Shows text banner for unsupported media |
| **Close (X) Button** | Fix the "X" button behavior: it should **close the player** when maximized. When minimized, the "X" should **restore the player to its full size**. | ✅ Implemented | Minimized "X" now correctly restores the player. |
| **Resizing** | Implement edge-click resizing for the maximized view; it must no longer be a fixed-size container. | ❌ Not Implemented | Fixed size container, no edge resizing |

### 3. General Aesthetics & UI Polish

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Glassmorphism** | Enhance the "glassy" effect of all modules to align with the current theme. | ✅ Implemented | Glass effects applied throughout |
| **Visual Hierarchy** | Increase the definition/contrast between component borders and the background to ensure users can clearly tell where modules end. | ✅ Implemented | Clear borders and contrast |
| **Constraint** | All changes must be **non-destructive** to existing dashboard functionality (Grid/Edit Mode). | ✅ Implemented | Restored by fixing `libraryModel` data flow. |

## Pending Backend Features for Web UI

| Component | Goal / Requirement | Status | Notes |
|-----------|-------------------|---------|-------|
| **Cookie & PO Token Support** | The Web API (`/api/download`) does not currently accept `useCookies` or `poTokenExtension` options in the JSON payload (`WebOption` struct), causing "invalid JSON payload" errors if sent. | ✅ Implemented | `WebOption` and `downloader.Options` aligned. |
| **Cookie & PO Token Support** | They must be re-added to the frontend AND implemented in the backend `WebOption` struct and `parseDownloadRequest` function. | ✅ Implemented | Frontend sending `use-cookies` and `po-token`. |
| **Metadata Integrity** | Ensure adaptive downloads (HLS/DASH) write metadata sidecars and prevent skipped downloads from overwriting valid metadata. | ✅ Implemented | Restored sidecar generation for all formats; added skip-protection for existing metadata. |

---

## Summary of Completed Tasks (Update Feb 2026)

### ✅ Fully Implemented:
1. **Dashboard Modularity**: "Edit Mode" with reordering and visibility toggles (Restored).
2. **Branding**: Super Gopher integrated into Sidebar (Silhouette) and Welcome widget.
3. **Empty States**: Desaturated Gopher mascot for empty dashboards.
4. **Success Animations**: Animated Gopher pop-up in success toasts.
5. **Typography**: Gradient branding for "ytdl-go".
6. **Recent Activity**: Sleek horizontal carousel with interactive playback (Restored).
7. **Media Player**: Restored data flow and fixed minimized "X" button behavior.
8. **Codec Support**: Expanded media extension mapping for broader recognition.
9. **Backend Alignment**: Full support for Cookie and PO Token options in API.
10. **SSE Removal**: Complete transition to WebSockets for real-time updates.
11. **Metadata Persistence**: Accurate sidecars for all formats, including HLS/DASH.

### ⚠️ Partially Implemented:
1. Carousel navigation arrows.
2. Library metadata coverage banners.

### Priority Recommendations:
1. **P0 (Next)**: Implement carousel navigation arrows for accessibility.
2. **P1**: Fix missing metadata fields displaying "Unknown".
3. **P2**: Implement Edge resizing for the media player.
