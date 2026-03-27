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
  - pkg/types/errors_test.go
autonomous: true
requirements:
  - FOUND-01
  - FOUND-02
  - FOUND-04

must_haves:
  truths:
    - "RequestRateLimiter interface exists with Allow() bool method"
    - "TokenBucketLimiter implements RequestRateLimiter with configurable rate and burst"
    - "ErrTooManyPendingRequests typed error (RequestError) exists and supports errors.Is() via Err sentinel"
    - "ErrMaxRetriesExceeded typed error (ReconnectError) exists and supports errors.Is() via Err sentinel"
    - "ReconnectConfig has MaxRetries field with explicit precedence documentation; DefaultReconnectConfig() sets MaxRetries=10"
    - "MaxRetries precedence over MaxAttempts is documented with exact zero/negative semantics"
  artifacts:
    - path: "pkg/types/types.go"
      provides: "RequestRateLimiter interface, TokenBucketLimiter struct, MaxRetries field on ReconnectConfig with precedence docs"
      contains: "RequestRateLimiter"
    - path: "pkg/types/errors.go"
      provides: "ErrTooManyPendingRequests sentinel + typed TooManyPendingRequests error, ErrMaxRetriesExceeded sentinel + typed MaxRetriesExceeded error"
      contains: "ErrTooManyPendingRequests"
    - path: "pkg/types/rate_limiter_test.go"
      provides: "Tests for TokenBucketLimiter behavior"
      exports: []
    - path: "pkg/types/types_test.go"
      provides: "Tests for MaxRetries default, precedence rules, and ReconnectConfig"
    - path: "pkg/types/errors_test.go"
      provides: "Tests for new typed errors and sentinel Is() support"
  key_links:
    - from: "pkg/types/types.go"
      to: "pkg/types/errors.go"
      via: "shared package -- typed errors reference sentinels"
      pattern: "RequestRateLimiter"
---

<objective>
Define shared type contracts, interfaces, and typed errors that Plans 02 and 03 consume.

Purpose: Plans 02 and 03 both depend on new types in pkg/types. This plan creates those contracts first so downstream plans implement against known interfaces. Addresses review concern: typed errors instead of plain sentinels, and MaxRetries/MaxAttempts precedence with exact semantics.

Output: New RequestRateLimiter interface + TokenBucketLimiter implementation, MaxRetries field on ReconnectConfig with documented precedence rules, and typed error variables ErrTooManyPendingRequests + ErrMaxRetriesExceeded that integrate with the existing error hierarchy.
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

From pkg/types/errors.go (lines 89-97) -- existing ReconnectErrorCode:
```go
type ReconnectErrorCode string

const (
    ReconnectErrMaxAttempts    ReconnectErrorCode = "MAX_RECONNECT_ATTEMPTS"
    ReconnectErrMaxAuthRetries ReconnectErrorCode = "MAX_AUTH_RETRIES"
    ReconnectErrDisabled       ReconnectErrorCode = "RECONNECT_DISABLED"
)
```

From pkg/types/errors.go (lines 120-134) -- existing typed error pattern:
```go
type OpenClawError interface {
    error
    Code() string
    Retryable() bool
    Details() any
    Unwrap() error
}

type BaseError struct {
    code      string
    message   string
    retryable bool
    details   any
    err       error
}
```

From pkg/types/errors.go (lines 193-205) -- RequestError pattern to follow:
```go
type RequestError struct {
    *BaseError
}

func NewRequestError(code, message string, retryable bool, details any) *RequestError {
    return &RequestError{&BaseError{
        code: code, message: message,
        retryable: retryable, details: details,
    }}
}
```

From pkg/types/errors.go (lines 223-236) -- ReconnectError pattern to follow:
```go
type ReconnectError struct {
    *BaseError
}

func NewReconnectError(code, message string, retryable bool, details any) *ReconnectError {
    return &ReconnectError{&BaseError{
        code: code, message: message,
        retryable: retryable, details: details,
    }}
}
```

