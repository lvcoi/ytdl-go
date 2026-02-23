# Widget Developer Guide

This guide explains how to create and register new widgets for the ytdl-go dashboard.

## Widget Contract

A dashboard widget is a standard SolidJS component.

- **Props:** Widgets receive no grid-related props. They just fill their container.
- **Dimensions:** The widget should style itself to fill `100%` width and height.
- **Overflow:** The widget must handle its own overflow (e.g., `overflow-hidden`, `overflow-y-auto`).
- **Data:** Widgets should fetch their own data or accept data props if they are generic (passed via `renderWidget` in `DashboardView.jsx`).

### Basic Template

```jsx
// src/components/dashboard/MyNewWidget.jsx
import { onMount, createSignal } from 'solid-js';

export default function MyNewWidget() {
    return (
        <div class="h-full w-full bg-surface-primary rounded-[2rem] p-6 border border-white/10 flex flex-col">
            <h3 class="text-lg font-bold text-white mb-2">My Widget</h3>
            <div class="flex-1 overflow-y-auto text-gray-400">
                Widget content goes here...
            </div>
        </div>
    );
}
```

## Registration Process

To make your widget available in the dashboard:

1.  **Register Metadata:**
    Add an entry to `WIDGET_REGISTRY` in `frontend/src/components/dashboard/widgetRegistry.js`:

    ```js
    'my-widget': {
        id: 'my-widget',
        label: 'My New Widget',
        icon: 'activity', // Lucide icon name
        defaultW: 4,      // Default width in columns (1-16)
        defaultH: 3,      // Default height in rows (80px per row)
        minW: 4,          // Minimum width
        minH: 2           // Minimum height
    },
    ```

2.  **Render Component:**
    Add a case to the `renderWidget` function in `frontend/src/components/DashboardView.jsx`:

    ```jsx
    import MyNewWidget from './dashboard/MyNewWidget';

    // ... inside renderWidget switch ...
    case 'my-widget':
        return <MyNewWidget />;
    ```

3.  **Default Layout (Optional):**
    If you want the widget to appear by default for new users, add it to `DEFAULT_LAYOUT_WIDGETS` in `widgetRegistry.js`.

## Sizing Guidelines

The grid uses **16 columns** and **80px rows** (plus 12px gaps).

- **1 Column** ≈ 75px wide (on desktop).
- **1 Row** = 80px high.

**Recommended Sizes:**

| Size | W x H | Approx Pixels | Use Case |
| :--- | :--- | :--- | :--- |
| **Small Tile** | 4 x 2 | 300 x 172 | Simple stat, toggle button |
| **Square** | 4 x 4 | 300 x 356 | Chart, list preview |
| **Wide Banner** | 8 x 2 | 600 x 172 | Header, input form |
| **Large Panel** | 8 x 4 | 600 x 356 | Detailed list, complex chart |

**Minimum Sizes:**
Always set `minW` and `minH` based on the absolute minimum space your content needs to be legible. The grid will prevent users from resizing below this.

## Conventions

- **Styling:** Use Tailwind CSS.
- **Theme:** Use `bg-surface-primary` for background, `border-white/10` for borders.
- **Border Radius:** Use `rounded-[2rem]` to match the dashboard aesthetic.
- **Icons:** Use the `Icon` component wrapper for Lucide icons.
