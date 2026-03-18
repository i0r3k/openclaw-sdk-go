# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OpenClaw SDK Go is a Go implementation of the OpenClaw WebSocket SDK, migrated from TypeScript. It provides a feature-complete WebSocket client with connection management, event handling, request/response patterns, and automatic reconnection.

**Module Path**: `github.com/i0r3k/openclaw-sdk-go`

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
├── client.go          # Main client API with option pattern configuration
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

## Migration Context

This is a Go migration of `openclaw-sdk-typescript`. The design document at `docs/specs/2026-03-18-typescript-to-go-migration-design.md` details architectural decisions. Phase plans in `docs/plans/` document the incremental implementation strategy.

## Important Notes

- **Go version**: 1.21+
- **External dependencies**: Minimal, only `github.com/gorilla/websocket`
- **Standard library preferred**: Uses `net/http`, `context`, `sync`, `crypto/tls` from stdlib
- **Feature parity**: Functionally equivalent to TypeScript SDK, not API-identical
