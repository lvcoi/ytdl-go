# Specification: BgUtils Integration

## Overview
Integrate `BgUtils` into the `ytdl-go` backend to handle PO Token and bot attestation. This will improve the reliability of YouTube requests and reduce the frequency of 403 Forbidden errors or bot detection blocks.

## Functional Requirements
- **PO Token Generation:** Integrate `BgUtils` to generate valid Proof of Origin (PO) tokens.
- **Bot Attestation:** Implement bot attestation logic using `BgUtils` to satisfy YouTube's security requirements.
- **Integration with `kkdai/youtube`:** Ensure the generated tokens and attestation data are correctly passed to the underlying YouTube library.
- **Fallback Mechanism:** Maintain existing download logic as a fallback if `BgUtils` fails or is unavailable.

## Non-Functional Requirements
- **Performance:** Token generation should not significantly delay the start of a download.
- **Reliability:** The integration should handle network errors and updates to the `BgUtils` library gracefully.

## Acceptance Criteria
- [ ] Downloads that previously failed with bot detection errors now succeed.
- [ ] PO tokens are successfully generated and logged (in debug mode).
- [ ] The system falls back to standard downloading if `BgUtils` integration is disabled or fails.
