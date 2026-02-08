---
title: "Frontend Development"
weight: 30
---

# Frontend Development

This guide covers SolidJS-specific development practices for ytdl-go's web UI.

## Table of Contents

- [Development Workflow](#development-workflow)
- [Component Development](#component-development)
- [State Management](#state-management)
- [Styling Guidelines](#styling-guidelines)
- [Testing](#testing)
- [Build and Integration](#build-and-integration)

## Development Workflow

### 1. Start Development Server

```bash
cd frontend
npm run dev
```

- Vite dev server starts on `http://localhost:5173`
- Hot module reload enabled
- API calls proxy to backend

### 2. Configure Backend Proxy

By default, API requests proxy to `http://127.0.0.1:8080`. To use a different backend:

```bash
VITE_API_PROXY_TARGET=http://127.0.0.1:9090 npm run dev
```

### 3. Start Backend Server

In a separate terminal:

```bash
cd ..
go run . -web -port 8080
```

Or use a custom port:

```bash
go run . -web -port 9090
```

### 4. Make Changes

- Edit files in `frontend/src/`
- Changes auto-reload in browser
- Check browser console for errors

### 5. Build for Production

```bash
npm run build
```

Output goes to `../internal/web/assets/`

### 6. Test Full-Stack Build

```bash
cd ..
./build.sh
./ytdl-go -web
```

Open `http://localhost:8080` to test embedded frontend.

## Component Development

### Component Structure

Keep components small and focused:

```jsx
// components/DownloadButton.jsx
import { createSignal } from 'solid-js';

export default function DownloadButton(props) {
  const [loading, setLoading] = createSignal(false);
  
  const handleClick = async () => {
    setLoading(true);
    try {
      await props.onDownload();
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <button
      onClick={handleClick}
      disabled={loading()}
      class="px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"
    >
      {loading() ? 'Downloading...' : 'Download'}
    </button>
  );
}
```

### File Naming Conventions

- **Components:** PascalCase (e.g., `DownloadView.jsx`)
- **Utilities:** camelCase (e.g., `apiClient.js`)
- **Store:** camelCase (e.g., `appStore.jsx`)

### Props Handling

Use destructuring for clarity:

```jsx
export default function MediaCard({ title, artist, size, onPlay }) {
  return (
    <div class="card">
      <h3>{title}</h3>
      <p>{artist}</p>
      <span>{size}</span>
      <button onClick={onPlay}>Play</button>
    </div>
  );
}
```

### Conditional Rendering

Use `<Show>` and `<For>`:

```jsx
import { Show, For } from 'solid-js';

export default function ItemList({ items, loading }) {
  return (
    <div>
      <Show when={!loading()} fallback={<LoadingSpinner />}>
        <For each={items()}>
          {(item) => <ItemCard item={item} />}
        </For>
      </Show>
    </div>
  );
}
```

## State Management

### Using the Store

Access store via context:

```jsx
import { useAppStore } from '../store/appStore';

export default function MyComponent() {
  const [store, actions] = useAppStore();
  
  const handleSubmit = () => {
    actions.submitDownload(store.download.urlInput, store.settings);
  };
  
  return (
    <div>
      <input
        value={store.download.urlInput}
        onInput={(e) => actions.setUrlInput(e.target.value)}
      />
      <button onClick={handleSubmit}>Submit</button>
    </div>
  );
}
```

### Adding New State

1. **Update initial state** in `appStore.jsx`:
   ```js
   const initialState = {
     // ...
     myNewFeature: {
       value: '',
       enabled: false,
     },
   };
   ```

2. **Add actions:**
   ```js
   const actions = {
     // ...
     setMyValue(value) {
       setState('myNewFeature', 'value', value);
     },
     toggleMyFeature() {
       setState('myNewFeature', 'enabled', (prev) => !prev);
     },
   };
   ```

3. **Use in components:**
   ```jsx
   const [store, actions] = useAppStore();
   actions.setMyValue('new value');
   ```

## Styling Guidelines

### Tailwind Utilities First

Prefer Tailwind utility classes:

```jsx
// Good
<button class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
  Download
</button>

// Avoid (unless necessary)
<button class="custom-button">Download</button>
```

### Consistent Spacing

Use Tailwind's spacing scale (4px base):

- `p-1` = 4px padding
- `p-2` = 8px padding
- `p-4` = 16px padding
- `p-8` = 32px padding

### Responsive Design

Use responsive modifiers:

```jsx
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  {/* Items */}
</div>
```

### Custom CSS (Sparingly)

Only add custom CSS for highly specific needs:

```css
/* index.css */

/* Custom scrollbar */
::-webkit-scrollbar {
  width: 8px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
}

::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 4px;
}
```

### Icons

Use Lucide icons:

```jsx
import { createEffect } from 'solid-js';
import { createIcons, Download, Play, Pause } from 'lucide';

export default function MediaPlayer() {
  createEffect(() => {
    createIcons({
      icons: { Download, Play, Pause },
    });
  });
  
  return (
    <div>
      <button><i data-lucide="play"></i></button>
      <button><i data-lucide="pause"></i></button>
      <button><i data-lucide="download"></i></button>
    </div>
  );
}
```

> **Important:** Re-initialize icons in `createEffect` when DOM changes significantly.

## Testing

### Manual Testing

1. **Test in browser console** - No red errors
2. **Test interactions** - All buttons and forms work
3. **Test SSE connection** - Progress updates display
4. **Test error handling** - Invalid inputs show errors
5. **Test responsive design** - Works on mobile sizes

### Testing SSE

Use browser DevTools Network tab:

1. Start a download
2. Open Network tab
3. Filter: EventStream
4. Click the connection
5. View messages in real-time

### Testing API Calls

Use browser console:

```js
// Test download endpoint
fetch('/api/download', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    urls: ['https://www.youtube.com/watch?v=dQw4w9WgXcQ'],
    options: { quality: 'best' }
  })
})
.then(r => r.json())
.then(console.log);
```

## Build and Integration

### Production Build

```bash
npm run build
```

**Output:** `../internal/web/assets/`

**Generated files:**
- `index.html` - HTML entry point
- `assets/*.js` - JavaScript bundles (hashed)
- `assets/*.css` - Stylesheets (hashed)

### Build Optimization

Vite automatically:
- Minifies JavaScript and CSS
- Tree-shakes unused code
- Generates source maps (dev only)
- Hashes filenames for cache busting

### Full-Stack Build

From repository root:

```bash
./build.sh
```

This:
1. Builds frontend (`npm run build`)
2. Embeds assets in Go binary (`//go:embed`)
3. Builds final executable

### Verifying Embedded Assets

```bash
# Build with embedded frontend
./build.sh

# Run server
./ytdl-go -web

# Test in browser
open http://localhost:8080
```

## Code Style Guidelines

### Component Structure

```jsx
// 1. Imports
import { createSignal, Show } from 'solid-js';
import { useAppStore } from '../store/appStore';

// 2. Component definition
export default function MyComponent(props) {
  // 3. Store access
  const [store, actions] = useAppStore();
  
  // 4. Local state
  const [count, setCount] = createSignal(0);
  
  // 5. Handlers
  const handleClick = () => {
    setCount(count() + 1);
  };
  
  // 6. Render
  return (
    <div>
      <h1>{props.title}</h1>
      <p>Count: {count()}</p>
      <button onClick={handleClick}>Increment</button>
    </div>
  );
}
```

### Prefer Composition

```jsx
// Good: Composable components
export default function DownloadView() {
  return (
    <div>
      <UrlInput />
      <OptionsForm />
      <ProgressDisplay />
    </div>
  );
}

// Avoid: Monolithic components
export default function DownloadView() {
  // 500+ lines in one component
}
```

### Keep Logic Separate

```jsx
// utils/apiClient.js
export async function startDownload(urls, options) {
  const response = await fetch('/api/download', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ urls, options }),
  });
  return response.json();
}

// components/DownloadView.jsx
import { startDownload } from '../utils/apiClient';

export default function DownloadView() {
  const handleDownload = async () => {
    const result = await startDownload(urls, options);
    // Handle result
  };
  // ...
}
```

## Adding New Dependencies

### Before Adding

1. **Check necessity** - Do we really need it?
2. **Check bundle size** - How much will it add?
3. **Check maintenance** - Is it actively maintained?
4. **Discuss in issue** - Get team consensus

### Allowed

- Small utility libraries (< 10KB)
- UI helpers (date formatting, etc.)
- Well-maintained packages

### Avoid

- Heavy component libraries (MUI, Bootstrap)
- Large data visualization libraries (unless tree-shaken)
- Redundant functionality (we already have Tailwind for styling)

### Installing

```bash
npm install package-name

# Or for dev dependencies
npm install -D package-name
```

### Updating Dependencies

```bash
# Check for updates
npm outdated

# Update specific package
npm update package-name

# Update all (carefully)
npm update
```

## Common Issues

### Issue: Icons Not Displaying

**Problem:** Icons don't appear after conditional rendering

**Solution:** Re-initialize icons in `createEffect`:

```jsx
createEffect(() => {
  createIcons({ icons: { Download, Play } });
});
```

### Issue: SSE Not Connecting

**Problem:** EventSource errors in console

**Solution:**
- Check backend is running
- Verify proxy configuration in `vite.config.js`
- Check browser DevTools Network tab
- Ensure jobId is valid

### Issue: State Not Updating

**Problem:** Component doesn't re-render on state change

**Solution:**
- Use signals: `count()` not `count`
- Ensure store actions are called correctly
- Check for missing `createEffect` or `createMemo`

## Related Documentation

- [Getting Started](getting-started) - Development environment setup
- [Code Style](code-style) - JavaScript coding standards
- [Frontend Structure](../architecture/frontend-structure) - Architecture deep dive
- [API Reference](../api-reference) - Backend API documentation
