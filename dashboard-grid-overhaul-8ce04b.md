# Dashboard Grid Overhaul

Rebuild the dashboard from a broken 4-column Tailwind-class grid into a 16-column placement-based grid engine with Grafana-style push collision, working drag-and-drop with ghost preview, animated resize (horizontal + vertical), grid-line overlay, a widget drawer, layout presets, undo/redo, and full test coverage — while fixing the reactivity bugs and testing violations from the prior code review. Includes ConcurrencyWidget integration and a widget developer guide for future extensibility.

---

## Resolved Design Decisions

| Decision               | Answer                                                               |
|------------------------|----------------------------------------------------------------------|
| **Collision handling** | Grafana-style push — dragged/resized widgets push neighbors downward |
| **Responsive**         | Desktop-only — no mobile breakpoints needed                          |
| **Widget min sizes**   | Per-widget, based on content legibility (see registry below)         |
| **Row unit height**    | 80px                                                                 |
| **Undo/redo**          | In scope — edit-mode action history                                  |

---

## Architecture

### 1. CSS Grid with Explicit Placement (replaces Tailwind col-span)

The current `Grid` uses `grid-cols-4` with `col-span-N` classes — can't support arbitrary placement, gaps, or vertical sizing. Replacement uses **CSS Grid with inline `grid-column` / `grid-row` styles** per widget:

```css
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(16, 1fr);
  grid-auto-rows: 80px;
  gap: 12px;
}
```

Each widget placed via SolidJS dynamic `style` prop:

```js
style={{
  'grid-column': `${x + 1} / span ${width}`,
  'grid-row': `${y + 1} / span ${height}`,
}}
```

Per Tailwind v4 docs, dynamic per-element values require inline styles — Tailwind arbitrary-value classes (`col-[16_/_span_16]`) are for known-at-build-time values, not runtime-computed per-widget coordinates. Tailwind is still used for everything else (transitions, colors, spacing, etc.).

**Why 16 columns:** Divisible by 1, 2, 4, 8, 16 — maximum layout flexibility.

### 2. Ghost Preview Pattern (Drag & Resize Animation)

Standard for dashboard editors (Grafana, Home Assistant):

1. **Start:** Original widget dims (opacity ~0.5). Semi-transparent ghost appears at current grid position.
2. **Move:** Ghost snaps to valid grid cells. CSS `transition` on grid placement provides smooth animation. Colliding neighbors are pushed down (Grafana-style).
3. **Drop:** Ghost disappears, widget animates to new position via CSS transition.

Zero-dependency — pure CSS transitions on `grid-column`/`grid-row` properties.

### 3. Grafana-Style Push Collision

When a widget is dragged or resized into space occupied by another widget:

- The displaced widget is pushed **downward** (increased `y`) by the minimum amount to clear the overlap.
- This cascades: if the pushed widget overlaps a third, that one is also pushed.
- On drop, a compaction pass pulls all widgets upward to close unnecessary vertical gaps.
- Push is previewed in real-time during drag (ghost + displaced widgets animate to pushed positions).

### 4. Widget Data Model (v3)

```js
// Single widget
{ id: 'welcome', enabled: true, x: 0, y: 0, width: 16, height: 3 }

// localStorage schema
{
  version: 3,
  activeLayoutId: 'default',
  layouts: {
    'default': { name: 'Default', widgets: [...], isFactory: true },
    'primary': { name: 'Primary', widgets: [...], isPrimary: true },
    'uuid-1':  { name: 'My Custom', widgets: [...] },
  }
}
```

The `span` field is dropped. Migration from v2: multiply `x` and `width` by 4 (4-col → 16-col).

### 5. Widget Registry (content-based minimum sizes)

Minimum sizes derived from actual widget content at 80px row height (~75px per column unit):

