# Implementation Plan: BgUtils Integration

## Phase 1: Research and Setup
- [ ] Task: Research `BgUtils` Go integration patterns and `kkdai/youtube` token injection points.
- [ ] Task: Add `BgUtils` as a dependency to the project.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Research and Setup' (Protocol in workflow.md)

## Phase 2: Core Implementation
- [ ] Task: Implement PO Token generation utility.
    - [ ] Write tests for token generation.
    - [ ] Implement `BgUtils` wrapper for token generation.
- [ ] Task: Implement Bot Attestation logic.
    - [ ] Write tests for attestation logic.
    - [ ] Implement attestation using `BgUtils`.
- [ ] Task: Integrate with Downloader.
    - [ ] Write tests for downloader integration.
    - [ ] Update `internal/downloader` to use the new token/attestation logic.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Core Implementation' (Protocol in workflow.md)

## Phase 3: Verification and Polling
- [ ] Task: Verify integration with real YouTube URLs known to require attestation.
- [ ] Task: Ensure TUI and Web UI correctly handle any new error states related to bot attestation.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Verification and Polling' (Protocol in workflow.md)
