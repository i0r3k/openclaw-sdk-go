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
    - "When pending map reaches limit, SendRequest returns typed TooManyPendingRequestsError wrapping ErrTooManyPendingRequests"
    - "When no rate limiter configured, SendRequest works as before (backward compatible)"
    - "When pending limit is 0 (default), no limit is enforced (backward compatible)"
    - "WithRateLimit option sets the limiter on ClientConfig"
    - "WithMaxPending option configures pending request limit via ClientConfig"
    - "client.SendRequest does NOT hold c.mu while waiting for response -- state snapshotted under lock, wait released"
    - "respCh is closed only by the cleanup function -- no double-close from Clear/Close"
  artifacts:
    - path: "pkg/client.go"
      provides: "RateLimiter field, MaxPending field, WithRateLimit/WithMaxPending options, reduced-scope mutex in SendRequest"
      contains: "RateLimiter"
    - path: "pkg/managers/request.go"
      provides: "maxPending field, pending limit check, fixed channel ownership"
      contains: "maxPending"
    - path: "pkg/managers/request_test.go"
      provides: "Tests for pending request limit and channel ownership"
      min_lines: 50
    - path: "pkg/client_test.go"
      provides: "Tests for WithRateLimit option, rate-limited SendRequest, reduced mutex scope"
  key_links:
    - from: "pkg/client.go"
      to: "pkg/types/types.go"
      via: "ClientConfig.RateLimiter field referencing RequestRateLimiter interface"
      pattern: "RateLimiter.*RequestRateLimiter"
    - from: "pkg/client.go"
      to: "pkg/types/errors.go"
      via: "rate limit error and TooManyPendingRequestsError"
      pattern: "NewRequestError.*RATE_LIMITED|TooManyPendingRequests"
    - from: "pkg/managers/request.go"
      to: "pkg/types/errors.go"
      via: "NewTooManyPendingRequestsError with configurable limit"
      pattern: "TooManyPendingRequests|ErrTooManyPendingRequests"
---

<objective>
Implement client-side rate limiting and pending request limits in the request path. Fix existing concurrency hazards (client mutex scope, channel double-close) that would undermine the new features.

Purpose: Prevent server rejection under burst load (FOUND-01) and prevent unbounded memory growth from pending requests (FOUND-04). Also fixes review-identified [HIGH] issues: client mutex held during response wait, and RequestManager double-close on respCh.

Output: Working rate limiting, pending request limiting, fixed mutex scope, fixed channel ownership, with full test coverage including race detector.
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

type TooManyPendingRequestsError struct { *BaseError }
func NewTooManyPendingRequestsError(limit int) *TooManyPendingRequestsError

type MaxRetriesExceededError struct { *BaseError }
func NewMaxRetriesExceededError(maxRetries int) *MaxRetriesExceededError
```

From pkg/client.go (existing, CRITICAL -- current SendRequest at line 563):
```go
func (c *client) SendRequest(ctx context.Context, req *protocol.RequestFrame) (*protocol.ResponseFrame, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    // ... holds c.mu the ENTIRE time, including while waiting for response from RequestManager
    return c.managers.request.SendRequest(ctx, req, sendFunc)
}
```

From pkg/managers/request.go (existing, CRITICAL -- current SendRequest at line 54):
```go
func (rm *RequestManager) SendRequest(ctx context.Context, req *protocol.RequestFrame, sendFunc func(*protocol.RequestFrame) error) (*protocol.ResponseFrame, error) {
    respCh := make(chan *protocol.ResponseFrame, 1)

    rm.mu.Lock()
    rm.pending[req.ID] = &pendingRequest{responseCh: respCh}
    // ... timeout setup ...
    rm.mu.Unlock()

    cleanup := func() {
        rm.mu.Lock()
        delete(rm.pending, req.ID)
        // ...
        close(respCh)  // <-- PROBLEM: also closed by Clear() and Close()
        rm.mu.Unlock()
    }
    defer cleanup()
    // ... send, wait on respCh ...
}
```

From pkg/managers/request.go (existing -- double-close sites):
```go
// Clear() line 162-164: closes respCh for each pending request
for id, req := range rm.pending {
    close(req.responseCh)  // <-- CLOSE #1
}

