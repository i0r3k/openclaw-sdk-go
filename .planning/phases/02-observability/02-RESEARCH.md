# Phase 2: Observability - Research

**Researched:** 2026-03-28
**Domain:** Go WebSocket SDK observability (connection health metrics, per-request timeouts, event priority dispatch)
**Confidence:** HIGH (internal codebase patterns confirmed, no external API research needed)

## Summary

Phase 2 delivers connection health visibility and graceful degradation under load. Three of four requirements need work (OBS-01, OBS-02, OBS-03); OBS-04 is already implemented. The key technical challenges are: (1) exposing a thread-safe attempt counter from `ReconnectManager.run()` goroutine, (2) layering variadic options over an existing `SendRequest` signature that already handles context deadlines, and (3) implementing priority-based event dropping without restructuring the existing `Emit()`+dispatcher pattern.

**Primary recommendation:** Implement `ConnectionMetrics` aggregation via a new `client.GetMetrics()` method; implement per-request timeout via `WithRequestTimeout` option that wraps ctx before passing to `RequestManager`; implement event priority via separate per-priority input channels feeding a priority-selecting dispatcher goroutine.

---

## User Constraints (from CONTEXT.md)

### Locked Decisions

- **OBS-01 Latency = tick-based**: `Latency = tickInterval * staleMultiplier` as baseline estimate (no RTT probing)
- **OBS-01 Metrics struct fields**: `Latency`, `LastTickAge`, `ReconnectCount`, `IsStale`
- **OBS-01 ReconnectCount source**: `ReconnectManager.AttemptCount()` method returning local `attempt` counter
- **OBS-02 SendRequest signature**: `SendRequest(ctx, req, opts ...RequestOption)` variadic options
- **OBS-02 WithRequestTimeout**: Wraps ctx with `context.WithTimeout`, overwrites existing ctx deadline (explicit caller choice)
- **OBS-02 Placement**: Options parsed in `client.SendRequest`, wrapped ctx passed to `RequestManager`
- **OBS-03 3 priority levels**: `EventPriorityHigh=2`, `EventPriorityMedium=1`, `EventPriorityLow=0`
- **OBS-03 Event.Priority field**: Default `EventPriorityMedium` for backward compat
- **OBS-03 Priority assignment**: HIGH=Error/Disconnect/StateChange/Gap; MEDIUM=Tick/Response/Connect; LOW=Message/Request
- **OBS-03 Drop order**: LOW first, then MEDIUM, then HIGH; HIGH events only drop as last resort
- **OBS-04**: Already implemented (skip)

### Claude's Discretion (research/planner decides)

- Exact internal channel structure for priority-based event dispatch (separate per-priority channels vs. single channel with priority tagging)
- How `ReconnectManager.AttemptCount()` is implemented (atomic int vs. mutex-protected int)
- Whether `GetMetrics()` should return a copy (defensive) or direct reference (zero-copy)
- Whether `TickMonitor` needs a `GetTickInterval()` method to support tick-based latency estimation

### Deferred Ideas (OUT OF SCOPE)

- Actual RTT measurement via ping/pong (Phase 4)
- More than 3 priority levels (future)
- Prometheus/OpenTelemetry metrics export (Phase 5)

---

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| OBS-01 | `ConnectionMetrics` struct with Latency, LastTickAge, ReconnectCount, IsStale; `GetMetrics()` on OpenClawClient | TickMonitor has `GetTimeSinceLastTick()` and `IsStale()`; ReconnectManager needs `AttemptCount()`; metrics aggregation in `client.GetMetrics()` |
| OBS-02 | `SendRequest(ctx, req, opts ...RequestOption)` with `WithRequestTimeout(d)` option | `client.SendRequest` already has reduced mutex scope; `RequestManager` already checks ctx deadline; layering options on top is straightforward |
| OBS-03 | Event priority levels; when EventChannel full, drop low-priority first | `EventManager` uses single `chan Event` with `emitTimer` bounded-wait; restructure to priority-aware dispatch with drop logic |
| OBS-04 | `EventBufferSize` configurable via client option | Already implemented: `ClientConfig.EventBufferSize` and `WithEventBufferSize()` exist in `pkg/client.go` |

---

## Standard Stack

