# TypeScript to Go SDK Migration Plan

> Complete merged plan v4 — incorporating all revisions from v1, v2, v3, and v4.
> Based on Architect + Critic consensus after 4 rounds of review.

## Document History

| Version | Changes |
|---------|---------|
| v1 | Initial 6-phase plan |
| v2 | Added M1/M2 milestones, Phase 6 split, Rollback strategy, parallelization |
| v3 | Fixed Decision 2 (Go methods vs TS getters), verified error code sources, added Phase 1.5 |
| v4 | Fixed wire-send mechanism (steps 1.4a/1.4b), added wire bytes test (1.5.2), rewrote Phase 4 error matching |

---

## Context

This is a migration of new features from `openclaw-sdk-typescript` (commit d9d9751..HEAD) to `openclaw-sdk-go`. The TypeScript SDK has undergone significant refactoring including protocol format changes, 8 new API namespaces, and 30+ bug fixes. The Go SDK currently uses an outdated protocol format and lacks all API namespaces.

### Key Reference Files
- TypeScript SDK: `/Users/linyang/workspace/my-projects/openclaw-sdk-typescript/`
- Go SDK: `/Users/linyang/workspace/my-projects/openclaw-sdk-go/`

---

## Current State vs Target State

| Component | Go SDK Current | TypeScript SDK Target |
|-----------|--------------|---------------------|
| **RequestFrame** | `{ type:"gateway", payload:{ RequestID, Method, Params, Timestamp } }` | `{ type:"req", id, method, params }` (no wrapper) |
| **ResponseFrame** | `{ type:"gateway", payload:{ RequestID, Success, Result, Error, Timestamp } }` | `{ type:"res", id, ok, payload?, error?, progress? }` |
| **EventFrame** | `{ type:"gateway", payload:{ EventType, Data, Timestamp } }` | `{ type:"event", event, payload?, seq?, stateVersion? }` |
| **GatewayFrame** | Wrapper struct (top-level `type` + `timestamp`) | Deprecated, frames sent flat |
| **ErrorShape** | `{ Code, Message }` | `{ code, message, details?, retryable?, retryAfterMs? }` |
| **ConnectionState** | `connected/authenticated/failed` | `handshaking/ready/closed` |
| **ProtocolNegotiator** | String version `[]string{"1.0"}` | Integer version `min=3, max=3` |
| **PolicyManager** | `maxReconnectAttempts, pingInterval` | `maxPayload, maxBufferedBytes, tickIntervalMs` |
| **TickMonitor** | Basic ticker implementation | `staleMultiplier, onStale, onRecovered` callbacks |
| **GapDetector** | Simple Record(seq) | Configurable + GapRecoveryConfig |
| **API Namespaces** | Does not exist | 8 namespaces (Chat, Agents, Sessions, Config, Cron, Nodes, Skills, DevicePairing) |
| **API Params** | Does not exist | 142 types (546 lines) |
| **Connection Types** | Does not exist | `ConnectParams`, `HelloOk`, `Snapshot` |
| **Error System** | 6 coarse-grained codes | Detailed classification (Auth/Connection/Protocol/Request/Gateway/Reconnect) |

**Completed phases (old format):**
- Phase 1: Setup/foundation
- Phase 2: Auth module
- Phase 3: Protocol types (old format)
- Phase 4: Transport module
- Phase 5: Connection module (partial)
- Phase 6: Events module
- Phase 7: Managers module
- Phase 8: Utils module
- Phase 9: Main client (old format)

**Phases requiring rewrite/creation:**
- Phase 1 rewrite: Protocol layer update
- Phase 1.5 new: Wire compatibility tests
- Phase 2 new: API parameter types
- Phase 3 new: API Namespace packages
- Phase 4 enhancement: Error system (source-verified)
- Phase 5 rewrite: Connection protocol (ConnectParams/HelloOk/Snapshot)
- Phase 6.1 rewrite: Client core
- Phase 6.2 new: API Namespaces integration
- Phase 7 enhancement: RequestManager
- Phase 8 enhancement: TickMonitor
- Phase 9 enhancement: GapDetector
- Phase 10 new: Integration tests

---

## Target State

Wire protocol fully compatible with TypeScript SDK (type=req/res/event, id, ok, payload, etc.).

Go idioms体现在:
- Option pattern configuration
- context.Context lifecycle management
- sync.Mutex thread safety
- Go error interface (errors.Is/As)

---

## Milestone Structure

### M1: Wire Protocol (Phase 1, 1.5, 4, 5, 7, 8, 9)
Core protocol layer migration + error system + connection protocol + infrastructure.

### M2: API Surface (Phase 2, 3, 6.1, 6.2, 10)
API parameter types + Namespace packages + Client integration + integration tests.

**After Phase 1.5 passes, M1 and M2 can proceed in parallel.**

---

## Phase Structure

| Milestone | Phase | Focus | Key Deliverables |
|-----------|-------|-------|------------------|
| M1 | Phase 1 | Protocol Foundation | Frame format refactor (`id`, `ok`, `progress`, `seq`, `stateVersion`), wire-send mechanism |
| M1 | Phase 1.5 | Wire Compatibility Tests | JSON snapshot + actual wire bytes tests (Gate: must pass) |
| M2 | Phase 2 | API Parameter Types | 142 types from TypeScript `api-params.ts` |
| M2 | Phase 3 | API Namespace Packages | 8 packages: Chat, Agents, Sessions, Config, Cron, Nodes, Skills, DevicePairing |
| M1 | Phase 4 | Error System | 37 error codes, matching TS `createErrorFromResponse` logic |
| M1 | Phase 5 | Connection Types | ConnectParams, HelloOk, Snapshot, Policy |
| M1 | Phase 6.1 | Client Core | TickMonitor, GapDetector, ProtocolNegotiator, PolicyManager integration |
| M1 | Phase 6.2 | API Integration | 8 API namespaces wired to client |
| M1 | Phase 7 | RequestManager | Progress callbacks, server-initiated requests |
| M1 | Phase 8 | TickMonitor | staleMultiplier, onStale/onRecovered callbacks |
| M1 | Phase 9 | GapDetector | GapRecoveryConfig, recovery flow |
| M2 | Phase 10 | Integration Tests | End-to-end tests with mock WebSocket server |

