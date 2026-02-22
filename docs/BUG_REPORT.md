# ytdl-go: Bug Report and Actionable Items

This document lists identified bugs, potential issues, and areas for improvement in the `ytdl-go` project. For each issue, actionable items are provided, along with a list of tangential documents that may need to be updated once the fix is complete.

---

## Backend

### File: `internal/app/runner.go`

*   **Issue:** Potential goroutine leak on context cancellation.
    *   **Description:** If the context is canceled while a worker goroutine is blocked trying to send a result to the `results` channel, the goroutine may not exit cleanly.
    *   **Actionable Items:**
        1.  In the worker goroutine, use a `select` statement to attempt to send the result, with a `default` case to handle the situation where the `results` channel is full.
        2.  If the `results` channel is full, check if the context has been canceled. If it has, exit the goroutine.
    *   **Tangential Documents:**
        *   `docs/ARCHITECTURE.md`: Update the concurrency model diagram and description to reflect the changes.

*   **Issue:** `goto` statement makes control flow harder to understand.
    *   **Description:** The `goto done` statement can be replaced with more conventional control flow constructs.
    *   **Actionable Items:**
        1.  Refactor the code to use a `break` statement to exit the loop.
    *   **Tangential Documents:** None.

### File: `internal/downloader/downloader.go`

*   **Issue:** Global YouTube client is not thread-safe.
    *   **Description:** Modifying the `youtube.DefaultClient` global variable can lead to race conditions.
    *   **Actionable Items:**
        1.  Instantiate a new `youtube.Client` in the `ProcessWithManager` function.
        2.  Pass the client instance to all the functions that need it.
        3.  Remove the code that modifies the global `youtube.DefaultClient`.
    *   **Tangential Documents:**
        *   `docs/ARCHITECTURE.md`: Update the module breakdown and data flow diagrams.

*   **Issue:** `ProcessWithManager` function is too large and complex.
    *   **Description:** The function handles too many different cases, making it hard to maintain.
    *   **Actionable Items:**
        1.  Create separate functions for handling playlists, direct URLs, and YouTube videos.
        2.  The `ProcessWithManager` function should be responsible for detecting the URL type and calling the appropriate handler function.
    *   **Tangential Documents:**
        *   `docs/ARCHITECTURE.md`: Update the module breakdown and data flow diagrams.

*   **Issue:** Tightly coupled TUI and downloader logic.
    *   **Description:** The `renderFormats` function is responsible for both rendering the TUI and initiating the download.
    *   **Actionable Items:**
        1.  Refactor `renderFormats` to only be responsible for displaying the available formats and returning the user's selection.
        2.  The download logic should be handled by a separate function that is called after the user has made their selection.
    *   **Tangential Documents:**
        *   `docs/ARCHITECTURE.md`: Update the TUI model description.

### File: `internal/web/server.go`

*   **Issue:** Large handler functions.
    *   **Description:** The `ListenAndServe` function contains several large handler functions defined as closures.
    *   **Actionable Items:**
        1.  Create a new `internal/web/handlers` package.
        2.  For each API endpoint, create a new file in the `handlers` package with a dedicated handler function.
        3.  The `ListenAndServe` function should be responsible for setting up the router and registering the handlers.
    *   **Tangential Documents:**
        *   `docs/ARCHITECTURE.md`: Update the web server module description.

*   **Issue:** Potential path traversal vulnerability in `resolveMediaPath`.
    *   **Description:** The `resolveMediaPath` function is security-sensitive and may not handle all edge cases correctly.
    *   **Actionable Items:**
        1.  Research and select a well-tested, third-party library for handling file paths and preventing path traversal attacks.
        2.  Replace the custom `resolveMediaPath` function with the chosen library.
    *   **Tangential Documents:**
        *   `SECURITY.md`: Document the new library and the steps taken to mitigate path traversal vulnerabilities.

---

## Frontend

### File: `frontend/src/App.jsx`

*   **Issue:** `App.jsx` is a monolithic component.
    *   **Description:** The file is too large and contains too much logic, making it difficult to maintain.
    *   **Actionable Items:**
        1.  Create a `frontend/src/components/layout` directory.
        2.  Create `Sidebar.jsx` and `Header.jsx` components in the new directory.
        3.  Move the relevant JSX and logic from `App.jsx` into the new components.
        4.  Refactor `App.jsx` to be a top-level layout component that composes the other components.
    *   **Tangential Documents:**
        *   `frontend/docs/ARCHITECTURE.md`: Update the component hierarchy diagram.

### File: `frontend/src/store/appStore.jsx`

*   **Issue:** Monolithic state logic.
    *   **Description:** All state management logic is in one file.
    *   **Actionable Items:**
        1.  Create a `frontend/src/store/modules` directory.
        2.  Create separate files for managing the `library`, `settings`, `player`, and `download` state.
        3.  Each file should export a SolidJS store and the actions that can be performed on that store.
        4.  Refactor `appStore.jsx` to be a top-level store that combines the other stores.
    *   **Tangential Documents:**
        *   `frontend/docs/ARCHITECTURE.md`: Update the state management diagram and description.

*   **Issue:** Lack of type safety.
    *   **Description:** The code uses plain JavaScript objects for the state, which can lead to bugs.
    *   **Actionable Items:**
        1.  Introduce TypeScript to the project.
        2.  Create type definitions for the application state and the props for each component.
        3.  Convert the `.jsx` files to `.tsx`.
    *   **Tangential Documents:**
        *   `CONTRIBUTING.md`: Add instructions for setting up a TypeScript development environment.
        *   `frontend/docs/ARCHITECTURE.md`: Update the architecture to reflect the use of TypeScript.