No new external dependencies for Phase 2. The SDK uses only `github.com/gorilla/websocket v1.5.3` plus Go standard library.

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/gorilla/websocket` | v1.5.3 | WebSocket client | Already dependency |
| Go standard library | 1.24+ | `sync/atomic`, `context`, `time`, `sync` | Built into runtime |

**Version verification:** N/A - no new dependencies added in Phase 2.

---

## Architecture Patterns

### Recommended Project Structure

```
pkg/
├── types/
│   ├── types.go        # Add: ConnectionMetrics, EventPriority type+consts
│   └── types_test.go
├── events/
│   └── tick.go         # Add: GetTickIntervalMs() for OBS-01
├── managers/
│   ├── reconnect.go    # Add: AttemptCount() method for OBS-01
│   ├── reconnect_test.go
│   ├── event.go        # Restructure for priority dispatch (OBS-03)
│   └── event_test.go
└── client.go          # Add: GetMetrics(), SendRequest variadic opts, ConnectionMetrics struct
```

### Pattern 1: ReconnectManager.AttemptCount() via atomic counter

**What:** Expose the `attempt` counter from `run()` goroutine via a thread-safe getter.

**Implementation approach (atomic):**
```go
// ReconnectManager struct -- add field:
attemptCount atomic.Int64

// In run(), at each attempt iteration:
rm.attemptCount.Add(1)

// Expose getter:
func (rm *ReconnectManager) AttemptCount() int64 {
    return rm.attemptCount.Load()
}
```

**Why:** `attempt` is currently a local variable in `run()`, incremented each loop iteration. Exposing it requires promoting it to a struct field. `sync/atomic.Int64` is the idiomatic Go approach for a simple incrementing counter with no complex synchronization needs. Alternative: mutex-protected int (more boilerplate, same correctness). This approach requires zero locking on the read path.

**Source:** Confirmed from `pkg/managers/reconnect.go:97` (local `attempt` variable in `run()` goroutine).

---

### Pattern 2: Per-Request Timeout via Functional Options

**What:** Extend `SendRequest` to accept variadic options that modify request behavior (e.g., timeout).

**Implementation:**
```go
// In pkg/client.go -- client.SendRequest:
func (c *client) SendRequest(ctx context.Context, req *protocol.RequestFrame, opts ...RequestOption) (*protocol.ResponseFrame, error) {
    // Apply options
    cfg := &requestConfig{}
    for _, opt := range opts {
        opt(cfg)
    }

    // Wrap ctx with timeout if set
    if cfg.timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, cfg.timeout)
        defer cancel()
    }
    // ... rest of existing SendRequest logic
}

// In pkg/managers/request.go -- RequestManager.SendRequest signature unchanged (still takes ctx):
func (rm *RequestManager) SendRequest(ctx context.Context, req *protocol.RequestFrame, sendFunc func(*protocol.RequestFrame) error) (*protocol.ResponseFrame, error)
```

**Why:** D-08 explicitly says options are parsed in `client.SendRequest`, wrapped ctx passed to `RequestManager`. The `RequestManager.SendRequest` already handles ctx deadline checking at line 81 of `pkg/managers/request.go`. By the time ctx reaches `RequestManager`, it already has the timeout deadline set. The `RequestOption` functional option is consistent with the existing `ClientOption` pattern in the codebase.

**Key insight from D-07:** `WithRequestTimeout` overwrites any existing ctx deadline. This is the explicit contract -- the caller's timeout wins, not additive timeouts. This is simpler to reason about.

---

### Pattern 3: Event Priority via Per-Priority Channels + Dispatcher

**What:** Replace single `events chan Event` with three priority-specific input channels feeding a dispatcher goroutine that selects HIGH first.

**Implementation:**
```go
type EventManager struct {
    // Replace single events channel with three priority channels:
    priorityHigh   chan types.Event
    priorityMedium chan types.Event
    priorityLow    chan types.Event
    events        chan types.Event  // Output channel (buffered, from dispatch loop)

    // ... existing fields ...
}

// In NewEventManager:
em.priorityHigh = make(chan types.Event, bufferSize/4)    // Quarter buffer
em.priorityMedium = make(chan types.Event, bufferSize/4)
em.priorityLow = make(chan types.Event, bufferSize/2)    // Half buffer