From pkg/connection/tls.go (lines 21-37) -- sentinel error pattern:
```go
var ErrCertificateExpired = errors.New("certificate has expired")
var ErrCertificateNotYetValid = errors.New("certificate is not yet valid")
// ... more sentinels
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Add RequestRateLimiter interface, TokenBucketLimiter, and MaxRetries to types</name>
  <files>pkg/types/types.go, pkg/types/rate_limiter_test.go, pkg/types/types_test.go</files>
  <read_first>
    - pkg/types/types.go (see current ReconnectConfig struct and DefaultReconnectConfig)
    - pkg/types/errors.go (see existing error patterns)
  </read_first>
  <acceptance_criteria>
    - pkg/types/types.go contains "RequestRateLimiter" interface with "Allow() bool" method
    - pkg/types/types.go contains "TokenBucketLimiter" struct with NewTokenBucketLimiter constructor
    - pkg/types/types.go ReconnectConfig struct has "MaxRetries int" field between MaxAttempts and InitialDelay
    - DefaultReconnectConfig() returns MaxRetries=10
    - ReconnectConfig has doc comment explaining MaxRetries precedence: "MaxRetries > 0 takes precedence over MaxAttempts; MaxRetries == 0 falls back to MaxAttempts; both zero means infinite"
    - go test ./pkg/types/ -run "TestTokenBucketLimiter|TestDefaultReconnectConfig|TestReconnectConfig" -v -count=1 exits 0
    - go test -race ./pkg/types/ exits 0
  </acceptance_criteria>
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

    Addresses review concern [HIGH]: MaxRetries overlaps with MaxAttempts -- adds explicit precedence documentation with exact nil/zero/negative semantics.

    1. Modify ReconnectConfig struct in pkg/types/types.go to add MaxRetries field with detailed precedence comment:

    ```go
    // ReconnectConfig holds configuration for automatic reconnection.
    type ReconnectConfig struct {
        // MaxAttempts is the legacy retry budget field.
        // Deprecated: use MaxRetries instead. MaxAttempts=0 means infinite (default).
        MaxAttempts int

        // MaxRetries sets the maximum number of reconnection attempts (FOUND-02).
        //
        // Precedence rules:
        //   - MaxRetries > 0: use MaxRetries as the limit (ignores MaxAttempts)
        //   - MaxRetries == 0 AND MaxAttempts > 0: fall back to MaxAttempts
        //   - MaxRetries == 0 AND MaxAttempts == 0: unlimited retries (backward compat)
        //   - MaxRetries < 0: treated as 0 (same as unset)
        //   - MaxAttempts < 0: treated as 0 (same as unset)
        MaxRetries        int
        InitialDelay      time.Duration
        MaxDelay          time.Duration
        BackoffMultiplier float64
    }
    ```

    Update DefaultReconnectConfig():
    ```go
    func DefaultReconnectConfig() ReconnectConfig {
        return ReconnectConfig{
            MaxAttempts:       0, // 0 = infinite (legacy, deprecated)
            MaxRetries:        10, // FOUND-02: sensible production default
            InitialDelay:      1 * time.Second,
            MaxDelay:          60 * time.Second,
            BackoffMultiplier: 1.618,
        }
    }
    ```

    2. Add RequestRateLimiter interface and TokenBucketLimiter AFTER DefaultReconnectConfig function:

    ```go
    // RequestRateLimiter controls request throughput (FOUND-01).
    // Implementations must be safe for concurrent use by multiple goroutines.
    type RequestRateLimiter interface {
        Allow() bool
    }

    // TokenBucketLimiter implements RequestRateLimiter using a token bucket algorithm.
    // Rate is in tokens per second; burst is the maximum tokens that can accumulate.
    // The limiter starts with a full bucket (burst tokens available immediately).
    type TokenBucketLimiter struct {
        rate     float64
        burst    int
        tokens   float64
        lastTime time.Time
        mu       sync.Mutex
    }

    func NewTokenBucketLimiter(rate float64, burst int) *TokenBucketLimiter {
        return &TokenBucketLimiter{
            rate:     rate,
            burst:    burst,
            tokens:   float64(burst),
            lastTime: time.Now(),
        }
    }

    // Allow consumes one token if available. Returns true if allowed, false if rate limited.
    // Thread-safe: safe to call from multiple goroutines.
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
       - TestTokenBucketLimiter_AllowWithinBurst
       - TestTokenBucketLimiter_TokenRefill
       - TestTokenBucketLimiter_BurstCap
       - TestTokenBucketLimiter_ConcurrentSafety
       - TestTokenBucketLimiter_InterfaceCompliance (var _ RequestRateLimiter = (*TokenBucketLimiter)(nil))

    4. Add tests to pkg/types/types_test.go:
       - TestDefaultReconnectConfig_MaxRetries: verify DefaultReconnectConfig().MaxRetries == 10
       - TestDefaultReconnectConfig_MaxAttemptsLegacy: verify MaxAttempts == 0 (infinite legacy)
       - TestReconnectConfig_MaxRetriesPrecedence: table-driven test for precedence rules (MaxRetries>0 wins, fallback to MaxAttempts, both zero = infinite, negative = treated as zero)
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/types/ -run "TestTokenBucketLimiter|TestDefaultReconnectConfig|TestReconnectConfig" -v -race -count=1</automated>
  </verify>
  <done>
    - pkg/types/types.go contains RequestRateLimiter interface with Allow() bool method
    - pkg/types/types.go contains TokenBucketLimiter struct with NewTokenBucketLimiter constructor
    - pkg/types/types.go ReconnectConfig struct has MaxRetries int field with full precedence documentation
    - DefaultReconnectConfig() returns MaxRetries=10
    - Precedence rules documented: MaxRetries>0 wins, MaxRetries==0 falls back to MaxAttempts, both==0 means infinite
    - All new tests pass with go test -race
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Add typed errors ErrTooManyPendingRequests and ErrMaxRetriesExceeded</name>
  <files>pkg/types/errors.go, pkg/types/errors_test.go</files>
  <read_first>
    - pkg/types/errors.go (see existing error patterns at lines 120-280 for typed error style, lines 193-205 for RequestError, lines 223-236 for ReconnectError)
    - pkg/connection/tls.go (lines 21-37 for sentinel error pattern)
  </read_first>
  <acceptance_criteria>
    - pkg/types/errors.go contains "var ErrTooManyPendingRequests = errors.New" sentinel
    - pkg/types/errors.go contains "var ErrMaxRetriesExceeded = errors.New" sentinel
    - pkg/types/errors.go contains typed TooManyPendingRequests error that wraps ErrTooManyPendingRequests via BaseError.Unwrap()
    - pkg/types/errors.go contains typed MaxRetriesExceededError that wraps ErrMaxRetriesExceeded via BaseError.Unwrap()
    - errors.Is(typedErr, ErrTooManyPendingRequests) returns true
    - errors.Is(typedErr, ErrMaxRetriesExceeded) returns true
    - Typed errors implement OpenClawError interface (Code(), Retryable(), Details(), Unwrap())
    - go test ./pkg/types/ -run "TestErrTooManyPendingRequests|TestErrMaxRetriesExceeded" -v -count=1 exits 0
  </acceptance_criteria>
  <behavior>
    - Test 1: errors.Is(ErrTooManyPendingRequests, ErrTooManyPendingRequests) returns true
    - Test 2: errors.Is(ErrMaxRetriesExceeded, ErrMaxRetriesExceeded) returns true
    - Test 3: NewTooManyPendingRequestsError() returns an error that satisfies errors.Is(err, ErrTooManyPendingRequests)
    - Test 4: NewMaxRetriesExceededError(10) returns an error that satisfies errors.Is(err, ErrMaxRetriesExceeded)
    - Test 5: Typed errors implement OpenClawError interface
    - Test 6: NewTooManyPendingRequestsError().Retryable() returns true (transient condition)
    - Test 7: NewMaxRetriesExceededError(10).Retryable() returns false (budget exhausted)
    - Test 8: NewMaxRetriesExceededError(10).Code() returns "MAX_RETRIES_EXCEEDED"
  </behavior>
  <action>
    Per FOUND-04 (ErrTooManyPendingRequests) and FOUND-02 (ErrMaxRetriesExceeded).

    Addresses review concern [MEDIUM]: Sentinel errors don't match existing typed-error pattern in pkg/types/errors.go:223. Solution: provide BOTH sentinel vars (for errors.Is) AND typed constructors that wrap the sentinel, following the dual pattern. This gives callers the choice of simple errors.Is() matching OR typed error inspection via OpenClawError interface.

    Add to pkg/types/errors.go, AFTER the ReconnectError section (after line ~236) but BEFORE the TimeoutError section:

    ```go
    // --- Foundation Phase Errors (FOUND-02, FOUND-04) ---

    // ErrTooManyPendingRequests is the sentinel error for pending request limit exceeded.
    // Use errors.Is(err, ErrTooManyPendingRequests) to check for this condition.
    var ErrTooManyPendingRequests = errors.New("too many pending requests")

    // ErrMaxRetriesExceeded is the sentinel error for retry budget exhaustion.
    // Use errors.Is(err, ErrMaxRetriesExceeded) to check for this condition.
    var ErrMaxRetriesExceeded = errors.New("max retries exceeded")

    // TooManyPendingRequestsError is returned when the pending request map exceeds its limit (FOUND-04).
    // This is a transient condition -- the caller can retry when pending requests complete.
    type TooManyPendingRequestsError struct {
        *BaseError
    }

    // NewTooManyPendingRequestsError creates a new TooManyPendingRequestsError.
    // The error wraps ErrTooManyPendingRequests so errors.Is(err, ErrTooManyPendingRequests) works.
    func NewTooManyPendingRequestsError(limit int) *TooManyPendingRequestsError {
        return &TooManyPendingRequestsError{&BaseError{
            code:      "TOO_MANY_PENDING_REQUESTS",
            message:   fmt.Sprintf("pending request limit (%d) exceeded", limit),
            retryable: true, // Transient: caller can retry after pending requests complete
            details:   map[string]int{"limit": limit},
            err:       ErrTooManyPendingRequests,
        }}
    }

    // MaxRetriesExceededError is returned when reconnection attempts exceed MaxRetries (FOUND-02).
    // This is a terminal condition -- no more retries will be attempted.
    type MaxRetriesExceededError struct {
        *BaseError
    }

    // NewMaxRetriesExceededError creates a new MaxRetriesExceededError.
    // The error wraps ErrMaxRetriesExceeded so errors.Is(err, ErrMaxRetriesExceeded) works.
    func NewMaxRetriesExceededError(maxRetries int) *MaxRetriesExceededError {
        return &MaxRetriesExceededError{&BaseError{
            code:      "MAX_RETRIES_EXCEEDED",
            message:   fmt.Sprintf("max retries exceeded: %d attempts", maxRetries),
            retryable: false, // Terminal: budget exhausted
            details:   map[string]int{"max_retries": maxRetries},
            err:       ErrMaxRetriesExceeded,
        }}
    }
    ```

    IMPORTANT: The `err` field in BaseError is the "cause" field. Check BaseError struct -- the `err` field is set via the literal and returned by the `Unwrap()` method. Verify by reading BaseError.Unwrap() implementation. The typed errors MUST wrap the sentinel so errors.Is() chains work.

    Ensure "fmt" is in the imports for pkg/types/errors.go (it already is).

    Add tests to pkg/types/errors_test.go:
    - TestErrTooManyPendingRequests_SentinelIs: errors.Is(ErrTooManyPendingRequests, ErrTooManyPendingRequests)
    - TestErrTooManyPendingRequests_TypedIs: errors.Is(NewTooManyPendingRequestsError(256), ErrTooManyPendingRequests) == true
    - TestErrTooManyPendingRequests_OpenClawError: typed err implements OpenClawError interface
    - TestErrTooManyPendingRequests_Retryable: NewTooManyPendingRequestsError(256).Retryable() == true
    - TestErrMaxRetriesExceeded_SentinelIs: errors.Is(ErrMaxRetriesExceeded, ErrMaxRetriesExceeded)
    - TestErrMaxRetriesExceeded_TypedIs: errors.Is(NewMaxRetriesExceededError(10), ErrMaxRetriesExceeded) == true
    - TestErrMaxRetriesExceeded_OpenClawError: typed err implements OpenClawError interface
    - TestErrMaxRetriesExceeded_NotRetryable: NewMaxRetriesExceededError(10).Retryable() == false
    - TestErrMaxRetriesExceeded_Code: NewMaxRetriesExceededError(10).Code() == "MAX_RETRIES_EXCEEDED"
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/types/ -run "TestErrTooManyPendingRequests|TestErrMaxRetriesExceeded" -v -race -count=1</automated>
  </verify>
  <done>
    - pkg/types/errors.go contains ErrTooManyPendingRequests sentinel var
    - pkg/types/errors.go contains ErrMaxRetriesExceeded sentinel var
    - pkg/types/errors.go contains TooManyPendingRequestsError typed struct with constructor
    - pkg/types/errors.go contains MaxRetriesExceededError typed struct with constructor
    - Typed errors wrap sentinels so errors.Is() works for both sentinel and typed
    - Typed errors implement OpenClawError interface (Code, Retryable, Details, Unwrap)
    - TooManyPendingRequestsError is retryable (transient)
    - MaxRetriesExceededError is not retryable (terminal)
    - Existing tests still pass: go test ./pkg/types/ -count=1
  </done>
