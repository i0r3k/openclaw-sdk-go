// Package protocol provides protocol frame types and utilities for OpenClaw SDK.
//
// This package defines the wire protocol for WebSocket communication:
//   - FrameType: Types of protocol frames (request, response, event)
//   - Frame structures: RequestFrame, ResponseFrame, EventFrame
//   - Error types: ErrorShape, StateVersion
//   - Factory functions: NewRequestFrame, NewResponseFrame, NewEventFrame
package protocol

import (
	"encoding/json"
)

// FrameType represents the type of frame in the protocol.
// Each frame type has a specific role in the communication flow.
type FrameType string

const (
	FrameTypeRequest  FrameType = "req"
	FrameTypeResponse FrameType = "res"
	FrameTypeEvent    FrameType = "event"
)

// IsValid checks if FrameType is a valid constant
func (f FrameType) IsValid() bool {
	switch f {
	case FrameTypeRequest, FrameTypeResponse, FrameTypeEvent:
		return true
	}
	return false
}

// RequestFrame represents an outbound request frame.
// Wire format: {"type":"req","id":"...","method":"...","params":{}}
type RequestFrame struct {
	Type   FrameType       `json:"type"`
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// ResponseFrame represents an inbound response frame.
// Wire format: {"type":"res","id":"...","ok":true,"payload":{}} or {"type":"res","id":"...","ok":false,"error":{}}
type ResponseFrame struct {
	Type     FrameType       `json:"type"`
	ID       string          `json:"id"`
	Ok       bool            `json:"ok"`
	Payload  json.RawMessage `json:"payload,omitempty"`
	Error    *ErrorShape     `json:"error,omitempty"`
	Progress bool            `json:"progress,omitempty"`
}

// EventFrame represents an event frame from the server.
// Wire format: {"type":"event","event":"tick","payload":{},"seq":42,"stateVersion":{"presence":10,"health":5}}
type EventFrame struct {
	Type         FrameType       `json:"type"`
	Event        string          `json:"event"`
	Payload      json.RawMessage `json:"payload,omitempty"`
	Seq          uint64          `json:"seq,omitempty"`
	StateVersion *StateVersion   `json:"stateVersion,omitempty"`
}

// StateVersion represents the version of server state.
// Used in EventFrame for gap detection and sync.
type StateVersion struct {
	Presence uint64 `json:"presence"`
	Health   uint64 `json:"health"`
}

// ErrorShape represents an error in a response or event.
// Wire format: {"code":"ERROR_CODE","message":"human readable","details":{},"retryable":true,"retryAfterMs":5000}
type ErrorShape struct {
	Code         string          `json:"code"`
	Message      string          `json:"message"`
	Details      json.RawMessage `json:"details,omitempty"`
	Retryable    *bool           `json:"retryable,omitempty"`
	RetryAfterMs *int64          `json:"retryAfterMs,omitempty"`
}

// NewRequestFrame creates a new request frame with the new wire format.
// The Type field is automatically set to "req".
func NewRequestFrame(id, method string, params json.RawMessage) *RequestFrame {
	return &RequestFrame{
		Type:   FrameTypeRequest,
		ID:     id,
		Method: method,
		Params: params,
	}
}

// NewResponseFrameSuccess creates a successful response frame.
func NewResponseFrameSuccess(id string, payload json.RawMessage) *ResponseFrame {
	return &ResponseFrame{
		Type:    FrameTypeResponse,
		ID:      id,
		Ok:      true,
		Payload: payload,
	}
}

// NewResponseFrameError creates an error response frame.
func NewResponseFrameError(id string, err *ErrorShape) *ResponseFrame {
	return &ResponseFrame{
		Type:  FrameTypeResponse,
		ID:    id,
		Ok:    false,
		Error: err,
	}
}

// NewResponseFrameProgress creates a progress update response frame.
func NewResponseFrameProgress(id string, payload json.RawMessage) *ResponseFrame {
	return &ResponseFrame{
		Type:     FrameTypeResponse,
		ID:       id,
		Ok:       true,
		Payload:  payload,
		Progress: true,
	}
}

// NewEventFrame creates a new event frame.
func NewEventFrame(event string, payload json.RawMessage) *EventFrame {
	return &EventFrame{
		Type:    FrameTypeEvent,
		Event:   event,
		Payload: payload,
	}
}

// NewEventFrameWithSeq creates a new event frame with sequence number.
func NewEventFrameWithSeq(event string, payload json.RawMessage, seq uint64) *EventFrame {
	return &EventFrame{
		Type:    FrameTypeEvent,
		Event:   event,
		Payload: payload,
		Seq:     seq,
	}
}

// NewEventFrameWithStateVersion creates a new event frame with sequence and state version.
func NewEventFrameWithStateVersion(event string, payload json.RawMessage, seq uint64, sv *StateVersion) *EventFrame {
	return &EventFrame{
		Type:         FrameTypeEvent,
		Event:        event,
		Payload:      payload,
		Seq:          seq,
		StateVersion: sv,
	}
}

// NewStateVersion creates a new state version.
func NewStateVersion(presence, health uint64) *StateVersion {
	return &StateVersion{
		Presence: presence,
		Health:   health,
	}
}

// NewErrorShape creates a new error shape.
func NewErrorShape(code, message string) *ErrorShape {
	return &ErrorShape{
		Code:    code,
		Message: message,
	}
}

// NewErrorShapeWithDetails creates a new error shape with details.
func NewErrorShapeWithDetails(code, message string, details json.RawMessage) *ErrorShape {
	return &ErrorShape{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewRetryableErrorShape creates a new retryable error shape.
func NewRetryableErrorShape(code, message string, retryable bool, retryAfterMs int64) *ErrorShape {
	return &ErrorShape{
		Code:         code,
		Message:      message,
		Retryable:    &retryable,
		RetryAfterMs: &retryAfterMs,
	}
}
