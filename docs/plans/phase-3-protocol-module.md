# Phase 3: Protocol Module

**Files:**
- Create: `protocol/types.go`, `protocol/types_test.go`
- Create: `protocol/validation.go`, `protocol/validation_test.go`

**Depends on:** Phase 1 (types.go, errors.go)

---

## Task 3.1: Protocol Types

- [ ] **Step 1: Create protocol directory and types.go**

```bash
mkdir -p protocol
```

```go
// protocol/types.go
package protocol

import (
	"encoding/json"
	"time"
)

// FrameType represents the type of frame
type FrameType string

const (
	FrameTypeGateway   FrameType = "gateway"
	FrameTypeRequest   FrameType = "request"
	FrameTypeResponse  FrameType = "response"
	FrameTypeEvent     FrameType = "event"
	FrameTypeError     FrameType = "error"
)

// IsValid checks if FrameType is a valid constant
func (f FrameType) IsValid() bool {
	switch f {
	case FrameTypeGateway, FrameTypeRequest, FrameTypeResponse, FrameTypeEvent, FrameTypeError:
		return true
	}
	return false
}

// GatewayFrame is the main frame type
type GatewayFrame struct {
	Type      FrameType          `json:"type"`
	Timestamp time.Time          `json:"timestamp"`
	Payload   json.RawMessage    `json:"payload,omitempty"`
}

// RequestFrame represents a request frame
type RequestFrame struct {
	RequestID  string          `json:"requestId"`
	Method     string          `json:"method"`
	Params     json.RawMessage `json:"params,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
}