---

### Phase 1: Protocol Layer Refactor

**Status**: Pending implementation

**TypeScript Source Verification**:
- RequestFrame: `src/protocol/frames.ts` defines `{ type:'req', id, method, params }`
- ResponseFrame: `src/protocol/frames.ts` defines `{ type:'res', id, ok, payload?, error?, progress? }`
- EventFrame: `src/protocol/frames.ts` defines `{ type:'event', event, payload?, seq?, stateVersion? }`

**File Changes**:

1. **`pkg/protocol/types.go`** -- Rewrite
   - Remove `FrameTypeGateway` constant and `GatewayFrame` wrapper struct
   - Add `FrameType` constants: `FrameTypeRequest = "req"`, `FrameTypeResponse = "res"`, `FrameTypeEvent = "event"`
   - Rewrite `RequestFrame`: `{ Type:"req", ID:string, Method:string, Params:json.RawMessage }` (flat, no wrapper, no timestamp)
   - Rewrite `ResponseFrame`: `{ Type:"res", ID:string, Ok:bool, Payload?:json.RawMessage, Error?:*ErrorShape, Progress?:bool }`
   - Rewrite `EventFrame`: `{ Type:"event", Event:string, Payload?:json.RawMessage, Seq?:uint64, StateVersion?:*StateVersion }`
   - Add `StateVersion`: `{ Presence:uint64, Health:uint64 }`
   - Add `ErrorShape`: `{ Code, Message, Details?, Retryable?, RetryAfterMs? }`
   - Update factory functions: `NewRequestFrame`, `NewResponseFrame`, `NewEventFrame`
   - **Acceptance**: JSON serialization output matches TypeScript (see Phase 1.5)

2. **`pkg/protocol/validation.go`** -- Rewrite
   - Update validation logic to match new format
   - RequestFrame validation: `ID` not empty, `Method` format correct, no `RequestID`/`Timestamp`
   - ResponseFrame validation: `ID` not empty, `Ok` and `Error` mutually exclusive
   - EventFrame validation: `Event` not empty, no `EventType`/`Data`/`Timestamp`
   - Note: Remove all references to old field names `RequestID`/`Success`/`Result`/`EventType`/`Data`/`Timestamp`

3. **`pkg/protocol/types_test.go`** -- Rewrite
   - Update all test cases to match new field names
   - Add JSON serialization/deserialization tests
   - Test: serialization has no gateway wrapper, field names match TS

4. **`pkg/protocol/errors.go`** -- New file
   - Move `ErrorCodes` constants from `pkg/types/errors.go` (protocol level)
   - Add `ErrorCode` type alias
   - **Source**: `openclaw-sdk-typescript/src/protocol/errors.ts` -- 5 codes: `NOT_LINKED`, `NOT_PAIRED`, `AGENT_TIMEOUT`, `INVALID_REQUEST`, `UNAVAILABLE`

5. **`pkg/client.go:SendRequest`** -- Modify sendFunc (wire-send mechanism)
   - Remove `GatewayFrame` wrapper logic
   - Send `RequestFrame` JSON directly via `json.Marshal(r)`
   - Old field names: `RequestID`, `Method`, `Params`, `Timestamp`
   - New field names: `Type:"req"`, `ID`, `Method`, `Params` (no timestamp)
   - **Acceptance**: marshal output is `{"type":"req","id":"...","method":"...","params":{}}` (no gateway wrapper)

6. **`pkg/transport/websocket.go`** -- Confirm no wrapping
   - Check `Send(data []byte)` method: pure passthrough to `sendCh`
   - Check `conn.WriteMessage` in `writePump`: writes data directly, no extra JSON wrapper
   - **Acceptance**: `Transport.Send` receives exactly the same `data` sent from client.go

---

### Phase 1.5: Wire Compatibility Tests

**Status**: Pending implementation

**Purpose**: Verify that Go actually sends bytes to WebSocket that match TypeScript wire format exactly.

#### 1.5.1 JSON Snapshot Test

File: **`pkg/protocol/wire_compat_test.go`**

- `TestRequestFrame_Serialize_MatchesTS` -- `{"type":"req","id":"req-abc123","method":"chat.list","params":{}}`
- `TestResponseFrame_Serialize_MatchesTS` -- success/error variants
- `TestEventFrame_Serialize_MatchesTS` -- with seq and stateVersion
- `TestErrorShape_AllFields` -- code, message, details, retryable, retryAfterMs

**Snapshot fixtures** (from TS source):
```go
tsRequestJSON := `{"type":"req","id":"req-abc123","method":"chat.list","params":{}}`
tsResponseJSON := `{"type":"res","id":"req-abc123","ok":true,"payload":{"chats":[]}}`
tsErrorResponseJSON := `{"type":"res","id":"req-abc123","ok":false,"error":{"code":"AUTH_TOKEN_EXPIRED","message":"Token expired","retryable":true}}`
tsEventJSON := `{"type":"event","event":"tick","payload":{"ts":1234567890},"seq":42,"stateVersion":{"presence":10,"health":5}}`
```

#### 1.5.2 Actual Wire Bytes Test (End-to-End)

File: **`pkg/protocol/wire_bytes_test.go`**

**Purpose**: Capture actual sent bytes via mock WebSocket server, verify no gateway wrapper.

