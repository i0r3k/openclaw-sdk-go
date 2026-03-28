---
phase: 01-foundation
plan: 02
subsystem: core
tags: [concurrency, rate-limiting, go, websocket, channel-safety]

requires:
  - phase: 01-foundation-01
    provides: TooManyPendingRequestsError, ErrTooManyPendingRequests, TokenBucketLimiter, RequestRateLimiter

provides:
  - RequestManager maxPending field with SetMaxPending() method
  - TooManyPendingRequestsError returned when pending limit exceeded
  - Clear() and Close() signal via nil send (no double-close)
  - ClientConfig RateLimiter and MaxPending fields
  - WithRateLimit and WithMaxPending option functions
  - Reduced client mutex scope (rate limit outside lock, snapshot+release before wait)
  - TokenBucketLimiter re-exported from main package

affects: [01-foundation, 02-features, 03-api]

tech-stack:
  added: []
  patterns:
    - "Channel ownership pattern: only SendRequest cleanup closes channels, Clear/Close signal with nil"
    - "Snapshot-release pattern: mutex held briefly for state snapshot, released before long wait"
    - "Option pattern: functional options for rate limiter and pending limit configuration"
    - "Token bucket rate limiting: Allow() bool interface with burst capacity"

key-files:
  created:
    - pkg/managers/request_pending_limit_test.go
    - pkg/client_rate_limit_test.go
  modified:
    - pkg/managers/request.go
    - pkg/client.go
    - pkg/transport/websocket.go

key-decisions:
  - "maxPending=0 means unlimited (backward compatible default for RequestManager)"
  - "Rate limit check happens outside client mutex, before connection check"
  - "Clear()/Close() send nil on respCh instead of closing -- cleanup() in SendRequest closes"
  - "RequestManager default maxPending in NewClient is 256"
  - "RATE_LIMITED error is retryable (transient load condition)"

patterns-established:
  - "Rule: Never close a channel you don't own. Clear/Close signal with non-blocking send, SendRequest cleanup closes."
  - "Rule: Hold mutex for the minimum time needed. Snapshot state, release, then do the work."

requirements-completed: [FOUND-01, FOUND-04]

duration: ~11h 43m
completed: 2026-03-28T14:36:29Z
---

# Phase 1 Plan 2: Foundation Hardening Summary

**Client-side rate limiting with TokenBucketLimiter, pending request limits with TooManyPendingRequestsError, reduced mutex scope, and fixed channel ownership for concurrent safety**

## Performance

- **Duration:** ~11h 43m
- **Started:** 2026-03-28T02:52:58Z
- **Completed:** 2026-03-28T14:36:29Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- RequestManager now enforces configurable pending request limits (FOUND-04)
- Channel double-close eliminated: Clear()/Close() signal with nil send, cleanup() closes
- Client mutex reduced scope: rate limit checked outside lock, state snapshotted, lock released before waiting
- Client-level rate limiting via RequestRateLimiter interface with TokenBucketLimiter implementation (FOUND-01)
- All 14 managers tests and 34 pkg tests pass with `-race`

## Task Commits

1. **Task 1: Fix RequestManager channel ownership and add pending request limit** - `e8f5252` (feat)
2. **Task 2: Reduce client mutex scope and add rate limiter to SendRequest path** - `d0b2234` (feat)

**Plan metadata:** `c3a9f84` (docs: complete plan)

## Files Created/Modified

- `pkg/managers/request.go` - Added maxPending field, SetMaxPending(), pending limit check in SendRequest, nil respCh handling, Clear()/Close() signal fix
- `pkg/managers/request_pending_limit_test.go` - 7 new tests for pending limits and channel ownership
- `pkg/client.go` - Added RateLimiter/MaxPending fields, WithRateLimit/WithMaxPending options, reduced mutex scope in SendRequest, TokenBucketLimiter re-exports
- `pkg/client_rate_limit_test.go` - 8 new tests for rate limiting options and behavior
- `pkg/transport/websocket.go` - Fixed pre-existing build errors (unused types import, missing os import)

## Decisions Made

- maxPending=0 means unlimited (backward compatible default for RequestManager)
- Rate limit check outside client mutex, before connection check (returns RATE_LIMITED immediately)
- Clear()/Close() send nil on respCh instead of closing channels; cleanup() in SendRequest closes
- RequestManager default maxPending in NewClient is 256
- RATE_LIMITED error is retryable=true (transient load condition)
- ClientConfig.MaxPending=0 falls back to 256 in NewClient wiring

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Pre-existing build errors in pkg/transport/websocket.go**
- **Found during:** Task 1 verification
- **Issue:** websocket.go had unused `types` import and used-but-not-imported `os` package
- **Fix:** Removed unused import, added missing `os` import
- **Files modified:** pkg/transport/websocket.go
- **Verification:** go vet passed, tests compiled and passed
- **Committed in:** e8f5252 (part of Task 1)

**2. [Rule 1 - Bug] Go vet failed due to stale build cache**
- **Found during:** Task 1 pre-commit hook failure
- **Issue:** `go vet` failed with `validator.SetLogger undefined` on pre-existing code; cleared build cache resolved it
- **Fix:** `go clean -cache` to clear stale artifacts
- **Verification:** `go vet ./...` passed cleanly after cache clear
- **Committed in:** e8f5252 (cleared cache before commit)

**3. [Rule 1 - Bug] Go fmt required blank line in client.go**
- **Found during:** Task 2 pre-commit hook
- **Issue:** go fmt required blank line between type alias and var block
- **Fix:** Go fmt auto-fixed (stashed and re-applied)
- **Files modified:** pkg/client.go
- **Committed in:** d0b2234 (part of Task 2)

---

**Total deviations:** 3 auto-fixed (3 blocking/1 bug)
**Impact on plan:** All auto-fixes were pre-existing code quality issues unrelated to plan tasks. No scope creep.

## Issues Encountered

- **Variable shadowing in Go tests**: Local `client` variable shadows package name, causing "client is not a type" compile errors. Fixed by renaming all local variables to `cli`.
- **Blocking sendFunc design flaw**: Tests using blocking sendFunc with `<-make(chan struct{})` failed because Go requires `return nil` in named-return functions. Fixed using `go func() { <-ch }(); return nil` pattern.
- **Double-close race in tests**: Original tests didn't account for channel closing order. Fixed with proper nil-signaling pattern and context-based unblocking using `context.WithTimeout`.

## Next Phase Readiness

- Rate limiting infrastructure in place for FOUND-01
- Pending request limits in place for FOUND-04
- Channel ownership fixed -- concurrent request handling is safe
- Client mutex reduced scope enables true concurrent SendRequests
- Ready for Phase 2: API stability with fuzz testing and edge cases
- No blockers or concerns remaining from this plan

---
*Phase: 01-foundation*
*Completed: 2026-03-28*