// Emit() -- routes to priority channel:
func (em *EventManager) Emit(event types.Event) {
    var priorityCh chan types.Event
    switch event.Priority {
    case EventPriorityHigh:
        priorityCh = em.priorityHigh
    case EventPriorityMedium:
        priorityCh = em.priorityMedium
    default:
        priorityCh = em.priorityLow
    }
    select {
    case priorityCh <- event:
    default:
        // Buffer full -- try drain lower priorities per D-14
        if event.Priority > EventPriorityLow {
            // Try to drop from lower priority to make room
            em.drainLowerPriority(event.Priority)
            select {
            case priorityCh <- event:
                return
            default:
            }
        }
        em.logger.Warn("event dropped", "type", event.Type, "priority", event.Priority)
    }
}

// Dispatch loop -- selects HIGH first:
func (em *EventManager) dispatch() {
    for {
        select {
        case <-em.ctx.Done():
            return
        case e := <-em.priorityHigh:
            em.events <- e
        case e := <-em.priorityMedium:
            select {
            case em.events <- e:
            case <-em.priorityHigh:
                em.events <- e // prefer HIGH
            }
        case e := <-em.priorityLow:
            select {
            case em.events <- e:
            case <-em.priorityHigh:
                em.events <- e
            case <-em.priorityMedium:
                em.events <- e
            }
        }
    }
}
```

**Alternative considered (single channel + priority tag):** Simpler API (no dispatcher goroutine) but harder to implement "drop LOW before MEDIUM" behavior deterministically when the single channel is full. With separate channels, the dispatcher can select HIGH first naturally.

**Why the dispatcher approach:** Go's `select` statement with multiple channel receives naturally implements priority: if HIGH channel has data, it is always selected first regardless of whether MEDIUM or LOW also have data waiting. The dispatcher goroutine adds one goroutine overhead but cleanly separates priority selection from event emission.

**Drop behavior:** Per D-14, drop order is LOW first, then MEDIUM, then HIGH (HIGH never drops unless catastrophic). The `drainLowerPriority()` helper drops the oldest event from lower-priority channels to make room when the current event's priority channel is full.

**Source:** Confirmed from `pkg/managers/event.go:22` (single `events` channel), lines 33-34 (emitTimer pattern).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Thread-safe counter | Manual mutex + int | `sync/atomic.Int64` | Simpler, lock-free reads, idiomatic Go |
| Context timeout | Custom timeout channel | `context.WithTimeout(ctx, d)` | Context cancellation is standard pattern, composes with existing ctx handling |
| Priority dispatch | Single channel + conditional drop | Per-priority channels + select | Go's select naturally implements priority; deterministic drop order |

---

## Common Pitfalls

### Pitfall 1: ReconnectManager.AttemptCount() - goroutine visibility

**What goes wrong:** If `attempt` is just a local variable in `run()`, promoting it to a struct field requires proper synchronization for the main goroutine to see updates from the `run()` goroutine.

**Why it happens:** Go's memory model guarantees visibility only through synchronization primitives (channels, mutexes, atomic operations). A plain struct field written by one goroutine and read by another without synchronization is a data race.

**How to avoid:** Use `sync/atomic.Int64` for the counter field. Reads (`Load()`) and writes (`Add()`) are atomic operations that provide the necessary visibility guarantees.

**Warning signs:** `go test -race` fails with "race: ..." on the attempt counter access.

**Source:** Confirmed from `pkg/managers/reconnect.go:97` (`attempt := 0` local in `run()` goroutine).

---

### Pitfall 2: client.GetMetrics() - tickMonitor nil before Connect

**What goes wrong:** `client.tickMonitor` is initialized inside `Connect()` method, after the WebSocket handshake and `processServerInfo()`. Calling `GetMetrics()` before `Connect()` (or on a client that never connected) would return zero/invalid metrics or nil-pointer dereference.

**Why it happens:** Looking at `pkg/client.go:528-544`, `tickMonitor` is created inside `Connect()` based on server policy negotiated during handshake. It is not set during `NewClient()`.

**How to avoid:** `GetMetrics()` must check if `tickMonitor` is nil and return zero-values for `Latency` and `IsStale` in that case. `ReconnectCount` from `managers.reconnect.AttemptCount()` is still valid (even before first connect attempt, it should be 0).

**Warning signs:** Calling `GetMetrics()` on an unconnected client panics or returns garbage.

**Source:** Confirmed from `pkg/client.go:528-544` (tickMonitor initialization in Connect).

---

### Pitfall 3: WithRequestTimeout overwrites existing ctx deadline

**What goes wrong:** Per D-07, `WithRequestTimeout` explicitly overwrites any existing ctx deadline. A caller who sets a 10s deadline on ctx and then calls `WithRequestTimeout(5s)` will get a 5s effective deadline, not a 10s one. This is the documented behavior, but callers who expect additive timeouts will be surprised.

**Why it happens:** `context.WithTimeout(ctx, d)` always creates a new context with deadline `now + d`, ignoring any existing deadline on ctx.

**How to avoid:** The documentation in D-07 is explicit. Implementors should ensure the error returned when timeout fires is a `TimeoutError` (not a generic context deadline exceeded error), so callers can distinguish timeout-induced failures.

**Warning signs:** Tests where caller sets ctx with deadline and expects longer effective timeout.

**Source:** Confirmed from CONTEXT.md D-07 ("If the caller already set a deadline on ctx, it is overwritten by WithRequestTimeout").

---

### Pitfall 4: Event priority restructuring breaking existing API

**What goes wrong:** Changing `EventManager` to use multiple channels or a dispatcher goroutine changes the event delivery semantics. Existing code subscribing to `Events()` expects a single channel.

**Why it happens:** The dispatcher goroutine adds an extra hop for all events. If the dispatcher panics or blocks, events stop flowing. The `Events()` return type is `<-chan Event` (receive-only), which is good for encapsulation.

**How to avoid:**
1. Keep `Events()` returning `<-chan Event` (single output channel)
2. The dispatcher is internal -- the public API is unchanged
3. Ensure dispatcher goroutine is started in `Start()` and cleaned up in `Close()`
4. Use `sync.WaitGroup` to confirm dispatcher exits on Close

**Warning signs:** `go test -race` shows channel races on the events channel.

**Source:** Confirmed from `pkg/managers/event.go:82-85` (Events() returns receive-only channel).

---

## Code Examples

### ConnectionMetrics struct (OBS-01)

```go
// pkg/types/types.go -- add after EventHandler type

