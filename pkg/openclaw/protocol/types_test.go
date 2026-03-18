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
