# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OpenClaw SDK Go is a Go implementation of the OpenClaw WebSocket SDK, migrated from TypeScript. It provides a feature-complete WebSocket client with connection management, event handling, request/response patterns, and automatic reconnection.

**Module Path**: `github.com/frisbee-ai/openclaw-sdk-go`

## Development Commands

### Build and Test
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./pkg/transport

# Run single test
go test -run TestWebSocketTransportDial ./pkg/transport

# Build the module
go build ./...
```

### Code Quality
```bash
# Format code
gofmt -w -l .

# Static analysis
go vet ./...

# Lint with golangci-lint
golangci-lint run

# Run all pre-commit hooks manually
pre-commit run --all-files
```

### Dependencies
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Architecture

### High-Level Structure

```
pkg/
├── api/              # API types and helpers
├── client.go         # Main client API with option pattern configuration
├── types/             # Shared types (ConnectionState, Event, errors, Logger)
├── auth/              # Authentication (CredentialsProvider, AuthHandler)
├── transport/         # WebSocket transport layer (gorilla/websocket)
├── protocol/          # Protocol frames (RequestFrame, ResponseFrame, validation)
├── connection/        # Connection state machine, policies, TLS validator
├── events/            # Tick monitor, gap detector
├── managers/          # High-level managers (event, request, connection, reconnect)
└── utils/             # Timeout manager
```

### Key Design Patterns

**1. Option Pattern**: Client configuration uses functional options (`WithURL()`, `WithAuthHandler()`, etc.) for flexible, readable construction.

**2. Context + Channel Hybrid**:
   - `context.Context` for lifecycle management and cancellation
   - Buffered channels for event delivery (prevents deadlocks)
   - **Critical Rule**: Never send to a channel while holding a lock

**3. Graceful Shutdown**: All managers implement `Close()` with proper goroutine cleanup using `sync.WaitGroup`.

**4. Re-export Pattern**: `pkg/client.go` re-exports types from subpackages for convenience, allowing users to import a single package.

### Manager Coordination

The `client` struct coordinates four managers:
- **EventManager**: Pub/sub event dispatch with channel-based delivery
- **RequestManager**: Correlates requests with responses using pending request map
- **ConnectionManager**: Wraps transport, manages state transitions
- **ReconnectManager**: Fibonacci backoff reconnection (optional)

### Thread-Safety Patterns

- **State Machine**: `ConnectionStateMachine` uses `sync.RWMutex`, releases lock before sending to events channel
- **Client Methods**: Top-level operations (Connect, Disconnect, SendRequest) use `sync.Mutex` for exclusive access
- **Channel Safety**: All channels used by managers are buffered to prevent blocking sends

### Protocol Layer

The SDK uses JSON-serialized frames:
- `RequestFrame`: Outbound requests with `RequestID`, `Action`, `Payload`
- `ResponseFrame`: Inbound responses correlated by `RequestID`
- `GatewayFrame`: Gateway messages (events, errors)

## Testing Strategy

- **Unit tests**: Every package has `*_test.go` files
- **Coverage target**: 80%+
- **TDD workflow**: Write tests first (RED), implement (GREEN), refactor (IMPROVE)
- **Table-driven tests**: Preferred for multiple test cases

## Pre-commit Hooks

The project uses pre-commit hooks configured in `.pre-commit-config.yaml`:
- `go-fmt`: Format code
- `go-vet`: Static analysis
- `golangci-lint`: Fast Go linters with auto-fix
- `go-test`: Run all tests

## Key Files

| File | Purpose |
|------|---------|
| `pkg/client.go` | Main SDK entry point, exports all public types |
| `pkg/client_test.go` | Client integration tests |
| `pkg/connection/` | Connection state machine and TLS validation |
| `pkg/managers/` | Event, request, connection, reconnect managers |

## Gotchas

- **Channel + Lock Rule**: Never send to a channel while holding a lock. Release lock BEFORE sending.
- **Buffered Channels**: All event channels are buffered to prevent deadlocks.
- **Feature Parity**: This SDK is functionally equivalent to TypeScript SDK but not API-identical.

## Migration Context

This is a Go migration of `openclaw-sdk-typescript`. The design document at `docs/specs/2026-03-18-typescript-to-go-migration-design.md` details architectural decisions. Phase plans in `docs/plans/` document the incremental implementation strategy.

## Important Notes

- **Go version**: 1.21+
- **External dependencies**: Minimal, only `github.com/gorilla/websocket`
- **Standard library preferred**: Uses `net/http`, `context`, `sync`, `crypto/tls` from stdlib
- **Feature parity**: Functionally equivalent to TypeScript SDK, not API-identical

<!-- GSD:project-start source:PROJECT.md -->
## Project

**OpenClaw SDK Go**

OpenClaw SDK Go is a feature-complete WebSocket client library for Go, migrated from TypeScript. It provides a robust SDK for connecting to the OpenClaw gateway with connection management, event handling, request/response patterns, and automatic reconnection. End users are Go applications that need to interact with the OpenClaw platform via WebSocket.

**Core Value:** **Go developers can integrate the OpenClaw platform in under 10 lines of code** — the SDK handles connection lifecycle, authentication, protocol framing, event dispatch, and reconnection transparently.

### Constraints

- **Go 1.21+ runtime** — No breaking changes to Go compatibility
- **No CGO** — Pure Go, no external C dependencies
- **Minimal dependencies** — Only `gorilla/websocket`; stdlib preferred
- **Library distribution** — GoReleaser configured for library mode (no binaries)
- **API compatibility** — Once v1.0.0 released, breaking changes require major version bump
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.24 - Core SDK implementation
## Runtime
- Go standard runtime (no external runtime dependencies)
- Pure Go - no CGO required
- Go modules (go.mod/go.sum)
- Lockfile: present
## Frameworks
- No framework - pure library SDK
- WebSocket: `github.com/gorilla/websocket v1.5.3` - WebSocket client implementation
- Go testing package (built-in `testing.T`)
- Fuzz testing via `testing.F` in `pkg/protocol/fuzz_test.go`
- GoReleaser v2 - Release automation for library distribution
- pre-commit hooks - Local hooks for format/vet/lint/test
## Key Dependencies
- `github.com/gorilla/websocket v1.5.3` - WebSocket protocol implementation
## Configuration
- No runtime environment configuration
- SDK is configured programmatically via option pattern
- `.golangci.yaml` - Linter configuration (excludes test files from errcheck/unused)
- `.goreleaser.yaml` - Library release configuration (no binary builds)
- `.pre-commit-config.yaml` - Local hooks: gofmt, go vet, golangci-lint, go test
## Platform Requirements
- Go 1.24+
- golangci-lint (for linting)
- pre-commit (optional, for local hooks)
- Go 1.21+ (runtime compatibility)
- No platform-specific requirements
## Tooling
| Tool | Version/Config | Purpose |
|------|---------------|---------|
| gofmt | built-in | Code formatting |
| go vet | built-in | Static analysis |
| golangci-lint | latest via GitHub Action | Fast linting with auto-fix |
| GoReleaser | v2 | Library release automation |
| codecov | v6 | Coverage reporting |
## CI/CD Pipeline
- `['1.24']` - Only Go 1.24 tested in CI
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Go source files: lowercase with underscores (`websocket_test.go`, `event_manager.go`)
- Test files: `*_test.go` suffix co-located with source
- Package name matches directory name
- Exported functions: `PascalCase` (`NewClient`, `SendRequest`)
- Unexported functions: `camelCase` (`buildConnectParams`, `processServerInfo`)
- Test functions: `PascalCase` with descriptive names (`TestEventManager_Subscribe`)
- Local variables: `camelCase` (`connectParams`, `authHandler`)
- Constants: `SCREAMING_SNAKE_CASE` for actual constants (`StateDisconnected`, `EventConnect`)
- Error variables: `err` prefix (`err`, `errCh`)
- Channel variables: `Ch` suffix or descriptive (`done`, `errCh`, `recvCh`)
- Structs: `PascalCase` (`ClientConfig`, `EventManager`)
- Interfaces: `PascalCase` with `er` suffix where idiomatic (`OpenClawClient`, `Transport`)
- Type aliases: preserve pattern of underlying type
## Code Style
- Tool: `gofmt` (runs automatically via pre-commit hook)
- Config: `.golangci.yaml` with `golangci-lint`
- Pre-commit hook: `gofmt -w -l .`
- Tool: `golangci-lint run --fix --timeout=5m`
- Exclusions: Test files (`_test.go`) exclude `errcheck` and `unused` linters
- Config: `.golangci.yaml`
- Standard Go formatting (gofmt handles this)
- No explicit limit; gofmt handles wrapping
## Import Organization
## Error Handling
- `BaseError` - core error with code, message, retryable flag
- `OpenClawError` interface - `Code()`, `Message()`, `Retryable()`, `Unwrap()`, `Details()`
- Specific error types: `AuthError`, `ConnectionError`, `ProtocolError`, `RequestError`, `GatewayError`, `ReconnectError`, `TimeoutError`, `CancelledError`, `AbortError`
- `NewAPIError(*ErrorShape)` - factory that routes error codes to correct type
## Logging
- `Debug(msg string, keyvals ...any)`
- `Info(msg string, keyvals ...any)`
- `Warn(msg string, keyval ...any)`
- `Error(msg string, keyvals ...any)`
- `DefaultLogger` - structured logging to stdout
- `NopLogger` - no-op implementation
- `WithContext`/`FromContext` for context-scoped loggers
## Comments
- Exported functions and types: package-level doc comments
- Non-obvious logic: inline comments explaining WHY
- Bug workarounds: comment explaining the issue
## Function Design
- Context as first parameter when used: `func(ctx context.Context, ...)`
- Options pattern for configuration: `func WithURL(url string) ClientOption`
- Error return: `error` as last return value
- Named returns only when clarity benefits
- Interface return types for flexibility
## Module Design
- Single package re-export pattern in `pkg/client.go` for convenience
- Types re-exported from subpackages: `type ConnectionState = types.ConnectionState`
- `pkg/client.go` acts as main entry point re-exporting all public types
- `pkg/` contains all implementation packages
- API subpackages: `pkg/api/` for API method groups
## Thread-Safety Patterns
- `sync.RWMutex` for connection state
- Release lock BEFORE sending to channels
- All event channels are buffered to prevent deadlocks
- Default buffer size: 100
## Context Usage
## Graceful Shutdown
## Pre-commit Hooks
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## Pattern Overview
- **Option Pattern**: Client configuration via functional options (`WithURL()`, `WithAuthHandler()`, etc.)
- **Context + Channel Hybrid**: `context.Context` for lifecycle/cancellation, buffered channels for event delivery
- **Manager Coordination**: Four managers (event, request, connection, reconnect) coordinated by a single `client` struct
- **Graceful Shutdown**: All managers implement `Close()` with proper goroutine cleanup via `sync.WaitGroup`
## Layers
- Purpose: Main public API, client factory, manager coordination
- Location: `pkg/client.go`
- Contains: `OpenClawClient` interface, `client` struct, `NewClient()`, all API namespace accessors
- Depends on: All subpackages
- Used by: End users
- Purpose: Typed API method wrappers per domain (chat, agents, sessions, etc.)
- Location: `pkg/api/*.go`
- Contains: `ChatAPI`, `AgentsAPI`, `SessionsAPI`, `ConfigAPI`, `CronAPI`, `NodesAPI`, `SkillsAPI`, `DevicePairingAPI`, `BrowserAPI`, `ChannelsAPI`, `PushAPI`, `ExecApprovalsAPI`, `SystemAPI`, `SecretsAPI`, `UsageAPI`
- Depends on: `pkg/protocol` for types, `pkg/api` for `RequestFn`
- Used by: `pkg/client.go`
- Purpose: Wire protocol frame definitions, validation, and API method signatures
- Location: `pkg/protocol/types.go`, `pkg/protocol/validation.go`, `pkg/protocol/api_*.go`
- Contains: `RequestFrame`, `ResponseFrame`, `EventFrame`, `ErrorShape`, `StateVersion`, API param/result types
- Depends on: Standard library only
- Used by: `pkg/api/`, `pkg/managers/request.go`
- Purpose: High-level coordination for events, requests, connections, reconnection
- Location: `pkg/managers/event.go`, `pkg/managers/request.go`, `pkg/managers/connection.go`, `pkg/managers/reconnect.go`
- Contains: `EventManager`, `RequestManager`, `ConnectionManager`, `ReconnectManager`
- Depends on: `pkg/transport`, `pkg/connection`, `pkg/types`, `pkg/events`
- Used by: `pkg/client.go`
- Purpose: Connection state machine, handshake types, TLS validation, policy management
- Location: `pkg/connection/state.go`, `pkg/connection/connection_types.go`, `pkg/connection/policies.go`, `pkg/connection/tls.go`
- Contains: `ConnectionStateMachine`, `ConnectParams`, `HelloOk`, `Snapshot`, `Policy`, `TLSConfig`
- Depends on: `pkg/types`
- Used by: `pkg/managers/connection.go`
- Purpose: Low-level WebSocket I/O
- Location: `pkg/transport/websocket.go`
- Contains: `WebSocketTransport`, `Transport` interface, `WebSocketConfig`, `TLSConfig`
- Depends on: `github.com/gorilla/websocket`, `pkg/connection` (for TLS config)
- Used by: `pkg/managers/connection.go`
- Purpose: Connection health monitoring and gap detection
- Location: `pkg/events/tick.go`
- Contains: `TickMonitor`, `GapDetector`
- Depends on: Standard library only
- Used by: `pkg/client.go`
- Purpose: Authentication handler and credentials provider interfaces
- Location: `pkg/auth/handler.go`, `pkg/auth/provider.go`
- Contains: `AuthHandler`, `CredentialsProvider`, `StaticAuthHandler`, `StaticCredentialsProvider`
- Depends on: Standard library only
- Used by: `pkg/client.go` (via `ClientConfig.CredentialsProvider`)
- Purpose: Shared core types (connection states, event types, errors, logging)
- Location: `pkg/types/types.go`, `pkg/types/errors.go`, `pkg/types/logger.go`
- Contains: `ConnectionState`, `EventType`, `Event`, `EventHandler`, `ReconnectConfig`, error types, `Logger` interface
- Depends on: Standard library only
- Used by: All layers (re-exported from `pkg/client.go`)
- Purpose: Utility functions (timeout management)
- Location: `pkg/utils/timeout.go`
- Depends on: Standard library only
- Used by: `pkg/managers/request.go`
## Data Flow
## Key Abstractions
- Purpose: Abstract WebSocket operations
- Examples: `WebSocketTransport` implementation
- Pattern: Interface with `Send()`, `Receive()`, `Errors()`, `Close()`, `IsConnected()`
- Purpose: Enforce valid connection state transitions
- Examples: `ConnectionStateMachine`
- Pattern: `sync.RWMutex` protecting state, buffered channel for state change events, `validTransitions` map
- Purpose: Correlate request/response by RequestID
- Examples: Pending request map with context channels
- Pattern: Map of RequestID -> pendingRequest with `sync.Cond` for notification
- Purpose: Pub/sub event dispatch
- Examples: Handler map by EventType
- Pattern: Buffered channel for events, handler registration with auto-increment keys, atomic ID generation
## Entry Points
- Location: `pkg/client.go`
- Triggers: User instantiates via `NewClient(opts...)`, then calls `Connect(ctx)`, `SendRequest(ctx, req)`, `Subscribe(eventType, handler)`
- Responsibilities: Client lifecycle, manager coordination, API namespace provision, re-export types for convenience
- Location: `pkg/transport/websocket.go`
- Triggers: Called by `ConnectionManager`
- Responsibilities: WebSocket dial, read/write goroutines, ping/pong, close handling
- Location: `pkg/api/chat.go`, `pkg/api/agents.go`, etc.
- Triggers: User calls `client.<Namespace>().<Method>(ctx, params)`
- Responsibilities: Method-specific request construction, response parsing
## Error Handling
- Custom error types with `ErrorCode` string constants (`pkg/types/errors.go`)
- Error creation via `New<Type>Error(code, message, retryable, cause)` constructors
- Errors propagated through context and channels, not thrown
- `RequestManager.SendRequest()` returns protocol errors as Go errors
- `EventManager.Emit()` drops events after timeout without failing
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd:quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd:debug` for investigation and bug fixing
- `/gsd:execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd:profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
