---
phase: 01-foundation
plan: 02
type: tdd
wave: 2
depends_on:
  - 01-foundation-01
files_modified:
  - pkg/managers/request.go
  - pkg/managers/request_test.go
  - pkg/client.go
  - pkg/client_test.go
autonomous: true
requirements:
  - FOUND-01
  - FOUND-04

must_haves:
  truths:
    - "When rate limiter denies, SendRequest returns error immediately without sending to transport"
    - "When pending map reaches limit, SendRequest returns ErrTooManyPendingRequests immediately"
    - "When no rate limiter configured, SendRequest works as before (backward compatible)"
    - "When pending limit is 0 (default), no limit is enforced (backward compatible)"
    - "WithRateLimit option sets the limiter on ClientConfig"
  artifacts:
    - path: "pkg/client.go"
      provides: "RateLimiter field on ClientConfig, WithRateLimit option, rate check in SendRequest"
      contains: "RateLimiter"
    - path: "pkg/managers/request.go"
      provides: "maxPending field, pending limit check in SendRequest"
      contains: "maxPending"
    - path: "pkg/managers/request_test.go"
      provides: "Tests for pending request limit enforcement"
      min_lines: 40
    - path: "pkg/client_test.go"
      provides: "Tests for WithRateLimit option and rate-limited SendRequest"
  key_links:
    - from: "pkg/client.go"
      to: "pkg/types/types.go"
      via: "ClientConfig.RateLimiter field referencing RequestRateLimiter interface"
      pattern: "RateLimiter.*RequestRateLimiter"
    - from: "pkg/client.go"
      to: "pkg/types/errors.go"
      via: "import and return of rate limit error"
      pattern: "NewRequestError.*RATE_LIMITED"
    - from: "pkg/managers/request.go"
      to: "pkg/types/errors.go"
      via: "return ErrTooManyPendingRequests"
      pattern: "ErrTooManyPendingRequests"
---

<objective>
Implement client-side rate limiting and pending request limits in the request path.

Purpose: Prevent server rejection under burst load (FOUND-01) and prevent unbounded memory growth from pending requests (FOUND-04). Rate limiting is checked at client.SendRequest level (per GA-2 decision), pending limit is checked at RequestManager level.

Output: Working rate limiting and pending request limiting with full test coverage.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/01-foundation/01-PLAN.md

<interfaces>
<!-- From Plan 01 outputs -->

From pkg/types/types.go (new additions):
```go
type RequestRateLimiter interface {
    Allow() bool
}

type TokenBucketLimiter struct { ... }
func NewTokenBucketLimiter(rate float64, burst int) *TokenBucketLimiter
```

From pkg/types/errors.go (new additions):
```go
var ErrTooManyPendingRequests = errors.New("too many pending requests")
var ErrMaxRetriesExceeded = errors.New("max retries exceeded")
```

From pkg/types/types.go (existing):
```go
type ReconnectConfig struct {
    MaxAttempts       int
    MaxRetries        int    // NEW from Plan 01
    InitialDelay      time.Duration
    MaxDelay          time.Duration
    BackoffMultiplier float64
}
```
</interfaces>

From pkg/client.go (existing, key structures):
```go
type ClientConfig struct {
    URL                 string
    ClientID            string
    // ... other fields ...
    Logger              Logger
    ReconnectEnabled    bool
    ReconnectConfig     *ReconnectConfig
    // NO RateLimiter field yet
}

type ClientOption func(*ClientConfig) error

func (c *client) SendRequest(ctx context.Context, req *protocol.RequestFrame) (*protocol.ResponseFrame, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    // ... connection check, payload validation, then:
    return c.managers.request.SendRequest(ctx, req, sendFunc)
}
```