</task>

</tasks>

<verification>
go test ./pkg/types/ -race -count=1
go vet ./pkg/types/
grep -c "RequestRateLimiter" pkg/types/types.go  # should be >= 2 (interface + comment)
grep -c "ErrTooManyPendingRequests" pkg/types/errors.go  # should be >= 3 (sentinel + typed + comment)
grep -c "ErrMaxRetriesExceeded" pkg/types/errors.go  # should be >= 3 (sentinel + typed + comment)
grep -c "MaxRetries" pkg/types/types.go  # should be >= 5 (field + default + precedence docs)
grep -c "MaxRetries > 0" pkg/types/types.go  # should be >= 1 (precedence rule documented)
grep -c "TooManyPendingRequestsError" pkg/types/errors.go  # should be >= 2
grep -c "MaxRetriesExceededError" pkg/types/errors.go  # should be >= 2
</verification>

<success_criteria>
1. RequestRateLimiter interface with Allow() bool exists in pkg/types/types.go
2. TokenBucketLimiter correctly implements the interface with token bucket algorithm
3. MaxRetries int field exists on ReconnectConfig with default value 10
4. MaxRetries precedence rules documented with exact zero/negative semantics
5. ErrTooManyPendingRequests and ErrMaxRetriesExceeded sentinel errors exist
6. Typed TooManyPendingRequestsError and MaxRetriesExceededError exist implementing OpenClawError
7. Typed errors wrap sentinels so errors.Is() chain works
8. All tests pass with go test -race ./pkg/types/
9. No existing tests broken
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-01-SUMMARY.md`
</output>
