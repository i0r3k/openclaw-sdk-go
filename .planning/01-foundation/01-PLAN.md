---
phase: 01-foundation
plan: 01
type: tdd
wave: 1
depends_on: []
files_modified:
  - pkg/types/types.go
  - pkg/types/errors.go
  - pkg/types/rate_limiter_test.go
  - pkg/types/types_test.go
autonomous: true
requirements:
  - FOUND-01
  - FOUND-02
  - FOUND-04

must_haves:
  truths:
    - "RequestRateLimiter interface exists with Allow() bool method"
    - "TokenBucketLimiter implements RequestRateLimiter with configurable rate and burst"
    - "ErrTooManyPendingRequests sentinel error exists and supports errors.Is()"
    - "ErrMaxRetriesExceeded sentinel error exists and supports errors.Is()"
    - "ReconnectConfig has MaxRetries field; DefaultReconnectConfig() sets MaxRetries=10"
  artifacts:
    - path: "pkg/types/types.go"
      provides: "RequestRateLimiter interface, TokenBucketLimiter struct, MaxRetries field on ReconnectConfig"
      contains: "RequestRateLimiter"
    - path: "pkg/types/errors.go"
      provides: "ErrTooManyPendingRequests, ErrMaxRetriesExceeded sentinel errors"
      contains: "ErrTooManyPendingRequests"
    - path: "pkg/types/rate_limiter_test.go"
      provides: "Tests for TokenBucketLimiter behavior"
      exports: []
    - path: "pkg/types/types_test.go"
      provides: "Tests for MaxRetries default and ReconnectConfig"
  key_links:
    - from: "pkg/types/types.go"
      to: "pkg/types/errors.go"
      via: "shared package"
      pattern: "RequestRateLimiter"
---

<objective>
Define shared type contracts, interfaces, and sentinel errors that Plans 02 and 03 consume.

Purpose: Plans 02 and 03 both depend on new types in pkg/types. This plan creates those contracts first so downstream plans implement against known interfaces rather than exploring the codebase.

Output: New RequestRateLimiter interface + TokenBucketLimiter implementation, MaxRetries field on ReconnectConfig, and sentinel error variables ErrTooManyPendingRequests + ErrMaxRetriesExceeded.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/01-foundation/CONTEXT.md
@.planning/01-foundation/01-RESEARCH.md
@pkg/types/types.go
@pkg/types/errors.go
@pkg/types/logger.go

<interfaces>
<!-- Current state of files being modified -->

From pkg/types/types.go (lines 58-79):
```go
type ReconnectConfig struct {
    MaxAttempts       int
    InitialDelay      time.Duration
    MaxDelay          time.Duration
    BackoffMultiplier float64
}

func DefaultReconnectConfig() ReconnectConfig {
    return ReconnectConfig{
        MaxAttempts:       0, // 0 = infinite
        InitialDelay:      1 * time.Second,
        MaxDelay:          60 * time.Second,
        BackoffMultiplier: 1.618,
    }
}
```

From pkg/types/errors.go (lines 89-97):
```go
type ReconnectErrorCode string

const (
    ReconnectErrMaxAttempts    ReconnectErrorCode = "MAX_RECONNECT_ATTEMPTS"
    ReconnectErrMaxAuthRetries ReconnectErrorCode = "MAX_AUTH_RETRIES"
    ReconnectErrDisabled       ReconnectErrorCode = "RECONNECT_DISABLED"
)
```

