# Phase 1: Project Setup and Foundation

**Project Structure:** Go module in root, source files in `pkg/openclaw/` directory

**Files:**
- Create: `go.mod` (root, no external deps yet)
- Create: `pkg/openclaw/types.go`, `pkg/openclaw/types_test.go`
- Create: `pkg/openclaw/errors.go`, `pkg/openclaw/errors_test.go`
- Create: `pkg/openclaw/logger.go`, `pkg/openclaw/logger_test.go`

---

## Task 1.1: Initialize Go Module

- [ ] **Step 1: Create go.mod**

```bash
go mod init github.com/frisbee-ai/openclaw-sdk-go
```

```go
// go.mod
module github.com/frisbee-ai/openclaw-sdk-go

go 1.21
```

- [ ] **Step 2: Commit (no dependencies yet - defer to Phase 4)**

```bash
git add go.mod
git commit -m "chore: initialize go module"
```

> **Note**: Dependencies (gorilla/websocket) will be added in Phase 4 when transport is implemented.

---

## Task 1.2: Create Basic Types

- [ ] **Step 1: Write types.go**

```go
package openclaw

import "time"

// ConnectionState represents the state of the connection
type ConnectionState string

const (
	StateDisconnected      ConnectionState = "disconnected"
	StateConnecting        ConnectionState = "connecting"
	StateConnected         ConnectionState = "connected"
	StateAuthenticating    ConnectionState = "authenticating"
	StateAuthenticated     ConnectionState = "authenticated"
	StateReconnecting      ConnectionState = "reconnecting"
	StateFailed            ConnectionState = "failed"
)

// EventType represents the type of event
type EventType string

const (
	EventConnect      EventType = "connect"
	EventDisconnect   EventType = "disconnect"
	EventError        EventType = "error"
	EventMessage      EventType = "message"
	EventRequest      EventType = "request"
	EventResponse     EventType = "response"
	EventTick         EventType = "tick"
	EventGap          EventType = "gap"
	EventStateChange  EventType = "stateChange"
)

// Event represents a generic event
type Event struct {
	Type      EventType
	Payload   interface{}
	Err       error
	Timestamp time.Time
}

// EventHandler is a function that handles events
type EventHandler func(Event)

// ReconnectConfig holds reconnection settings
type ReconnectConfig struct {
	MaxAttempts       int
	InitialDelay     time.Duration
	MaxDelay         time.Duration
	BackoffMultiplier float64
}

// DefaultReconnectConfig returns sensible defaults
// Note: InitialDelay must be <= MaxDelay for valid backoff
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxAttempts:       0, // 0 = infinite
		InitialDelay:     1 * time.Second,
		MaxDelay:         60 * time.Second,
		BackoffMultiplier: 1.618,
	}
}

// Validate validates the reconnect configuration
func (r ReconnectConfig) Validate() error {
	if r.InitialDelay > r.MaxDelay {
		return &ValidationError{&BaseError{ErrCodeValidation, "InitialDelay must be <= MaxDelay", nil}}
	}
	return nil
}
```

- [ ] **Step 2: Write types_test.go**

```go
package openclaw

import (
	"testing"
	"time"
)

func TestConnectionState(t *testing.T) {
	states := []ConnectionState{
		StateDisconnected,
		StateConnecting,
		StateConnected,
		StateAuthenticating,
		StateAuthenticated,
		StateReconnecting,
		StateFailed,
	}

	for _, s := range states {
		if s == "" {
			t.Error("state should not be empty")
		}
	}
}

func TestEventType(t *testing.T) {
	types := []EventType{
		EventConnect,
		EventDisconnect,
		EventError,
		EventMessage,
		EventRequest,
		EventResponse,
		EventTick,
		EventGap,
		EventStateChange,
	}

	for _, et := range types {
		if et == "" {
			t.Error("event type should not be empty")
		}
	}
}

func TestDefaultReconnectConfig(t *testing.T) {
	cfg := DefaultReconnectConfig()

	if cfg.MaxAttempts != 0 {
		t.Errorf("expected MaxAttempts=0 (infinite), got %d", cfg.MaxAttempts)
	}
	if cfg.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay=1s, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 60*time.Second {
		t.Errorf("expected MaxDelay=60s, got %v", cfg.MaxDelay)
	}
	if cfg.BackoffMultiplier != 1.618 {
		t.Errorf("expected BackoffMultiplier=1.618, got %f", cfg.BackoffMultiplier)
	}
}
```

- [ ] **Step 3: Verify it compiles and tests pass**

Run: `go build ./... && go test -v ./...`

- [ ] **Step 4: Commit**

