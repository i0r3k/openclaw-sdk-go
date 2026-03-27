# Phase 1: Foundation Hardening - Research

**Researched:** 2026-03-28
**Domain:** Go WebSocket SDK -- concurrency safety, rate limiting, retry budgets, TLS validation
**Confidence:** HIGH

## Summary

Phase 1 hardens the OpenClaw SDK Go for production use by adding five critical safety mechanisms: client-side rate limiting (FOUND-01), retry budgets (FOUND-02), TLS CRL validation stub (FOUND-03), pending request limits (FOUND-04), and InsecureSkipVerify warnings (FOUND-05). Each requirement targets a specific unbounded or unsafe behavior that could cause production incidents.

The current codebase has a clean architecture with clear separation across managers (request, reconnect, connection, event) and a well-established option pattern for client configuration. The primary implementation risk is preserving backward compatibility while adding new constraints. All five changes are localized to specific files with clear insertion points identified in the code.

**Primary recommendation:** Implement all five requirements as additive-only changes -- new interfaces, new struct fields with zero-value defaults preserving old behavior, and new error sentinels. No existing function signatures change. The `golang.org/x/time/rate` dependency should be avoided in favor of a simple internal token bucket to maintain the project's minimal-dependency principle.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Rate limiting location**: `RequestManager.SendRequest()` -- pre-send check before serialize
- **Rate limiting interface pattern**: `RequestRateLimiter` interface with `Allow()` method + `WithRateLimit()` option
- **Retry budget field**: `MaxRetries` added to `ReconnectConfig` (distinct from `MaxAttempts` for clarity)
- **Pending request limit location**: `RequestManager.SendRequest()` -- check map size before adding
- **Pending limit error**: `ErrTooManyPendingRequests` returned immediately when limit reached
- **InsecureSkipVerify warning location**: `TlsValidator.GetTLSConfig()` -- log warning after config built
- **TLS CRL approach**: Stub with explicit comment (not implementing actual CRL fetching for v1)

### Gray Areas (Claude's Discretion)
- GA-1: Rate limiter interface -- recommends Option A (simple `Allow() bool`)
- GA-2: Rate limiter placement -- recommends Option B (`client.SendRequest()` checks before `RequestManager.SendRequest`)
- GA-3: Default pending request limit -- recommends 256
- GA-4: Default MaxRetries value -- recommends Option B (add `MaxRetries=10`, keep `MaxAttempts=0`)
- GA-5: InsecureSkipVerify warning log level -- recommends WARN

### Deferred Ideas (OUT OF SCOPE)
- Per-namespace rate limiting
- Actual CRL/OCSP fetching implementation
- Middleware/interceptor hooks
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| FOUND-01 | Client-side rate limiting -- `RequestRateLimiter` interface with `WithRateLimit()` option | Internal token bucket limiter implementing `Allow() bool` interface; `WithRateLimit(limiter)` functional option; check at `client.SendRequest()` before delegating to `RequestManager` |
| FOUND-02 | Retry budget -- `MaxRetries` field in `ReconnectConfig`, default 10 | Add `MaxRetries int` field to `types.ReconnectConfig`; check in `ReconnectManager.run()` alongside `MaxAttempts`; return `ErrMaxRetriesExceeded` sentinel |
| FOUND-03 | TLS CRL validation -- stub with explicit comment | `CheckCertificateRevocation` already exists as stub (lines 250-280 of tls.go); add explicit v1 limitation comment and clear documentation |
| FOUND-04 | Pending request limit -- max pending with `ErrTooManyPendingRequests` | Add `maxPending int` field to `RequestManager`; check `len(rm.pending) >= rm.maxPending` before adding; `ErrTooManyPendingRequests` sentinel error |
| FOUND-05 | InsecureSkipVerify warning -- log at connection time | `TlsValidator.GetTLSConfig()` needs Logger injection; call `logger.Warn(...)` after building config when `InsecureSkipVerify==true` |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gorilla/websocket | v1.5.3 | WebSocket transport | Already in project; de facto Go WebSocket library |
| Go stdlib `sync` | 1.24+ | Mutex, Cond for thread safety | Standard for Go concurrency |
| Go stdlib `crypto/tls` | 1.24+ | TLS config and certificate validation | Standard for Go TLS |
| Go stdlib `crypto/x509` | 1.24+ | Certificate parsing and CRL types | Standard for certificate inspection |
| Go stdlib `time` | 1.24+ | Timer-based token bucket | Avoids external dependency |
| Go stdlib `testing` | 1.24+ | Unit and table-driven tests | Standard Go test framework |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Internal token bucket | `golang.org/x/time/rate` v0.15.0 | x/time/rate is battle-tested but adds external dependency; project principle is minimal deps; a ~30-line token bucket is sufficient |
| Sentinel error variables | `errors.New()` inline | Sentinel vars enable `errors.Is()` matching; consistent with existing error pattern in `pkg/types/errors.go` |