// ConnectionMetrics holds connection health metrics.
// All fields are snapshots at call time -- no history retained.
type ConnectionMetrics struct {
    Latency        time.Duration // tick-based estimate: tickInterval * staleMultiplier
    LastTickAge    time.Duration // actual time since last tick received
    ReconnectCount int           // total reconnection attempts made (snapshot)
    IsStale        bool          // whether connection is currently stale
}
```

**Source:** Based on `pkg/events/tick.go:163-171` (GetTimeSinceLastTick returns int64 ms), `pkg/events/tick.go:139` (IsStale()), and `pkg/managers/reconnect.go` (attempt counter in run goroutine).

### TickMonitor.GetTickIntervalMs() (OBS-01 helper)

```go
// pkg/events/tick.go -- add method

// GetTickIntervalMs returns the configured tick interval in milliseconds.
func (tm *TickMonitor) GetTickIntervalMs() int64 {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    return tm.tickIntervalMs
}
```

**Source:** Confirmed from `pkg/events/tick.go:18` (tickIntervalMs field).

### client.GetMetrics() implementation (OBS-01)

```go
// pkg/client.go -- add method

// GetMetrics returns a snapshot of connection health metrics.
func (c *client) GetMetrics() ConnectionMetrics {
    c.mu.Lock()
    defer c.mu.Unlock()

    var latency time.Duration
    var lastTickAge time.Duration
    var isStale bool

    if c.tickMonitor != nil {
        // Latency = tickInterval * staleMultiplier (baseline estimate)
        intervalMs := c.tickMonitor.GetTickIntervalMs()
        // staleMultiplier is private, need to expose or compute from IsStale behavior
        // Alternative: expose staleMultiplier via GetStaleMultiplier() method
        // For now, compute: if IsStale=true, latency > expected interval
        lastTickAge = time.Duration(c.tickMonitor.GetTimeSinceLastTick()) * time.Millisecond
        isStale = c.tickMonitor.IsStale()
        // Latency estimate: use LastTickAge as upper bound when stale
        if isStale {
            latency = lastTickAge
        } else {
            latency = time.Duration(intervalMs) * time.Millisecond
        }
    }

    reconnectCount := 0
    if c.managers.reconnect != nil {
        reconnectCount = int(c.managers.reconnect.AttemptCount())
    }

    return ConnectionMetrics{
        Latency:        latency,
        LastTickAge:    lastTickAge,
        ReconnectCount: reconnectCount,
        IsStale:        isStale,
    }
}
```

**Note:** `staleMultiplier` is private in `TickMonitor`. The latency calculation in D-03 uses `tickInterval * staleMultiplier` as a baseline. Since `staleMultiplier` cannot be read from outside, two options: (1) add `GetStaleMultiplier()` method to `TickMonitor`, or (2) use `LastTickAge` as the latency estimate when stale, and `tickIntervalMs * Millisecond` when healthy. Option 2 is simpler and matches the spec's intent (tick-based estimate).

### RequestOption type and WithRequestTimeout (OBS-02)

```go
// pkg/managers/request.go -- add at top (near RequestOptions struct)