```bash
git add pkg/openclaw/types.go pkg/openclaw/types_test.go
git commit -m "feat: add common types and constants"
```

---

## Task 1.3: Create Error Types

- [ ] **Step 1: Write errors.go**

```go
package openclaw

import "errors"

// ErrorCode represents an error code
type ErrorCode string

const (
	ErrCodeConnection   ErrorCode = "CONNECTION_ERROR"
	ErrCodeAuth        ErrorCode = "AUTH_ERROR"
	ErrCodeTimeout     ErrorCode = "TIMEOUT"
	ErrCodeProtocol    ErrorCode = "PROTOCOL_ERROR"
	ErrCodeValidation  ErrorCode = "VALIDATION_ERROR"
	ErrCodeTransport   ErrorCode = "TRANSPORT_ERROR"
	ErrCodeUnknown     ErrorCode = "UNKNOWN"
)

// OpenClawError is the base error interface
type OpenClawError interface {
	error
	Code() ErrorCode
	Unwrap() error
}

// BaseError is the base error struct
type BaseError struct {
	code    ErrorCode
	message string
	err     error
}

func (e *BaseError) Error() string { return e.message }
func (e *BaseError) Code() ErrorCode { return e.code }
func (e *BaseError) Unwrap() error { return e.err }

// ConnectionError represents a connection error
type ConnectionError struct {
	*BaseError
}

// AuthError represents an authentication error
type AuthError struct {
	*BaseError
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	*BaseError
}

// ProtocolError represents a protocol error
type ProtocolError struct {
	*BaseError
}

// ValidationError represents a validation error
type ValidationError struct {
	*BaseError
}

// TransportError represents a transport error
type TransportError struct {
	*BaseError
}

// NewError creates a new error with the given code, message, and cause
func NewError(code ErrorCode, message string, err error) OpenClawError {
	return &BaseError{code: code, message: message, err: err}
}

// NewConnectionError creates a new connection error
func NewConnectionError(message string, err error) OpenClawError {
	return &ConnectionError{&BaseError{ErrCodeConnection, message, err}}
}

// NewAuthError creates a new authentication error
func NewAuthError(message string, err error) OpenClawError {
	return &AuthError{&BaseError{ErrCodeAuth, message, err}}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, err error) OpenClawError {
	return &TimeoutError{&BaseError{ErrCodeTimeout, message, err}}
}

// NewProtocolError creates a new protocol error
func NewProtocolError(message string, err error) OpenClawError {
	return &ProtocolError{&BaseError{ErrCodeProtocol, message, err}}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) OpenClawError {
	return &ValidationError{&BaseError{ErrCodeValidation, message, err}}
}

// NewTransportError creates a new transport error
func NewTransportError(message string, err error) OpenClawError {
	return &TransportError{&BaseError{ErrCodeTransport, message, err}}
}

// Is checks if the error matches the given code
// Uses standard library errors.Is() with custom unwrap
func Is(err error, code ErrorCode) bool {
	var e OpenClawError
	if As(err, &e) {
		return e.Code() == code
	}
	return false
}

// As casts the error to OpenClawError
// Uses standard library errors.As() for proper type matching
func As(err error, target *OpenClawError) bool {
	return errors.As(err, target)
}
```

> **Note**: Added `import "errors"` to use standard library's error unwrapping.

- [ ] **Step 2: Write comprehensive errors_test.go**

```go
package openclaw

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	err := NewError(ErrCodeConnection, "test error", nil)
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got '%s'", err.Error())
	}
	if err.Code() != ErrCodeConnection {
		t.Errorf("expected CONNECTION_ERROR, got %s", err.Code())
	}
}

func TestIs(t *testing.T) {
	err := NewConnectionError("connection failed", nil)
	if !Is(err, ErrCodeConnection) {
		t.Error("expected Is to return true for matching code")
	}
	if Is(err, ErrCodeAuth) {
		t.Error("expected Is to return false for non-matching code")
	}
}

func TestAs(t *testing.T) {
	baseErr := NewError(ErrCodeConnection, "test", nil)
	var target OpenClawError
	if !As(baseErr, &target) {
		t.Error("expected As to return true")
	}
	if target.Code() != ErrCodeConnection {
		t.Error("expected to extract error with correct code")
	}
}

func TestErrorUnwrap(t *testing.T) {
	original := errors.New("original error")
	err := NewConnectionError("wrapped error", original)

	if !errors.Is(err, original) {
		t.Error("expected unwrap to return original error")
	}
}

func TestAllErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err     OpenClawError
		expected ErrorCode
	}{
		{"ConnectionError", NewConnectionError("test", nil), ErrCodeConnection},
		{"AuthError", NewAuthError("test", nil), ErrCodeAuth},
		{"TimeoutError", NewTimeoutError("test", nil), ErrCodeTimeout},
		{"ProtocolError", NewProtocolError("test", nil), ErrCodeProtocol},
		{"ValidationError", NewValidationError("test", nil), ErrCodeValidation},
		{"TransportError", NewTransportError("test", nil), ErrCodeTransport},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.err.Code())
			}
		})
	}
}
```