**No new external dependencies required.** All five requirements can be implemented with Go standard library alone.

## Architecture Patterns

### Recommended Project Structure (Changes Only)
```
pkg/
├── types/
│   ├── errors.go        # Add ErrTooManyPendingRequests, ErrMaxRetriesExceeded
│   └── types.go         # Add RequestRateLimiter interface, MaxRetries field
├── managers/
│   ├── request.go       # Add maxPending field, check in SendRequest
│   └── reconnect.go     # Add MaxRetries check in run()
├── connection/
│   └── tls.go           # Add Logger field to TlsValidator, warn in GetTLSConfig
├── ratelimit/
│   └── token_bucket.go  # NEW: simple token bucket implementing RequestRateLimiter
└── client.go            # Add rateLimiter field, WithRateLimit option, check in SendRequest
```

### Pattern 1: Functional Option for Rate Limiter
**What:** Follow existing `ClientOption func(*ClientConfig) error` pattern
**When to use:** FOUND-01 -- adding rate limiter configuration
**Example:**
```go
// Source: follows existing pattern from pkg/client.go WithURL, WithLogger, etc.

// RequestRateLimiter controls request throughput.
type RequestRateLimiter interface {
    Allow() bool
}

// WithRateLimit sets a rate limiter for the client.
func WithRateLimit(limiter RequestRateLimiter) ClientOption {
    return func(c *ClientConfig) error {
        c.RateLimiter = limiter
        return nil
    }
}
```

### Pattern 2: Backward-Compatible Struct Field Addition
**What:** Add new fields with zero-value defaults that preserve old behavior
**When to use:** FOUND-02 (MaxRetries), FOUND-04 (maxPending)
**Example:**
```go
// Source: Go blog "Keeping Your Modules Compatible"
// Adding a field with zero-value meaning "use existing behavior" is safe.

type ReconnectConfig struct {
    MaxAttempts       int           // existing, 0 = infinite
    MaxRetries        int           // NEW, 0 = use MaxAttempts behavior
    InitialDelay      time.Duration
    MaxDelay          time.Duration
    BackoffMultiplier float64
}

// In ReconnectManager.run(), check MaxRetries first:
// if rm.config.MaxRetries > 0 && attempt >= rm.config.MaxRetries { return }
// Then fall back to MaxAttempts:
// if rm.config.MaxAttempts > 0 && attempt >= rm.config.MaxAttempts { return }
```

### Pattern 3: Sentinel Error with errors.Is() Support
**What:** Package-level error variables for specific failure conditions
**When to use:** FOUND-04 (ErrTooManyPendingRequests), FOUND-02 (ErrMaxRetriesExceeded)
**Example:**
```go
// Source: follows existing pattern in pkg/connection/tls.go (ErrCertificateExpired, etc.)

var ErrTooManyPendingRequests = errors.New("too many pending requests")

// Usage in RequestManager.SendRequest():
// if rm.maxPending > 0 && len(rm.pending) >= rm.maxPending {
//     return nil, ErrTooManyPendingRequests
// }
```

### Pattern 4: Simple Token Bucket (Internal)
**What:** Lightweight rate limiter without external dependencies
**When to use:** FOUND-01 -- default rate limiter implementation
**Example:**
```go
// Source: standard token bucket algorithm
type TokenBucketLimiter struct {
    rate     float64       // tokens per second
    burst    int           // max tokens
    tokens   float64       // current tokens
    lastTime time.Time     // last refill time
    mu       sync.Mutex
}

func (l *TokenBucketLimiter) Allow() bool {
    l.mu.Lock()
    defer l.mu.Unlock()
    now := time.Now()
    elapsed := now.Sub(l.lastTime).Seconds()
    l.tokens += elapsed * l.rate
    if l.tokens > float64(l.burst) {
        l.tokens = float64(l.burst)
    }
    l.lastTime = now
    if l.tokens >= 1 {
        l.tokens--
        return true
    }
    return false
}
```