type requestConfig struct {
    timeout time.Duration
}

type RequestOption func(*requestConfig)

func WithRequestTimeout(d time.Duration) RequestOption {
    return func(cfg *requestConfig) {
        cfg.timeout = d
    }
}
```

**Source:** Based on existing `RequestOptions` struct at `pkg/managers/request.go:18-22` and `ClientOption` pattern in `pkg/client.go:194-196`.

### EventPriority type and constants (OBS-03)

```go
// pkg/types/types.go -- add after EventHandler type

type EventPriority int

const (
    EventPriorityLow    EventPriority = 0
    EventPriorityMedium EventPriority = 1
    EventPriorityHigh   EventPriority = 2
)
```

**Source:** Per D-09, D-10 from CONTEXT.md.

---

## State of the Art

This SDK uses Go's standard concurrency patterns. No external libraries for observability.

| Aspect | Approach Used | Alternative Considered |
|--------|--------------|------------------------|
| Metrics exposure | Return struct snapshot via GetMetrics() | Channel-based streaming, Prometheus client |
| Counter thread-safety | sync/atomic.Int64 | sync.Mutex-protected int |
| Per-request timeout | Context wrapping via functional option | Custom timeout field on RequestOptions |
| Event priority | Per-priority channels + dispatcher goroutine | Single channel with priority tag + conditional drop |
| Staleness detection | Tick-based (TickMonitor) | Ping/pong probing (Phase 4) |

---

## Open Questions

1. **TickMonitor staleMultiplier exposure**
   - What we know: `staleMultiplier` is private in `TickMonitor`; latency formula needs `tickInterval * staleMultiplier`
   - What's unclear: Whether to add `GetStaleMultiplier()` method or compute latency differently
   - Recommendation: Add `GetStaleMultiplier()` method to `TickMonitor` for clean API. Return value is constant after construction, so `RLock` is sufficient.

2. **Priority channel buffer sizes**
   - What we know: Total buffer size is `bufferSize` from config (default 100)
   - What's unclear: How to partition the buffer among HIGH/MEDIUM/LOW channels
   - Recommendation: Equal split or weighted (e.g., HIGH=25%, MEDIUM=25%, LOW=50%). Higher split for HIGH is appropriate given HIGH events should never drop.

3. **Gap between tickMonitor initialization and first tick**
   - What we know: `tickMonitor` created in `Connect()` after server info processed; first tick may take time
   - What's unclear: What `LastTickAge` should return before first tick received
   - Recommendation: Return 0 with `IsStale=false` before first tick (matches `GetTimeSinceLastTick()` behavior returning 0 when `lastTickTime==0`)

4. **ReconnectManager.AttemptCount() before any reconnect attempts**
   - What we know: `attemptCount` starts at 0; increments each reconnect loop iteration
   - What's unclear: Whether `AttemptCount()` should reset on successful reconnect
   - Recommendation: No reset -- counter is a lifetime total of reconnection attempts, consistent with D-03 ("ReconnectCount is a snapshot"). Reset on new connection is a separate feature if needed.

---

## Environment Availability

> Step 2.6: SKIPPED (no external dependencies identified for Phase 2)

Phase 2 is purely internal code changes (new types, modified methods, restructured event dispatch). No new external tools, services, or runtimes required.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go built-in `testing.T` |
| Config file | None -- standard Go test layout |
| Quick run command | `go test ./pkg/... -run "Test(OBS|Event|Priority|Metrics|Timeout)" -v -count=1` |
| Full suite command | `go test ./pkg/... -race -count=1` |

### Phase Requirements to Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| OBS-01 | GetMetrics returns correct Latency when tickMonitor running | unit | `go test -run TestGetMetrics_Latency ./pkg/ -v` | `pkg/client_test.go` (likely needs new tests) |
| OBS-01 | GetMetrics returns correct ReconnectCount | unit | `go test -run TestGetMetrics_ReconnectCount ./pkg/ -v` | `pkg/client_test.go` (likely needs new tests) |
| OBS-01 | GetMetrics returns IsStale correctly | unit | `go test -run TestGetMetrics_IsStale ./pkg/ -v` | `pkg/client_test.go` (likely needs new tests) |
| OBS-01 | AttemptCount returns thread-safe count | unit | `go test -run TestReconnectManager_AttemptCount ./pkg/managers/ -v -race` | `pkg/managers/reconnect_test.go` (needs new tests) |
| OBS-02 | SendRequest with WithRequestTimeout respects timeout | unit | `go test -run TestSendRequest_WithRequestTimeout ./pkg/ -v` | `pkg/client_test.go` (needs new tests) |
| OBS-02 | SendRequest without options works (backward compat) | unit | `go test -run TestSendRequest_NoOptions ./pkg/ -v` | existing tests cover basic path |
| OBS-03 | HIGH priority events never drop when MEDIUM and LOW full | unit | `go test -run TestEventManager_PriorityHighNeverDrops ./pkg/managers/ -v` | `pkg/managers/event_test.go` (needs new tests) |
| OBS-03 | LOW events drop first when buffer full | unit | `go test -run TestEventManager_PriorityDropOrder ./pkg/managers/ -v` | `pkg/managers/event_test.go` (needs new tests) |
| OBS-03 | Default Event priority is MEDIUM | unit | `go test -run TestEvent_DefaultPriority ./pkg/ -v` | `pkg/types/types_test.go` (needs new tests) |
| OBS-04 | EventBufferSize option works | unit | `go test -run TestWithEventBufferSize ./pkg/ -v` | already implemented, verify |

### Sampling Rate

- **Per task commit:** `go test ./pkg/... -run "<task tests>" -v -count=1`
- **Per wave merge:** `go test ./pkg/... -race -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

