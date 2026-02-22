---
title: "Code Style Guide"
weight: 40
---

# Code Style Guide

This document defines coding standards for Go and JavaScript code in ytdl-go.

## Table of Contents

- [Go Style Guide](#go-style-guide)
- [JavaScript Style Guide](#javascript-style-guide)
- [General Principles](#general-principles)

## Go Style Guide

### Formatting

Use standard Go formatting tools:

```bash
# Format code
go fmt ./...

# Or use gofmt directly
gofmt -w .
```

**Rules:**
- **Tabs for indentation** (not spaces)
- **No trailing whitespace**
- **Line length:** Aim for < 100 characters, hard limit 120

### Naming Conventions

#### Variables and Functions

```go
// Use camelCase for private
var downloadCount int
func processURL(url string) error { ... }

// Use PascalCase for exported
var MaxRetries int
func DownloadVideo(url string) error { ... }
```

#### Constants

```go
// Use PascalCase for exported constants
const DefaultQuality = "best"
const MaxConcurrency = 10

// Use camelCase for private constants
const retryDelay = 1 * time.Second
```

#### Types

```go
// Use PascalCase for exported types
type Video struct { ... }
type DownloadOptions struct { ... }

// Use camelCase for private types
type progressState struct { ... }
```

#### Acronyms

Keep acronyms capitalized when exported:

```go
// Good
type HTTPClient struct { ... }
type URLParser interface { ... }
func ParseURL(url string) error { ... }

// Avoid
type HttpClient struct { ... }
type UrlParser interface { ... }
func ParseUrl(url string) error { ... }
```

### Documentation

Document all exported identifiers:

```go
// Video represents a YouTube video with metadata and available formats.
type Video struct {
    ID          string
    Title       string
    Formats     []Format
}

// Download downloads a video from the given URL.
// It returns an error if the download fails.
func Download(ctx context.Context, url string) error {
    // Implementation
}
```

**Comment style:**
- Start with the name of the identifier
- Use complete sentences
- Explain what, not how (code shows how)

### Error Handling

#### Return Errors, Don't Panic

```go
// Good
func readFile(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read %s: %w", path, err)
    }
    return data, nil
}

// Avoid (unless truly exceptional)
func readFile(path string) []byte {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err)  // Don't panic!
    }
    return data
}
```

#### Wrap Errors with Context

```go
if err := downloadFile(url); err != nil {
    return fmt.Errorf("failed to download %s: %w", url, err)
}
```

#### Check Errors Immediately

```go
// Good
f, err := os.Open(filename)
if err != nil {
    return err
}
defer f.Close()

// Avoid
f, err := os.Open(filename)
defer f.Close()
if err != nil {  // defer called even if err != nil!
    return err
}
```

### Context Usage

Always pass `context.Context` as the first parameter:

```go
// Good
func Download(ctx context.Context, url string, opts Options) error { ... }

// Avoid
func Download(url string, opts Options, ctx context.Context) error { ... }
```

Check context cancellation in loops:

```go
for _, item := range items {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        process(item)
    }
}
```

### Concurrency

#### Use Channels for Communication

```go
// Good
results := make(chan Result, len(items))
for _, item := range items {
    go func(i Item) {
        results <- process(i)
    }(item)
}

// Collect results
for range items {
    result := <-results
    handleResult(result)
}
```

#### Protect Shared State

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

Or use atomic operations:

```go
type Counter struct {
    count atomic.Int64
}

func (c *Counter) Inc() {
    c.count.Add(1)
}
```

### Struct Organization

Group related fields, separate with blank lines:

```go
type Video struct {
    // Identity
    ID    string
    Title string
    
    // Metadata
    Author   string
    Duration time.Duration
    
    // Available formats
    Formats []Format
    
    // Timestamps
    UploadDate time.Time
    FetchedAt  time.Time
}
```

### Function Parameters

Keep parameter lists short (< 5 parameters). Use options structs for many parameters:

```go
// Good
type DownloadOptions struct {
    Quality    string
    AudioOnly  bool
    Format     string
    Jobs       int
    Timeout    time.Duration
}

func Download(ctx context.Context, url string, opts DownloadOptions) error { ... }

// Avoid
func Download(ctx context.Context, url, quality, format string, audioOnly bool, jobs int, timeout time.Duration) error { ... }
```

### Early Returns

Return early to reduce nesting:

```go
// Good
func process(item Item) error {
    if item.IsEmpty() {
        return nil
    }
    
    if err := validate(item); err != nil {
        return err
    }
    
    return process(item)
}

// Avoid
func process(item Item) error {
    if !item.IsEmpty() {
        if err := validate(item); err == nil {
            return process(item)
        } else {
            return err
        }
    }
    return nil
}
```

## JavaScript Style Guide

### Formatting

Use Prettier (if configured) or consistent manual formatting:

**Rules:**
- **2 spaces for indentation**
- **Semicolons: Optional (but be consistent)**
- **Line length:** Aim for < 80 characters
- **Single quotes** for strings (unless template literals)

### Naming Conventions

#### Variables and Functions

```js
// Use camelCase
const downloadCount = 0;
function processUrl(url) { ... }

// Constants: UPPER_SNAKE_CASE
const MAX_RETRIES = 3;
const API_BASE_URL = '/api';
```

#### Components

```jsx
// Use PascalCase for components
export default function DownloadView() { ... }
export default function MediaPlayer() { ... }
```

#### Files

```
components/DownloadView.jsx     // PascalCase for components
utils/apiClient.js              // camelCase for utilities
store/appStore.jsx              // camelCase for stores
```

### Component Structure

Organize components consistently:

```jsx
// 1. Imports
import { createSignal, Show } from 'solid-js';
import { useAppStore } from '../store/appStore';

// 2. Component
export default function MyComponent(props) {
  // 3. Store/context
  const [store, actions] = useAppStore();
  
  // 4. Local state
  const [count, setCount] = createSignal(0);
  
  // 5. Effects
  createEffect(() => {
    console.log('Count changed:', count());
  });
  
  // 6. Event handlers
  const handleClick = () => {
    setCount(count() + 1);
  };
  
  // 7. Render
  return (
    <div>
      <p>Count: {count()}</p>
      <button onClick={handleClick}>Increment</button>
    </div>
  );
}
```

### Props

Use destructuring for clarity:

```jsx
// Good
export default function MediaCard({ title, artist, size, onPlay }) {
  return <div>...</div>;
}

// Avoid
export default function MediaCard(props) {
  return <div>{props.title} - {props.artist}</div>;
}
```

### Conditional Rendering

Use SolidJS primitives:

```jsx
// Good
<Show when={!loading()} fallback={<Spinner />}>
  <Content />
</Show>

// Avoid
{!loading() ? <Content /> : <Spinner />}
```

### Lists

Use `<For>` for lists:

```jsx
// Good
<For each={items()}>
  {(item) => <ItemCard item={item} />}
</For>

// Avoid
{items().map(item => <ItemCard item={item} />)}
```

### Event Handlers

Use arrow functions or function declarations:

```jsx
// Good
const handleClick = () => {
  console.log('clicked');
};

// Good
function handleSubmit(event) {
  event.preventDefault();
  // ...
}

// Avoid inline for complex logic
<button onClick={() => {
  // 20 lines of logic
}}>
```

### API Calls

Separate API logic from components:

```js
// utils/apiClient.js
export async function startDownload(urls, options) {
  const response = await fetch('/api/download', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ urls, options }),
  });
  
  if (!response.ok) {
    throw new Error('Download failed');
  }
  
  return response.json();
}

// components/DownloadView.jsx
import { startDownload } from '../utils/apiClient';

const handleDownload = async () => {
  try {
    const result = await startDownload(urls, options);
    // Handle success
  } catch (error) {
    // Handle error
  }
};
```

### Comments

Comment complex logic, not obvious code:

```jsx
// Good
// Exponential backoff: 1s, 2s, 4s, 8s, ...
const delay = Math.pow(2, attempt) * 1000;

// Unnecessary
// Set count to 0
const [count, setCount] = createSignal(0);
```

## General Principles

### 1. Keep It Simple

Write simple, readable code over clever code:

```go
// Good
if len(items) == 0 {
    return nil
}

// Avoid (too clever)
if len(items) < 1 {
    return nil
}
```

### 2. Don't Repeat Yourself (DRY)

Extract repeated code into functions:

```go
// Before
fmt.Printf("Error downloading %s: %v\n", url1, err1)
fmt.Printf("Error downloading %s: %v\n", url2, err2)

// After
func logError(url string, err error) {
    fmt.Printf("Error downloading %s: %v\n", url, err)
}
logError(url1, err1)
logError(url2, err2)
```

### 3. Single Responsibility

Each function/component should do one thing:

```go
// Bad: Does too much
func downloadAndProcessAndSave(url string) error {
    // 100+ lines
}

// Good: Focused functions
func download(url string) ([]byte, error) { ... }
func process(data []byte) ([]byte, error) { ... }
func save(data []byte, path string) error { ... }
```

### 4. Fail Fast

Check preconditions early:

```go
func process(input string) error {
    if input == "" {
        return errors.New("input cannot be empty")
    }
    
    // Process input
    return nil
}
```

### 5. Use Meaningful Names

Names should reveal intent:

```go
// Good
const maxConcurrentDownloads = 5
const downloadTimeout = 180 * time.Second

// Avoid
const max = 5
const t = 180
```

### 6. Avoid Magic Numbers

Use named constants:

```go
// Good
const (
    kilobyte = 1024
    megabyte = 1024 * kilobyte
)
chunkSize := 2 * megabyte

// Avoid
chunkSize := 2097152  // What does this mean?
```

### 7. Test-Friendly Code

Write code that's easy to test:

```go
// Good (testable)
func sanitizeFilename(filename string) string {
    // Pure function, easy to test
}

// Harder to test
func sanitizeAndSaveFile(filename string, data []byte) error {
    // Multiple responsibilities, harder to test
}
```

## Code Review Checklist

Before submitting code, check:

- [ ] Code is formatted (`go fmt`, Prettier)
- [ ] No linting errors (`golangci-lint run`)
- [ ] All exported identifiers documented
- [ ] Error handling is appropriate
- [ ] No magic numbers (use constants)
- [ ] Functions are focused (single responsibility)
- [ ] Tests pass (`go test ./...`)
- [ ] No obvious bugs or race conditions

## Related Documentation

- [Backend Development](backend) - Go-specific practices
- [Frontend Development](frontend) - JavaScript-specific practices
- [Pull Request Process](pull-request-process) - PR guidelines