```go
func TestSendRequest_WireBytes_ActualOutput(t *testing.T) {
    // 1. Start mock WebSocket server
    server := httptest.NewServer(websocket.Handler(func(w http.ResponseWriter, r *http.Request) {
        conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
        defer conn.Close()

        // 2. Read actual bytes sent
        _, msg, _ := conn.ReadMessage()
        var frame map[string]interface{}
        json.Unmarshal(msg, &frame)

        // 3. Verify: no gateway wrapper
        require.Nil(t, frame["payload"], "RequestFrame should NOT be wrapped in GatewayFrame")
        require.Equal(t, "req", frame["type"])
        require.NotEmpty(t, frame["id"])
        require.NotEmpty(t, frame["method"])

        // 4. Verify field names match TS (id, method, params not requestId, Method, Params)
        require.Contains(t, frame, "id", "must use 'id' not 'requestId'")
        require.Contains(t, frame, "method", "must use 'method' not 'Method'")
        require.NotContains(t, frame, "requestId", "must not use legacy 'requestId' field")
        require.NotContains(t, frame, "timestamp", "must not include timestamp in wire format")
    }))
    defer server.Close()

    // 3. Create client and send request
    // 4. Assert: msg == expected json bytes (no wrapper)
}
```

**Test Cases**:
- `TestSendRequest_WireBytes_NoGatewayWrapper` -- Verify no `{"type":"gateway","payload":"..."}` wrapper
- `TestSendRequest_WireBytes_FieldNames` -- Verify `id`/`method`/`params` not `requestId`/`Method`/`Params`
- `TestSendRequest_WireBytes_NoTimestamp` -- Verify wire format has no timestamp field
- `TestSendRequest_WireBytes_EmptyParams` -- Verify `params:{}` not `params:null`
- `TestSendRequest_WireBytes_WithParams` -- Verify complex params serialization

**Gate Function**: ALL of these tests must PASS or rollback.

---

### Phase 2: API Parameter Types

**Status**: Pending implementation

**TypeScript Source Verification**:
- `src/protocol/api-params.ts` -- 142 types (546 lines)
- `src/protocol/api-common.ts` -- shared types

**File Changes**:

1. **`pkg/protocol/api_params.go`** -- New file (~500 lines)
   -移植 142 types from TypeScript `src/protocol/api-params.ts`
   - Organize by namespace: Agent, NodePair, DevicePair, Config, Wizard, Talk, Channels, WebLogin, Skills, Cron, Sessions, Node, Chat, Browser, Push, Usage, Doctor, Secrets, TTS, Voice, Logs, Update, Diagnostics
   - Use Go struct tags: `json:"fieldName,omitempty"`
   - All optional fields use pointer or omitempty

2. **`pkg/protocol/api_common.go`** -- New file (~100 lines)
   -移植 shared types from TypeScript `src/protocol/api-common.ts`
   - `AgentSummary`, `WizardStep`, `CronJob`, `CronRunLogEntry`, `TtsVoicesResult`, `VoiceWakeStatusResult`, `BrowserListResult`, `DoctorCheckResult`, `SecretsListResult`, `ChatListResult`

---

### Phase 3: API Namespace Packages

**Status**: Pending implementation

**TypeScript Source Verification**:
- `src/api/chat.ts`, `src/api/agents.ts`, `src/api/sessions.ts`, `src/api/config.ts`, `src/api/cron.ts`, `src/api/nodes.ts`, `src/api/skills.ts`, `src/api/devicePairing.ts`

**Design Pattern**:
```go
// Each API namespace takes a RequestFn in constructor
type RequestFn func(ctx context.Context, method string, params interface{}) (interface{}, error)

type ChatAPI struct {
    request RequestFn
}

func (api *ChatAPI) Inject(ctx context.Context, params ChatInjectParams) error
func (api *ChatAPI) List(ctx context.Context, params *ChatListParams) (ChatListResult, error)
func (api *ChatAPI) History(ctx context.Context, params ChatHistoryParams) (ChatHistoryResult, error)
func (api *ChatAPI) Delete(ctx context.Context, params ChatDeleteParams) error
func (api *ChatAPI) Title(ctx context.Context, params ChatTitleParams) (ChatTitleResult, error)
```

**File Changes**:

1. **`pkg/api/shared.go`** -- New file
   - `RequestFn` type: `func(ctx context.Context, method string, params interface{}) (interface{}, error)`

2. **`pkg/api/chat.go`** -- New file
   - `ChatAPI` struct: `request RequestFn`
   - Methods: `Inject`, `List`, `History`, `Delete`, `Title`
   - Each method calls `api.request("chat.method", params)`

3. **`pkg/api/agents.go`** -- New file
   - `AgentsAPI` struct + `Files` sub-struct
   - Methods: `Identity`, `Wait`, `Create`, `Update`, `Delete`, `List`, `Files.List`, `Files.Get`, `Files.Set`

4. **`pkg/api/sessions.go`** -- New file
   - `SessionsAPI` struct
   - Methods: `List`, `Preview`, `Resolve`, `Patch`, `Reset`, `Delete`, `Compact`, `Usage`

5. **`pkg/api/config.go`** -- New file
   - `ConfigAPI` struct
   - Methods: `Get`, `Set`, `Apply`, `Patch`, `Schema`

6. **`pkg/api/cron.go`** -- New file
   - `CronAPI` struct
   - Methods: `List`, `Status`, `Add`, `Update`, `Remove`, `Run`, `Runs`

7. **`pkg/api/nodes.go`** -- New file
   - `NodesAPI` struct + `Pairing` sub-struct
   - Methods: `List`, `Invoke`, `Event`, `PendingDrain`, `PendingEnqueue`, `Pairing.List/Approve/Reject/Verify`

