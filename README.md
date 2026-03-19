# OpenClaw SDK Go

[![OpenClaw SDK](https://img.shields.io/badge/OpenClaw-SDK-orange?logo=github)](https://openclaw.ai)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

> Feature-complete WebSocket SDK for Go with automatic reconnection, event handling, and request/response correlation.

OpenClaw SDK Go is a Go implementation migrated from the TypeScript version, providing a fully-featured WebSocket client with connection management, event handling, request/response patterns, and automatic reconnection.

## Features

- **Connection Management** - Automatic connection state management with disconnect handling
- **Event System** - Publish/subscribe pattern for event handling
- **Request/Response** - Automatic request-response correlation with timeout support
- **Auto-Reconnect** - Intelligent reconnection with Fibonacci backoff
- **TLS Support** - Configurable TLS connection options
- **Thread-Safe** - All public APIs are concurrency-safe
- **Context Support** - Full `context.Context` integration for cancellation and timeouts
- **Extensible Logging** - Built-in logger interface with custom implementation support

## Installation

```bash
go get github.com/frisbee-ai/openclaw-sdk-go
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    openclaw "github.com/frisbee-ai/openclaw-sdk-go/pkg"
    "github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
    "github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

func main() {
    // Create client
    client, err := openclaw.NewClient(
        openclaw.WithURL("ws://localhost:8080/ws"),
        openclaw.WithReconnect(true),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Subscribe to events
    client.Subscribe(types.EventConnect, func(e types.Event) {
        fmt.Println("Connected!")
    })

    // Connect to server
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // Send request
    resp, err := client.SendRequest(ctx, &protocol.RequestFrame{
        RequestID: "req-001",
        Action:    "ping",
        Payload:   map[string]interface{}{"message": "hello"},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Response: %v\n", resp)
}
```

### Enable Auto-Reconnect

```go
client, err := openclaw.NewClient(
    openclaw.WithURL("wss://api.example.com/ws"),
    openclaw.WithReconnect(true),
    openclaw.WithReconnectConfig(&types.ReconnectConfig{
        MaxAttempts: 10,                // Maximum reconnection attempts
        InitialDelay: 2 * time.Second,  // Initial delay
        MaxDelay:     60 * time.Second, // Maximum delay
    }),
)
```

### TLS Configuration

```go
client, err := openclaw.NewClient(
    openclaw.WithURL("wss://secure.example.com/ws"),
    openclaw.WithTLSConfig(&transport.TLSConfig{
        InsecureSkipVerify: false,
        CertFile:          "/path/to/client.crt",
        KeyFile:           "/path/to/client.key",
        CAFile:            "/path/to/ca.crt",
    }),
)
```

### Custom Logger

```go
type MyLogger struct{}

func (l *MyLogger) Debug(msg string, args ...any) {
    log.Printf("[DEBUG] %s %v", msg, args)
}

func (l *MyLogger) Info(msg string, args ...any) {
    log.Printf("[INFO] %s %v", msg, args)
}

func (l *MyLogger) Warn(msg string, args ...any) {
    log.Printf("[WARN] %s %v", msg, args)
}

func (l *MyLogger) Error(msg string, args ...any) {
    log.Printf("[ERROR] %s %v", msg, args)
}

client, err := openclaw.NewClient(
    openclaw.WithURL("ws://localhost:8080/ws"),
    openclaw.WithLogger(&MyLogger{}),
)
```

## API Documentation

### Client Options

| Option | Type | Description |
|--------|------|-------------|
| `WithURL(url string)` | string | WebSocket server URL |
| `WithAuthHandler(handler)` | AuthHandler | Authentication handler |
| `WithReconnect(enabled bool)` | bool | Enable auto-reconnect |
| `WithReconnectConfig(cfg)` | *ReconnectConfig | Reconnect configuration |
| `WithLogger(logger)` | Logger | Custom logger |
| `WithHeader(header)` | map[string][]string | Custom HTTP headers |
| `WithTLSConfig(cfg)` | *TLSConfig | TLS configuration |
| `WithEventBufferSize(n)` | int | Event buffer size |

### Connection States

```go
const (
    StateDisconnected   ConnectionState = "disconnected"
    StateConnecting     ConnectionState = "connecting"
    StateConnected      ConnectionState = "connected"
    StateAuthenticating ConnectionState = "authenticating"
    StateAuthenticated  ConnectionState = "authenticated"
    StateReconnecting   ConnectionState = "reconnecting"
    StateFailed         ConnectionState = "failed"
)
```

### Event Types

```go
const (
    EventConnect     EventType = "connect"
    EventDisconnect  EventType = "disconnect"
    EventError       EventType = "error"
    EventMessage     EventType = "message"
    EventRequest     EventType = "request"
    EventResponse    EventType = "response"
    EventTick        EventType = "tick"
    EventGap         EventType = "gap"
    EventStateChange EventType = "stateChange"
)
```

## Project Structure

```
openclaw-sdk-go/
├── pkg/
│   ├── client.go          # Main client API (Option pattern configuration)
│   ├── types/             # Shared types (ConnectionState, Event, errors, Logger)
│   ├── auth/              # Authentication (CredentialsProvider, AuthHandler)
│   ├── transport/         # WebSocket transport layer (gorilla/websocket)
│   ├── protocol/          # Protocol frames (RequestFrame, ResponseFrame, validation)
│   ├── connection/        # Connection state machine, policies, TLS validator
│   ├── events/            # Tick monitor, Gap detector
│   ├── managers/          # High-level managers (event, request, connection, reconnect)
│   └── utils/             # Timeout manager
└── examples/
    ├── cmd/               # CLI example
    └── server/            # Echo server example
```

## Design Patterns

### Option Pattern

Client configuration uses functional options for flexible, readable construction:

```go
client, err := openclaw.NewClient(
    openclaw.WithURL("ws://..."),
    openclaw.WithTimeout(30*time.Second),
    openclaw.WithLogger(myLogger),
)
```

### Context + Channel Hybrid

- `context.Context` for lifecycle management and cancellation
- Buffered channels for event delivery (prevents deadlocks)
- **Critical Rule**: Never send to a channel while holding a lock

### Graceful Shutdown

All managers implement `Close()` with proper goroutine cleanup using `sync.WaitGroup`.

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./pkg/transport

# Run single test
go test -run TestWebSocketTransportDial ./pkg/transport
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

### Examples

Run built-in examples:

```bash
# Start echo server
go run examples/server/main.go

# Run client in another terminal
go run examples/cmd/main.go
```

## Dependencies

This project minimizes external dependencies:

- `github.com/gorilla/websocket` - Industry standard WebSocket library for Go

> Note: Go's standard library `net/http` does not include WebSocket support. The `gorilla/websocket` package is the de facto standard for Go WebSocket implementations.

## Migration Notes

This is a Go implementation migrated from `openclaw-sdk-typescript`. While the Go SDK follows Go idioms rather than the TypeScript API, it maintains **functional equivalence**:

- All features from TypeScript SDK are available in Go
- Same protocol wire format
- Same authentication flow
- Same reconnection behavior (Fibonacci backoff)
- Same event types and semantics

Users migrating from TypeScript will find equivalent functionality with Go-idiomatic APIs.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details

## Resources

- [Design Document](docs/specs/2026-03-18-typescript-to-go-migration-design.md) - Architecture decisions
- [Implementation Plans](docs/plans/) - Phased implementation plans
- [GoDoc](https://pkg.go.dev/github.com/frisbee-ai/openclaw-sdk-go) - API reference

---

Copyright © 2026 @frisbee-ai
