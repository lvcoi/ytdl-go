# Implementation Plan: WebUI Overhaul & Restoration (Feb 2026)

### ## Approach
The strategy is to move from **Critical Functional Fixes** (Media Player, Backend API) to **UX/Branding Polish** (Super Gopher, Dashboard Edit Mode) and finally **View Enhancements** (Library & Download carousels).

### ## Task List

#### Phase 1: Critical Fixes
- [ ] **Player:** Fix "X" button behavior (Restore vs Close).
- [ ] **Player:** Improve media type/extension detection for broader codec support.
- [ ] **Backend:** Add `UseCookies` and `PoToken` to `WebOption` in `server.go`.
- [ ] **Frontend:** Ensure `DownloadView` settings are correctly passed to the API.

#### Phase 2: Dashboard & Branding
- [ ] **Branding:** Replace Lightning icon with Super Gopher silhouette.
- [ ] **Branding:** Apply gradient typography to "ytdl-go" header.
- [ ] **Modularity:** Implement Dashboard "Edit Mode" (wobble/drag handles).
- [ ] **Persistence:** Save widget layout to `localStorage`.
- [ ] **Widgets:** Replace "Welcome" icon with Super Gopher asset.

#### Phase 3: Library & Media Management
- [ ] **UX:** Implement horizontal Carousel for "Recent Additions".
- [ ] **Stats:** Make Library Statistics configurable/toggleable.
- [ ] **Navigation:** Expand Media section to include distinct History and Playlists views.

### ## Timeline
Estimated Total: 5.5 Hours.