8. **`pkg/api/skills.go`** -- New file
   - `SkillsAPI` struct + `Tools` sub-struct
   - Methods: `Status`, `Tools.Catalog`, `Bins`, `Install`, `Update`

9. **`pkg/api/device_pairing.go`** -- New file
   - `DevicePairingAPI` struct
   - Methods: `List`, `Approve`, `Reject`

10. **`pkg/api/api.go`** -- New file
    - Re-export all API types and constructors

---

### Phase 4: Error System Enhancement

**Status**: Pending implementation

**TypeScript Source Verification**:
- `src/errors.ts` -- `createErrorFromResponse` function (lines 386-441)
- Error codes source verified: **All 37 codes come from `src/errors.ts`**

**Matching Logic (TS `createErrorFromResponse` Source)**:

```typescript
// AUTH_* or CHALLENGE_* -> prefix match -> AuthError
// CONNECTION_* or TLS_FINGERPRINT_MISMATCH -> prefix + exact -> ConnectionError
// PROTOCOL_* -> prefix -> ProtocolError
// METHOD_NOT_FOUND | INVALID_PARAMS | INTERNAL_ERROR -> exact -> RequestError
// Everything else (including REQUEST_TIMEOUT, REQUEST_CANCELLED, REQUEST_ABORTED) -> GatewayError
```

**File Changes**:

1. **`pkg/types/errors.go`** -- Major expansion
   - Add error code classification (37 codes, from `src/errors.ts`):
     - `AuthErrorCode`: 9 codes (CHALLENGE_EXPIRED, CHALLENGE_FAILED, AUTH_TOKEN_EXPIRED, AUTH_TOKEN_MISMATCH, AUTH_RATE_LIMITED, AUTH_DEVICE_REJECTED, AUTH_PASSWORD_INVALID, AUTH_NOT_SUPPORTED, AUTH_CONFIGURATION_ERROR)
     - `ConnectionErrorCode`: 7 codes (TLS_FINGERPRINT_MISMATCH, CONNECTION_STALE, CONNECTION_CLOSED, CONNECT_TIMEOUT, CONNECTION_REFUSED, NETWORK_ERROR, PROTOCOL_ERROR)
     - `ProtocolErrorCode`: 4 codes (PROTOCOL_UNSUPPORTED, PROTOCOL_NEGOTIATION_FAILED, INVALID_FRAME, FRAME_TOO_LARGE)
     - `RequestErrorCode`: 3 codes (METHOD_NOT_FOUND, INVALID_PARAMS, INTERNAL_ERROR) -- **NOTE: Does NOT include REQUEST_TIMEOUT/REQUEST_CANCELLED/REQUEST_ABORTED**
     - `GatewayErrorCode`: 8 codes (AGENT_NOT_FOUND, AGENT_BUSY, NODE_NOT_FOUND, NODE_OFFLINE, SESSION_NOT_FOUND, SESSION_EXPIRED, PERMISSION_DENIED, QUOTA_EXCEEDED)
     - `ReconnectErrorCode`: 3 codes (MAX_RECONNECT_ATTEMPTS, MAX_AUTH_RETRIES, RECONNECT_DISABLED)
   - **Rewrite `NewAPIError` matching logic** (match TS `createErrorFromResponse`):
     ```go
     func NewAPIError(shape *protocol.ErrorShape) error {
         code := strings.ToUpper(shape.Code)

         // 1. Auth errors: AUTH_* or CHALLENGE_* (prefix)
         if strings.HasPrefix(code, "AUTH_") || strings.HasPrefix(code, "CHALLENGE_") {
             return NewAuthError(shape.Code, shape.Message, shape.Retryable, shape.Details)
         }

         // 2. Connection errors: CONNECTION_* or TLS_FINGERPRINT_MISMATCH (prefix + exact)
         if strings.HasPrefix(code, "CONNECTION_") || code == "TLS_FINGERPRINT_MISMATCH" {
             return NewConnectionError(shape.Code, shape.Message, shape.Retryable, shape.Details)
         }

         // 3. Protocol errors: PROTOCOL_* (prefix)
         if strings.HasPrefix(code, "PROTOCOL_") {
             return NewProtocolError(shape.Code, shape.Message, shape.Retryable, shape.Details)
         }

         // 4. Request errors: exact match only
         if code == "METHOD_NOT_FOUND" || code == "INVALID_PARAMS" || code == "INTERNAL_ERROR" {
             return NewRequestError(shape.Code, shape.Message, shape.Retryable, shape.Details)
         }

         // 5. All other codes (including REQUEST_TIMEOUT, REQUEST_CANCELLED, REQUEST_ABORTED)
         //    fall through to GatewayError -- matching TS behavior
         return NewGatewayError(shape.Code, shape.Message, shape.Retryable, shape.Details)
     }
     ```
   - Add `RetryableError` interface
   - Add `APIError` struct holding full ErrorShape
   - Add type guards: `IsAuthError`, `IsConnectionError`, `IsProtocolError`, `IsRequestError`, `IsGatewayError`
   - Add `TimeoutError`, `CancelledError`, `AbortError`, `GatewayError`, `ReconnectError` types

2. **`pkg/types/errors_test.go`** -- Rewrite
   - Test error code mapping (37 codes)
   - **Key boundary tests**:
     - `AUTH_TOKEN_EXPIRED` -> AuthError (prefix match)
     - `CHALLENGE_EXPIRED` -> AuthError (prefix match)
     - `CONNECTION_TIMEOUT` -> ConnectionError (prefix match)
     - `TLS_FINGERPRINT_MISMATCH` -> ConnectionError (exact match, NOT prefix)
     - `PROTOCOL_NEGOTIATION_FAILED` -> ProtocolError (prefix match)
     - `METHOD_NOT_FOUND` -> RequestError (exact match)
     - `INVALID_PARAMS` -> RequestError (exact match)
     - `INTERNAL_ERROR` -> RequestError (exact match)
     - `REQUEST_TIMEOUT` -> GatewayError (**NOT RequestError -- fall through**)
     - `REQUEST_CANCELLED` -> GatewayError (**NOT RequestError -- fall through**)
     - `REQUEST_ABORTED` -> GatewayError (**NOT RequestError -- fall through**)
     - `UNKNOWN_ERROR_CODE` -> GatewayError (fallback)
     - `auth_token_expired` (lowercase) -> AuthError (case insensitive)
   - Test Retryable/Details fields
   - Test `errors.Is`/`errors.As` compatibility

