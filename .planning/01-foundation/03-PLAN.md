---
phase: 01-foundation
plan: 03
type: tdd
wave: 2
depends_on:
  - 01-foundation-01
files_modified:
  - pkg/managers/reconnect.go
  - pkg/managers/reconnect_test.go
  - pkg/connection/tls.go
  - pkg/connection/tls_test.go
autonomous: true
requirements:
  - FOUND-02
  - FOUND-03
  - FOUND-05

must_haves:
  truths:
    - "ReconnectManager stops after MaxRetries attempts and calls onReconnectFailed"
    - "MaxRetries takes precedence over MaxAttempts when both set"
    - "MaxRetries=0 falls back to MaxAttempts behavior (backward compatible)"
    - "CheckCertificateRevocation has explicit v1 limitation comment block"
    - "TlsValidator logs WARN when InsecureSkipVerify=true and logger is set"
    - "TlsValidator does not log when InsecureSkipVerify=false"
    - "TlsValidator does not panic when logger is nil"
  artifacts:
    - path: "pkg/managers/reconnect.go"
      provides: "MaxRetries check with precedence over MaxAttempts in run() loop"
      contains: "MaxRetries"
    - path: "pkg/managers/reconnect_test.go"
      provides: "Tests for MaxRetries behavior and fallback"
      min_lines: 40
    - path: "pkg/connection/tls.go"
      provides: "SetLogger method, InsecureSkipVerify WARN log, CRL stub documentation"
      contains: "SetLogger"
    - path: "pkg/connection/tls_test.go"
      provides: "Tests for InsecureSkipVerify warning and CRL stub documentation"
  key_links:
    - from: "pkg/managers/reconnect.go"
      to: "pkg/types/types.go"
      via: "rm.config.MaxRetries field access"
      pattern: "MaxRetries"
    - from: "pkg/managers/reconnect.go"
      to: "pkg/types/errors.go"
      via: "onReconnectFailed callback with error indicating max retries"
      pattern: "onReconnectFailed"
    - from: "pkg/connection/tls.go"
      to: "pkg/types/logger.go"
      via: "Logger interface field and Warn method"
      pattern: "logger\\.Warn"
---

<objective>
Implement retry budget with MaxRetries field and TLS hardening (InsecureSkipVerify warning + CRL stub documentation).

Purpose: Bound reconnection attempts to prevent infinite retry loops (FOUND-02), make operators aware of insecure TLS configuration (FOUND-05), and document the CRL checking limitation (FOUND-03).

Output: MaxRetries enforcement in ReconnectManager, InsecureSkipVerify warning logging in TlsValidator, and documented CRL stub.
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

From pkg/types/types.go (new additions from Plan 01):
```go
type ReconnectConfig struct {
    MaxAttempts       int           // existing: 0 = infinite
    MaxRetries        int           // NEW: 0 = use MaxAttempts; takes precedence when > 0
    InitialDelay      time.Duration
    MaxDelay          time.Duration
    BackoffMultiplier float64
}

func DefaultReconnectConfig() ReconnectConfig {
    // MaxRetries: 10, MaxAttempts: 0
}
```

From pkg/types/errors.go (new additions from Plan 01):
```go
var ErrMaxRetriesExceeded = errors.New("max retries exceeded")
```

From pkg/types/logger.go (existing):
```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```
</interfaces>

From pkg/managers/reconnect.go (existing, critical section):
```go
// In run() method, line 120 - current MaxAttempts check:
if rm.config.MaxAttempts > 0 && attempt >= rm.config.MaxAttempts {
    return
}

// Currently does NOT call onReconnectFailed when MaxAttempts reached!
// Plan must ADD this callback AND add MaxRetries precedence.
```