// ResponseFrame represents a response frame
type ResponseFrame struct {
	RequestID string          `json:"requestId"`
	Success   bool            `json:"success"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *ResponseError `json:"error,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// ResponseError represents an error in a response
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// EventFrame represents an event frame
type EventFrame struct {
	EventType string          `json:"eventType"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewRequestFrame creates a new request frame
func NewRequestFrame(requestID, method string) *RequestFrame {
	return &RequestFrame{
		RequestID: requestID,
		Method:    method,
		Timestamp: time.Now(),
	}
}

// NewResponseFrame creates a new response frame
func NewResponseFrame(requestID string, success bool) *ResponseFrame {
	return &ResponseFrame{
		RequestID: requestID,
		Success:   success,
		Timestamp: time.Now(),
	}
}

// NewEventFrame creates a new event frame
func NewEventFrame(eventType string) *EventFrame {
	return &EventFrame{
		EventType: eventType,
		Timestamp: time.Now(),
	}
}
```

- [ ] **Step 2: Write comprehensive tests**

```go
// protocol/types_test.go
package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGatewayFrameSerialization(t *testing.T) {
	frame := GatewayFrame{
		Type:      FrameTypeGateway,
		Timestamp: time.Now(),
		Payload:   json.RawMessage(`{"key":"value"}`),
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded GatewayFrame
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decoded.Type != frame.Type {
		t.Errorf("expected %s, got %s", frame.Type, decoded.Type)
	}
}

func TestFrameTypeIsValid(t *testing.T) {
	tests := []struct {
		frameType FrameType
		expected  bool
	}{
		{FrameTypeGateway, true},
		{FrameTypeRequest, true},
		{FrameTypeResponse, true},
		{FrameTypeEvent, true},
		{FrameTypeError, true},
		{FrameType("invalid"), false},
		{FrameType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.frameType), func(t *testing.T) {
			if got := tt.frameType.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRequestFrame(t *testing.T) {
	frame := NewRequestFrame("req-1", "test.method")
	if frame.RequestID != "req-1" {
		t.Errorf("expected req-1, got %s", frame.RequestID)
	}
	if frame.Method != "test.method" {
		t.Errorf("expected test.method, got %s", frame.Method)
	}
	if frame.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestResponseFrame(t *testing.T) {
	frame := NewResponseFrame("req-1", true)
	if frame.RequestID != "req-1" {
		t.Errorf("expected req-1, got %s", frame.RequestID)
	}
	if !frame.Success {
		t.Error("expected Success=true")
	}
}

func TestEventFrame(t *testing.T) {
	frame := NewEventFrame("connection.established")
	if frame.EventType != "connection.established" {
		t.Errorf("expected connection.established, got %s", frame.EventType)
	}
}
```

- [ ] **Step 3: Run tests and commit**

Run: `go test -v ./protocol/...`
Commit: `git add protocol/ && git commit -m "feat: add protocol types with validation"`

---

## Task 3.2: Protocol Validation

- [ ] **Step 1: Write validation.go**

```go
// protocol/validation.go
package protocol

import (
	"errors"
	"strings"
)

// ValidationError represents a validation error (uses Phase 1 error pattern)
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validator validates protocol frames
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateGatewayFrame validates a gateway frame
func (v *Validator) ValidateGatewayFrame(frame *GatewayFrame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	if frame.Type == "" {
		return &ValidationError{Field: "Type", Message: "is required"}
	}
	if !frame.Type.IsValid() {
		return &ValidationError{Field: "Type", Message: "is not a valid frame type"}
	}
	return nil
}

// ValidateRequestFrame validates a request frame
func (v *Validator) ValidateRequestFrame(frame *RequestFrame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	if frame.RequestID == "" {
		return &ValidationError{Field: "RequestID", Message: "is required"}
	}
	if frame.Method == "" {
		return &ValidationError{Field: "Method", Message: "is required"}
	}
	// Validate method format (namespace.method or namespace.sub.method)
	parts := strings.Split(frame.Method, ".")
	if len(parts) < 2 {
		return &ValidationError{Field: "Method", Message: "must be in format 'namespace.method'"}
	}
	for _, part := range parts {
		if part == "" {
			return &ValidationError{Field: "Method", Message: "must be in format 'namespace.method'"}
		}
	}
	return nil
}

// ValidateResponseFrame validates a response frame
func (v *Validator) ValidateResponseFrame(frame *ResponseFrame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	if frame.RequestID == "" {
		return &ValidationError{Field: "RequestID", Message: "is required"}
	}
	// Success and Error are mutually exclusive
	if frame.Success && frame.Error != nil {
		return &ValidationError{Field: "Error", Message: "must be nil when Success is true"}
	}
	if !frame.Success && frame.Error == nil {
		return &ValidationError{Field: "Error", Message: "is required when Success is false"}
	}
	return nil
}

// ValidateEventFrame validates an event frame
func (v *Validator) ValidateEventFrame(frame *EventFrame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	if frame.EventType == "" {
		return &ValidationError{Field: "EventType", Message: "is required"}
	}
	return nil
}
```

- [ ] **Step 2: Write comprehensive tests**

```go
// protocol/validation_test.go
package protocol

import (
	"testing"
)

func TestValidator_ValidateGatewayFrame(t *testing.T) {
	v := NewValidator()

	// nil test
	err := v.ValidateGatewayFrame(nil)
	if err == nil {
		t.Error("expected error for nil frame")
	}

	// valid frame
	err = v.ValidateGatewayFrame(&GatewayFrame{Type: FrameTypeGateway})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// empty type
	err = v.ValidateGatewayFrame(&GatewayFrame{})
	if err == nil {
		t.Error("expected error for empty type")
	}

	// invalid type
	err = v.ValidateGatewayFrame(&GatewayFrame{Type: FrameType("invalid")})
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestValidator_ValidateRequestFrame(t *testing.T) {
	v := NewValidator()

	// valid frame
	err := v.ValidateRequestFrame(&RequestFrame{
		RequestID: "123",
		Method:   "test.method",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// missing RequestID
	err = v.ValidateRequestFrame(&RequestFrame{Method: "test"})
	if err == nil {
		t.Error("expected error for missing RequestID")
	}

	// missing Method
	err = v.ValidateRequestFrame(&RequestFrame{RequestID: "123"})
	if err == nil {
		t.Error("expected error for missing Method")
	}

	// invalid method format
	err = v.ValidateRequestFrame(&RequestFrame{RequestID: "123", Method: "invalid"})
	if err == nil {
		t.Error("expected error for invalid method format")
	}
}

func TestValidator_ValidateResponseFrame(t *testing.T) {
	v := NewValidator()

	// valid success frame
	err := v.ValidateResponseFrame(&ResponseFrame{
		RequestID: "123",
		Success:   true,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// valid error frame
	err = v.ValidateResponseFrame(&ResponseFrame{
		RequestID: "123",
		Success:   false,
		Error:     &ResponseError{Code: "ERR001", Message: "error"},
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// success with error
	err = v.ValidateResponseFrame(&ResponseFrame{
		RequestID: "123",
		Success:   true,
		Error:     &ResponseError{Code: "ERR001", Message: "error"},
	})
	if err == nil {
		t.Error("expected error for success with error")
	}

	// failure without error
	err = v.ValidateResponseFrame(&ResponseFrame{
		RequestID: "123",
		Success:   false,
	})
	if err == nil {
		t.Error("expected error for failure without error")
	}
}

func TestValidator_ValidateEventFrame(t *testing.T) {
	v := NewValidator()

	// valid frame
	err := v.ValidateEventFrame(&EventFrame{EventType: "test.event"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// nil frame
	err = v.ValidateEventFrame(nil)
	if err == nil {
		t.Error("expected error for nil frame")
	}

	// empty event type
	err = v.ValidateEventFrame(&EventFrame{})
	if err == nil {
		t.Error("expected error for empty event type")
	}
}
```

- [ ] **Step 3: Run tests and commit**

Run: `go test -v ./protocol/...`
Commit: `git add protocol/ && git commit -m "feat: add protocol validation with comprehensive tests"`

---

## Phase 3 Complete

After this phase, you should have:
- `protocol/types.go` - Protocol frame types with validation helpers
- `protocol/types_test.go` - Comprehensive types tests
- `protocol/validation.go` - Frame validation with all frame types
- `protocol/validation_test.go` - Comprehensive validation tests

All code should compile and tests should pass.
