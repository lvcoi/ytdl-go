# Web UI Implementation Task List

## Status Legend

- ✅ Completed and committed
- 🟡 Implemented locally (not committed)
- ⏳ Pending

## PR12 Implementation Task List

1. **PR12-A: Inactive Control UX Feedback Pass**
   - Scope: mark non-wired controls as explicitly inactive and show consistent "Coming Soon" tooltips.
   - Controls:
     - Settings: `Re-sync` cookies button
     - Settings: `PO Token Extension` toggle
     - Library item card: `External Link` button
   - Status: 🟡 Implemented locally (not committed)

2. **PR12-B: YouTube Auth Cookie Re-sync Wiring**
   - Scope: implement manual re-sync action end-to-end (backend endpoint + UI trigger + error/success feedback).
   - Deliverables:
     - Add API contract for re-sync action in docs.
     - Wire `Settings` re-sync button to backend.
     - Show operation result to user (success/failure state in settings panel).
   - Status: ⏳ Pending

3. **PR12-C: PO Token Extension Runtime Controls**
   - Scope: wire settings toggle to persisted configuration and runtime behavior.
   - Deliverables:
     - Persist toggle state through existing app settings flow.
     - Reflect actual backend/provider runtime status in UI.
     - Guard unsupported states with clear user-facing errors.
   - Status: ⏳ Pending

4. **PR12-D: Library External Link Action**
   - Scope: implement real behavior for the external-link control on library media cards.
   - Deliverables:
     - Decide link target priority (source URL vs local file path fallback).
     - Ensure sanitized URL/path handling before launch.
     - Add disabled fallback if no valid link is available per media item.
   - Status: ⏳ Pending

5. **PR12-E: Validation + Docs**
   - Scope: finish PR12 with tests and docs updates for newly wired controls.
   - Deliverables:
     - Add/adjust frontend tests for disabled/enabled states and click behavior.
     - Update user-facing docs (`frontend/docs` + root docs if API changes).
     - Execute manual QA pass across Download, Library, and Settings views.
   - Status: ⏳ Pending