---

### Phase 5: Connection Protocol Types

**Status**: Pending implementation

**TypeScript Source Verification**:
- `src/protocol/connection.ts` -- `ConnectParams`, `HelloOk`, `Snapshot`, `Policy`
- `src/connection/protocol.ts` -- `ProtocolNegotiator`
- `src/connection/policies.ts` -- `PolicyManager`

**File Changes**:

1. **`pkg/connection/connection_types.go`** -- New file
   - `ConnectParams`: `{ MinProtocol, MaxProtocol, Client{ID, DisplayName, Version, Platform, DeviceFamily, ModelIdentifier, Mode, InstanceID}, Caps, Commands, Permissions, PathEnv, Role, Scopes, Device?, Auth?, Locale, UserAgent }`
   - `HelloOk`: `{ Type:"hello-ok", Protocol:int, Server{Version, ConnID}, Features{Methods, Events}, Snapshot, CanvasHostUrl?, Auth{DeviceToken, Role, Scopes, IssuedAtMs?}, Policy }`
   - `Snapshot`: `{ Presence, Health, StateVersion, UptimeMs, ConfigPath?, AuthMode?, Agents?, Nodes? }`
   - `Policy`: `{ MaxPayload, MaxBufferedBytes, TickIntervalMs }`
   - `PresenceEntry`: `{ Node, ... }`

2. **`pkg/connection/protocol.go`** -- Rewrite
   - `ProtocolNegotiator`: `{ range{min,max} int, negotiatedVersion int }`
   - `Negotiate(helloOk *HelloOk) (*NegotiatedProtocol, error)`: validate server.protocol in [min,max] range
   - Default: min=3, max=3
   - `Reset()`: clear negotiatedVersion

3. **`pkg/connection/policies.go`** -- Rewrite
   - `PolicyManager`: store `{ policy Policy, hasSetPolicy bool }`
   - `SetPolicies(policy *Policy)`: set policy
   - `HasPolicy() bool`: check if set
   - `GetMaxPayload() int64`, `GetMaxBufferedBytes() int64`, `GetTickIntervalMs() int64`
   - Defaults: maxPayload=1048576, maxBufferedBytes=65536, tickIntervalMs=30000

4. **`pkg/connection/state.go`** -- Rewrite
   - Update `validTransitions` map to match TypeScript:
     ```
     disconnected -> connecting
     connecting -> handshaking, disconnected, closed
     handshaking -> authenticating, reconnecting, disconnected, closed
     authenticating -> ready, reconnecting, disconnected, closed
     ready -> reconnecting, disconnected, closed
     reconnecting -> connecting, handshaking, authenticating, ready, disconnected, closed
     closed -> disconnected
     ```
   - Add `IsConnected()`, `IsReady()`, `Reset()` methods
   - Add `OnChange(listener)`, `OnListenerError(handler)` methods

---

### Phase 6.1: Client Core Rewrite

**Status**: Pending implementation

**TypeScript Source Verification**:
- `src/client.ts` (lines 1-720) -- Core connection and request logic

**File Changes**:

1. **`pkg/types/types.go`** -- Expand
   - Add `ConnectionState` constants: `StateHandshaking`, `StateReady`, `StateClosed` (replace/supplement existing)
   - Approach: maintain backward compatibility, add new states with unified mapping

2. **`pkg/client.go`** -- Major expansion (~300 lines changes)
   - **Expand ClientConfig**:
     - Add `ClientID`, `ClientVersion`, `Platform`, `Mode`, `InstanceID`
     - Add `Auth: { Token?, BootstrapToken?, DeviceToken?, Password? }`
     - Add `Device: { ID, PublicKey, Signature, SignedAt, Nonce }`
     - Add `Capabilities []string`
     - Add `TickMonitorConfig`, `GapDetectorConfig`
     - Add `RequestTimeoutMs`, `ConnectTimeoutMs`
     - Add `AutoReconnect`, `MaxReconnectAttempts`, `ReconnectDelayMs`
     - Add `Logger`

   - **Expand client struct**:
     - Add `tickMonitor *events.TickMonitor`
     - Add `gapDetector *events.GapDetector`
     - Add `protocolNegotiator *connection.ProtocolNegotiator`
     - Add `policyManager *connection.PolicyManager`
     - Add `serverInfo *connection.HelloOk`
     - Add `snapshot *connection.Snapshot`
     - Add `requestIDPrefix string`

   - **New Option functions**:
     - `WithClientID`, `WithClientVersion`, `WithPlatform`, `WithMode`
     - `WithAuth`, `WithDevice`, `WithCapabilities`
     - `WithTickMonitor`, `WithGapDetector`
     - `WithRequestTimeout`, `WithConnectTimeout`
     - `WithAutoReconnect`, `WithMaxReconnectAttempts`, `WithReconnectDelay`

   - **Rewrite Connect()**:
     - Build `ConnectParams` (minProtocol=3, maxProtocol=3)
     - Send connect handshake
     - Parse `HelloOk` response
     - Call `protocolNegotiator.Negotiate()`
     - Call `policyManager.SetPolicies()`
     - Optionally: create `TickMonitor`, create `GapDetector`

   - **Rewrite SendRequest()**:
     - Generate request ID: `req-${cryptoRandHex(16)}`
     - Build `{ type:"req", id, method, params }` JSON
     - Send and wait for response
     - Handle `response.ok=false` -> `NewAPIError(response.error)`
     - Support `onProgress` callback

   - **New methods** (not API namespace):
     - `Request(ctx, method, params, opts) (interface{}, error)`
     - `ServerInfo() *connection.HelloOk`
     - `Snapshot() *connection.Snapshot`
     - `Protocol() int`
     - `GetTickMonitor() *events.TickMonitor`
     - `GetGapDetector() *events.GapDetector`
     - `GetPolicy() *connection.Policy`
     - `OnStateChange(handler) func()`
     - `OnError(handler) func()`
     - `OnMessage(handler) func()`
     - `OnClosed(handler) func()`
     - `On(pattern, handler) func() unsubscribe`
     - `Once(pattern, handler) func() unsubscribe`
     - `Off(pattern?, handler?)`
     - `Abort(requestID string)`
     - `IsConnected() bool`
     - `ConnectionState() ConnectionState`

   - **Rewrite Disconnect()**:
     - Stop TickMonitor
     - Reset GapDetector
     - Clean up state

   - **Acceptance criteria**:
     - [ ] All methods compile
     - [ ] Connect/Disconnect flow tests pass
     - [ ] Request flow tests pass
     - [ ] Event subscribe/unsubscribe tests pass

