**Product Requirement Document & Engineering Specification: ytdl-go Dashboard Evolution v2**\-----Part 1: Product Requirement Document (PRD) \- ytdl-go Dashboard Evolution v2

**Goal:** Evolve the `ytdl-go` dashboard into a fully modular, user-defined grid system with a strong, integrated 'Super Gopher' brand identity and critical new features.1. Dashboard Layout & Customization (Core Architecture)

| Component | Goal / Requirement |
| ----- | ----- |
| **Grid System** | Utilize a responsive grid layout (e.g., CSS Grid/Masonry) to allow for a fully modular, user-defined layout. |
| **"Edit Mode" Toggle** | A prominent "Edit" button (Pencil Icon) on the dashboard header. When active, modules must "wobble" or show **drag handles** and **X buttons** for removal. |
| **Add Module** | A "+" button allows users to re-add previously removed modules from a complete list. |
| **Module Persistence** | User preferences for layout and active modules must be saved to local storage or the user configuration file. |
| **Header Alignment** | Ensure the `YT_AUTH_OK` status indicator and Advanced Mode toggle are vertically aligned and visually balanced in the top navigation bar. |

2\. Branding & Mascot Integration (The Super Gopher)

**Goal:** Inject personality and replace generic icons with the custom Super Gopher branding.

* **Welcome Back Component:** Replace the generic "Circle with Exclamation Point" icon with the **Super Gopher** image. He will serve as the primary greeter.  
* **Empty States:** If the "Queue" or "Recent Additions" modules are empty, display a desaturated or semi-transparent version of the Gopher with the message: *"Nothing here yet\! Feed me URLs\!"*  
* **Success Animations:** When a download finishes successfully, a small, animated Gopher could pop up as a toast notification, giving a "thumbs up" or holding the play button.

3\. The "Direct Download" Module (Priority Feature)

**Goal:** Enable immediate download action without requiring navigation to the staging "Deck/Queue."

* **UI Component:** A standalone, high-visibility input field placed prominently (defaulting to the top of the grid).  
* **Function:** Accepts a pasted URL and a **Download Now** button. This action skips the "Deck/Queue" and initiates the download immediately.  
* **Visual Cue:** Must be styled distinctively (e.g., "Hot Input") to separate it from the global search bar.

4\. Library & Statistics Module

**Goal:** Provide granular, user-configurable control over displayed library data.

* **Configurable Stats:** The "Library Stats" section must be modular and toggleable via a Settings page.  
* **Toggleable Data Points:**  
  * Total Movies / Video Files  
  * Total Songs / Music Tracks  
  * Total Artists & Albums  
  * Total Channels  
  * Total Podcasts & Podcast Stations  
* **Default State:** This module must be **Disabled/Hidden** by default to prevent initial UI clutter.

5\. Enhanced "Recent Additions"

**Goal:** Improve visual discovery of newly added media.

* **Component:** Implement a **Carousel Component** (horizontal scrolling list of thumbnails).  
* **Controls:**  
  * Manual navigation via Left/Right arrows.  
  * Auto-Play functionality that cycles automatically but pauses on hover/user interaction.  
* **Action:** Clicking a thumbnail must play the media or open the file location.

6\. The "On Deck" System (Scheduled Downloads)

**Goal:** Support batch processing and scheduling of downloads.

* **The Deck:** A staging list for URLs.  
* **Workflow:** User adds URLs \-\> URLs sit in "Deck" \-\> User clicks **"Process All"** or schedules them for a later time (e.g., 2 AM).

7\. Media Management & Playlists

* **Playlist Module:** Implement full **CRUD** (Create, Read, Update, Delete) capabilities for custom user playlists.  
* **Global Search:** Implement a function to filter the local library by title, artist, or album.

8\. Integrated YouTube Portal (Experimental)

* **Concept:** Implement embedded YouTube browsing (e.g., an Iframe) to avoid tab-switching.  
* **Action:** Inject a distinct **"Download" overlay button** into the embedded view to capture the current video URL directly.

Implementation Priority for Engineering

1. **P0 (Critical):** Direct Download Input, Dashboard Grid/Edit Mode, Header Alignment Fixes.  
2. **P1 (High):** Super Gopher Asset Integration (Welcome & Empty States), Carousel for Recents.  
3. **P2 (Medium):** Library Stats (Backend counters), Playlist Management.  
4. **P3 (Future):** YouTube Portal (Iframe/Browser) research.

\-----Part 2: Engineering Task List \- ytdl-go UI/UX Refinement

**Project:** `ytdl-go` Dashboard Evolution

**Status:** Updated for Media Player & Sidebar Sprint1. Sidebar & Navigation Updates

* **Branding:** Replace the generic lightning bolt icon with a **Super Gopher silhouette**.  
* **Typography:** Apply a **gradient effect** to the "ytdl-go" brand text.  
* **Iconography:** Update the "Dashboard" icon to a **grid pattern**.  
* **Menu Structure:**  
  * Rename "New Download" to **"Download"** for clarity.  
  * Expand the **Media** section to include: History, Playlists, and Details.  
  * Retain **Settings** as a listed item under the **System** section for discoverability.  
  * Maintain the **Extensions** section as a distinct footer element.

2\. Media Player Technical Specifications

* **Performance:** Optimize the click-and-drag event listener; current implementation is laggy/slow during movement.  
* **State Management & Visibility:**  
  * **Conditional Rendering:** The player component must remain unpopulated (**hidden**) until a media item is explicitly selected.  
  * **Data Handling:** If metadata (Artist, Album, etc.) is missing, fields must be left **blank** rather than displaying "Unknown."  
  * **Title Logic:** Default the title to the **file name** if metadata is unavailable.  
* **Control Logic:**  
  * Ensure **Play, Next, and Queue** controls are visible in both minimized and maximized states.  
  * **"Unsupported Media Type" Warning:** Replace the current text banner with a clickable informational icon that provides error details on request.  
* **Window Management:**  
  * **Close (X) Button:** Fix the "X" button behavior: it should **close the player** when maximized. When minimized, the "X" should **restore the player to its full size**.  
  * **Resizing:** Implement edge-click resizing for the maximized view; it must no longer be a fixed-size container.

3\. General Aesthetics & UI Polish

* **Glassmorphism:** Enhance the "glassy" effect of all modules to align with the current theme.  
* **Visual Hierarchy:** Increase the definition/contrast between component borders and the background to ensure users can clearly tell where modules end.  
* **Constraint:** All changes must be **non-destructive** to existing dashboard functionality (Grid/Edit Mode).

## Pending Backend Features for Web UI
- [ ] **Cookie & PO Token Support:** The Web API (`/api/download`) does not currently accept `useCookies` or `poTokenExtension` options in the JSON payload (`WebOption` struct), causing "invalid JSON payload" errors if sent. These were temporarily removed from the frontend `startDownload` call to unblock Phase 1. They must be re-added to the frontend AND implemented in the backend `WebOption` struct and `parseDownloadRequest` function.