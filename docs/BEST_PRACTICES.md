# ytdl-go: Best Practices

This document provides a set of specific and actionable best practices for developing the `ytdl-go` project. These practices are designed to ensure that the code is high-quality, maintainable, and consistent, leaving no room for interpretation.

## Backend (Go)

### Concurrent Download Management

*   **Worker Pools:** To manage concurrent downloads, use a fixed-size worker pool. The number of workers should be configurable via a command-line flag.

    *   **Example:**
        ```go
        // In main.go
        var numWorkers = flag.Int("jobs", 4, "Number of concurrent download jobs")

        // In runner.go
        func Run(ctx context.Context, urls []string, opts downloader.Options, jobs int) ([]Result, int) {
            // ...
            for i := 0; i < jobs; i++ {
                go worker(ctx, tasks, results)
            }
            // ...
        }
        ```

*   **Context for Cancellation:** Every function that performs a potentially long-running operation (e.g., an HTTP request, a file download) must accept a `context.Context` as its first argument. The function must listen for cancellation on the context's `Done()` channel and gracefully terminate when the context is canceled.

    *   **Example:**
        ```go
        func downloadFile(ctx context.Context, url string) error {
            req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
            // ...
        }
        ```

*   **Channels for Communication:** Use buffered channels for communication between goroutines. The buffer size should be carefully chosen to balance memory usage and performance. For example, the `results` channel in `runner.go` should have a buffer size equal to the number of URLs to be downloaded.

### Error Handling

*   **Custom Error Types:** For each distinct type of error, define a custom error type that implements the `error` interface. This allows for more granular error handling and testing.

    *   **Example:**
        ```go
        type ErrPrivateVideo struct {
            URL string
        }

        func (e *ErrPrivateVideo) Error() string {
            return fmt.Sprintf("video is private: %s", e.URL)
        }
        ```

*   **Error Wrapping:** Always wrap errors with additional context using `fmt.Errorf` and the `%w` verb. This creates a chain of errors that can be inspected to determine the root cause of the problem.

    *   **Example:**
        ```go
        if err != nil {
            return fmt.Errorf("failed to download video: %w", err)
        }
        ```

### General Go Best Practices

*   **Interfaces:** Use interfaces to decouple components and improve testability. For example, the `Downloader` should be an interface that can be implemented by different download strategies (e.g., a YouTube downloader, a direct URL downloader).
*   **Testing:** Every new function or method must have a corresponding unit test. Use the `testify` suite of packages (`assert`, `require`, `mock`) to write more expressive and concise tests.
*   **Dependencies:** Avoid global dependencies. All dependencies should be explicitly passed to the functions and types that need them.

---

## Frontend (SolidJS)

### UI/UX Design

*   **Controlled Components:** All form inputs must be controlled components. The value of the input should be stored in the component's state, and the `onChange` event should be used to update the state.
*   **Aria Attributes:** All interactive elements (buttons, links, form inputs, etc.) must have the appropriate ARIA attributes to ensure accessibility.
*   **Loading and Error States:** Every component that performs an asynchronous operation must have a clear loading state (e.g., a spinner) and an error state (e.g., an error message).

### State Management

*   **Store Structure:** The application state should be organized into a `stores` directory, with each store in its own file. Each store file should export the store and a set of actions for modifying the store.

    *   **Example (`stores/library.js`):**
        ```javascript
        import { createStore } from 'solid-js/store';

        const [library, setLibrary] = createStore({
          downloads: [],
        });

        export const libraryStore = library;

        export const addDownload = (download) => {
          setLibrary('downloads', (d) => [...d, download]);
        };
        ```

*   **Immutability:** The state must be treated as immutable. To update the state, create a new copy of the state with the desired changes, instead of modifying the existing state in place.

### General SolidJS Best Practices

*   **`For` vs. `Index`:** Use the `<For>` component for rendering lists of data. Use the `<Index>` component only when the list is static and the order of the items will not change.
*   **`Show` Component:** Use the `<Show>` component for conditional rendering. Avoid using the `&&` operator for conditional rendering, as it can lead to unexpected behavior.
*   **Props Destructuring:** When destructuring props, use the `splitProps` function to ensure that reactivity is preserved.

---

## Testing Practices (All Languages)

These rules apply to both backend (Go) and frontend (SolidJS) tests.

### Guard-Pair Testing

Every boolean guard or early-return condition requires **two** tests:

1.  **Block test:** Verify the guard prevents the action when the condition is not met.
2.  **Pass-through test:** Verify the action proceeds when the condition *is* met.

A guard that never opens is indistinguishable from a working guard if only the block side is tested. A passing block-only test can mask a bug where the guard is permanently closed.

*   **Example (bad — block-only):**
    ```javascript
    // Only proves saves don't happen on mount — never proves saves work after load
    it('does not save on mount', () => {
        render(() => <Component />);
        expect(localStorage.setItem).not.toHaveBeenCalled();
    });
    ```

*   **Example (good — guard pair):**
    ```javascript
    it('does not save on mount', () => {
        render(() => <Component />);
        expect(localStorage.setItem).not.toHaveBeenCalled();
    });

    it('saves after user modifies state', () => {
        render(() => <Component />);
        // trigger a user action that should persist
        fireEvent.click(screen.getByText('Toggle'));
        expect(localStorage.setItem).toHaveBeenCalled();
    });
    ```

### State-Transition Coverage

When a feature has distinct states (e.g., loading → loaded, idle → dragging → idle), tests must cover:

*   The **entry** into each state.
*   The **exit** from each state.
*   The **behavior** while in each state.

If a signal or flag controls a transition, at least one test must exercise the full round-trip (e.g., start drag → move → end drag → verify final state).

### Test Independence

*   Each test must set up its own preconditions. Do not rely on ordering or side effects from previous tests.
*   Mock state (e.g., `localStorage`, `fetch`) must be reset in `beforeEach` — not only in the global setup file. If a mock is defined in both the setup file and a test file, the test file's mock wins and the setup mock is wasted; prefer one authoritative location per mock.

## Definitions

*   **Issue:** An "issue" is any deviation from the best practices outlined in this document. This includes, but is not limited to:
    *   Code that does not follow the specified coding style.
    *   Missing or incomplete tests.
    *   Code that is overly complex or difficult to understand.
    *   Missing or incomplete documentation.
    *   Violations of the architectural principles.

*   **Improvement:** An "improvement" is any change that brings the code into closer alignment with the best practices outlined in this document. This includes, but is not limited to:
    *   Refactoring code to improve its readability, performance, or maintainability.
    *   Adding missing tests or improving existing tests.
    *   Adding or improving documentation.
    *   Fixing any of the "issues" defined above.

---

## Go, npm, and SolidJS Integration

*   **API Contract:** The API contract between the backend and the frontend must be formally defined in an OpenAPI (Swagger) specification. This specification should be used to generate the API client for the frontend.
*   **Build Script:** The `build.sh` script should be the single source of truth for building the entire application. It should be responsible for building the frontend, embedding the frontend assets into the Go binary, and building the final executable.
*   **Environment Variables:** Use environment variables for all configuration that differs between development and production. For example, the URL of the backend API should be an environment variable.
