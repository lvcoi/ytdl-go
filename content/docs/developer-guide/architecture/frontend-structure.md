---
title: "Frontend Structure"
weight: 30
---

# Frontend Structure

This document describes the high-level architecture of the ytdl-go web frontend.

## Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Component Structure](#component-structure)
- [State Management](#state-management)
- [Build & Integration Pipeline](#build--integration-pipeline)
- [API Communication](#api-communication)

## Overview

The frontend is a **Single Page Application (SPA)** built with [SolidJS](https://www.solidjs.com/). It compiles to static assets (`index.html`, `app.js`, `styles.css`) that are embedded directly into the Go binary at build time via `//go:embed`. This means the final distributed artifact is a single, self-contained executable.

### Key Features

- **Zero external dependencies at runtime** - Everything bundled in the binary
- **Hot reload development** - Fast iteration with Vite dev server
- **Reactive UI** - SolidJS fine-grained reactivity
- **Tailwind styling** - Utility-first CSS framework
- **Server-Sent Events** - Real-time download progress updates

## Tech Stack

| Layer | Technology | Purpose |
| ----- | ---------- | ------- |
| UI Framework | SolidJS | Reactive UI with fine-grained updates |
| Build Tool | Vite | Fast builds and hot module reload |
| Styling | Tailwind CSS | Utility-first CSS framework |
| Icons | Lucide (vanilla) | Clean, consistent iconography |
| Language | JavaScript (JSX) | No TypeScript overhead |

### Why SolidJS?

- **Performance** - No virtual DOM, compiles to vanilla JS
- **Reactivity** - Fine-grained reactive primitives
- **Small bundle size** - ~6KB runtime (vs React's ~40KB)
- **Familiar syntax** - JSX similar to React
- **Solid ecosystem** - Good tooling and community support

### Why No TypeScript?

- **Simplicity** - Reduces build complexity
- **Speed** - Faster builds without type checking
- **Size** - Slightly smaller final bundle
- **Maintenance** - Easier for contributors without TS experience

> **Note:** For larger projects, TypeScript is recommended. For ytdl-go's scope, vanilla JS with JSDoc comments provides sufficient type hints without the overhead.

## Component Structure

```text
frontend/
├── src/
│   ├── App.jsx              # Root component and top-level layout
│   ├── index.jsx            # Entry point, mounts App + store provider
│   ├── store/
│   │   └── appStore.jsx     # Central app state + localStorage persistence
│   └── components/
│       ├── DownloadView.jsx  # URL input, download options, result display
│       ├── LibraryView.jsx   # Downloaded media list, search/filter
│       ├── SettingsView.jsx  # Configuration (cookies, extensions)
│       └── Player.jsx        # Media playback
├── public/
│   └── favicon.ico          # Application icon
├── index.html               # HTML template
├── package.json             # Dependencies and scripts
├── vite.config.js           # Vite build configuration
├── tailwind.config.js       # Tailwind CSS configuration
└── postcss.config.js        # PostCSS configuration

Build output → ../internal/web/assets/
```

### Component Responsibilities

#### `App.jsx` - Root Component

The main layout coordinator:

- Tab navigation (Download, Library, Settings)
- Top-level error boundaries
- Global keyboard shortcuts
- Theme management (if implemented)

```jsx
export default function App() {
  const [store, actions] = useAppStore();
  
  return (
    <div class="app-container">
      <Header activeTab={store.ui.activeTab} onTabChange={actions.setActiveTab} />
      <main>
        <Show when={store.ui.activeTab === 'download'}>
          <DownloadView />
        </Show>
        <Show when={store.ui.activeTab === 'library'}>
          <LibraryView />
        </Show>
        <Show when={store.ui.activeTab === 'settings'}>
          <SettingsView />
        </Show>
      </main>
    </div>
  );
}
```

#### `DownloadView.jsx` - Download Interface

Handles the main download workflow:

- URL input textarea (supports multiple URLs)
- Download options form
- "Advanced" toggle for power users
- Real-time progress display via SSE
- Download result logs
- Duplicate file prompt handling

**Key Features:**
- Validates URLs before submission
- Shows/hides advanced options
- Displays progress bars for active downloads
- Handles SSE reconnection
- Responds to duplicate file prompts

#### `LibraryView.jsx` - Media Library

Displays downloaded media:

- Grid/list view of downloaded files
- Search and filter functionality
- Sort options (date, title, size)
- Media metadata display
- Play/delete actions

**Data Source:**
- Fetches from `/api/media/` endpoint
- Supports pagination
- Caches results for performance

#### `SettingsView.jsx` - Configuration

User preferences and settings:

- Default download options
- Output template configuration
- Cookie management
- File format preferences
- Duplicate file policy

**Persistence:**
- Saves to localStorage
- Applied as defaults for new downloads

#### `Player.jsx` - Media Playback

Simple media player:

- Video playback with HTML5 `<video>`
- Audio playback with HTML5 `<audio>`
- Basic controls (play, pause, seek, volume)
- Fullscreen support for video

## State Management

State is centralized in `store/appStore.jsx` using Solid's context + store primitives. The app persists key UI fields in localStorage so layout and form choices survive reloads.

### Store Structure

```js
const initialState = {
  ui: {
    activeTab: 'download',      // Current view
    isAdvanced: false,          // Power user mode toggle
  },
  settings: {
    output: '{title}.{ext}',    // Output template
    quality: 'best',            // Video quality
    audioOnly: false,           // Audio-only mode
    jobs: 1,                    // Concurrent downloads
    onDuplicate: 'prompt',      // Duplicate file policy
  },
  download: {
    urlInput: '',               // Current URL draft
    activeJobs: {},             // Runtime job status
    logs: [],                   // Download logs
  },
  library: {
    items: [],                  // Downloaded media
    filter: '',                 // Search filter
    sortBy: 'date',             // Sort field
  },
};
```

### Store Actions

```js
const actions = {
  // UI actions
  setActiveTab(tab) { ... },
  toggleAdvanced() { ... },
  
  // Settings actions
  updateSettings(updates) { ... },
  
  // Download actions
  setUrlInput(text) { ... },
  submitDownload(urls, options) { ... },
  updateJobProgress(jobId, progress) { ... },
  handleDuplicatePrompt(jobId, promptId, choice) { ... },
  
  // Library actions
  fetchLibrary() { ... },
  deleteMedia(id) { ... },
  filterLibrary(query) { ... },
};
```

### Persistence Strategy

Only durable state is persisted to localStorage:

**Persisted:**
- `ui.activeTab` - Remember last viewed tab
- `ui.isAdvanced` - Remember advanced mode preference
- `settings.*` - All user preferences
- `download.urlInput` - Preserve draft URLs

**Not Persisted (runtime only):**
- `download.activeJobs` - Active download status/progress
- `download.logs` - Download logs
- `library.items` - Fetched from server on demand

### Store Provider

```jsx
// index.jsx
import { AppStoreProvider } from './store/appStore';

render(() => (
  <AppStoreProvider>
    <App />
  </AppStoreProvider>
), document.getElementById('root'));
```

## Build & Integration Pipeline

### Development Workflow

1. **Start dev server:**
   ```bash
   cd frontend
   npm run dev
   ```
   - Vite dev server starts on `http://localhost:5173`
   - Hot module reload enabled
   - API calls proxy to backend via `vite.config.js`

2. **Configure backend proxy:**
   ```bash
   # Default: http://127.0.0.1:8080
   VITE_API_PROXY_TARGET=http://127.0.0.1:9090 npm run dev
   ```

3. **API proxy configuration:**
   ```js
   // vite.config.js
   export default defineConfig({
     server: {
       proxy: {
         '/api': {
           target: process.env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080',
           changeOrigin: true,
         },
       },
     },
   });
   ```

### Production Build

1. **Build static assets:**
   ```bash
   cd frontend
   npm run build
   ```
   - Output: `../internal/web/assets/`
   - Minified JS and CSS
   - Hashed filenames for cache busting

2. **Go embed:**
   ```go
   // internal/web/server.go
   //go:embed assets/*
   var assetsFS embed.FS
   
   func ServeAssets() http.Handler {
       return http.FileServer(http.FS(assetsFS))
   }
   ```

3. **Runtime:**
   - Go server serves SPA at `/`
   - API endpoints at `/api/*`
   - All assets bundled in binary

### Full-Stack Build

```bash
# From repository root
./build.sh

# Or manually:
cd frontend && npm run build && cd ..
go build -o ytdl-go .
```

## API Communication

The frontend communicates with the Go backend via JSON over HTTP and Server-Sent Events (SSE) for real-time updates.

### REST Endpoints

#### POST `/api/download`

Start a download job:

```js
async function startDownload(urls, options) {
  const response = await fetch('/api/download', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ urls, options }),
  });
  
  const data = await response.json();
  return data.jobId;
}
```

#### GET `/api/media/`

List downloaded media:

```js
async function fetchLibrary(offset = 0, limit = 200) {
  const response = await fetch(`/api/media/?offset=${offset}&limit=${limit}`);
  const data = await response.json();
  return data.items;
}
```

#### GET `/api/media/{filename}`

Serve media file:

```jsx
<video src={`/api/media/${filename}`} controls />
```

### Server-Sent Events (SSE)

Real-time download progress:

```js
function subscribeToJob(jobId, onUpdate) {
  const eventSource = new EventSource(`/api/download/progress?id=${jobId}`);
  
  eventSource.onmessage = (event) => {
    const data = JSON.parse(event.data);
    onUpdate(data);
  };
  
  eventSource.onerror = () => {
    console.error('SSE connection error');
    eventSource.close();
  };
  
  return () => eventSource.close();
}
```

**Event Types:**
- `snapshot` - Initial state on connect
- `status` - Job status change
- `register` - New task registered
- `progress` - Download progress update
- `finish` - Task completed
- `log` - Log message
- `duplicate` - Duplicate file prompt
- `duplicate-resolved` - Prompt resolved
- `done` - Job complete

See [API Reference](../../api-reference/events) for detailed event schemas.

### Error Handling

```js
async function apiCall(url, options) {
  try {
    const response = await fetch(url, options);
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Request failed');
    }
    
    return await response.json();
  } catch (err) {
    console.error('API Error:', err);
    // Show user-friendly error message
    showNotification('error', err.message);
    throw err;
  }
}
```

## Styling Approach

### Tailwind CSS

Utility-first CSS for rapid development:

```jsx
<button class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
  Download
</button>
```

**Benefits:**
- **No custom CSS files** - Everything inline with utilities
- **Consistent spacing** - Predefined scale (4px base)
- **Responsive design** - Built-in breakpoint modifiers
- **Purging** - Unused classes removed in production

### Custom CSS

Only for highly specific cases:

```css
/* index.css */

/* Custom scrollbar styling */
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

Lucide icons via `createIcons()`:

```jsx
import { createEffect } from 'solid-js';
import { createIcons, Download, Play } from 'lucide';

function MyComponent() {
  createEffect(() => {
    createIcons({
      icons: { Download, Play },
    });
  });
  
  return (
    <div>
      <i data-lucide="download"></i>
      <i data-lucide="play"></i>
    </div>
  );
}
```

> **Important:** Always re-initialize icons in `createEffect` if the DOM structure changes significantly (e.g., after conditional rendering).

## Performance Considerations

### Bundle Size

Target: **< 200KB** total (gzipped)

- SolidJS runtime: ~6KB
- Tailwind CSS (purged): ~20KB
- Application code: ~50KB
- Total: ~76KB (well under target)

### Code Splitting

Not currently implemented due to small bundle size. Consider if app grows:

```js
// Example: Lazy load LibraryView
import { lazy } from 'solid-js';

const LibraryView = lazy(() => import('./components/LibraryView'));
```

### SSE Connection Management

- **Single connection per job** - Avoid connection spam
- **Auto-reconnect** - Handle network interruptions
- **Close on unmount** - Prevent memory leaks

```js
onCleanup(() => {
  eventSource.close();
});
```

## Development Best Practices

### Component Structure

- **Keep components small** - Single responsibility
- **Use functional components** - No classes
- **Prefer composition** - Over prop drilling

### File Naming

- **PascalCase for components** - `DownloadView.jsx`
- **camelCase for utilities** - `apiClient.js`

### Styling

- **Tailwind utilities first** - Only custom CSS when necessary
- **Consistent spacing** - Use Tailwind scale (4px base)
- **Mobile-first** - Use responsive modifiers (`md:`, `lg:`)

### State Management

- **Centralize in store** - Avoid prop drilling
- **Persist settings** - Remember user preferences
- **Don't persist runtime state** - Fetch fresh from server

## Related Documentation

- [API Reference](../../api-reference) - REST endpoints and SSE events
- [Contributing Guide](../../contributing/frontend) - Frontend development workflow
- [Architecture Overview](overview) - System-wide architecture
