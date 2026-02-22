# Widget Authoring Guide

How to create a new dashboard widget for the ytdl-go Web UI.

---

## Overview

The dashboard uses a **16-column CSS Grid** with **80px row height** and **12px gap**. Widgets are placed via inline `grid-column`/`grid-row` styles. All widget metadata lives in a central registry that powers the grid engine, widget drawer, layout presets, drag/resize, and collision handling automatically.

Adding a new widget requires **exactly 3 steps**.

---

## Step 1: Create the Component

Create a new file in `frontend/src/components/dashboard/YourWidget.jsx` (or directly in `frontend/src/components/` for top-level widgets).

### Template

```jsx
export default function YourWidget(props) {
    return (
        <div class="rounded-[2rem] border border-white/10 bg-black/20 backdrop-blur-sm p-6 h-full flex flex-col">
            <h3 class="text-xs font-black uppercase tracking-widest text-gray-500 mb-4">
                Your Widget Title
            </h3>
            <div class="flex-1">
                {/* Widget content here */}
            </div>
        </div>
    );
}
```

### Styling Conventions

| Property | Value |
|----------|-------|
| Border radius | `rounded-[2rem]` |
| Border | `border border-white/10` |
| Background | `bg-black/20 backdrop-blur-sm` |
| Padding | `p-6` |
| Height | `h-full` (fill the grid cell) |
| Title style | `text-xs font-black uppercase tracking-widest text-gray-500` |

### Props Contract

Widgets receive props from `DashboardView.jsx`'s `renderWidget()` function. Common props available:

| Prop | Type | Description |
|------|------|-------------|
| `stats` | `Accessor<object>` | Library stats (totalItems, totalCreators, recentItems) |
| `onTabChange` | `function` | Navigate to another tab |
| `onPlay` | `function` | Play a media item |
| `onDownload` | `function` | Trigger a download |
| `onSettingsChange` | `function` | Update a setting |

You only receive the props you wire up in the `renderWidget()` switch (Step 3). Not all widgets need all props.

---

## Step 2: Register in Widget Registry

Open `frontend/src/components/dashboard/widgetRegistry.js` and add an entry to `WIDGET_REGISTRY`:

```js
'your-widget': {
    id: 'your-widget',
    label: 'Your Widget',
    icon: 'box',           // Lucide icon name
    defaultW: 6,           // Default width in grid columns (1–16)
    defaultH: 3,           // Default height in grid rows
    minW: 4,               // Minimum width (content legibility)
    minH: 2,               // Minimum height (content legibility)
},
```

### Choosing Sizes

- **Row height** is 80px. A `minH: 2` widget gets ~160px + 12px gap = 172px visible height.
- **Column width** depends on viewport. At 1440px wide: each column ≈ 83px. A `minW: 4` widget gets ~332px + gaps.
- Set `minW`/`minH` to the smallest size where your content is still **legible and usable** — no truncated text, no invisible controls.
- `defaultW`/`defaultH` should be a comfortable "looks good" size.
- **Constraint:** `minW ≤ defaultW` and `minH ≤ defaultH`.

### Default Layout (optional)

If the widget should appear in the factory default layout, add it to `DEFAULT_LAYOUT_WIDGETS`:

```js
{ id: 'your-widget', enabled: true, x: 0, y: 10, width: 6, height: 3 },
```

Set `enabled: false` if you want it available in the drawer but not shown by default.

---

## Step 3: Add Render Case

Open `frontend/src/components/DashboardView.jsx` and add a case to the `renderWidget()` switch:

```jsx
case 'your-widget':
    return <YourWidget stats={stats()} />;
```

Don't forget to add the import at the top of the file:

```jsx
import YourWidget from './dashboard/YourWidget';
```

---

## What You Get Automatically

After these 3 steps, the grid engine provides:

- **Drag-and-drop** with ghost preview and Grafana-style push collision
- **Resize** in all 8 directions with min-size enforcement
- **Widget drawer** listing your widget with add/remove toggle
- **Layout presets** save/load including your widget
- **Undo/redo** for all layout changes involving your widget
- **localStorage persistence** across page reloads
- **v2→v3 migration** (not applicable for new widgets, but existing layouts won't break)

---

## Testing Requirements

Per `docs/BEST_PRACTICES.md`:

1. **Render test** — widget mounts without errors.
2. **Guard-pair tests** — if your widget has any `if (!condition) return;` guards, write both a block test and a pass-through test.
3. **Mock in DashboardView tests** — add a `vi.mock()` entry in `DashboardView.test.jsx` so the DashboardView tests don't depend on your widget internals.

Example mock:

```js
vi.mock('./dashboard/YourWidget', () => ({
    default: () => <div data-testid="your-widget">YourWidget</div>
}));
```

---

## Existing Widgets Reference

| ID | Component | Min W×H | Default W×H | Notes |
|----|-----------|---------|-------------|-------|
| `welcome` | `WelcomeWidget` | 6×3 | 16×3 | Title + stats + nav buttons |
| `quick-download` | `QuickDownload` | 6×2 | 8×2 | URL input + download button |
| `active-downloads` | `ActiveDownloads` | 4×3 | 4×4 | Task list with progress bars |
| `recent-activity` | `RecentActivityWidget` | 6×4 | 12×4 | Horizontal card carousel |
| `stats` | `StatsWidget` | 4×3 | 4×3 | 2-column stat grid |
| `concurrency` | `ConcurrencyWidget` | 5×2 | 6×4 | Collapsible sliders + toggles |