```js
export const WIDGET_REGISTRY = {
  'welcome':          { label: 'Welcome',            icon: 'home',           defaultW: 16, defaultH: 3, minW: 6, minH: 3 },
  'quick-download':   { label: 'Quick Download',     icon: 'download-cloud', defaultW: 8,  defaultH: 2, minW: 6, minH: 2 },
  'active-downloads': { label: 'Active Downloads',   icon: 'download-cloud', defaultW: 4,  defaultH: 4, minW: 4, minH: 3 },
  'recent-activity':  { label: 'Recent Activity',    icon: 'history',        defaultW: 12, defaultH: 4, minW: 6, minH: 4 },
  'stats':            { label: 'Library Stats',      icon: 'bar-chart-2',    defaultW: 4,  defaultH: 3, minW: 4, minH: 3 },
  'concurrency':      { label: 'Concurrency',        icon: 'settings',       defaultW: 4,  defaultH: 4, minW: 4, minH: 2 },
};
```

| Widget              | Min W×H | Why                                                                                        |
|---------------------|---------|--------------------------------------------------------------------------------------------|
| **Welcome**         | 6×3     | h1 title + stats paragraph + 2 side-by-side buttons need ~440px; 3 rows for vertical stack |
| **QuickDownload**   | 6×2     | Input + download button side-by-side need ~440px; 2 rows (160px) fits header + form + hint |
| **ActiveDownloads** | 4×3     | Task items with progress bars need ~290px; 3 rows for header + 1 task + empty state        |
| **RecentActivity**  | 6×4     | Carousel cards are 288px wide; 4 rows for header + aspect-video thumbnail + text           |
| **Stats**           | 4×3     | Internal 2-col stat grid needs ~290px; 3 rows for header + stat cards with large numbers   |
| **Concurrency**     | 4×2     | Collapsed: header bar fits in 2 rows (160px); expanded state needs user to resize taller   |

### 6. Undo/Redo

Action history stack stored as SolidJS signals:

- `undoStack: WidgetState[]` — previous states, pushed on every drag/resize/add/remove/layout-load.
- `redoStack: WidgetState[]` — states popped by undo, cleared on new action.
- Keyboard: `Ctrl+Z` (undo), `Ctrl+Shift+Z` (redo) — active only in edit mode.
- Toolbar buttons with disabled state when stack is empty.
- Stack depth limit: 50 entries.

### 7. Widget Developer Guide

A `frontend/src/components/dashboard/WIDGETS.md` document defining the contract for creating new dashboard widgets. This enables other developers to add widgets (CPU monitor, network usage, etc.) without touching the grid engine.

---

## Sprint 1: Foundation — Bug Fixes & Grid Engine

### Task 1: Fix Reactivity Bug (loaded flag + queueMicrotask)

**Files:** `frontend/src/components/DashboardView.jsx`
**Context:** `loaded` is a plain `let` — invisible to SolidJS reactivity. The auto-save `createEffect` checks `if (!loaded) return;` but won't re-run when `loaded` changes since it's not a signal. The `queueMicrotask` workaround is fragile. Per SolidJS docs, `createEffect` only tracks signals accessed during execution.

**Step-by-Step:**

1. **Replace plain variable with signal:** Change `let loaded = false;` (line 23) to `const [hasLoaded, setHasLoaded] = createSignal(false);`.
2. **Update onMount — v2 path:** Replace `queueMicrotask(() => { loaded = true; });` (line 90) with `setHasLoaded(true);`.
3. **Update onMount — fallback path:** Replace `queueMicrotask(() => { loaded = true; });` (line 115) with `setHasLoaded(true);`.
4. **Update effect guard:** Change `if (!loaded) return;` (line 121) to `if (!hasLoaded()) return;`.

**Acceptance Criteria:**

- [x] `loaded` is a `createSignal`, not a plain variable.
- [x] No `queueMicrotask` calls remain in the component.
- [x] Auto-save effect fires reactively when `hasLoaded()` becomes true and `widgets()` changes.
- [x] Existing tests "does not overwrite saved layout with defaults on mount" and "persists layout changes after loading an existing v2 layout" still pass.