From pkg/types/errors.go (lines 21-38):
```go
var ErrCertificateExpired = errors.New("certificate has expired")
// ... sentinel pattern from connection/tls.go
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Add RequestRateLimiter interface, TokenBucketLimiter, and MaxRetries to types</name>
  <files>pkg/types/types.go, pkg/types/rate_limiter_test.go, pkg/types/types_test.go</files>
  <read_first>
    - pkg/types/types.go (see current ReconnectConfig struct and DefaultReconnectConfig)
    - pkg/types/errors.go (see sentinel error pattern)
  </read_first>
  <behavior>
    - Test 1: TokenBucketLimiter with rate=10, burst=5 allows exactly 5 calls then denies the 6th
    - Test 2: TokenBucketLimiter refills tokens after time passes (rate=1000, burst=10, drain all, wait >10ms, Allow() returns true)
    - Test 3: TokenBucketLimiter does not exceed burst cap (rate=1000000, burst=3, call Allow 100 times, count trues <= 3)
    - Test 4: DefaultReconnectConfig() returns MaxRetries=10
    - Test 5: ReconnectConfig with MaxRetries=0 preserves backward compat (zero value means "use MaxAttempts")
    - Test 6: TokenBucketLimiter is safe for concurrent use (goroutines calling Allow() simultaneously do not race)
  </behavior>
  <action>
    Per FOUND-01 and CONTEXT.md decisions (GA-1: simple Allow() bool interface):

    1. Add to pkg/types/types.go, AFTER the ReconnectConfig struct (after line 68):

    ```go
    // MaxRetries field (FOUND-02): takes precedence over MaxAttempts when > 0.
    // Add to ReconnectConfig struct:
    //   MaxRetries int  // NEW: 0 = use MaxAttempts behavior; takes precedence
    ```

    Modify the ReconnectConfig struct to add MaxRetries int field between MaxAttempts and InitialDelay:
    ```go
    type ReconnectConfig struct {
        MaxAttempts       int           // existing: 0 = infinite
        MaxRetries        int           // NEW (FOUND-02): 0 = use MaxAttempts; takes precedence when > 0
        InitialDelay      time.Duration
        MaxDelay          time.Duration
        BackoffMultiplier float64
    }
    ```

    Update DefaultReconnectConfig() to set MaxRetries: 10:
    ```go
    func DefaultReconnectConfig() ReconnectConfig {
        return ReconnectConfig{
            MaxAttempts:       0, // 0 = infinite (legacy)
            MaxRetries:        10,
            InitialDelay:      1 * time.Second,
            MaxDelay:          60 * time.Second,
            BackoffMultiplier: 1.618,
        }
    }
    ```

    2. Add RequestRateLimiter interface and TokenBucketLimiter to pkg/types/types.go, AFTER the DefaultReconnectConfig function (after line 79):

    ```go
    // RequestRateLimiter controls request throughput (FOUND-01).
    // Implementations must be safe for concurrent use.
    type RequestRateLimiter interface {
        Allow() bool
    }

    // TokenBucketLimiter implements RequestRateLimiter using a token bucket algorithm.
    // Rate is in tokens per second; burst is the maximum tokens that can accumulate.
    type TokenBucketLimiter struct {
        rate     float64
        burst    int
        tokens   float64
        lastTime time.Time
        mu       sync.Mutex
    }

    // NewTokenBucketLimiter creates a new token bucket rate limiter.
    // rate: tokens added per second. burst: maximum token capacity.
    func NewTokenBucketLimiter(rate float64, burst int) *TokenBucketLimiter {
        return &TokenBucketLimiter{
            rate:     rate,
            burst:    burst,
            tokens:   float64(burst),
            lastTime: time.Now(),
        }
    }

    // Allow consumes one token if available. Returns true if allowed, false if rate limited.
    func (l *TokenBucketLimiter) Allow() bool {
        l.mu.Lock()
        defer l.mu.Unlock()
        now := time.Now()
        elapsed := now.Sub(l.lastTime).Seconds()
        l.lastTime = now
        l.tokens += elapsed * l.rate
        if l.tokens > float64(l.burst) {
            l.tokens = float64(l.burst)
        }
        if l.tokens >= 1 {
            l.tokens--
            return true
        }
        return false
    }
    ```

    Add "sync" to the import block in types.go (currently only imports "time").

    3. Create NEW file pkg/types/rate_limiter_test.go with table-driven tests for TokenBucketLimiter:
       - TestTokenBucketLimiter_AllowWithinBurst: rate=10, burst=5, first 5 calls return true, 6th returns false
       - TestTokenBucketLimiter_TokenRefill: drain tokens, wait, verify refill
       - TestTokenBucketLimiter_BurstCap: high rate, low burst, verify cap not exceeded
       - TestTokenBucketLimiter_ConcurrentSafety: multiple goroutines calling Allow()
       - TestTokenBucketLimiter_InterfaceCompliance: verify *TokenBucketLimiter implements RequestRateLimiter

    4. Add tests to pkg/types/types_test.go:
       - TestDefaultReconnectConfig_MaxRetries: verify DefaultReconnectConfig().MaxRetries == 10
       - TestReconnectConfig_MaxRetriesZeroBackwardCompat: verify MaxRetries=0 zero value behavior
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/types/ -run "TestTokenBucketLimiter|TestDefaultReconnectConfig_MaxRetries|TestReconnectConfig_MaxRetriesZeroBackwardCompat" -v -count=1</automated>
  </verify>
  <done>
    - pkg/types/types.go contains "RequestRateLimiter interface" with Allow() bool method
    - pkg/types/types.go contains "TokenBucketLimiter" struct with NewTokenBucketLimiter constructor
    - pkg/types/types.go ReconnectConfig struct has MaxRetries int field
    - DefaultReconnectConfig() returns MaxRetries=10
    - All new tests pass with go test -race
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Add sentinel errors ErrTooManyPendingRequests and ErrMaxRetriesExceeded</name>
  <files>pkg/types/errors.go, pkg/types/errors_test.go</files>
  <read_first>
    - pkg/types/errors.go (see existing error patterns at lines 21-38 for sentinel style, and lines 193-280 for typed error style)
    - pkg/connection/tls.go (lines 21-37 for sentinel error pattern: ErrCertificateExpired, etc.)
  </read_first>
  <behavior>
    - Test 1: errors.Is(ErrTooManyPendingRequests, ErrTooManyPendingRequests) returns true
    - Test 2: errors.Is(ErrMaxRetriesExceeded, ErrMaxRetriesExceeded) returns true
    - Test 3: ErrTooManyPendingRequests.Error() returns "too many pending requests"
    - Test 4: ErrMaxRetriesExceeded.Error() returns "max retries exceeded"
  </behavior>
  <action>
    Per FOUND-04 (ErrTooManyPendingRequests) and FOUND-02 (ErrMaxRetriesExceeded):

    Add sentinel error variables to pkg/types/errors.go, BEFORE the ErrorShape struct (before line 103, after the ReconnectErrorCode const block at line 97). This keeps them near the top-level error definitions, consistent with the sentinel pattern in pkg/connection/tls.go:

    ```go
    // ErrTooManyPendingRequests is returned when the pending request map exceeds its limit.
    // This is an internal SDK error, not a wire-protocol error (FOUND-04).
    var ErrTooManyPendingRequests = errors.New("too many pending requests")

    // ErrMaxRetriesExceeded is returned when reconnection attempts exceed MaxRetries.
    // This is an internal SDK error (FOUND-02).
    var ErrMaxRetriesExceeded = errors.New("max retries exceeded")
    ```

    Note: These are plain sentinel errors (var = errors.New(...)), not typed error structs.
    This follows the pattern from pkg/connection/tls.go (ErrCertificateExpired, etc.)
    and is consistent with the research recommendation for internal SDK errors.

    Add tests to pkg/types/errors_test.go:
    - TestErrTooManyPendingRequests_Is: verify errors.Is works
    - TestErrTooManyPendingRequests_Message: verify error message string
    - TestErrMaxRetriesExceeded_Is: verify errors.Is works
    - TestErrMaxRetriesExceeded_Message: verify error message string
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/types/ -run "TestErrTooManyPendingRequests|TestErrMaxRetriesExceeded" -v -count=1</automated>
  </verify>
  <done>
    - pkg/types/errors.go contains "var ErrTooManyPendingRequests = errors.New"
    - pkg/types/errors.go contains "var ErrMaxRetriesExceeded = errors.New"
    - Both errors pass errors.Is() checks
    - Existing tests still pass: go test ./pkg/types/ -count=1
  </done>
</task>

</tasks>

<verification>
go test ./pkg/types/ -race -count=1
go vet ./pkg/types/
grep -c "RequestRateLimiter" pkg/types/types.go  # should be >= 2 (interface + comment)
grep -c "ErrTooManyPendingRequests" pkg/types/errors.go  # should be >= 2 (var + comment)
grep -c "ErrMaxRetriesExceeded" pkg/types/errors.go  # should be >= 2 (var + comment)
grep -c "MaxRetries" pkg/types/types.go  # should be >= 3 (struct field + default + comment)
</verification>

<success_criteria>
1. RequestRateLimiter interface with Allow() bool exists in pkg/types/types.go
2. TokenBucketLimiter correctly implements the interface with token bucket algorithm
3. MaxRetries int field exists on ReconnectConfig with default value 10
4. ErrTooManyPendingRequests and ErrMaxRetriesExceeded sentinel errors exist
5. All tests pass with go test -race ./pkg/types/
6. No existing tests broken
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-01-SUMMARY.md`
</output>