From pkg/connection/tls.go (existing):
```go
type TlsValidator struct {
    config *TLSConfig
    // NO logger field currently
}

func NewTlsValidator(config *TLSConfig) *TlsValidator {
    return &TlsValidator{config: config}
}

// In GetTLSConfig(), line 114-117 - builds config but NO warning:
config := &tls.Config{
    InsecureSkipVerify: v.config.InsecureSkipVerify,
    ServerName:         v.config.ServerName,
}

// CheckCertificateRevocation at line 250-280 - stub without clear limitation docs
// Current TODO comment exists but doesn't explain v1 scope limitation
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Add MaxRetries enforcement to ReconnectManager</name>
  <files>pkg/managers/reconnect.go, pkg/managers/reconnect_test.go</files>
  <read_first>
    - pkg/managers/reconnect.go (full file -- especially run() method lines 81-141)
    - pkg/managers/reconnect_test.go (existing test patterns)
    - pkg/types/types.go (for MaxRetries field added in Plan 01)
    - pkg/types/errors.go (for ErrMaxRetriesExceeded from Plan 01)
  </read_first>
  <acceptance_criteria>
    - pkg/managers/reconnect.go run() method checks MaxRetries BEFORE MaxAttempts
    - Precedence logic: `if maxRetries > 0 { use MaxRetries } else { use MaxAttempts }`
    - When limit reached, onReconnectFailed is called with error containing "max retries exceeded"
    - When limit reached, run() returns (stops loop)
    - MaxRetries=0 falls back to MaxAttempts behavior (backward compatible)
    - go test ./pkg/managers/ -run "TestReconnectManager_MaxRetries" -v exits 0
    - go test -race ./pkg/managers/ exits 0
  </acceptance_criteria>
  <behavior>
    - Test 1: MaxRetries=3, onReconnect always fails, stops after exactly 3 attempts
    - Test 2: MaxRetries=0, MaxAttempts=2, falls back to MaxAttempts behavior (stops after 2)
    - Test 3: MaxRetries=0, MaxAttempts=0, reconnects continue indefinitely (until stopped)
    - Test 4: MaxRetries=3, MaxAttempts=10, MaxRetries takes precedence (stops after 3, not 10)
    - Test 5: onReconnectFailed callback is called with error when MaxRetries exceeded
  </behavior>
  <action>
    Per FOUND-02 and CONTEXT.md locked decisions (MaxRetries added to ReconnectConfig, default 10, precedence over MaxAttempts):

    1. Replace the MaxAttempts check in run() method. Currently at line 120:
    ```go
    if rm.config.MaxAttempts > 0 && attempt >= rm.config.MaxAttempts {
        return
    }
    ```

    Replace with MaxRetries-first precedence logic (per GA-4):
    ```go
    // Check retry budget (FOUND-02)
    // MaxRetries takes precedence over MaxAttempts (GA-4)
    maxRetries := rm.config.MaxRetries
    if maxRetries <= 0 {
        // Fall back to MaxAttempts for backward compatibility
        maxRetries = rm.config.MaxAttempts
    }
    if maxRetries > 0 && attempt >= maxRetries {
        if onReconnectFailed != nil {
            onReconnectFailed(fmt.Errorf("max retries exceeded: %d", maxRetries))
        }
        return
    }
    ```

    IMPORTANT: This block goes in the same position as the old MaxAttempts check (after the onReconnectFailed callback for the individual attempt failure, around line 119). The variable names onReconnect and onReconnectFailed are already read under the lock earlier in the same select case (lines 102-105).

    2. Add "fmt" to the import list in reconnect.go (needed for fmt.Errorf). Check if it is already imported -- currently imports are: context, sync, time, and types. Add "fmt" to the import block.

    3. Add tests to pkg/managers/reconnect_test.go:

    - TestReconnectManager_MaxRetries_StopsAtLimit:
      ```go
      config := DefaultReconnectConfig()
      config.MaxRetries = 3
      config.MaxAttempts = 0 // disabled
      config.InitialDelay = 10 * time.Millisecond
      // Count attempts, always fail, verify stops at exactly 3
      ```

    - TestReconnectManager_MaxRetries_FallbackToMaxAttempts:
      ```go
      config := DefaultReconnectConfig()
      config.MaxRetries = 0  // fall back
      config.MaxAttempts = 2
      config.InitialDelay = 10 * time.Millisecond
      // Verify stops at 2 attempts
      ```

    - TestReconnectManager_MaxRetries_TakesPrecedence:
      ```go
      config := DefaultReconnectConfig()
      config.MaxRetries = 3
      config.MaxAttempts = 100  // should NOT reach this
      config.InitialDelay = 10 * time.Millisecond
      // Verify stops at 3, not 100
      ```

    - TestReconnectManager_MaxRetries_CallsFailedCallback:
      ```go
      // Verify onReconnectFailed is called with error containing "max retries exceeded"
      // when MaxRetries limit is reached
      ```

    Follow existing test patterns from reconnect_test.go: use sync.Mutex for shared state, time.Sleep for timing, and check attempt counts.
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/managers/ -run "TestReconnectManager_MaxRetries" -v -race -count=1</automated>
  </verify>
  <done>
    - ReconnectManager.run() checks MaxRetries with precedence over MaxAttempts
    - Stops after MaxRetries attempts and calls onReconnectFailed with descriptive error
    - MaxRetries=0 falls back to MaxAttempts (backward compatible)
    - All existing reconnect tests still pass
    - New tests pass with -race flag
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Add InsecureSkipVerify warning and CRL stub documentation to TlsValidator</name>
  <files>pkg/connection/tls.go, pkg/connection/tls_test.go</files>
  <read_first>
    - pkg/connection/tls.go (full file -- TlsValidator struct at line 41, GetTLSConfig at line 103, CheckCertificateRevocation at line 250)
    - pkg/connection/tls_test.go (existing test patterns)
    - pkg/types/logger.go (for Logger interface: Debug, Info, Warn, Error methods)
  </read_first>
  <acceptance_criteria>
    - pkg/connection/tls.go TlsValidator struct has "logger types.Logger" field
    - pkg/connection/tls.go has "func (v *TlsValidator) SetLogger(logger types.Logger)" method
    - pkg/connection/tls.go GetTLSConfig() calls v.logger.Warn() when InsecureSkipVerify=true and logger is not nil
    - Warn message contains "InsecureSkipVerify" and "not recommended for production"
    - pkg/connection/tls.go CheckCertificateRevocation has comment block containing "v1 limitation"
    - pkg/connection/tls.go imports "github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
    - go test ./pkg/connection/ -run "TestTlsValidator_InsecureSkipVerifyWarning|TestCheckCertificateRevocation" -v exits 0
    - go test -race ./pkg/connection/ exits 0
  </acceptance_criteria>
  <behavior>
    - Test 1: GetTLSConfig with InsecureSkipVerify=true and logger set calls Warn with message containing "InsecureSkipVerify"
    - Test 2: GetTLSConfig with InsecureSkipVerify=false does NOT call Warn
    - Test 3: GetTLSConfig with InsecureSkipVerify=true but nil logger does NOT panic
    - Test 4: CheckCertificateRevocation with valid cert returns nil (stub behavior unchanged)
  </behavior>
  <action>
    Per FOUND-05, FOUND-03, and CONTEXT.md decisions (GA-5: WARN level, SetLogger method, stub with explicit comment):

    1. Add types import to pkg/connection/tls.go. Currently imports do NOT include the types package. Add to the import block:
    ```go
    import (
        // ... existing imports ...
        "github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
    )
    ```

    2. Add logger field to TlsValidator struct (line 41-43):
    ```go
    type TlsValidator struct {
        config *TLSConfig     // TLS configuration to validate
        logger types.Logger   // Optional logger for warnings (FOUND-05)
    }
    ```

    3. Add SetLogger method after NewTlsValidator (after line 68):
    ```go
    // SetLogger sets the logger for the TLS validator.
    // When set, the validator logs warnings for insecure configurations (FOUND-05).
    func (v *TlsValidator) SetLogger(logger types.Logger) {
        v.logger = logger
    }
    ```

    4. Add InsecureSkipVerify warning in GetTLSConfig(). After building the config (after line 117 `config := &tls.Config{...}`), add:
    ```go
    // Warn when InsecureSkipVerify is enabled (FOUND-05)
    if v.config.InsecureSkipVerify && v.logger != nil {
        v.logger.Warn("TLS InsecureSkipVerify is enabled -- server certificate verification disabled; not recommended for production use")
    }
    ```

    IMPORTANT: This goes BEFORE the client certificate loading section (line 119+). The check must be: InsecureSkipVerify is true AND logger is not nil. Per GA-5, use Warn level.

    5. Update CheckCertificateRevocation comment (replace lines 250-257 TODO comment with explicit v1 limitation documentation):
    ```go
    // CheckCertificateRevocation checks if a certificate has been revoked.
    //
    // v1 limitation: This is a STUB implementation for v1.0. No actual CRL fetching
    // or OCSP checking is performed. The function always returns nil (no revocation detected).
    //
    // A production implementation would require:
    //   - Fetching CRLs from cert.CRLDistributionPoints (HTTP/HTTPS)
    //   - Parsing CRLs using x509.ParseRevocationList (Go 1.21+)
    //   - Performing OCSP checks against cert.OCSPServer endpoints
    //   - Caching revocation status with TTL to avoid repeated network calls
    //   - Handling timeouts and failures gracefully (fail-open vs fail-closed policy)
    //
    // See: REQUIREMENTS.md FOUND-03
    //
    // This stub exists so callers can integrate revocation checking into their validation
    // flow without API changes when a real implementation is added in a future release.
    ```

    The function body remains unchanged (returns nil for all cases).

    6. Add tests to pkg/connection/tls_test.go:

    - TestTlsValidator_InsecureSkipVerifyWarning_WithLogger:
      Create a bytes.Buffer, create DefaultLoggerWithWriter(buf), create TlsValidator with InsecureSkipVerify=true, call SetLogger, call GetTLSConfig, verify buf.String() contains "InsecureSkipVerify" and "WARN".

    - TestTlsValidator_InsecureSkipVerifyWarning_NotEnabled:
      Same setup but InsecureSkipVerify=false. Verify buf.String() does NOT contain "InsecureSkipVerify".

    - TestTlsValidator_InsecureSkipVerifyWarning_NilLogger:
      TlsValidator with InsecureSkipVerify=true but NO SetLogger call. Verify GetTLSConfig does not panic and returns valid config.

    - TestCheckCertificateRevocation_WithCRL:
      Existing test covers this. Add a check that the function has the v1 limitation comment:
      This is a documentation-only change, the test is just verifying the stub returns nil.
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/connection/ -run "TestTlsValidator_InsecureSkipVerifyWarning" -v -race -count=1 && grep -c "v1 limitation" pkg/connection/tls.go</automated>
  </verify>
  <done>
    - TlsValidator has logger field and SetLogger method
    - GetTLSConfig logs WARN when InsecureSkipVerify=true and logger is set
    - GetTLSConfig does not panic when logger is nil
    - GetTLSConfig does not warn when InsecureSkipVerify=false
    - CheckCertificateRevocation has explicit v1 limitation documentation
    - All existing TLS tests still pass
    - New tests pass with -race flag
  </done>
</task>

</tasks>

<verification>
go test ./pkg/managers/ ./pkg/connection/ -race -count=1
go vet ./pkg/managers/ ./pkg/connection/
grep -c "MaxRetries" pkg/managers/reconnect.go  # should be >= 3
grep -c "SetLogger" pkg/connection/tls.go  # should be >= 2
grep -c "v1 limitation" pkg/connection/tls.go  # should be >= 1
grep -c "InsecureSkipVerify" pkg/connection/tls.go  # should be >= 4
</verification>

<success_criteria>
1. ReconnectManager stops after MaxRetries attempts
2. MaxRetries takes precedence over MaxAttempts
3. MaxRetries=0 falls back to MaxAttempts (backward compatible)
4. onReconnectFailed callback called when limit reached
5. InsecureSkipVerify=true triggers WARN log when logger is set
6. No warning when InsecureSkipVerify=false
7. No panic when logger is nil
8. CRL stub has explicit v1 limitation documentation
9. All tests pass with -race flag
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-03-SUMMARY.md`
</output>