---

### Task 2: Create Widget Registry & Grid Constants

**Files:** `frontend/src/components/dashboard/widgetRegistry.js` (new)
**Context:** Widget metadata is scattered across `DashboardView.jsx` and `DEFAULT_WIDGETS`. Centralizing into a registry enables the widget drawer, enforces min sizes during resize, and provides default dimensions for newly-added widgets. Includes the existing `ConcurrencyWidget`.

**Step-by-Step:**

1. **Create `widgetRegistry.js`:** Export `WIDGET_REGISTRY` object with all 6 widgets (including `concurrency`) and content-based min sizes per the table above.
2. **Export `DEFAULT_LAYOUT_WIDGETS`:** Array of widget placement objects using 16-col coordinates. ConcurrencyWidget is included but `enabled: false` by default (available via drawer).
3. **Export grid constants:** `GRID_COLS = 16`, `GRID_ROW_HEIGHT_PX = 80`, `GRID_GAP_PX = 12`.

**Acceptance Criteria:**

- [x] `WIDGET_REGISTRY` contains entries for all 6 widgets with content-based min/default sizes.
- [x] `DEFAULT_LAYOUT_WIDGETS` uses 16-col coordinates.
- [x] ConcurrencyWidget is registered with `enabled: false` in the default layout.
- [x] Grid constants are exported and importable.

---

### Task 3: Build the 16-Column Grid Engine

**Files:** `frontend/src/components/Grid.jsx`, `frontend/index.css`
**Context:** The current `Grid` uses Tailwind `grid-cols-4` with class-based spans. Replace with a 16-column CSS Grid using explicit `grid-column`/`grid-row` inline styles per widget. Add grid-line overlay for edit mode.

**Step-by-Step:**

1. **Rewrite `Grid` component:** Replace Tailwind grid classes with inline style: `display: grid; grid-template-columns: repeat(16, 1fr); grid-auto-rows: 80px; gap: 12px;`. Accept `isEditMode`, `totalRows`, and `ref` (forwarded) props. Use `position: relative` for ghost overlay positioning.
2. **Add grid-lines overlay:** Inside `Grid`, render a `<Show when={props.isEditMode}>` child div with `position: absolute; inset: 0; pointer-events: none; z-index: 0;`. Use CSS `background-image` with `repeating-linear-gradient` to draw subtle grid lines matching the 16-col / 80px-row spacing. Lines: `rgba(255,255,255,0.06)`.
3. **Rewrite `GridItem` component:** Remove the `spanClass()` switch. Accept `x, y, width, height` props. Compute inline `style` for grid placement. Add CSS transition: `transition: grid-column 0.25s ease, grid-row 0.25s ease, grid-column-end 0.25s ease, grid-row-end 0.25s ease;`.
4. **Add `.dashboard-grid-lines` CSS** in `index.css` for the background pattern.
5. **Ensure widget content sits above grid lines** with `position: relative; z-index: 1;` on `GridItem`.

**Acceptance Criteria:**

- [x] Grid renders as a 16-column CSS Grid with 80px rows and 12px gap.
- [x] Widgets are placed via `grid-column` / `grid-row` inline styles.
- [x] Grid lines appear only in edit mode, behind widgets.
- [x] Widget position changes animate smoothly via CSS transition.
- [x] Gap between widgets is visible (12px in both axes).

---

### Task 4: Migrate DashboardView to New Grid Engine

**Files:** `frontend/src/components/DashboardView.jsx`
**Context:** DashboardView currently uses the old `Grid`/`GridItem` with `span` props and hardcoded 4-col coordinates. Update to use the new 16-col grid, import from `widgetRegistry.js`, add ConcurrencyWidget to the render switch, and add v2→v3 migration for localStorage.

**Step-by-Step:**

