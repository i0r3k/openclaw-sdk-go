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
  - pkg/managers/connection.go
  - pkg/managers/connection_test.go
  - pkg/transport/websocket.go
  - pkg/transport/websocket_test.go
  - pkg/client.go
  - pkg/client_test.go
autonomous: true
requirements:
  - FOUND-02
  - FOUND-03
  - FOUND-05

must_haves:
  truths:
    - "ReconnectManager stops after MaxRetries attempts and calls onReconnectFailed with typed error"
    - "MaxRetries takes precedence over MaxAttempts with documented zero/negative semantics"
    - "MaxRetries=0 falls back to MaxAttempts behavior (backward compatible)"
    - "ReconnectManager does NOT start reconnect loop immediately after healthy initial Connect"
    - "CheckCertificateRevocation has explicit v1 limitation comment block"
    - "InsecureSkipVerify warning logged through proper Logger at connection time, NOT to stderr"
    - "TLSConfig wired through ConnectionManager.Connect -> transport.Dial actual dial path"
    - "Logger threaded from client config through ConnectionManager to transport Dial"
  artifacts:
    - path: "pkg/managers/reconnect.go"
      provides: "MaxRetries check with precedence over MaxAttempts, only triggers on disconnect"
      contains: "MaxRetries"
    - path: "pkg/managers/reconnect_test.go"
      provides: "Tests for MaxRetries behavior, fallback, and no-reconnect-after-healthy-connect"
      min_lines: 60
    - path: "pkg/connection/tls.go"
      provides: "SetLogger method, InsecureSkipVerify WARN log, CRL stub documentation"
      contains: "SetLogger"
    - path: "pkg/connection/tls_test.go"
      provides: "Tests for InsecureSkipVerify warning and CRL stub documentation"
    - path: "pkg/managers/connection.go"
      provides: "TLSConfig wiring to transport.Dial, Logger pass-through"
      contains: "Dial.*config\|TLSConfig"
    - path: "pkg/transport/websocket.go"
      provides: "Logger parameter in Dial, stderr replaced with Logger.Warn"
      contains: "Logger"
    - path: "pkg/client.go"
      provides: "Reconnect triggered by disconnect event, not by Connect completion"
      contains: "OnDisconnect\|reconnect.*Start"
  key_links:
    - from: "pkg/client.go"
      to: "pkg/managers/reconnect.go"
      via: "reconnect.Start() called on disconnect event, not after Connect()"
      pattern: "reconnect\\.Start"
    - from: "pkg/managers/connection.go"
      to: "pkg/transport/websocket.go"
      via: "Dial(ctx, url, header, config) -- TLSConfig passed through"
      pattern: "transport\\.Dial"
    - from: "pkg/transport/websocket.go"
      to: "pkg/connection/tls.go"
      via: "TlsValidator with Logger, GetTLSConfig with warning"
      pattern: "TlsValidator"
    - from: "pkg/managers/reconnect.go"
      to: "pkg/types/errors.go"
      via: "onReconnectFailed callback with NewMaxRetriesExceededError"
      pattern: "MaxRetriesExceeded"
---

<objective>
Implement retry budget with MaxRetries field, TLS hardening (InsecureSkipVerify warning + CRL stub), and wire TLS/Logger through the actual connection path.

Purpose: Bound reconnection attempts (FOUND-02), make operators aware of insecure TLS configuration through proper logging (FOUND-05), document CRL limitation (FOUND-03), and ensure these features work in the actual dial path -- not just in isolated test files.

Output: MaxRetries enforcement in ReconnectManager, TLSConfig wired through ConnectionManager to transport.Dial, InsecureSkipVerify warning through Logger (not stderr), reconnect triggered only on disconnect (not after healthy connect), documented CRL stub.
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
    MaxAttempts       int           // Deprecated, 0 = infinite
    MaxRetries        int           // NEW: >0 takes precedence, 0 = fallback to MaxAttempts
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