From pkg/managers/request.go (existing):
```go
type RequestManager struct {
    pending  map[string]*pendingRequest
    timeouts map[string]context.CancelFunc
    mu       sync.Mutex
    ctx      context.Context
    cancel   context.CancelFunc
    closed   bool
    // NO maxPending field yet
}

func NewRequestManager(ctx context.Context) *RequestManager { ... }

func (rm *RequestManager) SendRequest(ctx context.Context, req *protocol.RequestFrame, sendFunc func(*protocol.RequestFrame) error) (*protocol.ResponseFrame, error) {
    // Currently: creates respCh, adds to pending, sends, waits
    // NO pending limit check
}
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Add pending request limit to RequestManager</name>
  <files>pkg/managers/request.go, pkg/managers/request_test.go</files>
  <read_first>
    - pkg/managers/request.go (full file, to see current SendRequest implementation and lock usage)
    - pkg/managers/request_test.go (existing tests to follow pattern)
    - pkg/types/errors.go (to see ErrTooManyPendingRequests from Plan 01)
  </read_first>
  <acceptance_criteria>
    - pkg/managers/request.go contains "maxPending int" field on RequestManager struct
    - pkg/managers/request.go contains "ErrTooManyPendingRequests" in a return statement
    - NewRequestManager constructor unchanged (backward compatible)
    - SetMaxPending method exists on RequestManager: `func (rm *RequestManager) SetMaxPending(max int)`
    - In SendRequest, after rm.mu.Lock() and before rm.pending[req.ID] = ..., there is a check: `if rm.maxPending > 0 && len(rm.pending) >= rm.maxPending`
    - go test ./pkg/managers/ -run "TestPendingLimit|TestPendingLimitZero" -v exits 0
    - go test -race ./pkg/managers/ exits 0
  </acceptance_criteria>
  <behavior>
    - Test 1: SendRequest returns ErrTooManyPendingRequests when pending map is full (maxPending=2, fill 2 requests, 3rd fails immediately)
    - Test 2: SendRequest succeeds when maxPending=0 (unlimited, backward compatible)
    - Test 3: SendRequest succeeds when pending count < maxPending (under limit works normally)
    - Test 4: Concurrent test: multiple goroutines cannot exceed maxPending (fill exactly maxPending+1, verify at most maxPending succeed)
  </behavior>
  <action>
    Per FOUND-04 and CONTEXT.md locked decisions (pending limit in RequestManager.SendRequest, ErrTooManyPendingRequests sentinel, default 256):

    1. Add `maxPending int` field to RequestManager struct (after `closed bool`). Zero value means unlimited (backward compatible).

    2. Add setter method:
    ```go
    // SetMaxPending sets the maximum number of concurrent pending requests.
    // A value of 0 means unlimited (default). Per FOUND-04, recommended default is 256.
    func (rm *RequestManager) SetMaxPending(max int) {
        rm.maxPending = max
    }
    ```

    3. In SendRequest method, AFTER `rm.mu.Lock()` (line 57) and BEFORE `rm.pending[req.ID] = ...` (line 58), add the pending limit check:
    ```go
    // Check pending request limit (FOUND-04)
    if rm.maxPending > 0 && len(rm.pending) >= rm.maxPending {
        rm.mu.Unlock()
        return nil, types.ErrTooManyPendingRequests
    }
    ```

    IMPORTANT: This check must happen AFTER acquiring the lock and BEFORE adding to the pending map. The lock is already held at this point. The mu.Unlock() before return is needed because the rest of SendRequest expects to hold the lock through the function body -- but since we return early, we must release it.

    Actually, re-reading the current code: rm.mu.Lock() is at line 57, and rm.mu.Unlock() is implicit through the rest of the function flow. The current code does NOT use defer for the initial lock -- it manually unlocks at line 68. So the pending check goes between lock (line 57) and the pending map write (line 58):
    ```go
    rm.mu.Lock()
    // ADD HERE: pending limit check
    if rm.maxPending > 0 && len(rm.pending) >= rm.maxPending {
        rm.mu.Unlock()
        return nil, types.ErrTooManyPendingRequests
    }
    rm.pending[req.ID] = &pendingRequest{...}
    ```

    4. Add import for types package if not already present (it IS already imported at line 15).

    5. Add tests to pkg/managers/request_test.go:
    - TestRequestManager_PendingLimit: Set maxPending=2, send 2 requests that don't get responses (they block), verify 3rd returns ErrTooManyPendingRequests immediately. Use `errors.Is` to check.
    - TestRequestManager_PendingLimitZero: maxPending=0 (default), send many requests, verify no limit enforced.
    - TestRequestManager_PendingLimitUnderLimit: maxPending=5, send 3 requests, verify all succeed normally (with mock responses).

    For the blocking request pattern, use goroutines: send first N requests in goroutines (they block waiting for responses), then send N+1 from the test goroutine and verify ErrTooManyPendingRequests.

    Test pattern from existing code:
    ```go
    import "errors"
    // ... in test:
    if !errors.Is(err, types.ErrTooManyPendingRequests) {
        t.Errorf("expected ErrTooManyPendingRequests, got: %v", err)
    }
    ```
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/managers/ -run "TestRequestManager_PendingLimit" -v -race -count=1</automated>
  </verify>
  <done>
    - RequestManager has maxPending int field
    - SetMaxPending method exists and works
    - SendRequest returns ErrTooManyPendingRequests when limit reached
    - maxPending=0 means unlimited (backward compatible)
    - All existing request_test.go tests still pass
    - New tests pass with -race flag
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Add rate limiter to ClientConfig and client.SendRequest</name>
  <files>pkg/client.go, pkg/client_test.go</files>
  <read_first>
    - pkg/client.go (full file -- ClientConfig struct at line 127, SendRequest at line 563, option pattern examples)
    - pkg/client_test.go (if exists, for existing test patterns)
    - pkg/types/types.go (for RequestRateLimiter interface from Plan 01)
    - pkg/types/errors.go (for error types used in rate limit error)
  </read_first>
  <acceptance_criteria>
    - pkg/client.go ClientConfig struct contains "RateLimiter RequestRateLimiter" field
    - pkg/client.go contains `func WithRateLimit(limiter RequestRateLimiter) ClientOption` function
    - pkg/client.go SendRequest method checks RateLimiter BEFORE delegating to RequestManager
    - Rate check code: `if c.config.RateLimiter != nil && !c.config.RateLimiter.Allow()` returns NewRequestError("RATE_LIMITED", ...)
    - pkg/client.go re-exports RequestRateLimiter and TokenBucketLimiter types
    - go test ./pkg/ -run "TestWithRateLimit|TestSendRequest_RateLimited" -v exits 0
    - go test -race ./pkg/ exits 0 (no regressions)
  </acceptance_criteria>
  <behavior>
    - Test 1: WithRateLimit option sets RateLimiter on config
    - Test 2: SendRequest returns RequestError when limiter.Allow() returns false
    - Test 3: SendRequest works normally when no limiter configured (backward compatible)
    - Test 4: SendRequest works normally when limiter.Allow() returns true
  </behavior>
  <action>
    Per FOUND-01 and CONTEXT.md locked decisions (GA-1: Allow() bool, GA-2: client-level check):

    1. Add RateLimiter field to ClientConfig struct (after ReconnectConfig field, around line 149):
    ```go
    RateLimiter        types.RequestRateLimiter  // Optional rate limiter for request throughput (FOUND-01)
    ```

    2. Add WithRateLimit option function (after the other With* functions, around line 290):
    ```go
    // WithRateLimit sets a rate limiter for outgoing requests (FOUND-01).
    // When configured, SendRequest checks the limiter before sending.
    // If Allow() returns false, SendRequest returns a rate limit error immediately.
    func WithRateLimit(limiter types.RequestRateLimiter) ClientOption {
        return func(c *ClientConfig) error {
            c.RateLimiter = limiter
            return nil
        }
    }
    ```

    3. Add rate limiter check in client.SendRequest method. Insert AFTER the connection nil check (line 567-569) and BEFORE the payload validation (line 572). This is the GA-2 decision: check at client level, before delegating to RequestManager:

    ```go
    // Rate limit check (FOUND-01): check before any transport work
    if c.config.RateLimiter != nil && !c.config.RateLimiter.Allow() {
        return nil, NewRequestError("RATE_LIMITED", "rate limit exceeded", true, nil)
    }
    ```

    Note: The error is retryable=true because rate limiting is a transient condition.
    Using NewRequestError (existing constructor) consistent with wire-protocol error style.

    4. Add re-exports in client.go for the new types from Plan 01 (in the type alias section around line 83 and var section around line 99):
    ```go
    // In the type alias block:
    type RequestRateLimiter = types.RequestRateLimiter
    type TokenBucketLimiter = types.TokenBucketLimiter

    // In the var block:
    var NewTokenBucketLimiter = types.NewTokenBucketLimiter
    ```

    5. Wire the pending request limit. In NewClient (around line 395 after creating RequestManager), set the default maxPending:
    ```go
    c.managers.request = managers.NewRequestManager(ctx)
    c.managers.request.SetMaxPending(256) // FOUND-04: default pending limit
    ```

    6. Add tests to pkg/client_test.go:

    - TestWithRateLimit_Option: Create client with WithRateLimit(NewTokenBucketLimiter(100, 10)), verify config.RateLimiter is not nil.
    - TestSendRequest_RateLimited: Create a mock limiter that always returns false from Allow(). Create a client, verify SendRequest returns error with code "RATE_LIMITED" and the error is a RequestError.
    - TestSendRequest_NoRateLimiter: Verify client without rate limiter works normally (backward compat).

    Mock limiter for tests:
    ```go
    type mockLimiter struct {
        allow bool
    }
    func (m *mockLimiter) Allow() bool { return m.allow }
    ```
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/ -run "TestWithRateLimit|TestSendRequest_RateLimited|TestSendRequest_NoRateLimiter" -v -race -count=1</automated>
  </verify>
  <done>
    - ClientConfig has RateLimiter field
    - WithRateLimit option function exists and works
    - client.SendRequest checks rate limiter before delegating to RequestManager
    - When limiter denies, returns RequestError with code RATE_LIMITED
    - When no limiter, works as before (backward compatible)
    - RequestManager created with default maxPending=256
    - Re-exports for RequestRateLimiter and TokenBucketLimiter exist
    - All existing tests still pass
  </done>
</task>

</tasks>

<verification>
go test ./pkg/... -race -count=1
go vet ./pkg/
grep -c "RateLimiter" pkg/client.go  # should be >= 3
grep -c "maxPending" pkg/managers/request.go  # should be >= 3
grep -c "SetMaxPending" pkg/managers/request.go  # should be >= 2
grep -c "RATE_LIMITED" pkg/client.go  # should be 1
</verification>

<success_criteria>
1. SendRequest returns error immediately when rate limiter denies (no transport call made)
2. SendRequest returns ErrTooManyPendingRequests when pending map full
3. Without rate limiter, SendRequest works as before
4. Without SetMaxPending (zero value), no pending limit enforced
5. Default pending limit is 256 (set in NewClient)
6. All tests pass with -race flag
7. No existing tests broken
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-02-SUMMARY.md`
</output>
