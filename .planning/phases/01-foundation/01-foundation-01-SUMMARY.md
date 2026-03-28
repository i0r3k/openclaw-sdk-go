---
phase: 01-foundation
plan: "01"
subsystem: foundation
tags: [go, websocket, rate-limiting, error-types, token-bucket]

requires: []
provides:
  - RequestRateLimiter interface with Allow() bool method
  - TokenBucketLimiter implementing token bucket algorithm (thread-safe)
  - MaxRetries int field on ReconnectConfig with explicit precedence documentation
  - DefaultReconnectConfig() sets MaxRetries=10 (production default)
  - ErrTooManyPendingRequests sentinel + TooManyPendingRequestsError typed error
  - ErrMaxRetriesExceeded sentinel + MaxRetriesExceededError typed error
affects: [01-foundation-02, 01-foundation-03, 02-observability]

tech-stack:
  added: []
  patterns:
    - Token bucket rate limiting algorithm
    - Sentinel error + typed error dual pattern with errors.Is() support
    - Explicit precedence documentation for configuration field overlap

key-files:
  created:
    - pkg/types/rate_limiter_test.go - TokenBucketLimiter tests (burst, refill, concurrent safety)
    - pkg/types/reconnect_config_test.go - MaxRetries default and precedence tests
    - pkg/types/foundation_errors_test.go - Foundation typed error tests
  modified:
    - pkg/types/types.go - Added RequestRateLimiter, TokenBucketLimiter, MaxRetries field
    - pkg/types/errors.go - Added ErrTooManyPendingRequests, ErrMaxRetriesExceeded, typed errors
    - pkg/types/types_test.go - Extended TestDefaultReconnectConfig to check MaxRetries=10

key-decisions:
  - "Option A (simple Allow() bool) chosen for RequestRateLimiter over Option B (retry-after feedback) per CONTEXT.md GA-1 recommendation"
  - "Backwards-compatible MaxRetries/MaxAttempts overlap: MaxRetries > 0 wins, zero falls back to MaxAttempts, both zero = infinite"
  - "Token bucket starts full (burst tokens available immediately) for immediate-first-request readiness"
  - "Token bucket uses sync.Mutex (not atomic) for simplicity since Allow() is not performance-critical"

patterns-established:
  - "Sentinel error + typed error dual pattern: sentinel enables errors.Is(), typed enables interface inspection"
  - "MaxRetries precedence rules documented with exact zero/negative semantics"
  - "TDD workflow: RED (failing tests) -> GREEN (pass implementation) -> go-fmt auto-fix -> commit"

requirements-completed: [FOUND-01, FOUND-02, FOUND-04]

# Phase 1 Plan 1: Foundation Types Summary

**RequestRateLimiter interface, TokenBucketLimiter, MaxRetries field, and typed error sentinels with OpenClawError implementation**

## Performance

- **Duration:** 6 min 25 sec
- **Started:** 2026-03-28T02:42:23Z
- **Completed:** 2026-03-28T02:48:48Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Defined `RequestRateLimiter` interface with `Allow() bool` method for rate limiting
- Implemented `TokenBucketLimiter` with thread-safe token bucket algorithm (rate/sec + burst)
- Added `MaxRetries` int field to `ReconnectConfig` with explicit precedence documentation over legacy `MaxAttempts`
- Set `DefaultReconnectConfig()` to `MaxRetries=10` (production default, backward compatible)
- Added `ErrTooManyPendingRequests` sentinel + `TooManyPendingRequestsError` typed error (FOUND-04)
- Added `ErrMaxRetriesExceeded` sentinel + `MaxRetriesExceededError` typed error (FOUND-02)
- Both typed errors implement `OpenClawError` interface and wrap sentinels for `errors.Is()` support

## Task Commits

Each task was committed atomically:

1. **Task 1: Add RequestRateLimiter, TokenBucketLimiter, MaxRetries** - `f36a470` (feat)
2. **Task 2: Add typed errors ErrTooManyPendingRequests and ErrMaxRetriesExceeded** - `fe2d433` (feat)
3. **Extend TestDefaultReconnectConfig** - `5773424` (test)