---

### Phase 6.2: API Namespaces Integration

**Status**: Pending implementation

**TypeScript Source Verification**:
- `src/client.ts` (lines 178-595) -- API namespace getters

**Key Difference (Decision 2)**:
- TypeScript uses **property getters**: `client.chat`, `client.agents`, etc.
- Go does not support property getters, uses **methods**: `client.Chat()`, `client.Agents()`, etc.
- **This is an intentional syntax deviation** from TypeScript: Go idiom first, no implicit state, methods more idiomatic.

**File Changes**:

1. **`pkg/client.go`** -- Expand (on top of Phase 6.1)
   - Add API namespace instances to client struct:
     ```go
     chat          *api.ChatAPI
     agents        *api.AgentsAPI
     sessions      *api.SessionsAPI
     config        *api.ConfigAPI
     cron          *api.CronAPI
     nodes         *api.NodesAPI
     skills        *api.SkillsAPI
     devicePairing *api.DevicePairingAPI
     ```

   - Add API namespace constructor calls (in `NewClient`):
     ```go
     requestFn := func(ctx context.Context, method string, params interface{}) (interface{}, error) {
         return c.Request(ctx, method, params, nil)
     }
     c.chat = api.NewChatAPI(requestFn)
     c.agents = api.NewAgentsAPI(requestFn)
     // ... remaining namespaces
     ```

   - Add namespace accessor methods (Go idiomatic syntax):
     ```go
     func (c *client) Chat() *api.ChatAPI          { return c.chat }
     func (c *client) Agents() *api.AgentsAPI       { return c.agents }
     func (c *client) Sessions() *api.SessionsAPI   { return c.sessions }
     func (c *client) Config() *api.ConfigAPI       { return c.config }
     func (c *client) Cron() *api.CronAPI           { return c.cron }
     func (c *client) Nodes() *api.NodesAPI         { return c.nodes }
     func (c *client) Skills() *api.SkillsAPI       { return c.skills }
     func (c *client) DevicePairing() *api.DevicePairingAPI { return c.devicePairing }
     ```

   - **Acceptance criteria**:
     - [ ] 8 namespace accessor methods exist
     - [ ] Each namespace has at least one method mock test
     - [ ] Uses `client.Chat()` not `client.Chat`

2. **`pkg/api/api_test.go`** -- New file
   - Mock tests for each namespace
   - Test requestFn correctly passed to each API

---

### Phase 7: RequestManager Enhancement

**Status**: Pending implementation

**File Changes**:

1. **`pkg/managers/request.go`** -- Rewrite
   - Add `progressCb map[string]func(interface{})`
   - Add `progressMu sync.Mutex`
   - Add `ResolveProgress(requestID string, payload interface{})`
   - Modify `SendRequest`: support progress callback, use `req.ID` not `req.RequestID`
   - Add `AbortRequest(requestID string)`
   - Add `RejectRequest(requestID string, err error)` (for server-initiated cancellation)
   - Add `Clear()` (clear all pending on disconnect)

2. **`pkg/managers/request_test.go`** -- Expand
   - Test progress callback
   - Test request cancellation
   - Test ID field

---

### Phase 8: TickMonitor Enhancement

**Status**: Pending implementation

**TypeScript Source Verification**:
- `src/events/tick.ts` -- `TickMonitorConfig`, `staleMultiplier`, `onStale`, `onRecovered`

**File Changes**:

1. **`pkg/events/tick.go`** -- Rewrite
   - Add `TickMonitorConfig` struct: `{ TickIntervalMs, StaleMultiplier?, OnStale?, OnRecovered?, GetTime? }`
   - Add fields: `staleMultiplier`, `lastTickTime`, `staleDetected`, `staleStartTime`, `getTime func()`
   - Add `IsStale() bool` (pure query, no side effects)
   - Add `CheckStale() bool` (trigger stale callbacks)
   - Add `GetTimeSinceLastTick() time.Duration`
   - Add `GetStaleDuration() time.Duration`
   - Add `GetStatus() TickStatus`
   - Add `RecordTick(ts time.Time)` method
   - Constructor changed to `NewTickMonitor(config TickMonitorConfig) (*TickMonitor, error)`

2. **`pkg/events/tick_test.go`** -- Expand
   - Test staleMultiplier logic
   - Test onStale/onRecovered callbacks
   - Test getTime injection