### Anti-Patterns to Avoid
- **Breaking function signatures:** Never change `NewRequestManager(ctx)` to `NewRequestManager(ctx, maxPending)`. Use a setter or functional option instead.
- **Holding lock during channel send:** The codebase rule is "never send to channel while holding lock." Rate limiting checks must happen BEFORE acquiring the pending map lock.
- **Returning error types that don't match existing hierarchy:** `ErrTooManyPendingRequests` should be a plain sentinel (for `errors.Is`), not a new error struct type, since it's a simple boolean condition.
- **Logging without a Logger:** `TlsValidator` currently has no Logger field. Must add one rather than using `fmt.Println` or the global `log` package.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Complex rate limiting with burst/distribution | Leaky bucket, sliding window | Simple token bucket | Burst semantics sufficient for WebSocket SDK; ~30 lines vs. hundreds |
| CRL/OCSP certificate revocation | Full CRL fetch + parse + cache | Stub with explicit comment | Per CONTEXT.md decision: v1 scope is stub only; crypto/x509.RevocationList type exists for future use |
| Thread-safe pending request counting | Atomic counters separate from map | `len(rm.pending)` under existing mutex | Map length is O(1) in Go; already under mutex in SendRequest |

**Key insight:** The existing mutex in `RequestManager` already protects the pending map. The pending limit check is a single `len()` call under that lock -- no additional synchronization needed.

## Common Pitfalls

### Pitfall 1: Rate Limiter Check Placement
**What goes wrong:** Checking rate limit inside `RequestManager.SendRequest()` after acquiring the map lock
**Why it happens:** CONTEXT.md says "pre-send check" but doesn't specify lock ordering
**How to avoid:** GA-2 recommends checking at `client.SendRequest()` BEFORE calling `rm.SendRequest()`. This avoids holding the RequestManager lock during rate limit computation.
**Warning signs:** Deadlock in concurrent load tests; `go test -race` failures.

### Pitfall 2: MaxRetries vs MaxAttempts Confusion
**What goes wrong:** Implementing `MaxRetries` as an alias for `MaxAttempts` instead of a distinct field
**Why it happens:** The names are similar and both control reconnect limits
**How to avoid:** Per GA-4: `MaxRetries=10` is the new production default. `MaxAttempts=0` remains infinite for backward compat. Check `MaxRetries > 0` FIRST in the reconnect loop, then fall back to `MaxAttempts`. Document that `MaxRetries` takes precedence.
**Warning signs:** Existing tests that rely on infinite reconnects start failing.

### Pitfall 3: TlsValidator Has No Logger
**What goes wrong:** Trying to call `logger.Warn()` in `GetTLSConfig()` when `TlsValidator` struct has no Logger field
**Why it happens:** `TlsValidator` was designed as a pure validator with no logging dependency
**How to avoid:** Add a `logger types.Logger` field to `TlsValidator`. Set it in `NewTlsValidator` or add a `SetLogger` method. The transport layer already receives a logger through config, so thread it through.
**Warning signs:** Nil pointer dereference in `GetTLSConfig` when `InsecureSkipVerify=true`.

### Pitfall 4: Pending Limit Check Race
**What goes wrong:** Checking `len(rm.pending)` before acquiring the mutex
**Why it happens:** Trying to do a "fast path" check without the lock
**How to avoid:** Check `len(rm.pending) >= rm.maxPending` AFTER acquiring `rm.mu.Lock()` but BEFORE adding to the map. This is the same lock acquisition point already in `SendRequest()`.
**Warning signs:** More than `maxPending` entries in the map under high concurrency.

### Pitfall 5: CRL Stub Ambiguity
**What goes wrong:** The existing stub returns `nil` without any documentation about what that means
**Why it happens:** Original implementation was a placeholder
**How to avoid:** Per FOUND-03: add explicit comment block explaining this is intentionally a stub for v1, what real implementation would require (CRL fetch from DistributionPoints, OCSP check, caching), and link to the requirement.
**Warning signs:** Someone sees the function and assumes revocation checking is actually happening.

## Code Examples

### FOUND-01: Rate Limiter Interface and Option
```go
// pkg/types/types.go -- add interface
type RequestRateLimiter interface {
    Allow() bool
}

// pkg/client.go -- add to ClientConfig
type ClientConfig struct {
    // ... existing fields ...
    RateLimiter types.RequestRateLimiter // NEW: optional rate limiter
}

// pkg/client.go -- add option
func WithRateLimit(limiter types.RequestRateLimiter) ClientOption {
    return func(c *ClientConfig) error {
        c.RateLimiter = limiter
        return nil
    }
}

// pkg/client.go -- check in SendRequest (GA-2: client level)
func (c *client) SendRequest(ctx context.Context, req *protocol.RequestFrame) (*protocol.ResponseFrame, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Rate limit check BEFORE delegating to RequestManager
    if c.config.RateLimiter != nil && !c.config.RateLimiter.Allow() {
        return nil, NewRequestError("RATE_LIMITED", "rate limit exceeded", true, nil)
    }

    // ... existing code ...
}
```