// Close() line 185-188: also closes respCh for each pending request
for id, req := range rm.pending {
    close(req.responseCh)  // <-- CLOSE #2 (double-close if cleanup runs first)
}
```

From pkg/client.go ClientConfig (line 127-150):
```go
type ClientConfig struct {
    URL                 string
    ClientID            string
    // ... other fields ...
    Logger              Logger
    ReconnectEnabled    bool
    ReconnectConfig     *ReconnectConfig
    AuthHandler         auth.AuthHandler
    // NO RateLimiter field yet
    // NO MaxPending field yet
}
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Fix RequestManager channel ownership and add pending request limit</name>
  <files>pkg/managers/request.go, pkg/managers/request_test.go</files>
  <read_first>
    - pkg/managers/request.go (full file -- SendRequest lines 54-100, Clear lines 154-169, Close lines 173-193)
    - pkg/managers/request_test.go (existing tests to follow pattern)
    - pkg/types/errors.go (for ErrTooManyPendingRequests and NewTooManyPendingRequestsError from Plan 01)
  </read_first>
  <acceptance_criteria>
    - pkg/managers/request.go contains "maxPending int" field on RequestManager struct
    - respCh closed ONLY by the SendRequest cleanup function -- Clear() and Close() do NOT close respCh
    - Clear() removes entries from pending map and cancels timeouts but does NOT close channels
    - Close() removes entries from pending map and cancels timeouts but does NOT close channels
    - NewRequestManager constructor unchanged (backward compatible -- maxPending defaults to 0)
    - SetMaxPending method exists: `func (rm *RequestManager) SetMaxPending(max int)`
    - In SendRequest, after rm.mu.Lock() and before rm.pending[req.ID]=..., there is a check: `if rm.maxPending > 0 && len(rm.pending) >= rm.maxPending`
    - SendRequest returns NewTooManyPendingRequestsError(rm.maxPending) when limit exceeded
    - go test ./pkg/managers/ -run "TestRequestManager_PendingLimit|TestRequestManager_ChannelOwnership" -v exits 0
    - go test -race ./pkg/managers/ exits 0
  </acceptance_criteria>
  <behavior>
    - Test 1: SendRequest returns error wrapping ErrTooManyPendingRequests when pending map is full (maxPending=2, fill 2, 3rd fails)
    - Test 2: SendRequest succeeds when maxPending=0 (unlimited, backward compatible)
    - Test 3: SendRequest succeeds when pending count < maxPending
    - Test 4: Clear() followed by cleanup does not double-close (no panic)
    - Test 5: Close() followed by cleanup does not double-close (no panic)
    - Test 6: Concurrent test: multiple goroutines cannot exceed maxPending (race detector)
  </behavior>
  <action>
    Per FOUND-04 and CONTEXT.md locked decisions. Addresses review concerns:
    - [HIGH] RequestManager has double-close risk on respCh: fix channel ownership so ONLY SendRequest cleanup closes respCh
    - [MEDIUM] Hard-coded pending limit: make configurable via SetMaxPending

    STEP 1: Fix channel ownership in RequestManager.

    The problem: `cleanup()` in SendRequest closes respCh, AND Clear()/Close() also iterate pending and close respCh. If cleanup runs first (deletes from map, closes channel), and then Clear/Close runs, it won't find the entry -- that's safe. But if Clear/Close runs first while SendRequest is still waiting on respCh, both will close the same channel -- panic.

    Fix: Clear() and Close() should NOT close respCh. Instead, they should:
    1. Delete the entry from the pending map
    2. Cancel the timeout context (which will cause SendRequest's `<-ctx.Done()` to fire)
    3. Send a nil response on respCh to unblock the waiting goroutine (non-blocking send to buffered channel)
    4. The cleanup function in SendRequest will then close respCh when it runs

    Modify Clear() (line 154-169):
    ```go
    func (rm *RequestManager) Clear() {
        rm.mu.Lock()
        defer rm.mu.Unlock()

        if rm.closed {
            return
        }

        for id, req := range rm.pending {
            // Signal the waiting goroutine with nil response (non-blocking)
            select {
            case req.responseCh <- nil:
            default:
            }
            delete(rm.pending, id)
        }
        for _, cancel := range rm.timeouts {
            cancel()
        }
    }
    ```

    Modify Close() (line 173-193):
    ```go
    func (rm *RequestManager) Close() error {
        rm.mu.Lock()
        defer rm.mu.Unlock()

        if rm.closed {
            return nil
        }
        rm.closed = true

        rm.cancel()
        for id, req := range rm.pending {
            // Signal the waiting goroutine with nil response (non-blocking)
            select {
            case req.responseCh <- nil:
            default:
            }
            delete(rm.pending, id)
        }
        for _, cancel := range rm.timeouts {
            cancel()
        }
        return nil
    }
    ```

    IMPORTANT: The cleanup function in SendRequest must handle receiving nil from respCh:
    ```go
    select {
    case resp := <-respCh:
        if resp == nil {
            // Channel was signaled by Clear/Close, not a real response
            return nil, ctx.Err()
        }
        return resp, nil
    case <-ctx.Done():
        // ...
    }
    ```

    STEP 2: Add maxPending field and pending limit check.

    Add `maxPending int` field to RequestManager struct (after `closed bool`). Zero value means unlimited.

    Add setter:
    ```go
    func (rm *RequestManager) SetMaxPending(max int) {
        rm.maxPending = max
    }
    ```

    In SendRequest, AFTER `rm.mu.Lock()` (line 57) and BEFORE `rm.pending[req.ID] = ...` (line 58):
    ```go
    if rm.maxPending > 0 && len(rm.pending) >= rm.maxPending {
        rm.mu.Unlock()
        return nil, types.NewTooManyPendingRequestsError(rm.maxPending)
    }
    ```

    STEP 3: Add tests.

    - TestRequestManager_PendingLimit_RejectsOverLimit: maxPending=2, send 2 blocking requests in goroutines, verify 3rd returns error wrapping ErrTooManyPendingRequests using errors.Is()
    - TestRequestManager_PendingLimit_ZeroUnlimited: maxPending=0, send many requests, verify no limit
    - TestRequestManager_PendingLimit_UnderLimit: maxPending=5, send 3, verify all succeed
    - TestRequestManager_ChannelOwnership_ClearNoDoubleClose: call Clear() while SendRequest is blocked, verify no panic and SendRequest returns error
    - TestRequestManager_ChannelOwnership_CloseNoDoubleClose: call Close() while SendRequest is blocked, verify no panic
    - TestRequestManager_PendingLimit_ConcurrentRace: multiple goroutines, verify at most maxPending succeed under -race

    For blocking requests: send requests in goroutines with a sendFunc that never returns (blocks), then test the N+1 call.
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/managers/ -run "TestRequestManager_PendingLimit|TestRequestManager_ChannelOwnership" -v -race -count=1</automated>
  </verify>
  <done>
    - RequestManager has maxPending int field
    - SetMaxPending method exists
    - SendRequest returns typed TooManyPendingRequestsError when limit reached
    - maxPending=0 means unlimited (backward compatible)
    - Channel ownership fixed: only SendRequest cleanup closes respCh
    - Clear() signals waiting goroutines instead of closing channels
    - Close() signals waiting goroutines instead of closing channels
    - No double-close possible
    - All existing tests still pass
    - New tests pass with -race flag
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Reduce client mutex scope and add rate limiter to SendRequest path</name>
  <files>pkg/client.go, pkg/client_test.go</files>
  <read_first>
    - pkg/client.go (full file -- ClientConfig struct at line 127, SendRequest at line 563, option pattern examples around line 290, NewClient around line 395)
    - pkg/client_test.go (if exists, for existing test patterns)
    - pkg/types/types.go (for RequestRateLimiter interface from Plan 01)
    - pkg/types/errors.go (for error types)
    - pkg/managers/connection.go (for ConnectionManager.Transport() method)
  </read_first>
  <acceptance_criteria>
    - pkg/client.go ClientConfig struct contains "RateLimiter RequestRateLimiter" field
    - pkg/client.go ClientConfig struct contains "MaxPending int" field
    - pkg/client.go contains `func WithRateLimit(limiter RequestRateLimiter) ClientOption`
    - pkg/client.go contains `func WithMaxPending(max int) ClientOption`
    - client.SendRequest does NOT hold c.mu while waiting for response -- snapshots state under lock, releases, then waits
    - Rate limiter check happens OUTSIDE c.mu (before lock acquisition)
    - Re-exports for RequestRateLimiter and TokenBucketLimiter exist in client.go
    - NewClient wires maxPending from config to RequestManager via SetMaxPending
    - go test ./pkg/ -run "TestWithRateLimit|TestSendRequest_RateLimited|TestSendRequest_MutexScope" -v exits 0
    - go test -race ./pkg/ exits 0 (no regressions)
  </acceptance_criteria>
  <behavior>
    - Test 1: WithRateLimit option sets RateLimiter on config
    - Test 2: WithMaxPending option sets MaxPending on config
    - Test 3: SendRequest returns error when limiter.Allow() returns false (checked outside mutex)
    - Test 4: SendRequest works normally when no limiter configured (backward compatible)
    - Test 5: SendRequest works normally when limiter.Allow() returns true
    - Test 6: Concurrent SendRequests can all be in-flight simultaneously (mutex not held during wait)
  </behavior>
  <action>
    Per FOUND-01, FOUND-04, and CONTEXT.md locked decisions (GA-1: Allow() bool, GA-2: client-level check).

    Addresses review concerns:
    - [HIGH] SendRequest holds client mutex while waiting -- reduce scope: snapshot state under lock, wait outside
    - [MEDIUM] Hard-coded pending limit 256 -- make configurable via ClientConfig option

    STEP 1: Add fields to ClientConfig:

    ```go
    // In ClientConfig struct, after ReconnectConfig:
    RateLimiter        RequestRateLimiter  // Optional rate limiter for request throughput (FOUND-01)
    MaxPending         int                 // Max concurrent pending requests; 0 = unlimited (FOUND-04, default 256)
    ```

    STEP 2: Add option functions:

    ```go
    func WithRateLimit(limiter RequestRateLimiter) ClientOption {
        return func(c *ClientConfig) error {
            c.RateLimiter = limiter
            return nil
        }
    }

    func WithMaxPending(max int) ClientOption {
        return func(c *ClientConfig) error {
            if max < 0 {
                return fmt.Errorf("max pending must be >= 0, got %d", max)
            }
            c.MaxPending = max
            return nil
        }
    }
    ```

    STEP 3: Reduce client mutex scope in SendRequest.

    Current code (line 563-600):
    ```go
    func (c *client) SendRequest(ctx context.Context, req *protocol.RequestFrame) (*protocol.ResponseFrame, error) {
        c.mu.Lock()
        defer c.mu.Unlock()
        // ... connection check, payload validation, marshal, send ...
        return c.managers.request.SendRequest(ctx, req, sendFunc)
    }
    ```

    Problem: c.mu is held the entire time including while waiting for the response from RequestManager. This serializes ALL client operations -- no concurrent SendRequests possible.

    New approach: snapshot needed state under the lock, release the lock, then do the work:
    ```go
    func (c *client) SendRequest(ctx context.Context, req *protocol.RequestFrame) (*protocol.ResponseFrame, error) {
        // Rate limit check OUTSIDE client mutex (FOUND-01)
        // Rate limiting is a pre-flight check independent of connection state
        if c.config.RateLimiter != nil && !c.config.RateLimiter.Allow() {
            return nil, NewRequestError("RATE_LIMITED", "rate limit exceeded", true, nil)
        }

        // Snapshot state under client mutex, then release
        c.mu.Lock()
        if c.managers.connection == nil || c.managers.connection.Transport() == nil {
            c.mu.Unlock()
            return nil, NewConnectionError("NOT_CONNECTED", "not connected", false, nil)
        }

        transport := c.managers.connection.Transport()
        policyMgr := c.policyManager
        c.mu.Unlock()
        // c.mu is now released -- concurrent SendRequests can proceed

        // Validate payload size against server policy
        var sendFunc func(*protocol.RequestFrame) error
        if policyMgr != nil && policyMgr.HasPolicy() {
            data, err := json.Marshal(req)
            if err != nil {
                return nil, NewProtocolError(string(ProtocolErrFrameTooLarge), "failed to marshal request", false, nil)
            }
            maxPayload := policyMgr.GetMaxPayload()
            if int64(len(data)) > maxPayload {
                return nil, NewProtocolError(
                    string(ProtocolErrFrameTooLarge),
                    fmt.Sprintf("request payload size %d exceeds server limit %d", len(data), maxPayload),
                    false, nil,
                )
            }
            sendFunc = func(r *protocol.RequestFrame) error {
                return transport.Send(data)
            }
        } else {
            data, err := json.Marshal(req)
            if err != nil {
                return nil, NewProtocolError(string(ProtocolErrFrameTooLarge), "failed to marshal request", false, nil)
            }
            sendFunc = func(r *protocol.RequestFrame) error {
                return transport.Send(data)
            }
        }

        return c.managers.request.SendRequest(ctx, req, sendFunc)
    }
    ```

    IMPORTANT: Read the current full SendRequest implementation carefully. The exact branching for policy vs non-policy paths must be preserved. The key change is: lock, snapshot `transport` and `policyMgr` into local vars, unlock, then proceed with existing logic using local vars instead of `c.managers.connection.Transport()` and `c.policyManager`.

    CAUTION: Verify that `c.config.RateLimiter` access is safe without the lock. Since ClientConfig is set at construction and never mutated after NewClient, reading c.config fields is safe without c.mu. If uncertain, read c.config.RateLimiter once under the lock during snapshot.

    STEP 4: Wire MaxPending in NewClient:

    After creating RequestManager in NewClient (around line 395):
    ```go
    c.managers.request = managers.NewRequestManager(ctx)
    // Wire pending request limit (FOUND-04)
    maxPending := c.config.MaxPending
    if maxPending == 0 {
        maxPending = 256 // sensible default
    }
    c.managers.request.SetMaxPending(maxPending)
    ```

    STEP 5: Add re-exports in client.go:
    ```go
    type RequestRateLimiter = types.RequestRateLimiter
    type TokenBucketLimiter = types.TokenBucketLimiter
    var NewTokenBucketLimiter = types.NewTokenBucketLimiter
    ```

    STEP 6: Add tests to pkg/client_test.go:

    - TestWithRateLimit_Option: Create config with WithRateLimit(NewTokenBucketLimiter(100, 10)), verify config.RateLimiter is not nil
    - TestWithMaxPending_Option: Create config with WithMaxPending(50), verify config.MaxPending == 50
    - TestWithMaxPending_NegativeReturnsError: WithMaxPending(-1) returns error
    - TestSendRequest_RateLimited: Mock limiter (allow=false), verify SendRequest returns RequestError with code RATE_LIMITED
    - TestSendRequest_NoRateLimiter: Client without limiter, verify SendRequest works (backward compat)
    - TestSendRequest_ConcurrentInFlight: Send multiple requests concurrently, verify they can all be in-flight (mutex not blocking), test with -race

    Mock limiter:
    ```go
    type mockLimiter struct {
        allow bool
    }
    func (m *mockLimiter) Allow() bool { return m.allow }
    ```
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/ -run "TestWithRateLimit|TestWithMaxPending|TestSendRequest_RateLimited|TestSendRequest_NoRateLimiter|TestSendRequest_Concurrent" -v -race -count=1</automated>
  </verify>
  <done>
    - ClientConfig has RateLimiter and MaxPending fields
    - WithRateLimit and WithMaxPending option functions exist and work
    - client.SendRequest checks rate limiter BEFORE acquiring client mutex
    - client.SendRequest releases client mutex before waiting for response
    - When limiter denies, returns RequestError with code RATE_LIMITED (retryable=true)
    - When no limiter, works as before (backward compatible)
    - RequestManager created with configured or default (256) maxPending
    - Re-exports for RequestRateLimiter and TokenBucketLimiter exist
    - All existing tests still pass
    - Concurrent SendRequest test passes with -race
  </done>
</task>

</tasks>

<verification>
go test ./pkg/... -race -count=1
go vet ./pkg/
grep -c "RateLimiter" pkg/client.go  # should be >= 3
grep -c "MaxPending" pkg/client.go  # should be >= 3
grep -c "maxPending" pkg/managers/request.go  # should be >= 3
grep -c "SetMaxPending" pkg/managers/request.go  # should be >= 2
grep -c "RATE_LIMITED" pkg/client.go  # should be 1
grep -c "select" pkg/managers/request.go  # Clear/Close should use select for non-blocking send
</verification>

<success_criteria>
1. SendRequest returns error immediately when rate limiter denies (no transport call made)
2. SendRequest returns typed TooManyPendingRequestsError when pending map full
3. Without rate limiter, SendRequest works as before
4. Without MaxPending configured, default 256 is used
5. client.SendRequest does NOT hold c.mu while waiting for response
6. Rate limit check happens outside client mutex
7. RequestManager respCh closed only by SendRequest cleanup (no double-close)
8. Clear() and Close() signal waiting goroutines without closing channels
9. All tests pass with -race flag
10. No existing tests broken
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-02-SUMMARY.md`
</output>