1. **Update imports:** Import `WIDGET_REGISTRY`, `DEFAULT_LAYOUT_WIDGETS`, `GRID_COLS` from `widgetRegistry.js`. Import `ConcurrencyWidget`.
2. **Replace `DEFAULT_WIDGETS`** with `DEFAULT_LAYOUT_WIDGETS`.
3. **Update localStorage keys:** Add `DASHBOARD_LAYOUT_KEY_V3 = 'ytdl-go:dashboard-layout:v3'`.
4. **Add v2→v3 migration:** In `onMount`, if v3 key is absent but v2 exists, multiply `x` and `width` by 4, keep `y`/`height`. Save as v3.
5. **Update `<Grid>` usage:** Pass `isEditMode={isEditMode()}` and `totalRows` (computed from max widget bottom edge).
6. **Update `<GridItem>` usage:** Pass `x={widget.x} y={widget.y} width={widget.width} height={widget.height}` instead of `span`.
7. **Add `concurrency` case to `renderWidget`:** Render `<ConcurrencyWidget />` (already exists at `frontend/src/components/ConcurrencyWidget.jsx`).
8. **Remove `span` references** from widget data.

**Acceptance Criteria:**

- [x] Dashboard renders using 16-col grid placement.
- [x] v2 layouts are migrated to v3 coordinates on first load.
- [x] ConcurrencyWidget renders when enabled.
- [x] `span` field is no longer used in widget data.
- [x] Default layout fills the grid sensibly at 16-col scale.

---

## Sprint 2: Interaction — Drag, Resize & Collision

### Task 5: Implement Grafana-Style Push Collision Engine

**Files:** `frontend/src/components/dashboard/gridCollision.js` (new)
**Context:** When widgets are dragged or resized, overlapping neighbors must be pushed downward (Grafana-style). This is a pure-logic module with no UI — it takes a widget array and a proposed placement, and returns the resolved widget positions after push + compaction.

**Step-by-Step:**

1. **Export `resolveCollisions(widgets, movedWidget)`:** Given the full widget array and one widget with a new position/size, push all overlapping widgets downward recursively. Return the new widget array.
2. **Export `compactLayout(widgets)`:** Pull all widgets upward to close vertical gaps while respecting collisions. Iterate top-to-bottom, moving each widget to the lowest `y` that doesn't overlap.
3. **Export `findOpenPosition(widgets, width, height, gridCols)`:** Scan the grid for the first open rectangle of the given size (used by widget drawer "Add" button).
4. **Export `widgetsOverlap(a, b)`:** Boolean helper — true if two widget rectangles overlap.

**Acceptance Criteria:**

- [x] `resolveCollisions` pushes overlapping widgets downward.
- [x] Push cascades (A pushes B, B pushes C).
- [x] `compactLayout` closes vertical gaps without creating new overlaps.
- [x] `findOpenPosition` returns a valid non-overlapping position.
- [x] All functions are pure (no side effects, no signals).

---

### Task 6: Implement Drag-and-Drop with Ghost Preview

**Files:** `frontend/src/components/DashboardView.jsx`, `frontend/src/components/Grid.jsx`
**Context:** The current drag implementation uses pixel-based `transform: translate()` which doesn't snap to grid and doesn't provide visual feedback. Replace with ghost-preview pattern using measured cell sizes and the collision engine.

**Step-by-Step:**

1. **Add a `ref` to the Grid container** and measure it on mount/resize (via `ResizeObserver`) to compute `cellWidth` and `cellHeight` in pixels.
2. **Add ghost state signal:** `const [ghostPos, setGhostPos] = createSignal(null);` — `{ x, y, width, height }` in grid units.
3. **Rewrite `handleDragStart`:** Record start client coords, widget's original grid position. Set ghost to widget's current position. Dim the original widget.
4. **Rewrite `handleDragMove`:** Compute delta in grid units using measured cell sizes. Clamp to grid bounds. Update ghost. Run `resolveCollisions` to preview pushed positions for other widgets.
5. **Rewrite `handleDragEnd`:** Commit ghost + pushed positions to widget state. Run `compactLayout`. Clear ghost. Push pre-move state to undo stack.
6. **Render ghost in `Grid`:** `<Show when={ghostPos()}>` div with grid placement, semi-transparent accent color, rounded corners.
7. **Guard:** `if (!isEditMode()) return;` on drag start.

