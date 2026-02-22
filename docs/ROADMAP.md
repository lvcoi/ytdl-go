# ytdl-go: AI-Powered MVP Roadmap

This document provides a structured, AI-first roadmap to guide the `ytdl-go` project to a stable and feature-rich Minimum Viable Product (MVP). The primary goal is to create a clear, actionable plan that an AI developer can follow to produce a high-quality, stable application, while minimizing regressions.

## Guiding Principles for the AI

*   **Simplicity and Stability Over premature feature creep:** Focus on making the core functionality robust and reliable before adding new features.
*   **Test-Driven Development (TDD):** For every new feature or bug fix, especially in the backend, a corresponding test must be created to prevent regressions.
*   **Incremental and Verifiable Steps:** Each step in this roadmap should be a small, verifiable change that can be tested and validated before moving on.
*   **User-Centric Design:** Both the TUI and Web UI should be intuitive and easy to use for non-technical users, with advanced features accessible but not intrusive.

---

## Phase 1: Stabilize the Backend

This is the most critical phase. A stable backend is the foundation for the entire application. The focus here is on reliability, error handling, and creating a robust testing suite.

### 1.1. Isolate and Refactor Core YouTube API Logic

*   **Goal:** Create a dedicated, well-documented, and easily testable package for all YouTube API interactions.
*   **Tasks:**
    1.  Create a new package: `internal/youtube_api`.
    2.  Move all code related to YouTube API requests, response parsing, and token handling from `internal/downloader` to this new package.
    3.  Refactor the code to be more modular and easier to understand. For example, have separate functions for getting video metadata, getting stream URLs, and handling API tokens.
    4.  Implement a robust error handling mechanism that can distinguish between different types of errors (e.g., network errors, API errors, private videos, etc.).

### 1.2. Implement a Comprehensive Test Suite for the YouTube API

*   **Goal:** Create a suite of tests that cover all aspects of the `youtube_api` package.
*   **Tasks:**
    1.  **Mocking:** Use mocking to simulate YouTube API responses. This will allow for testing without making actual network requests, which is faster and more reliable.
    2.  **Test Cases:** Create test cases for:
        *   Successful video metadata retrieval.
        *   Successful stream URL retrieval.
        *   Handling of various video types (public, private, unlisted, age-restricted).
        *   Handling of API errors (e.g., 403 Forbidden, 429 Too Many Requests).
        *   Token refresh logic (if applicable).
    3.  **Integration Tests:** Create a small number of integration tests that make real requests to the YouTube API to ensure that the mocked responses are still valid. These should be run sparingly.

### 1.3. Address PO Token Handling

*   **Goal:** Implement a reliable and extensible system for managing and utilizing PO tokens.
*   **Tasks:**
    1.  **Token Provider Interface:** Create an interface for a "PO Token Provider" that can be implemented by different sources (e.g., a local file, a remote server, or a browser extension).
    2.  **Configuration:** Allow the user to configure which PO Token Provider to use via the `settings.json` file.
    3.  **Extensibility:** Document how to create a new PO Token Provider, to encourage community contributions.

---

## Phase 2: TUI and CLI Refinement

The TUI and CLI are the primary interfaces for many users. They should be powerful, yet easy to use.

### 2.1. Simplify the TUI by Default

*   **Goal:** Make the TUI more approachable for new users.
*   **Tasks:**
    1.  **Simplified View:** By default, show a simplified view that only displays the most important information (e.g., quality, size, and format).
    2.  **Advanced View:** Add a keybinding (e.g., `a`) to toggle an "Advanced View" that shows all available information (e.g., codecs, bitrate, etc.).
    3.  **Help View:** Add a help view (e.g., `?`) that explains the keybindings and options.

### 2.2. Improve CLI Output and Error Reporting

*   **Goal:** Provide clearer and more informative output and error messages in the CLI.
*   **Tasks:**
    1.  **Human-Readable Errors:** When an error occurs, provide a clear, human-readable error message that explains what went wrong and how to fix it (if possible).
    2.  **JSON Output for Errors:** When the `--json` flag is used, ensure that all errors are also output in a structured JSON format.
    3.  **Progress Bar:** Improve the progress bar to show more information, such as the download speed, ETA, and the size of the file.

---

## Phase 3: Web UI (SolidJS) MVP

The SolidJS-based Web UI is the most user-friendly way to interact with `ytdl-go`. The goal is to get it to a stable MVP that is both functional and easy to use.

### 3.1. Refactor `App.jsx` into Smaller Components

*   **Goal:** Improve the maintainability and readability of the frontend code by breaking down the massive `App.jsx` component.
*   **Tasks:**
    1.  **Component Identification:** Identify logical sections of `App.jsx` that can be extracted into their own components. For example:
        *   `Sidebar.jsx`
        *   `Header.jsx`
        *   `Player.jsx`
        *   `DownloadView.jsx`
        *   `LibraryView.jsx`
        *   `SettingsView.jsx`
    2.  **State Management:** For each new component, determine what state it needs and how it will be passed down from the parent component. Use SolidJS stores and props effectively.
    3.  **Component Creation:** Create the new component files and move the relevant JSX and logic from `App.jsx` into them.

### 3.2. Implement a Robust API Client for the Frontend

*   **Goal:** Create a dedicated API client for the frontend to handle all communication with the backend.
*   **Tasks:**
    1.  **API Client Module:** Create a new file, `src/utils/apiClient.js`, that exports functions for making requests to the backend API.
    2.  **Endpoint Functions:** For each backend endpoint, create a corresponding function in the API client (e.g., `getMediaFiles()`, `downloadMedia()`, etc.).
    3.  **Error Handling:** The API client should handle all network errors and API errors, and return them in a consistent format.
    4.  **Use in Components:** Refactor the components to use the new API client instead of making `fetch` requests directly.

### 3.3. Web UI MVP Feature Checklist

*   **Goal:** Implement the core features required for a functional and user-friendly Web UI.
*   **Tasks:**
    1.  **Download View:**
        *   Input for YouTube URL.
        *   Basic format selection (e.g., "Best Video", "Best Audio").
        *   Download button.
        *   Real-time progress of downloads.
    2.  **Library View:**
        *   Display a list of downloaded media.
        *   Ability to play media in the player.
        *   Basic filtering and sorting.
    3.  **Player:**
        *   Playback controls (play, pause, seek).
        *   Volume control.
        *   Display basic media information (title, artist).
    4.  **Settings:**
        *   Configure download directory.
        *   Configure PO Token Provider.

---

## Phase 4: Advanced Features and Future Work

Once the MVP is stable, the following features can be implemented.

*   **Plugin System:** Implement a plugin system for both the backend and frontend, to allow for community contributions.
*   **Expanded Format Support:** Add support for more video and audio formats, as well as subtitles.
*   **Advanced Web UI Features:** Implement advanced features in the Web UI, such as playlist management, metadata editing, and a download queue.
*   **Cross-Platform Build and Packaging:** Create a simple build and packaging process for creating distributable binaries for Windows, macOS, and Linux.

This roadmap provides a clear path forward for the `ytdl-go` project. By following these steps, the AI can help to create a stable, feature-rich, and user-friendly application.