**Plan metadata:** `5773424` (docs: complete plan)

_Note: TDD tasks may have multiple commits (test -> feat -> refactor)_

## Files Created/Modified

- `pkg/types/types.go` - RequestRateLimiter interface, TokenBucketLimiter struct, MaxRetries field, DefaultReconnectConfig update
- `pkg/types/errors.go` - ErrTooManyPendingRequests sentinel, ErrMaxRetriesExceeded sentinel, TooManyPendingRequestsError, MaxRetriesExceededError, fmt import added
- `pkg/types/rate_limiter_test.go` - TokenBucketLimiter tests: burst cap, token refill, concurrent safety, interface compliance
- `pkg/types/reconnect_config_test.go` - MaxRetries default, precedence rule tests
- `pkg/types/foundation_errors_test.go` - Foundation typed error tests: sentinel Is, typed Is, OpenClawError, Retryable, Code, Details
- `pkg/types/types_test.go` - Extended TestDefaultReconnectConfig to check MaxRetries=10

## Decisions Made

- Chose Option A (simple `Allow() bool`) for RequestRateLimiter over Option B (retry-after feedback) per CONTEXT.md GA-1 recommendation
- Backwards-compatible MaxRetries/MaxAttempts overlap: MaxRetries > 0 wins, zero falls back to MaxAttempts, both zero = infinite
- Token bucket starts full (burst tokens available immediately) for immediate-first-request readiness
- Token bucket uses sync.Mutex (not atomic) for simplicity since Allow() is not performance-critical

## Deviations from Plan

None - plan executed exactly as written.

## Auto-Fixed Issues

**1. [Rule 1 - Bug] TestTokenBucketLimiter_BurstCap - rate too high causing false failures**
- **Found during:** Task 1 (RED phase)
- **Issue:** Test used rate=1M/sec which refilled tokens faster than test assumed burst cap enforcement
- **Fix:** Reduced rate to 100/sec with comment explaining why 100/sec is appropriate for burst cap testing
- **Files modified:** pkg/types/rate_limiter_test.go
- **Verification:** Test passes with rate=100/sec, burst=3, 100 rapid calls
- **Committed in:** f36a470 (Task 1 commit)

**2. [Rule 3 - Blocking] Missing fmt import in errors.go**
- **Found during:** Task 2 (GREEN phase)
- **Issue:** NewTooManyPendingRequestsError and NewMaxRetriesExceededError use fmt.Sprintf but fmt was not imported
- **Fix:** Added "fmt" to imports in errors.go
- **Files modified:** pkg/types/errors.go
- **Verification:** Build succeeds after import addition
- **Committed in:** fe2d433 (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking)
**Impact on plan:** Both auto-fixes necessary for test correctness and build. No scope creep.

## Issues Encountered

None beyond the two auto-fixed issues above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plans 02 and 03 can now implement against the known interfaces created in this plan:
  - `RequestRateLimiter` interface is available for Plan 02 to wire into RequestManager
  - `TokenBucketLimiter` ready for use via `NewTokenBucketLimiter(rate, burst)`
  - `MaxRetries` field on `ReconnectConfig` ready for Plan 03 to enforce in ReconnectManager
  - `ErrTooManyPendingRequests` and `ErrMaxRetriesExceeded` typed errors ready for Plans 02 and 03
- All 14 foundation error tests pass with race detection enabled

## Self-Check

- [x] RequestRateLimiter interface with Allow() bool method exists in pkg/types/types.go
- [x] TokenBucketLimiter correctly implements the interface with token bucket algorithm
- [x] MaxRetries int field exists on ReconnectConfig with default value 10
- [x] MaxRetries precedence rules documented with exact zero/negative semantics
- [x] ErrTooManyPendingRequests and ErrMaxRetriesExceeded sentinel errors exist
- [x] Typed TooManyPendingRequestsError and MaxRetriesExceededError exist implementing OpenClawError
- [x] Typed errors wrap sentinels so errors.Is() chain works
- [x] All tests pass with go test -race ./pkg/types/
- [x] No existing tests broken

---
_Phase: 01-foundation_
_Plan: 01_
_Completed: 2026-03-28_