Test infrastructure exists (`pkg/client_test.go`, `pkg/managers/event_test.go`, `pkg/managers/reconnect_test.go`). Phase 2 needs new tests added to these files:

- [ ] `pkg/managers/reconnect_test.go` -- `TestReconnectManager_AttemptCount` tests atomic counter behavior
- [ ] `pkg/client_test.go` -- `TestGetMetrics_*` tests for metrics aggregation
- [ ] `pkg/client_test.go` -- `TestSendRequest_WithRequestTimeout` for variadic options
- [ ] `pkg/managers/event_test.go` -- `TestEventManager_Priority*` tests for priority dispatch and drop behavior
- [ ] `pkg/types/types_test.go` -- `TestEvent_DefaultPriority` for default priority assignment

---

## Sources

### Primary (HIGH confidence - internal codebase patterns)

- `pkg/client.go` -- OpenClawClient interface, SendRequest signature, tickMonitor initialization, GetTickMonitor pattern
- `pkg/managers/event.go` -- EventManager struct, Emit() implementation, emitTimer pattern, dispatch() loop
- `pkg/managers/request.go` -- RequestOptions struct, timeout handling via ctx deadline
- `pkg/managers/reconnect.go` -- run() goroutine with local attempt counter, Start() idempotency
- `pkg/events/tick.go` -- TickMonitor struct, GetTimeSinceLastTick(), IsStale(), GetStaleDuration()
- `pkg/types/types.go` -- Event struct, EventType constants, EventHandler type
- `.planning/01-foundation/02-PLAN.md` -- Task 1 SendRequest mutex scope pattern
- `.planning/01-foundation/03-PLAN.md` -- Task 2 reconnect triggering pattern

### Secondary (HIGH confidence - decisions from discuss-phase)

- `.planning/phases/02-observability/02-CONTEXT.md` -- All OBS-01 through OBS-04 decisions documented

### No external sources required

Phase 2 is entirely about implementing documented decisions against existing internal codebase. No web search, Context7, or external documentation needed.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - no new dependencies, Go standard patterns only
- Architecture: HIGH - internal patterns confirmed from source, decisions locked
- Pitfalls: HIGH - all identified from code review of existing patterns

**Research date:** 2026-03-28
**Valid until:** 2026-04-28 (30 days -- stable codebase, no external API changes)