**Acceptance Criteria:**

- [x] Dragging shows a ghost preview snapped to grid cells.
- [x] Colliding widgets are pushed downward in real-time during drag.
- [x] On drop, layout is compacted and widget animates to final position.
- [x] Dragging is blocked outside edit mode.
- [x] Widget cannot be dragged outside grid bounds.
- [x] Document event listeners are cleaned up on mouseup and unmount.

---

### Task 7: Implement Resize with Ghost Preview (Horizontal + Vertical)

**Files:** `frontend/src/components/DashboardView.jsx`, `frontend/src/components/Grid.jsx`
**Context:** Resize handles exist in `GridItem` but the resize logic uses hardcoded 100px cell size and doesn't support vertical resizing. Replace with ghost-preview resize using measured cell dimensions. Enforce min sizes from `WIDGET_REGISTRY`. Use collision engine for push.

**Step-by-Step:**

1. **Rewrite `handleResizeStart`:** Record direction, widget's original position/size, start client coords. Set ghost to widget's current bounds.
2. **Rewrite `handleResizeMove`:** Compute delta in grid units. Apply directional logic (e/w/s/n/corners). Enforce min width/height from `WIDGET_REGISTRY[widgetId]`. Clamp to grid bounds. Update ghost. Run `resolveCollisions` for preview.
3. **Rewrite `handleResizeEnd`:** Commit ghost bounds + pushed positions. Compact. Clear ghost. Push pre-resize state to undo stack.
4. **Share ghost rendering** with drag (same ghost div, different source of coords).
5. **Vertical resize:** `n` and `s` handles modify `height` and `y` in 80px row units.

**Acceptance Criteria:**

- [x] Resizing shows a ghost preview snapped to grid units.
- [x] All 8 directions work (n, s, e, w, ne, nw, se, sw).
- [x] Vertical resizing changes widget height in row units.
- [x] Min sizes from `WIDGET_REGISTRY` are enforced.
- [x] Colliding widgets are pushed during resize.
- [x] Resize is blocked outside edit mode.

---

## Sprint 3: Features — Undo/Redo, Widget Drawer & Layout Presets

### Task 8: Implement Undo/Redo for Edit Mode

**Files:** `frontend/src/components/DashboardView.jsx`
**Context:** Users need to undo accidental drag/resize/add/remove actions in edit mode. Implement a simple state-snapshot undo/redo stack with keyboard shortcuts.

**Step-by-Step:**

1. **Add undo/redo signals:** `const [undoStack, setUndoStack] = createSignal([]);` and `const [redoStack, setRedoStack] = createSignal([]);`.
2. **Create `pushUndo(currentWidgets)`:** Pushes a deep copy of the current widget array onto the undo stack (max 50). Clears redo stack.
3. **Create `undo()`:** Pops from undo stack, pushes current state to redo stack, sets widgets to popped state.
4. **Create `redo()`:** Pops from redo stack, pushes current state to undo stack, sets widgets to popped state.
5. **Wire into drag/resize/add/remove:** Call `pushUndo` before committing each action.
6. **Add keyboard shortcuts:** `Ctrl+Z` → undo, `Ctrl+Shift+Z` → redo. Active only when `isEditMode()`. Use `onCleanup` to remove listener.
7. **Add toolbar buttons:** Undo/redo icons in the edit-mode toolbar, disabled when respective stack is empty.

**Acceptance Criteria:**

- [x] Undo reverts the last drag/resize/add/remove action.
- [x] Redo re-applies an undone action.
- [x] Redo stack is cleared on new action.
- [x] `Ctrl+Z` / `Ctrl+Shift+Z` work in edit mode only.
- [x] Stack is capped at 50 entries.
- [x] Toolbar buttons reflect stack state (enabled/disabled).