type MaxRetriesExceededError struct { *BaseError }
func NewMaxRetriesExceededError(maxRetries int) *MaxRetriesExceededError
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

From pkg/managers/reconnect.go (existing, full run() method):
```go
func (rm *ReconnectManager) run() {
    defer rm.wg.Done()
    prevDelay := time.Duration(0)
    delay := rm.config.InitialDelay
    attempt := 0

    for {
        attempt++
        timer := time.NewTimer(delay)
        select {
        case <-rm.ctx.Done():
            timer.Stop()
            return
        case <-rm.stopped:
            timer.Stop()
            return
        case <-timer.C:
            rm.mu.Lock()
            onReconnect := rm.onReconnect
            onReconnectFailed := rm.onReconnectFailed
            rm.mu.Unlock()

            if onReconnect == nil {
                return
            }

            err := onReconnect()
            if err == nil {
                return  // PROBLEM: returns immediately on success
            }
            if onReconnectFailed != nil {
                onReconnectFailed(err)
            }

            if rm.config.MaxAttempts > 0 && attempt >= rm.config.MaxAttempts {
                return
            }
            // ... backoff calculation ...
        }
    }
}
```

From pkg/managers/reconnect.go -- Start() at line 74:
```go
func (rm *ReconnectManager) Start() {
    rm.wg.Add(1)
    go rm.run()
}
```

From pkg/client.go -- Connect() at line 508:
```go
// At end of Connect(), after successful connection + handshake:
if c.config.ReconnectEnabled && c.managers.reconnect != nil {
    c.managers.reconnect.Start()  // PROBLEM: starts reconnect immediately after healthy connect
}
```

From pkg/managers/connection.go -- Connect() at line 73:
```go
t, err := transport.Dial(ctx, cm.config.URL, header, nil)
//                                                        ^^^^ nil config -- TLSConfig NOT wired!
```

From pkg/transport/websocket.go -- Dial() at line 109-110:
```go
if config.TLSConfig.InsecureSkipVerify {
    fmt.Fprintf(os.Stderr, "WARNING: InsecureSkipVerify...")  // PROBLEM: uses stderr, not Logger
}
```

From pkg/transport/websocket.go -- Dial() signature at line 82:
```go
func Dial(ctx context.Context, url string, header http.Header, config *WebSocketConfig) (*WebSocketTransport, error)
```

