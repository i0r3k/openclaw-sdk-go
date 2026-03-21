package protocol

import (
	"encoding/json"
	"testing"
)

func TestFrameTypeIsValid(t *testing.T) {
	tests := []struct {
		frameType FrameType
		expected  bool
	}{
		{FrameTypeRequest, true},
		{FrameTypeResponse, true},
		{FrameTypeEvent, true},
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
	frame := NewRequestFrame("req-1", "test.method", nil)
	if frame.ID != "req-1" {
		t.Errorf("expected req-1, got %s", frame.ID)
	}
	if frame.Method != "test.method" {
		t.Errorf("expected test.method, got %s", frame.Method)
	}
	if frame.Type != FrameTypeRequest {
		t.Errorf("expected req, got %s", frame.Type)
	}
}

func TestRequestFrameWithParams(t *testing.T) {
	params := json.RawMessage(`{"key":"value"}`)
	frame := NewRequestFrame("req-2", "test.method", params)
	if frame.ID != "req-2" {
		t.Errorf("expected req-2, got %s", frame.ID)
	}
	if string(frame.Params) != `{"key":"value"}` {
		t.Errorf("expected params, got %s", string(frame.Params))
	}
}

func TestResponseFrameSuccess(t *testing.T) {
	payload := json.RawMessage(`{"result":true}`)
	frame := NewResponseFrameSuccess("req-1", payload)
	if frame.ID != "req-1" {
		t.Errorf("expected req-1, got %s", frame.ID)
	}
	if !frame.Ok {
		t.Error("expected Ok=true")
	}
	if string(frame.Payload) != `{"result":true}` {
		t.Errorf("expected payload, got %s", string(frame.Payload))
	}
}

func TestResponseFrameError(t *testing.T) {
	errShape := &ErrorShape{Code: "ERR001", Message: "error"}
	frame := NewResponseFrameError("req-1", errShape)
	if frame.ID != "req-1" {
		t.Errorf("expected req-1, got %s", frame.ID)
	}
	if frame.Ok {
		t.Error("expected Ok=false")
	}
	if frame.Error == nil {
		t.Error("expected Error to be set")
	}
	if frame.Error.Code != "ERR001" {
		t.Errorf("expected ERR001, got %s", frame.Error.Code)
	}
}

func TestResponseFrameProgress(t *testing.T) {
	payload := json.RawMessage(`{"progress":50}`)
	frame := NewResponseFrameProgress("req-1", payload)
	if frame.ID != "req-1" {
		t.Errorf("expected req-1, got %s", frame.ID)
	}
	if !frame.Ok {
		t.Error("expected Ok=true for progress")
	}
	if !frame.Progress {
		t.Error("expected Progress=true")
	}
}

func TestEventFrame(t *testing.T) {
	payload := json.RawMessage(`{"data":"value"}`)
	frame := NewEventFrame("connection.established", payload)
	if frame.Event != "connection.established" {
		t.Errorf("expected connection.established, got %s", frame.Event)
	}
	if string(frame.Payload) != `{"data":"value"}` {
		t.Errorf("expected payload, got %s", string(frame.Payload))
	}
	if frame.Type != FrameTypeEvent {
		t.Errorf("expected event, got %s", frame.Type)
	}
}

func TestEventFrameWithSeq(t *testing.T) {
	payload := json.RawMessage(`{}`)
	frame := NewEventFrameWithSeq("tick", payload, 42)
	if frame.Seq != 42 {
		t.Errorf("expected 42, got %d", frame.Seq)
	}
}

func TestEventFrameWithStateVersion(t *testing.T) {
	payload := json.RawMessage(`{}`)
	sv := NewStateVersion(10, 5)
	frame := NewEventFrameWithStateVersion("tick", payload, 42, sv)
	if frame.Seq != 42 {
		t.Errorf("expected 42, got %d", frame.Seq)
	}
	if frame.StateVersion.Presence != 10 {
		t.Errorf("expected 10, got %d", frame.StateVersion.Presence)
	}
	if frame.StateVersion.Health != 5 {
		t.Errorf("expected 5, got %d", frame.StateVersion.Health)
	}
}

func TestStateVersion(t *testing.T) {
	sv := NewStateVersion(100, 200)
	if sv.Presence != 100 {
		t.Errorf("expected 100, got %d", sv.Presence)
	}
	if sv.Health != 200 {
		t.Errorf("expected 200, got %d", sv.Health)
	}
}

func TestErrorShape(t *testing.T) {
	err := NewErrorShape("ERR_CODE", "error message")
	if err.Code != "ERR_CODE" {
		t.Errorf("expected ERR_CODE, got %s", err.Code)
	}
	if err.Message != "error message" {
		t.Errorf("expected error message, got %s", err.Message)
	}
}

func TestErrorShapeWithDetails(t *testing.T) {
	details := json.RawMessage(`{"key":"value"}`)
	err := NewErrorShapeWithDetails("ERR_CODE", "error message", details)
	if err.Details == nil {
		t.Error("expected Details to be set")
	}
	if string(err.Details) != `{"key":"value"}` {
		t.Errorf("expected details, got %s", string(err.Details))
	}
}

func TestNewRetryableErrorShape(t *testing.T) {
	retryable := true
	retryAfterMs := int64(5000)
	err := NewRetryableErrorShape("ERR_CODE", "error message", retryable, retryAfterMs)
	if err.Retryable == nil || !*err.Retryable {
		t.Error("expected Retryable to be true")
	}
	if err.RetryAfterMs == nil || *err.RetryAfterMs != 5000 {
		t.Error("expected RetryAfterMs to be 5000")
	}
}

func TestRequestFrameJSONSerialization(t *testing.T) {
	params := json.RawMessage(`{}`)
	frame := NewRequestFrame("req-abc123", "chat.list", params)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify JSON output matches expected wire format
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decoded["type"] != "req" {
		t.Errorf("expected type=req, got %v", decoded["type"])
	}
	if decoded["id"] != "req-abc123" {
		t.Errorf("expected id=req-abc123, got %v", decoded["id"])
	}
	if decoded["method"] != "chat.list" {
		t.Errorf("expected method=chat.list, got %v", decoded["method"])
	}
	// No timestamp, requestId, etc.
	if _, ok := decoded["timestamp"]; ok {
		t.Error("should not have timestamp field")
	}
	if _, ok := decoded["requestId"]; ok {
		t.Error("should not have requestId field")
	}
}

func TestResponseFrameJSONSerialization(t *testing.T) {
	// Success case
	frame := NewResponseFrameSuccess("req-abc123", json.RawMessage(`{"chats":[]}`))
	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decoded["type"] != "res" {
		t.Errorf("expected type=res, got %v", decoded["type"])
	}
	if decoded["id"] != "req-abc123" {
		t.Errorf("expected id=req-abc123, got %v", decoded["id"])
	}
	if decoded["ok"] != true {
		t.Errorf("expected ok=true, got %v", decoded["ok"])
	}
}

func TestEventFrameJSONSerialization(t *testing.T) {
	sv := NewStateVersion(10, 5)
	frame := NewEventFrameWithStateVersion("tick", json.RawMessage(`{"ts":1234567890}`), 42, sv)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decoded["type"] != "event" {
		t.Errorf("expected type=event, got %v", decoded["type"])
	}
	if decoded["event"] != "tick" {
		t.Errorf("expected event=tick, got %v", decoded["event"])
	}
	if decoded["seq"].(float64) != 42 {
		t.Errorf("expected seq=42, got %v", decoded["seq"])
	}

	stateVersion := decoded["stateVersion"].(map[string]interface{})
	if stateVersion["presence"].(float64) != 10 {
		t.Errorf("expected presence=10, got %v", stateVersion["presence"])
	}
	if stateVersion["health"].(float64) != 5 {
		t.Errorf("expected health=5, got %v", stateVersion["health"])
	}
}