---

### Phase 9: GapDetector Enhancement

**Status**: Pending implementation

**TypeScript Source Verification**:
- `src/events/gap.ts` -- `GapRecoveryMode`, `GapInfo`, `GapRecoveryConfig`, `GapDetectorConfig`

**File Changes**:

1. **`pkg/events/gap.go`** -- Rewrite
   - Add `GapRecoveryMode` type: `"reconnect" | "snapshot" | "skip"`
   - Add `GapInfo` struct: `{ Expected, Received, DetectedAt }`
   - Add `GapRecoveryConfig` struct: `{ Mode, OnGap?, SnapshotEndpoint? }`
   - Add `GapDetectorConfig` struct: `{ Recovery, MaxGaps? }`
   - Rewrite `GapDetector` struct: `{ recovery, maxGaps, lastSequence, gaps, mu }`
   - Constructor changed to `NewGapDetector(config GapDetectorConfig) *GapDetector`
   - Add `RecordSequence(seq uint64)` (replace `Record`)
   - Add `HasGap() bool`
   - Add `GetGaps() []GapInfo` (return copy)
   - Add `GetLastSequence() uint64`
   - Internal: deferred side-effects pattern

---

### Phase 10: Integration Tests

**Status**: Pending implementation

**File Changes**:

1. **`pkg/integration_test.go`** -- Expand
   - Test new protocol format serialization/deserialization
   - Test API namespace methods (mock-based)
   - Test progress flow
   - Test error mapping
   - Test TickMonitor integration
   - Test GapDetector integration
   - Test protocol negotiation
   - Test PolicyManager
   - **Wire compatibility verification**: Use Phase 1.5 snapshot fixtures for end-to-end tests

2. **`pkg/api/api_test.go`** -- Expand (created in Phase 6.2)

---

## Rollback Strategy

### Rollback Trigger Conditions

After Phase 1, if the following validations fail, trigger rollback:

1. **Wire compatibility test failure** -- Phase 1.5 JSON snapshot test does not pass
2. **Real server incompatibility** -- Cannot establish connection with TypeScript version Gateway

### Rollback Execution Steps

1. **Preserve current code**: Do not delete, switch to new branch
   ```bash
   git checkout -b migration-wire-protocol  # skip if already created
   git stash  # stash Phase 1 changes
   ```

2. **Restore old protocol**: Switch back to main branch
   ```bash
   git checkout main
   ```

3. **Protocol compatibility investigation**:
   - Analyze specific differences between Go serialization output and TypeScript expectation
   - Determine if it's field naming, type serialization, or structural difference
   - Possible fixes:
     - JSON tag adjustment (field rename)
     - Type mapping (uint64 vs number)
     - Wrapper layer handling (whether to keep GatewayFrame)

4. **Revise Phase 1**: Based on analysis, adjust protocol layer implementation
   - Re-run Phase 1.5 tests
   - Proceed with M1 subsequent phases after passing

5. **Decision record**:
   - If TypeScript or Go has a bug, record as ADR
   - If protocol itself has multi-version compatibility requirements, update protocol negotiation logic

### Phase 1.5 Gate

Phase 1.5 is a **mandatory gate test**. If any test fails:
- Immediately stop M1 subsequent work
- Execute rollback strategy
- Do not proceed to Phase 2, 3, 4, etc.

---

## Pre-Mortem (5 Failure Scenarios)

### Scenario 1: Server incompatible with new protocol format
**Risk**: Gateway server still uses old protocol (type=gateway, requestId, success, result)
**Impact**: All request/response parsing fails
**Mitigation**:
- Phase 1.5 must pass to continue
- Add protocol version negotiation supporting old format fallback
- Add protocol detection logic in `pkg/connection/protocol.go`

### Scenario 2: API param type mismatch with Gateway
**Risk**: Subtle differences in field names/types between Go and TypeScript types causing runtime errors
**Impact**: Specific API calls fail
**Mitigation**:
- After Phase 2, verify field-by-field against TypeScript
- Write automated tests comparing TS and Go SDK results with identical params
- Use `json:"fieldName"` tags ensuring exact field name matching

### Scenario 3: Concurrency/race conditions in event handling chain
**Risk**: TypeScript uses EventEmitter, Go uses channel+mutex, different state update order
**Impact**: Events lost, duplicated, or out of order
**Mitigation**:
- Strictly follow "send channel outside lock" pattern
- Critical: GapDetector.RecordSequence and TickMonitor.RecordTick must be thread-safe
- Write concurrency stress tests in integration_test.go

### Scenario 4: Memory leaks in reconnect cycles
**Risk**: tick handler/gap detector not properly cleaned up during reconnect
**Impact**: goroutine leak, file descriptor exhaustion
**Mitigation**:
- Disconnect() must call tickMonitor.Stop() and gapDetector.Reset()
- Add tickHandlerUnsubscribe cleanup in client.go Disconnect()
- Write multiple connect/disconnect cycle tests in integration_test.go

### Scenario 5: Error code mismatch causing wrong error type inference
**Risk**: Go SDK returns error code not in TS expected 37 codes
**Impact**: `createErrorFromResponse` falls back to GatewayError, may lose semantics
**Mitigation**:
- After Phase 4, verify all 37 codes match TS
- Unknown code falls back to `GatewayError` (matches TS behavior)
- Write tests verifying unknown code -> GatewayError fallback

---

## Verification Steps

```bash
# Phase 1.5 Gate (must pass)
go test ./pkg/protocol/... -v -run "WireBytes|Snapshot"

# All phases
go build ./...
go test -cover ./...
golangci-lint run
pre-commit run --all-files

# Coverage target: 80%+
```

---

## Migration Order & Dependencies