### FOUND-02: MaxRetries in ReconnectManager
```go
// pkg/types/types.go -- add field
type ReconnectConfig struct {
    MaxAttempts       int           // existing: 0 = infinite
    MaxRetries        int           // NEW: 0 = use MaxAttempts; takes precedence
    InitialDelay      time.Duration
    MaxDelay          time.Duration
    BackoffMultiplier float64
}

// pkg/managers/reconnect.go -- check in run() loop, after line 118
// Replace the single check at line 120 with:
maxRetries := rm.config.MaxRetries
if maxRetries <= 0 {
    // Fall back to MaxAttempts for backward compatibility
    maxRetries = rm.config.MaxAttempts
}
if maxRetries > 0 && attempt >= maxRetries {
    if onReconnectFailed != nil {
        onReconnectFailed(NewReconnectError(
            string(ReconnectErrMaxAttempts),
            fmt.Sprintf("max retries exceeded: %d", maxRetries),
            false, nil,
        ))
    }
    return
}
```

### FOUND-04: Pending Request Limit
```go
// pkg/managers/request.go -- add field and constructor change
type RequestManager struct {
    pending    map[string]*pendingRequest
    timeouts   map[string]context.CancelFunc
    mu         sync.Mutex
    ctx        context.Context
    cancel     context.CancelFunc
    closed     bool
    maxPending int  // NEW: 0 = unlimited
}

// In SendRequest(), after acquiring lock (line 57), before adding to map:
if rm.maxPending > 0 && len(rm.pending) >= rm.maxPending {
    rm.mu.Unlock()
    return nil, ErrTooManyPendingRequests
}
```