---

### Task 9: Build Widget Drawer

**Files:** `frontend/src/components/dashboard/WidgetDrawer.jsx` (new), `frontend/src/components/DashboardView.jsx`
**Context:** Users need a way to add removed widgets back and see available widgets not on the dashboard. The drawer lists all widgets from `WIDGET_REGISTRY` (including ConcurrencyWidget) with their enabled/disabled status.

**Step-by-Step:**

1. **Create `WidgetDrawer.jsx`:** Slide-in panel from the right. Accepts `widgets`, `onToggleWidget`, `onAddWidget` props.
2. **List all registry widgets:** For each entry in `WIDGET_REGISTRY`, show label, icon, and status. Active widgets show "Remove"; inactive/missing show "Add".
3. **Add button:** Calls `findOpenPosition` from collision engine to place widget at first available spot with default size. Pushes undo state.
4. **Remove button:** Sets `enabled: false`. Pushes undo state.
5. **Wire into DashboardView:** Drawer toggle button in edit-mode toolbar. Render `<WidgetDrawer>` conditionally with slide animation.

**Acceptance Criteria:**

- [x] Drawer slides in from the right when toggled in edit mode.
- [x] All 6 registered widgets appear with correct active/available status.
- [x] "Add" places a widget at a valid non-overlapping position.
- [x] "Remove" hides the widget from the grid.
- [x] Drawer is not visible outside edit mode.
- [x] Add/remove actions are undoable.

---

### Task 10: Build Layout Presets System

**Files:** `frontend/src/components/dashboard/LayoutPresets.jsx` (new), `frontend/src/components/DashboardView.jsx`
**Context:** Users need to save, load, and manage multiple dashboard layouts. Three tiers: (1) "Default" — factory layout, immutable; (2) "Primary" — user's chosen default, loaded on startup; (3) Custom named layouts.

**Step-by-Step:**

1. **Create `LayoutPresets.jsx`:** Dropdown/popover in the edit-mode toolbar. Lists saved layouts with actions.
2. **Save current layout:** Text input for name → saves current `widgets()` as a new named layout with a generated UUID key.
3. **Load layout:** Click a layout name → `setWidgets(layout.widgets)`. Pushes undo state.
4. **Set as Primary:** Star icon → marks layout as startup default. Only one primary at a time.
5. **Delete layout:** Trash icon → removes custom layout. "Default" cannot be deleted.
6. **Restore Default:** Always-visible button → loads factory `DEFAULT_LAYOUT_WIDGETS`.
7. **Update localStorage schema:** Wrap all layout data in v3 envelope `{ version: 3, activeLayoutId, layouts: {...} }`.
8. **Update onMount:** Load "Primary" layout on startup (or "Default" if no primary set).

**Acceptance Criteria:**

- [x] Users can save the current layout with a custom name.
- [x] Users can load any saved layout.
- [x] Users can set any layout as "Primary" (loaded on startup).
- [x] Factory "Default" layout is always available and cannot be deleted.
- [x] Custom layouts can be deleted.
- [x] Layout data persists across page reloads.

---

## Sprint 4: Documentation, Testing & Cleanup

### Task 11: Write Widget Developer Guide

**Files:** `frontend/src/components/dashboard/WIDGETS.md` (new)
**Context:** Other developers need a clear contract for creating new dashboard widgets (CPU monitor, network usage, etc.) without touching the grid engine. This document defines the interface, registration process, and conventions.

**Step-by-Step:**