- [ ] **Step 3: Verify it compiles and tests pass**

Run: `go build ./... && go test -v ./...`

- [ ] **Step 4: Commit**

```bash
git add pkg/openclaw/errors.go pkg/openclaw/errors_test.go
git commit -m "feat: add error type hierarchy"
```

---

## Task 1.4: Create Logger Interface

- [ ] **Step 1: Write logger.go**

```go
package openclaw

import (
	"context"
	"io"
	"log"
	"os"
)

// Logger interface for customizable logging
// Follows standard Go logging patterns with level support
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// DefaultLogger uses stdlib log
type DefaultLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

// NewDefaultLogger creates a logger that writes to stdout
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		debug: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime),
		info:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		warn:  log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime),
		error: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime),
	}
}

// NewDefaultLoggerWithWriter creates a logger with custom writer
func NewDefaultLoggerWithWriter(w io.Writer) *DefaultLogger {
	return &DefaultLogger{
		debug: log.New(w, "[DEBUG] ", log.Ldate|log.Ltime),
		info:  log.New(w, "[INFO] ", log.Ldate|log.Ltime),
		warn:  log.New(w, "[WARN] ", log.Ldate|log.Ltime),
		error: log.New(w, "[ERROR] ", log.Ldate|log.Ltime),
	}
}

func (l *DefaultLogger) Debug(msg string, args ...any) { l.debug.Printf(msg, args...) }
func (l *DefaultLogger) Info(msg string, args ...any)  { l.info.Printf(msg, args...) }
func (l *DefaultLogger) Warn(msg string, args ...any)  { l.warn.Printf(msg, args...) }
func (l *DefaultLogger) Error(msg string, args ...any) { l.error.Printf(msg, args...) }

// NopLogger is a no-op implementation for testing
type NopLogger struct{}

func (l *NopLogger) Debug(msg string, args ...any) {}
func (l *NopLogger) Info(msg string, args ...any)  {}
func (l *NopLogger) Warn(msg string, args ...any)  {}
func (l *NopLogger) Error(msg string, args ...any) {}

// WithContext creates a context with logger
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) (Logger, bool) {
	logger, ok := ctx.Value(loggerKey{}).(Logger)
	return logger, ok
}

type loggerKey struct{}
```

- [ ] **Step 2: Write logger_test.go**

```go
package openclaw

import (
	"bytes"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLoggerWithWriter(buf)

	logger.Info("test message %s", "world")
	logger.Debug("debug message")
	logger.Warn("warning message")
	logger.Error("error message")

	output := buf.String()
	if output == "" {
		t.Error("expected logger output")
	}
}

func TestNopLogger(t *testing.T) {
	logger := &NopLogger{}
	// Should not panic
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
}

func TestLoggerInterface(t *testing.T) {
	// Verify DefaultLogger implements Logger
	var _ Logger = &DefaultLogger{}
	// Verify NopLogger implements Logger
	var _ Logger = &NopLogger{}
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	logger := &NopLogger{}

	ctx = WithContext(ctx, logger)
	retrieved, ok := FromContext(ctx)

	if !ok {
		t.Error("expected to retrieve logger from context")
	}
	if retrieved != logger {
		t.Error("expected to retrieve same logger")
	}
}
```

- [ ] **Step 3: Verify it compiles and tests pass**

Run: `go build ./... && go test -v ./...`

- [ ] **Step 4: Commit**

```bash
git add pkg/openclaw/logger.go pkg/openclaw/logger_test.go
git commit -m "feat: add Logger interface with context support"
```

---

## Phase 1 Complete

After this phase, you should have:
- `pkg/openclaw/go.mod` - Go module initialized (no external deps yet)
- `pkg/openclaw/types.go` - Common types and constants
- `pkg/openclaw/types_test.go` - Types tests
- `pkg/openclaw/errors.go` - Error type hierarchy
- `pkg/openclaw/errors_test.go` - Comprehensive error tests
- `pkg/openclaw/logger.go` - Logger interface with context support
- `pkg/openclaw/logger_test.go` - Logger tests

All code should compile and tests should pass.