### FOUND-05: InsecureSkipVerify Warning
```go
// pkg/connection/tls.go -- add logger to TlsValidator
type TlsValidator struct {
    config *TLSConfig
    logger types.Logger  // NEW: optional logger for warnings
}

func NewTlsValidator(config *TLSConfig) *TlsValidator {
    return &TlsValidator{config: config}
}

// Add SetLogger method
func (v *TlsValidator) SetLogger(logger types.Logger) {
    v.logger = logger
}

// In GetTLSConfig(), after building config (after line 117):
if v.config.InsecureSkipVerify && v.logger != nil {
    v.logger.Warn("TLS InsecureSkipVerify is enabled -- server certificate verification disabled; not recommended for production use")
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Unbounded pending requests | Bounded pending map with configurable limit | This phase | Prevents memory exhaustion under load |
| Infinite reconnect retries | Bounded retries with MaxRetries=10 default | This phase | Prevents permanent retry loops |
| Silent InsecureSkipVerify | Explicit WARN log when enabled | This phase | Operators aware of security risk |
| No rate limiting | Optional RateLimiter interface | This phase | Prevents server rejection under burst load |

**Deprecated/outdated:**
- `MaxAttempts` field in `ReconnectConfig`: Still functional but `MaxRetries` is the preferred field going forward. Document `MaxAttempts` as legacy.

## Open Questions

1. **Should `ErrTooManyPendingRequests` be a typed error or sentinel?**
   - What we know: Existing pattern uses typed errors (`RequestError`, `ConnectionError`) for wire-protocol errors and sentinel `var` for internal conditions (like `ErrCertificateExpired` in tls.go).
   - What's unclear: Whether the planner wants this to match the typed error hierarchy for API consistency.
   - Recommendation: Use a simple sentinel `var ErrTooManyPendingRequests = errors.New(...)` consistent with the `pkg/connection/tls.go` pattern. It's an internal SDK error, not a wire-protocol error.

2. **Should `TokenBucketLimiter` be in its own package or in `pkg/types`?**
   - What we know: Interface goes in `pkg/types`. Implementation could go with the interface or in a separate package.
   - What's unclear: Whether a separate `pkg/ratelimit` package is worth the import path for ~30 lines.
   - Recommendation: Put `TokenBucketLimiter` in `pkg/types` alongside the `RequestRateLimiter` interface. It's small enough and keeps the dependency graph flat. If it grows, extract later.

3. **How does `TlsValidator` receive a Logger when it's created?**
   - What we know: `TlsValidator` is created in the connection setup path. The `client` struct has `c.config.Logger`.
   - What's unclear: The exact wiring path from client config to TlsValidator.
   - Recommendation: Add `SetLogger(logger types.Logger)` method to `TlsValidator`. The caller (likely `ConnectionManager` or `client`) calls it after construction. This avoids changing the `NewTlsValidator` signature.

## Environment Availability

> This phase is purely code/config changes with no external dependencies beyond the Go toolchain already in use.

Step 2.6: SKIPPED (no external dependencies identified)

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Build + Test | Yes | go1.26.1 darwin/arm64 | -- |
| gorilla/websocket | Transport (unchanged) | Yes | v1.5.3 | -- |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `testify` (if present) |
| Config file | None -- Go convention |
| Quick run command | `go test ./pkg/... -count=1 -v` |
| Full suite command | `go test ./... -race -count=1` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FOUND-01 | Rate limiter blocks requests when `Allow()` returns false | unit | `go test ./pkg/... -run TestRateLimit -v` | No -- Wave 0 |
| FOUND-01 | `WithRateLimit()` option sets limiter on config | unit | `go test ./pkg/... -run TestWithRateLimit -v` | No -- Wave 0 |
| FOUND-02 | Reconnect stops after MaxRetries attempts | unit | `go test ./pkg/managers/... -run TestMaxRetries -v` | No -- Wave 0 |
| FOUND-02 | MaxRetries=0 falls back to MaxAttempts behavior | unit | `go test ./pkg/managers/... -run TestMaxRetriesFallback -v` | No -- Wave 0 |
| FOUND-03 | CRL stub has explicit v1 limitation comment | manual | `grep -c "v1 limitation" pkg/connection/tls.go` | N/A |
| FOUND-04 | SendRequest returns ErrTooManyPendingRequests when map full | unit | `go test ./pkg/managers/... -run TestPendingLimit -v` | No -- Wave 0 |
| FOUND-04 | Pending limit of 0 means unlimited | unit | `go test ./pkg/managers/... -run TestPendingLimitUnlimited -v` | No -- Wave 0 |
| FOUND-05 | Warn logged when InsecureSkipVerify=true | unit | `go test ./pkg/connection/... -run TestInsecureSkipVerifyWarning -v` | No -- Wave 0 |
| FOUND-05 | No warning when InsecureSkipVerify=false | unit | `go test ./pkg/connection/... -run TestNoInsecureWarning -v` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./pkg/... -count=1`
- **Per wave merge:** `go test ./... -race -count=1`
- **Phase gate:** Full suite green + `go vet ./...` clean

### Wave 0 Gaps
- [ ] `pkg/managers/request_test.go` -- add tests for FOUND-01 (rate limit) and FOUND-04 (pending limit)
- [ ] `pkg/managers/reconnect_test.go` -- add tests for FOUND-02 (MaxRetries)
- [ ] `pkg/connection/tls_test.go` -- add tests for FOUND-05 (InsecureSkipVerify warning)
- [ ] `pkg/types/types_test.go` -- add tests for `RequestRateLimiter` interface compliance
- [ ] `pkg/client_test.go` -- add tests for `WithRateLimit` option and `SendRequest` rate check

## Sources

### Primary (HIGH confidence)
- Source code analysis: `pkg/managers/request.go`, `pkg/managers/reconnect.go`, `pkg/connection/tls.go`, `pkg/types/types.go`, `pkg/types/errors.go`, `pkg/client.go`
- Go stdlib `crypto/x509` package documentation (certificate types, CRLDistributionPoints, OCSPServer)
- Go stdlib `sync` package (Mutex patterns)

### Secondary (MEDIUM confidence)
- Go Wiki: "Rate limiters in Go" -- `Allow() bool` as canonical interface pattern
- Go blog: "Keeping Your Modules Compatible" -- backward-compatible struct field additions
- `golang.org/x/time/rate` v0.15.0 documentation -- token bucket API reference (researched but not used)

### Tertiary (LOW confidence)
- None -- all findings verified against source code

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new external dependencies; all stdlib or internal
- Architecture: HIGH -- insertion points identified in source code with line numbers
- Pitfalls: HIGH -- derived from codebase analysis (lock ordering, Logger absence, field precedence)
- Rate limiter design: HIGH -- follows Go community standard `Allow() bool` pattern
- Backward compatibility: HIGH -- follows official Go blog guidance on struct field additions

**Research date:** 2026-03-28
**Valid until:** 2026-04-28 (stable -- no fast-moving dependencies)
