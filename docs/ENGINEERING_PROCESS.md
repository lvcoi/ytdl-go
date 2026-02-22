# ytdl-go: Engineering Process

This document outlines the standardized engineering process for the `ytdl-go` project. This process is designed to be followed by any developer, including an AI, to ensure consistency, quality, and to minimize regressions.

## Guiding Principles

*   **Clarity and Simplicity:** The process should be easy to understand and follow.
*   **Automation:** Automate as much of the process as possible to reduce manual effort and errors.
*   **Consistency:** Every change, no matter how small, should follow the same process.
*   **Quality:** The process should be designed to produce high-quality, well-tested code.

---

## The Development Lifecycle

All changes to the codebase, including bug fixes, new features, and refactoring, must follow this lifecycle.

### 1. Issue Creation and Assignment

*   **Bug Fixes:** Before working on a bug fix, an issue must be created in the project's issue tracker. The issue should include a clear description of the bug, steps to reproduce it, and any relevant logs or screenshots. The issue should be assigned to the developer who will be working on the fix.
*   **New Features:** For new features, a design document should be created and approved before any code is written. The design document should outline the goals of the feature, the proposed implementation, and any potential risks or trade-offs. Once the design is approved, an issue should be created and assigned.

### 2. Branching

*   All work must be done in a feature branch, created from the `main` branch.
*   The branch name should be descriptive and include the issue number, for example: `feature/123-add-subtitle-support` or `bugfix/456-fix-goroutine-leak`.

### 3. Development

*   **Test-Driven Development (TDD):** For every bug fix or new feature, a corresponding test must be created *before* the code is written. The test should initially fail, and then pass once the code is implemented correctly.
*   **Backend (Go):**
    *   All new code must be accompanied by unit tests.
    *   Run `go test ./...` to run all tests.
    *   Run `go fmt ./...` to format the code.
    *   Run `go vet ./...` to run the vet tool.
*   **Frontend (SolidJS):**
    *   All new components must be accompanied by unit tests.
    *   Run `npm test` to run all tests.
    *   Run `npm run lint` to lint the code.
*   **Commit Messages:** Commit messages should be clear, concise, and follow the Conventional Commits specification.

### 4. Code Review

*   When the development is complete, a pull request (PR) should be opened to merge the feature branch into the `main` branch.
*   The PR description should include a clear explanation of the changes, a link to the issue, and any relevant testing instructions.
*   At least one other developer must review and approve the PR before it can be merged. For AI developers, a human must review the code.
*   The reviewer should check for correctness, style, and adherence to the project's best practices.

### 5. Merging and Deployment

*   Once the PR is approved, it can be merged into the `main` branch.
*   The `main` branch is automatically deployed to the staging environment.
*   After the changes have been verified in the staging environment, they can be deployed to production by creating a new release.

---

## Tooling and Automation

*   **Continuous Integration (CI):** A CI pipeline is configured to run on every push to every branch. The pipeline will:
    *   Run all backend and frontend tests.
    *   Run the linters.
    *   Build the application.
*   **Dependabot:** Dependabot is configured to automatically open PRs to update the project's dependencies. These PRs must be reviewed and tested before being merged.
*   **Code Coverage:** The CI pipeline will report the code coverage for both the backend and frontend. The code coverage must not decrease.
