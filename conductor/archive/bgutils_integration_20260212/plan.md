# Implementation Plan: BgUtils Integration

## Phase 1: Research and Setup
- [x] Task: Research `BgUtils` Go integration patterns and `kkdai/youtube` token injection points.
- [x] Task: Add `BgUtils` as a dependency to the project.
- [x] Task: Conductor - User Manual Verification 'Phase 1: Research and Setup' (Protocol in workflow.md)

## Phase 2: Core Implementation
- [x] Task: Implement PO Token generation utility.
    - [x] Write tests for token generation.
    - [x] Implement `BgUtils` wrapper for token generation.
- [x] Task: Implement Bot Attestation logic.
    - [x] Write tests for attestation logic.
    - [x] Implement attestation using `BgUtils`.
- [x] Task: Integrate with Downloader.
    - [x] Write tests for downloader integration.
    - [x] Update `internal/downloader` to use the new token/attestation logic.
- [x] Task: Conductor - User Manual Verification 'Phase 2: Core Implementation' (Protocol in workflow.md)

## Phase 3: Verification and Polling
- [x] Task: Verify integration with real YouTube URLs known to require attestation.
- [x] Task: Ensure TUI and Web UI correctly handle any new error states related to bot attestation.
- [x] Task: Conductor - User Manual Verification 'Phase 3: Verification and Polling' (Protocol in workflow.md)