```
Phase 1 (Protocol) ─── Phase 1.5 (Wire Test Gate) ──┐
                                                    │
Phase 2 (API Params) ──────────────────────────────┼── Phase 3 (API Namespaces)
                                                    │
Phase 5 (Connection Types) ────────────────────────┤
                                                    │
Phase 8 (TickMonitor) ─────────────────────────────┤
                                                    │
Phase 9 (GapDetector) ─────────────────────────────┤
                                                    │
Phase 7 (RequestManager) ──────────────────────────┼── Phase 6.1 (Client Core)
                                                    │
Phase 4 (Error System) ────────────────────────────┤
                                                    │
                                                   Phase 6.2 (API Namespaces Integration)
                                                    │
                                                   Phase 10 (Integration Tests)
```

**After Phase 1.5 passes, M1 and M2 can proceed in parallel**:
- **M1 path**: Phase 5 -> Phase 7 -> Phase 8 -> Phase 9 -> Phase 6.1
- **M2 path**: Phase 2 -> Phase 3 -> Phase 6.2

**Phase 6.1 and Phase 6.2 can be implemented in parallel** (client core and API namespaces are independent).

---

## Key Decisions

### Decision 1: Wire Compatibility vs Go Idioms

| Choice | Approach A: Full Compatibility | Approach B: Internal Compatibility |
|--------|------------------------------|-----------------------------------|
| RequestFrame | `{ type:"req", id, method, params }` | Same (JSON wire) |
| ResponseFrame | `{ type:"res", id, ok, payload, error, progress }` | Same |
| EventFrame | `{ type:"event", event, payload, seq, stateVersion }` | Same |
| Internal Go struct | Directly use wire format | Add internal high-level types |

**Selected: Approach A** -- Use TypeScript wire format directly as Go structs for Gateway communication.

### Decision 2: API Namespace Access Method

| Choice | Approach A: Methods | Approach B: Properties (TS) |
|--------|-------------------|----------------------------|
| Syntax | `client.Chat()` | `client.Chat` |
| Go idiom | Higher (no implicit state) | Not supported |
| Consistency | Consistent with `client.State()` etc. | N/A |

**Selected: Approach A** -- Go does not support property getters, method calls are more idiomatic.

**Deviation note**: TypeScript SDK uses `get chat()`, `get agents()` etc. property getters (`src/client.ts:502-595`). Go uses `client.Chat()`, `client.Agents()` methods - this is an **intentional syntax deviation** because:
1. Go does not support property getters
2. Methods return the same instance on each call, no implicit state issues
3. Maintains consistency with existing APIs like `client.State()`, `client.Events()`

### Decision 3: ConnectionState Compatibility

| Choice | Approach A: Replace | Approach B: Supplement+Map |
|--------|--------------------|---------------------------|
| Existing constants | Delete | Keep+add TS states |
| Breaking changes | Yes | No |

**Selected: Approach B** -- Maintain backward compatibility, add TypeScript states.

---

## ADR: Error Code Source

**Decision**: The 37 error codes in Phase 4 error system all come from TypeScript SDK `src/errors.ts`.

**Drivers**:
1. Maintain consistency with TypeScript SDK error type system
2. Ensure `createErrorFromResponse` logic is equivalent in Go and TS
3. Avoid creating non-existent error codes

**Alternatives considered**:
- Extract codes from Gateway server documentation: documentation does not exist or is outdated
- Observe codes from actual traffic: may miss edge cases
- Extract from TS source: only authoritative source

**Why chosen**: TypeScript SDK source is the only authoritative specification source, containing complete `createErrorFromResponse` matching logic.

**Consequences**:
- Go SDK error codes 100% aligned with TS
- `pkg/types/errors.go` needs major refactoring to support detailed classification
- `pkg/protocol/errors.go` has 5 additional wire-level codes (from `src/protocol/errors.ts`)

**Follow-ups**:
- If TS SDK adds new error codes, Go SDK must sync
- Consider centralized error code management for version alignment

---

## Appendix: TypeScript Source Index

For verifying phase source files.

| Verification Item | TypeScript Source Path | Lines/Notes |
|-------------------|----------------------|-------------|
| Wire Frame format | `src/protocol/frames.ts` | GatewayFrame, RequestFrame, ResponseFrame, EventFrame |
| API param types | `src/protocol/api-params.ts` | 142 types, 546 lines |
| API shared types | `src/protocol/api-common.ts` | AgentSummary, WizardStep, etc. |
| Connection types | `src/protocol/connection.ts` | ConnectParams, HelloOk, Snapshot, Policy |
| Protocol negotiation | `src/connection/protocol.ts` | ProtocolNegotiator |
| Policy management | `src/connection/policies.ts` | PolicyManager |
| Error types (SDK) | `src/errors.ts` | 37 error codes, 8 error classes |
| Error types (Wire) | `src/protocol/errors.ts` | 5 error codes: NOT_LINKED, NOT_PAIRED, AGENT_TIMEOUT, INVALID_REQUEST, UNAVAILABLE |
| Client implementation | `src/client.ts` | OpenClawClient class, 1037 lines |
| Chat API | `src/api/chat.ts` | ChatAPI class |
| Agents API | `src/api/agents.ts` | AgentsAPI class |
| Sessions API | `src/api/sessions.ts` | SessionsAPI class |
| Config API | `src/api/config.ts` | ConfigAPI class |
| Cron API | `src/api/cron.ts` | CronAPI class |
| Nodes API | `src/api/nodes.ts` | NodesAPI class |
| Skills API | `src/api/skills.ts` | SkillsAPI class |
| Device Pairing API | `src/api/devicePairing.ts` | DevicePairingAPI class |
| TickMonitor | `src/events/tick.ts` | TickMonitorConfig, staleMultiplier |
| GapDetector | `src/events/gap.ts` | GapRecoveryMode, GapInfo, GapRecoveryConfig |