From pkg/transport/websocket.go -- WebSocketConfig struct:
```go
type WebSocketConfig struct {
    ReadBufferSize    int
    WriteBufferSize   int
    HandshakeTimeout  time.Duration
    ReadTimeout       time.Duration
    WriteTimeout      time.Duration
    ChannelBufferSize int
    TLSConfig         *TLSConfig
}
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Fix reconnect triggering and add MaxRetries enforcement</name>
  <files>pkg/managers/reconnect.go, pkg/managers/reconnect_test.go, pkg/client.go, pkg/client_test.go</files>
  <read_first>
    - pkg/managers/reconnect.go (full file -- especially run() method lines 81-141, Start() line 74)
    - pkg/managers/reconnect_test.go (existing test patterns)
    - pkg/client.go (Connect() lines 452-511, Disconnect() lines 515+, reconnect setup lines 431-441)
    - pkg/client_test.go (existing client tests)
    - pkg/types/types.go (for MaxRetries field and precedence docs from Plan 01)
    - pkg/types/errors.go (for NewMaxRetriesExceededError from Plan 01)
  </read_first>
  <acceptance_criteria>
    - pkg/managers/reconnect.go run() method checks MaxRetries BEFORE MaxAttempts using precedence rules from Plan 01
    - Precedence logic: `maxRetries := rm.config.MaxRetries; if maxRetries <= 0 { maxRetries = rm.config.MaxAttempts }; if maxRetries < 0 { maxRetries = 0 }`
    - When limit reached, onReconnectFailed called with NewMaxRetriesExceededError(maxRetries)
    - When limit reached, run() returns (stops loop)
    - MaxRetries=0 falls back to MaxAttempts (backward compatible)
    - Both MaxRetries=0 and MaxAttempts=0 means infinite (backward compatible)
    - Negative values treated as 0
    - pkg/client.go Connect() does NOT call reconnect.Start() after healthy connect
    - pkg/client.go has disconnect-event handler that starts reconnect manager
    - go test ./pkg/managers/ -run "TestReconnectManager_MaxRetries|TestReconnectManager_NoStartOnHealthy" -v exits 0
    - go test -race ./pkg/managers/ exits 0
  </acceptance_criteria>
  <behavior>
    - Test 1: MaxRetries=3, onReconnect always fails, stops after exactly 3 attempts
    - Test 2: MaxRetries=0, MaxAttempts=2, falls back to MaxAttempts (stops after 2)
    - Test 3: MaxRetries=0, MaxAttempts=0, reconnects continue (until ctx cancelled)
    - Test 4: MaxRetries=3, MaxAttempts=10, MaxRetries takes precedence (stops after 3)
    - Test 5: MaxRetries=-1 treated as 0, falls back to MaxAttempts
    - Test 6: onReconnectFailed called with MaxRetriesExceededError wrapping ErrMaxRetriesExceeded
    - Test 7: After healthy Connect(), reconnect.Start() is NOT called
    - Test 8: After disconnect event, reconnect.Start() IS called (if ReconnectEnabled)
  </behavior>
  <action>
    Per FOUND-02 and CONTEXT.md locked decisions. Addresses review concerns:
    - [HIGH] Reconnect starts immediately after healthy Connect -- fix so it only starts on disconnect
    - [HIGH] MaxRetries precedence over MaxAttempts needs exact nil/zero/negative semantics
    - [MEDIUM] "MaxRetries takes precedence" too vague without exact semantics

    STEP 1: Fix MaxRetries enforcement in run() method.

    Replace the MaxAttempts check in run() (line 120-122) with MaxRetries-first precedence:

    ```go
    // Check retry budget (FOUND-02)
    // Precedence: MaxRetries > 0 wins over MaxAttempts
    // MaxRetries <= 0 falls back to MaxAttempts
    // Both <= 0 means infinite
    maxRetries := rm.config.MaxRetries
    if maxRetries <= 0 {
        maxRetries = rm.config.MaxAttempts
    }
    if maxRetries > 0 && attempt >= maxRetries {
        if onReconnectFailed != nil {
            onReconnectFailed(types.NewMaxRetriesExceededError(maxRetries))
        }
        return
    }
    ```

    This block goes in the SAME position as the old MaxAttempts check (after the onReconnectFailed callback for the individual attempt failure, line 119). The `onReconnectFailed` variable is already read under the lock earlier in the same select case (lines 102-105).

    Add "fmt" to imports if not present -- actually, NewMaxRetriesExceededError handles formatting internally, so no fmt import needed. Just need types import (already present at line 15).

    STEP 2: Fix reconnect triggering in client.

    Current problem: pkg/client.go line 508 calls `c.managers.reconnect.Start()` at the end of Connect(), right after a healthy connection. This means the reconnect loop starts immediately, and the first timer tick fires an onReconnect attempt even though the connection is healthy.

    Fix: Remove `c.managers.reconnect.Start()` from the end of Connect(). Instead, start reconnect on disconnect event.

    In Connect() (around line 508), REMOVE:
    ```go
    if c.config.ReconnectEnabled && c.managers.reconnect != nil {
        c.managers.reconnect.Start()
    }
    ```

    Instead, the reconnect manager should be started when a disconnect event occurs. There are two approaches:

    Option A (recommended): Subscribe to StateDisconnected transition and start reconnect:

    In NewClient(), after setting up the reconnect manager (around line 441), add a disconnect handler:
    ```go
    // Start reconnect on disconnect, not after initial connect (FOUND-02 fix)
    if c.config.ReconnectEnabled {
        c.managers.reconnect.SetOnReconnect(func() error {
            return c.managers.connection.Reconnect(ctx)
        })
        c.managers.reconnect.SetOnReconnectFailed(func(err error) {
            c.managers.event.Emit(Event{
                Type:      EventError,
                Err:       err,
                Timestamp: time.Now(),
            })
        })

        // Subscribe to disconnect events to trigger reconnect
        c.managers.event.Subscribe(EventDisconnect, func(event Event) {
            c.mu.Lock()
            reconnect := c.managers.reconnect
            c.mu.Unlock()
            if reconnect != nil {
                reconnect.Start()
            }
        })
    }
    ```

    CAUTION: Read the existing event subscription pattern and disconnect event emission in the codebase first. Check if EventDisconnect is already emitted by ConnectionManager.Disconnect() or by the state machine transition. Also check if there is already a disconnect handler -- we do not want to add a duplicate. Read pkg/managers/connection.go Disconnect() method and the state machine event emissions.

    If the state machine already emits events on state transitions, and there is already a Disconnect event, this approach works directly. If not, we may need to add an EventDisconnect emission in ConnectionManager.Disconnect().

    IMPORTANT: ReconnectManager.Start() spawns a new goroutine each call. We need to ensure Start() is idempotent or guard against calling it multiple times. Check if Start() already has this protection -- currently it does NOT (it always does rm.wg.Add(1) and go rm.run()). Add a guard:
    ```go
    func (rm *ReconnectManager) Start() {
        rm.mu.Lock()
        defer rm.mu.Unlock()
        // Prevent starting if already running or stopped
        select {
        case <-rm.stopped:
            return // already stopped
        default:
        }
        rm.wg.Add(1)
        go rm.run()
    }
    ```

    Actually, a cleaner approach: use an `atomic.Bool` or check `rm.stopped` channel. But the simplest fix that doesn't change the API: use `sync.Once` for start, or add a `running` flag. Read the code carefully to choose the minimal change.

    STEP 3: Add tests.

    To pkg/managers/reconnect_test.go:
    - TestReconnectManager_MaxRetries_StopsAtLimit
    - TestReconnectManager_MaxRetries_FallbackToMaxAttempts
    - TestReconnectManager_MaxRetries_TakesPrecedence
    - TestReconnectManager_MaxRetries_NegativeTreatedAsZero
    - TestReconnectManager_MaxRetries_CallsFailedWithTypedError: verify errors.Is(err, types.ErrMaxRetriesExceeded) on the callback error
    - TestReconnectManager_BothZeroInfinite: both MaxRetries=0 and MaxAttempts=0, loop continues until context cancelled

    Follow existing test patterns: use sync.Mutex for shared state, time.Sleep for timing, context cancellation for stopping.

    To pkg/client_test.go (or reconnect_test.go if client test is not feasible without full integration):
    - Test that verifies Connect() does NOT call Start() on reconnect manager (check goroutine count or mock)
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/managers/ -run "TestReconnectManager_MaxRetries" -v -race -count=1 && go test ./pkg/ -run "TestConnect.*Reconnect" -v -race -count=1</automated>
  </verify>
  <done>
    - ReconnectManager.run() checks MaxRetries with precedence over MaxAttempts
    - MaxRetries=0 falls back to MaxAttempts, both zero = infinite
    - Negative values treated as zero
    - Stops after MaxRetries attempts and calls onReconnectFailed with NewMaxRetriesExceededError
    - errors.Is(callbackErr, ErrMaxRetriesExceeded) returns true
    - Connect() does NOT start reconnect after healthy connection
    - Reconnect starts on disconnect event (not after initial Connect)
    - Start() has guard against multiple invocations
    - All existing tests still pass
    - New tests pass with -race flag
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Wire TLS/Logger through actual connection path and add InsecureSkipVerify warning</name>
  <files>pkg/transport/websocket.go, pkg/transport/websocket_test.go, pkg/managers/connection.go, pkg/managers/connection_test.go, pkg/connection/tls.go, pkg/connection/tls_test.go</files>
  <read_first>
    - pkg/transport/websocket.go (full file -- Dial function lines 82-160, WebSocketConfig struct, TLSConfig struct, stderr usage at line 109-111)
    - pkg/managers/connection.go (full file -- Connect() at line 54, especially line 73 transport.Dial call with nil config)
    - pkg/connection/tls.go (full file -- TlsValidator struct at line 39, GetTLSConfig at line 103, CheckCertificateRevocation stub)
    - pkg/connection/tls_test.go (existing test patterns)
    - pkg/types/logger.go (for Logger interface)
  </read_first>
  <acceptance_criteria>
    - pkg/transport/websocket.go Dial() function accepts Logger in WebSocketConfig (or new parameter) and uses it instead of fmt.Fprintf(os.Stderr, ...)
    - pkg/managers/connection.go Connect() passes TLSConfig and Logger through to transport.Dial instead of nil
    - pkg/connection/tls.go TlsValidator struct has "logger types.Logger" field
    - pkg/connection/tls.go has "func (v *TlsValidator) SetLogger(logger types.Logger)" method
    - pkg/connection/tls.go GetTLSConfig() calls v.logger.Warn() when InsecureSkipVerify=true and logger is not nil
    - pkg/transport/websocket.go does NOT use fmt.Fprintf(os.Stderr, ...) for InsecureSkipVerify warning -- uses Logger
    - pkg/connection/tls.go CheckCertificateRevocation has comment block containing "v1 limitation"
    - go test ./pkg/transport/ ./pkg/connection/ ./pkg/managers/ -run "TestInsecureSkipVerify|TestDial.*TLS|TestConnect.*TLS" -v exits 0
    - go test -race ./pkg/transport/ ./pkg/connection/ exits 0
  </acceptance_criteria>
  <behavior>
    - Test 1: GetTLSConfig with InsecureSkipVerify=true and logger set calls Warn with "InsecureSkipVerify"
    - Test 2: GetTLSConfig with InsecureSkipVerify=false does NOT call Warn
    - Test 3: GetTLSConfig with InsecureSkipVerify=true but nil logger does NOT panic
    - Test 4: Dial with TLSConfig.InsecureSkipVerify=true and Logger logs Warn through Logger (not stderr)
    - Test 5: Dial with TLSConfig.InsecureSkipVerify=true and nil Logger still works (no panic)
    - Test 6: ConnectionManager.Connect passes TLSConfig from its config to transport.Dial
    - Test 7: CheckCertificateRevocation returns nil (stub behavior unchanged)
  </behavior>
  <action>
    Per FOUND-05, FOUND-03, and CONTEXT.md decisions (GA-5: WARN level, SetLogger method, stub with explicit comment).

    Addresses review concerns:
    - [HIGH] TLSConfig not wired into connection path -- ConnectionManager.Connect calls transport.Dial(..., nil)
    - [HIGH] Insecure warning goes to stderr in transport -- need proper Logger wiring through dial path
    - [HIGH] Missing files: client.go, connection.go, websocket.go MUST be in scope

    This task wires TLS/Logger through the ENTIRE connection path: client -> ConnectionManager -> transport.Dial -> TlsValidator.

    STEP 1: Add Logger to WebSocketConfig in pkg/transport/websocket.go.

    The WebSocketConfig struct needs a Logger field so Dial() can use it:

    ```go
    type WebSocketConfig struct {
        ReadBufferSize    int
        WriteBufferSize   int
        HandshakeTimeout  time.Duration
        ReadTimeout       time.Duration
        WriteTimeout      time.Duration
        ChannelBufferSize int
        TLSConfig         *TLSConfig
        Logger            types.Logger // Optional logger for warnings (FOUND-05)
    }
    ```

    Add types import to websocket.go:
    ```go
    "github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
    ```

    STEP 2: Replace stderr with Logger in Dial() function.

    In Dial() function, replace the stderr warning (lines 109-111):
    ```go
    // OLD:
    if config.TLSConfig.InsecureSkipVerify {
        fmt.Fprintf(os.Stderr, "WARNING: InsecureSkipVerify...")
    }

    // NEW:
    if config.TLSConfig.InsecureSkipVerify {
        if config.Logger != nil {
            config.Logger.Warn("TLS InsecureSkipVerify is enabled -- server certificate verification disabled; not recommended for production use")
        }
    }
    ```

    This removes the fmt.Fprintf(os.Stderr, ...) call entirely. After this change, check if fmt and os are still needed in the import block -- they likely are (fmt is used elsewhere in Dial, os may not be needed anymore). Clean up unused imports.

    ALSO: The TlsValidator created at line 121 should receive the Logger:
    ```go
    validator := connection.NewTlsValidator(tlsConfig)
    if config.Logger != nil {
        validator.SetLogger(config.Logger)
    }
    ```

    STEP 3: Add SetLogger to TlsValidator in pkg/connection/tls.go.

    Add types import:
    ```go
    import (
        // ... existing imports ...
        types "github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
    )
    ```

    Wait -- importing pkg/types from pkg/connection creates a dependency. Check if this creates a circular import. pkg/types should have no dependency on pkg/connection, so this should be safe. Verify by checking pkg/types imports.

    Add logger field to TlsValidator:
    ```go
    type TlsValidator struct {
        config *TLSConfig
        logger types.Logger // Optional logger for warnings (FOUND-05)
    }
    ```

    Add SetLogger method:
    ```go
    func (v *TlsValidator) SetLogger(logger types.Logger) {
        v.logger = logger
    }
    ```

    Add InsecureSkipVerify warning in GetTLSConfig() -- after building the config, before client cert loading:
    ```go
    if v.config.InsecureSkipVerify && v.logger != nil {
        v.logger.Warn("TLS InsecureSkipVerify is enabled -- server certificate verification disabled; not recommended for production use")
    }
    ```

    STEP 4: Update CheckCertificateRevocation comment in pkg/connection/tls.go.

    Replace existing TODO/stub comment with explicit v1 limitation documentation:
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
    ```

    STEP 5: Wire TLSConfig through ConnectionManager to transport.Dial.

    In pkg/managers/connection.go Connect() method, replace line 73:
    ```go
    // OLD:
    t, err := transport.Dial(ctx, cm.config.URL, header, nil)

    // NEW:
    wsConfig := &transport.WebSocketConfig{
        TLSConfig: cm.config.TLSConfig,
        Logger:    cm.config.Logger,
    }
    t, err := transport.Dial(ctx, cm.config.URL, header, wsConfig)
    ```

    First, check what cm.config looks like. Read ConnectionManager struct and its config field. The config needs to have TLSConfig and Logger fields. If ConnectionManager does not have these, they need to be passed through from client.go.

    Read the ConnectionManager struct definition and how it's created in NewClient(). Check what fields are available on cm.config. The ConnectionManager likely receives its config from the client -- we may need to ensure TLSConfig and Logger are included in what gets passed to ConnectionManager.

    IMPORTANT: Read pkg/managers/connection.go struct definition to understand what config fields exist before writing the wiring code. The goal is to pass the WebSocketConfig (with TLSConfig + Logger) through to transport.Dial instead of nil.

    STEP 6: Add tests.

    To pkg/connection/tls_test.go:
    - TestTlsValidator_InsecureSkipVerifyWarning_WithLogger: mock logger, InsecureSkipVerify=true, verify Warn called
    - TestTlsValidator_InsecureSkipVerifyWarning_NotEnabled: InsecureSkipVerify=false, verify no Warn call
    - TestTlsValidator_InsecureSkipVerifyWarning_NilLogger: InsecureSkipVerify=true, no SetLogger, verify no panic
    - TestCheckCertificateRevocation_V1Stub: verify returns nil and has v1 limitation comment

    To pkg/transport/websocket_test.go:
    - TestDial_InsecureSkipVerifyWarning_UsesLogger: verify Logger.Warn called instead of stderr (use mock logger, capture output)

    To pkg/managers/connection_test.go:
    - TestConnectionManager_Connect_PassesTLSConfig: verify TLSConfig is passed to transport.Dial (may require mock transport or checking dial config)

    Mock logger for tests:
    ```go
    type mockLogger struct {
        msgs []string
        mu    sync.Mutex
    }
    func (m *mockLogger) Debug(msg string, args ...any) {}
    func (m *mockLogger) Info(msg string, args ...any)  {}
    func (m *mockLogger) Warn(msg string, args ...any) {
        m.mu.Lock()
        m.msgs = append(m.msgs, msg)
        m.mu.Unlock()
    }
    func (m *mockLogger) Error(msg string, args ...any) {}
    ```
  </action>
  <verify>
    <automated>cd /Users/linyang/workspace/my-projects/openclaw-sdk-go && go test ./pkg/transport/ ./pkg/connection/ ./pkg/managers/ -run "TestTlsValidator_InsecureSkipVerifyWarning|TestDial.*TLS|TestCheckCertificateRevocation|TestConnectionManager_Connect" -v -race -count=1 && grep -c "v1 limitation" pkg/connection/tls.go && grep -c "fmt.Fprintf(os.Stderr" pkg/transport/websocket.go</automated>
  </verify>
  <done>
    - TlsValidator has logger field and SetLogger method
    - GetTLSConfig logs WARN via Logger when InsecureSkipVerify=true and logger is set
    - GetTLSConfig does not panic when logger is nil
    - GetTLSConfig does not warn when InsecureSkipVerify=false
    - CheckCertificateRevocation has explicit v1 limitation documentation
    - transport/websocket.go Dial() uses Logger.Warn instead of fmt.Fprintf(os.Stderr, ...)
    - WebSocketConfig has Logger field
    - ConnectionManager.Connect() passes TLSConfig and Logger through to transport.Dial (not nil)
    - TLS warning flows through: client config -> ConnectionManager -> transport.Dial -> Logger.Warn
    - All existing tests still pass
    - New tests pass with -race flag
  </done>
