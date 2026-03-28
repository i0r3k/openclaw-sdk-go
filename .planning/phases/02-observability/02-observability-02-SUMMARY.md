# Phase 02 Plan 02: Observability - Wave 2 Summary

**Plan:** 02-observability-02
**Phase:** 02-observability
**Status:** COMPLETE
**Date:** 2026-03-28
**Commit:** 5803615

## One-Liner

OBS-02: Per-request timeout via variadic RequestOption functional options; OBS-03: Priority-based event dispatch with 3-tier channels (HIGH/MEDIUM/LOW) and graceful degradation.

## Objective

Implement OBS-02: Per-request timeout via variadic SendRequest options, and OBS-03: Event priority levels with graceful degradation. Both requirements share the same wave since they modify different subsystems and have no ordering dependency.

## Tasks Executed

| # | Task | Status | Commit |
|---|------|--------|--------|
| 1 | Add RequestOption type and WithRequestTimeout to client | DONE | 5803615 |
| 2 | Add Priority field to Event and restructure EventManager | DONE | 5803615 |
| 3 | Wire event priority assignment based on event type | DONE | 5803615 |

## Key Changes

### OBS-02: Per-Request Timeout (pkg/client.go)

**RequestOption functional option pattern:**
- `type requestConfig struct { timeout time.Duration }`
- `type RequestOption func(*requestConfig)`
- `func WithRequestTimeout(d time.Duration) RequestOption`

**SendRequest signature updated:**
- Old: `SendRequest(ctx context.Context, req *protocol.RequestFrame) (*protocol.ResponseFrame, error)`
- New: `SendRequest(ctx context.Context, req *protocol.RequestFrame, opts ...RequestOption) (*protocol.ResponseFrame, error)`

**Implementation:**
- Options parsed in `client.SendRequest`, wrapped ctx passed to `RequestManager`
- `WithRequestTimeout(d)` wraps ctx with `context.WithTimeout` (overwrites existing deadline per D-07)
- Backward compatible: `SendRequest(ctx, req)` continues to work unchanged

### OBS-03: Event Priority Levels (pkg/types/types.go, pkg/managers/event.go)

**Event struct updated:**
```go
type Event struct {
    Priority  EventPriority // NEW: default EventPriorityMedium when zero
    Type      EventType
    Payload   any
    Err       error
    Timestamp time.Time
}
```

**Priority channels structure:**
- `priorityHigh chan types.Event` - 25% of buffer
- `priorityMedium chan types.Event` - 25% of buffer
- `priorityLow chan types.Event` - 50% of buffer
- `events chan types.Event` - output channel from dispatcher

**Buffer partition:** HIGH=25%, MEDIUM=25%, LOW=50% (OBS-03)

**Dispatcher priority selection:**
1. HIGH events always processed first when available
2. MEDIUM events when HIGH is empty
3. LOW events when HIGH and MEDIUM are both empty

**Auto-assignment based on event type (D-11 through D-13):**
- HIGH: `EventError`, `EventDisconnect`, `EventStateChange`, `EventGap`
- MEDIUM: `EventTick`, `EventResponse`, `EventConnect`
- LOW: `EventMessage`, `EventRequest`

**Drop behavior (D-14):**
- When priority channel is full, drain lower priority channels first
- If still full, drop immediately (non-blocking)
- HIGH events never cause MEDIUM/LOW to be dropped (only MEDIUM can drain LOW)

## Files Modified

| File | Change |
|------|--------|
| pkg/client.go | Add requestConfig, RequestOption, WithRequestTimeout; update SendRequest signature |
| pkg/types/types.go | Add Priority field to Event struct |
| pkg/managers/event.go | Restructure with priority channels, dispatcher, drainLowerPriority |
| pkg/managers/event_test.go | Update backpressure test for non-blocking drop behavior |
| pkg/managers/event_priority_test.go | NEW: priority-specific tests |

## Files Created

| File | Purpose |
|------|---------|
| pkg/managers/event_priority_test.go | Tests for priority channels, auto-assignment, HIGH never drops |

## Success Criteria Verification

- [x] SendRequest accepts variadic RequestOption arguments
- [x] WithRequestTimeout(d) wraps ctx with deadline (overwrites existing per D-07)
- [x] Backward compatible: SendRequest(ctx, req) still works
- [x] Event struct has Priority field with default EventPriorityMedium
- [x] EventManager uses three priority channels with dispatcher
- [x] Buffer partition: HIGH=25%, MEDIUM=25%, LOW=50%
- [x] Dispatcher selects HIGH first, then MEDIUM, then LOW
- [x] Drop order: LOW first, then MEDIUM, then HIGH
- [x] Priority auto-assigned based on event type
- [x] All tests pass with -race flag

## Test Results

```
go test ./pkg/... -race -count=1
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg	2.248s
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg/api	1.464s
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg/auth	1.802s
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg/connection	1.985s
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg/events	2.541s
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg/managers	5.812s
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol	2.687s
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg/transport	63.698s
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg/types	3.246s
ok  	github.com/frisbee-ai/openclaw-sdk-go/pkg/utils	2.990s
```

## Deviations from Plan

1. **Task consolidation:** Tasks 1 and 2 were committed together since they both passed tests and were part of the same wave. No functional impact.

2. **Backpressure test updated:** `TestEventManager_Emit_BackpressureTimeout` was updated to reflect the new non-blocking drop behavior. The old test expected blocking when channel was full; the new design drains lower priorities first then drops immediately (non-blocking).

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Non-blocking Emit | Priority-based design drains lower channels instead of blocking; simpler and more predictable |
| Two goroutines in Start() | One for priority dispatcher, one for dispatchLoop reading from output channel |
| EventPriority default 0 (LOW) | Matches Go zero-value convention; EventManager auto-assigns MEDIUM when Priority==0 |

## Dependencies

- Depends on: 02-observability-01 (ConnectionMetrics, EventPriority type already existed)
- Provides: OBS-02 (per-request timeout), OBS-03 (priority event dispatch)

## Next Steps

- Phase 02 Plan 03: Additional refinements if needed
- Move to Phase 03: Client struct refactor
