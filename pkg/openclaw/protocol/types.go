package protocol

import (
	"encoding/json"
	"time"
)

// FrameType represents the type of frame
type FrameType string

const (
	FrameTypeGateway   FrameType = "gateway"
	FrameTypeRequest  FrameType = "request"
	FrameTypeResponse FrameType = "response"
	FrameTypeEvent    FrameType = "event"
	FrameTypeError    FrameType = "error"
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
	Type      FrameType       `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// RequestFrame represents a request frame
type RequestFrame struct {
	RequestID string          `json:"requestId"`
	Method    string          `json:"method"`
	Params    json.RawMessage `json:"params,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
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