1. **Document the widget contract:** A widget is a SolidJS component that receives no grid-related props — it just renders its content and fills its container (`h-full w-full` or equivalent). The grid engine handles all sizing and positioning.
2. **Document the registration process:** Add an entry to `WIDGET_REGISTRY` in `widgetRegistry.js` with `id`, `label`, `icon`, `defaultW`, `defaultH`, `minW`, `minH`. Add a `case` to `renderWidget()` in `DashboardView.jsx`. That's it.
3. **Document sizing guidelines:** How to determine `minW`/`minH` based on content (reference the analysis table). Recommend designing for the minimum size first, then scaling up gracefully.
4. **Document conventions:** Use Tailwind for styling, `h-full` on root element, handle overflow internally (scroll or truncate), use `splitProps` if accepting custom props.
5. **Provide a template:** Minimal example widget with the required structure.

**Acceptance Criteria:**

- [x] `WIDGETS.md` documents the widget interface contract.
- [x] Registration process is described step-by-step with file paths.
- [x] Sizing guidelines reference the min-size analysis methodology.
- [x] A copy-pastable template widget is included.

---

### Task 12: Comprehensive Test Coverage

**Files:** `frontend/src/components/DashboardView.test.jsx`, `frontend/src/components/Grid.test.jsx`, `frontend/src/components/dashboard/gridCollision.test.js` (new), `frontend/src/components/dashboard/WidgetDrawer.test.jsx` (new), `frontend/src/components/dashboard/LayoutPresets.test.jsx` (new), `frontend/src/components/dashboard/widgetRegistry.test.js` (new)
**Context:** Per `BEST_PRACTICES.md`, every guard needs block + pass-through tests, and state transitions need entry/exit/behavior coverage.

**Step-by-Step:**

1. **Widget Registry tests:** All entries have required fields, min ≤ default sizes, unique IDs, ConcurrencyWidget present.
2. **Grid Collision tests:** `resolveCollisions` pushes correctly, cascades, `compactLayout` closes gaps, `findOpenPosition` finds valid spots, `widgetsOverlap` edge cases.
3. **Grid component tests:** 16-col grid renders with correct inline styles. GridItem computes correct placement. Grid lines appear only in edit mode.
4. **Reactivity fix tests:** Guard-pair for `hasLoaded` — block (no save on mount) + pass-through (saves after loaded + widget change).
5. **Drag tests:** Guard-pair for `isEditMode`. State-transition: idle → dragging → idle. Ghost appears/disappears. Bounds clamping. Push collision during drag.
6. **Resize tests:** Guard-pair for `isEditMode`. Directions (e, s minimum). Min size enforcement. Vertical resize. Ghost preview.
7. **Undo/redo tests:** Undo reverts, redo re-applies, new action clears redo, stack cap at 50, keyboard shortcuts only in edit mode.
8. **Widget Drawer tests:** Visibility toggle, add/remove widget, all 6 registry entries shown.
9. **Layout Presets tests:** Save/load/delete, primary designation, default immutability, persistence.

**Acceptance Criteria:**

- [x] Every `if (!isEditMode()) return;` guard has both block and pass-through tests.
- [x] Drag state transition (idle → dragging → idle) is tested end-to-end.
- [x] Resize state transition is tested for at least 2 directions (e, s).
- [x] Collision engine has unit tests for push, cascade, compact, and overlap.
- [x] Undo/redo is tested for at least 2 action types.
- [x] Widget drawer add/remove is tested.
- [x] Layout preset CRUD is tested.
- [x] All tests pass: `cd frontend && npm test`.

---

### Task 13: Build Pipeline Compliance

**Files:** `internal/web/assets/` (verify only), `frontend/` (source of truth)
**Context:** Build artifacts in `internal/web/assets/` are generated by `npm run build` (vite.config.js line 220). All changes go through `frontend/` source only.

**Step-by-Step:**

1. **Verify** no manual edits to `internal/web/assets/` during this sprint.
2. **Run `cd frontend && npm test`** — all tests pass.
3. **Run `cd frontend && npm run build`** — regenerate assets from source.
4. **Verify** built assets reflect all dashboard changes.

**Acceptance Criteria:**

- [x] No files in `internal/web/assets/` were hand-edited.
- [x] `npm test` passes with 0 failures.
- [x] `npm run build` succeeds and outputs to `internal/web/assets/`.