</task>

</tasks>

<verification>
go test ./pkg/... -race -count=1
go vet ./pkg/
grep -c "MaxRetries" pkg/managers/reconnect.go  # >= 3
grep -c "MaxRetriesExceeded" pkg/managers/reconnect.go  # >= 1
grep -c "SetLogger" pkg/connection/tls.go  # >= 2
grep -c "v1 limitation" pkg/connection/tls.go  # >= 1
grep -c "Logger" pkg/transport/websocket.go  # >= 2 (struct field + usage)
grep -c "fmt.Fprintf(os.Stderr" pkg/transport/websocket.go  # must be 0 (replaced with Logger)
grep -c "TLSConfig" pkg/managers/connection.go  # >= 1 (wired through)
grep -c "reconnect.Start" pkg/client.go  # check placement -- should NOT be in Connect()
</verification>

<success_criteria>
1. ReconnectManager stops after MaxRetries attempts with typed MaxRetriesExceededError
2. MaxRetries takes precedence over MaxAttempts with documented semantics
3. MaxRetries=0 falls back to MaxAttempts (backward compatible)
4. Negative values treated as zero
5. onReconnectFailed callback receives typed error wrapping ErrMaxRetriesExceeded
6. Reconnect does NOT start immediately after healthy Connect()
7. Reconnect starts on disconnect event
8. InsecureSkipVerify=true triggers WARN via Logger (not stderr)
9. TLSConfig wired through ConnectionManager -> transport.Dial
10. Logger wired through entire dial path
11. No warning when InsecureSkipVerify=false
12. No panic when logger is nil
13. CRL stub has explicit v1 limitation documentation
14. All tests pass with -race flag
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-03-SUMMARY.md`
</output>
