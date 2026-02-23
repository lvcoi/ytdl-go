# PR #133 Review Spec

## Problem Statement

PR #133 ("Fix response body leak in retry transport and context-unaware sleep in doWithRetry") addresses two bugs in the HTTP retry infrastructure. This spec documents the review findings and the criteria for the PR to be merge-ready.

## Summary of Changes

The PR contains a single commit (`bd97844`) with changes to 3 files (128 additions, 2 deletions against the `web-ui` parent branch):

### 1. `internal/downloader/retry.go` — Response body leak fix

Two new `lastResp.Body.Close()` calls added to `RoundTrip`:

- **On non-retryable transport error**: When a retryable response was stored in `lastResp` and the next attempt returns a non-retryable error (e.g., DNS not-found), `lastResp.Body` was never closed. Fixed.
- **On successful/non-retryable status**: When a retryable response was stored and the next attempt succeeds, `lastResp.Body` was never closed. Fixed.

### 2. `internal/downloader/direct.go` — Context-aware retry sleep

`doWithRetry` replaced `time.Sleep(...)` with `sleepWithContext(req.Context(), ...)`, and skips the sleep on the final attempt.

### 3. `internal/downloader/retry_test.go` — 3 new tests + `trackingBody` helper

- `TestRetryTransport_ClosesRetryableBodyOnSuccess`
- `TestRetryTransport_ClosesRetryableBodyOnNonRetryableError`
- `TestDoWithRetry_RespectsContextCancellation`

## Review Findings

### Correctness: PASS

All exit paths in `RoundTrip` properly handle `lastResp`:

| Exit path | `lastResp` handling | Status |
|---|---|---|
| Context cancelled during sleep | `.Body.Close()` | OK |
| Clone fails, lastResp exists | Returned to caller | OK |
| Clone fails, no lastResp | Returns `lastErr` | OK |
| Non-retryable transport error | `.Body.Close()` | **Fixed** |
| Retryable transport error | Kept; closed on next iteration or exit | OK |
| Non-retryable status (success) | `.Body.Close()` | **Fixed** |
| Retryable status | Closes previous, stores new | OK |
| Exhausted retries, lastResp exists | Returned to caller | OK |
| Exhausted retries, no lastResp | Returns `lastErr` | OK |

The `doWithRetry` fix correctly:
- Skips sleep on the final attempt (avoids unnecessary delay before returning)
- Returns context error immediately on cancellation

### Tests: PASS

- New tests directly verify the two fixed bugs
- `trackingBody` helper is minimal and correct
- No race conditions (retry loop is sequential; `bodies` slice access is single-threaded)
- `TestDoWithRetry_RespectsContextCancellation` uses a real HTTP server — slightly heavier than needed but functional

### Issues Found

#### Issue 1 (Medium): PR targets wrong base branch

The branch was created from `web-ui` but the PR targets `main`. This causes the PR diff to include all `web-ui` changes (20k+ lines), making it unreviewable via GitHub's diff view. The PR base should be changed to `web-ui`, or the branch should be rebased onto `main`.

#### Issue 2 (Low): `TestDoWithRetry_RespectsContextCancellation` flakiness risk

The test cancels context after 50ms and asserts `calls < 10`. With 300ms backoff delays this is safe in practice, but on extremely resource-starved CI runners, the first HTTP round-trip itself could take >50ms, making the cancellation timing unpredictable. Consider using a mock transport instead of a real HTTP server for deterministic behavior.

#### Issue 3 (Informational): `doWithRetry` doesn't drain response body on non-2xx

`doWithRetry` returns `resp` on any `err == nil`, including non-2xx status codes. The caller (`downloadDirectFile`) does `defer resp.Body.Close()`, so this is fine. No action needed.

## Acceptance Criteria

The review is complete when:

1. All inline review comments are posted on the PR
2. The wrong-base-branch issue is flagged to the author
3. The spec accurately reflects the review findings

## Recommended Actions

1. **Change PR base branch** from `main` to `web-ui` (or rebase onto `main` if that's the intended merge target)
2. **Optional**: Replace the real HTTP server in `TestDoWithRetry_RespectsContextCancellation` with a mock transport for determinism
3. **Approve** the code changes themselves — the bug fixes are correct and well-tested
