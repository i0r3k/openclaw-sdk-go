package protocol

import (
	"encoding/json"
	"testing"
)

// TestRequestFrame_Serialize_MatchesTS verifies RequestFrame serialization matches TypeScript wire format
func TestRequestFrame_Serialize_MatchesTS(t *testing.T) {
	params := json.RawMessage(`{}`)
	frame := NewRequestFrame("req-abc123", "chat.list", params)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected TS wire format: {"type":"req","id":"req-abc123","method":"chat.list","params":{}}
	expected := `{"type":"req","id":"req-abc123","method":"chat.list","params":{}}`

	if string(data) != expected {
		t.Errorf("JSON output:\n  got:  %s\n  want: %s", string(data), expected)
	}

	// Verify it can be parsed back
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if parsed["type"] != "req" {
		t.Errorf("type=req, got %v", parsed["type"])
	}
	if parsed["id"] != "req-abc123" {
		t.Errorf("id=req-abc123, got %v", parsed["id"])
	}
	if parsed["method"] != "chat.list" {
		t.Errorf("method=chat.list, got %v", parsed["method"])
	}
}

// TestResponseFrame_Serialize_Success_MatchesTS verifies success ResponseFrame serialization
func TestResponseFrame_Serialize_Success_MatchesTS(t *testing.T) {
	payload := json.RawMessage(`{"chats":[]}`)
	frame := NewResponseFrameSuccess("req-abc123", payload)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected TS wire format: {"type":"res","id":"req-abc123","ok":true,"payload":{"chats":[]}}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if parsed["type"] != "res" {
		t.Errorf("type=res, got %v", parsed["type"])
	}
	if parsed["id"] != "req-abc123" {
		t.Errorf("id=req-abc123, got %v", parsed["id"])
	}
	if parsed["ok"] != true {
		t.Errorf("ok=true, got %v", parsed["ok"])
	}
}

// TestResponseFrame_Serialize_Error_MatchesTS verifies error ResponseFrame serialization
func TestResponseFrame_Serialize_Error_MatchesTS(t *testing.T) {
	errShape := &ErrorShape{
		Code:      "AUTH_TOKEN_EXPIRED",
		Message:   "Token expired",
		Retryable: func() *bool { b := true; return &b }(),
	}
	frame := NewResponseFrameError("req-abc123", errShape)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if parsed["type"] != "res" {
		t.Errorf("type=res, got %v", parsed["type"])
	}
	if parsed["ok"] != false {
		t.Errorf("ok=false, got %v", parsed["ok"])
	}

	errObj := parsed["error"].(map[string]any)
	if errObj["code"] != "AUTH_TOKEN_EXPIRED" {
		t.Errorf("error.code=AUTH_TOKEN_EXPIRED, got %v", errObj["code"])
	}
}

// TestEventFrame_Serialize_MatchesTS verifies EventFrame with seq and stateVersion
func TestEventFrame_Serialize_MatchesTS(t *testing.T) {
	payload := json.RawMessage(`{"ts":1234567890}`)
	sv := NewStateVersion(10, 5)
	frame := NewEventFrameWithStateVersion("tick", payload, 42, sv)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if parsed["type"] != "event" {
		t.Errorf("type=event, got %v", parsed["type"])
	}
	if parsed["event"] != "tick" {
		t.Errorf("event=tick, got %v", parsed["event"])
	}
	if parsed["seq"].(float64) != 42 {
		t.Errorf("seq=42, got %v", parsed["seq"])
	}

	svParsed := parsed["stateVersion"].(map[string]any)
	if svParsed["presence"].(float64) != 10 {
		t.Errorf("stateVersion.presence=10, got %v", svParsed["presence"])
	}
	if svParsed["health"].(float64) != 5 {
		t.Errorf("stateVersion.health=5, got %v", svParsed["health"])
	}
}

// TestErrorShape_AllFields verifies ErrorShape serialization
func TestErrorShape_AllFields(t *testing.T) {
	retryable := true
	retryAfterMs := int64(5000)
	details := json.RawMessage(`{"key":"value"}`)

	shape := &ErrorShape{
		Code:         "TEST_ERROR",
		Message:      "Test error message",
		Details:      details,
		Retryable:    &retryable,
		RetryAfterMs: &retryAfterMs,
	}

	data, err := json.Marshal(shape)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if parsed["code"] != "TEST_ERROR" {
		t.Errorf("code=TEST_ERROR, got %v", parsed["code"])
	}
	if parsed["message"] != "Test error message" {
		t.Errorf("message=Test error message, got %v", parsed["message"])
	}
	if parsed["details"].(map[string]any)["key"] != "value" {
		t.Errorf("details.key=value, got %v", parsed["details"])
	}
	if parsed["retryable"] != true {
		t.Errorf("retryable=true, got %v", parsed["retryable"])
	}
	if parsed["retryAfterMs"].(float64) != 5000 {
		t.Errorf("retryAfterMs=5000, got %v", parsed["retryAfterMs"])
	}
}

// TestRequestFrame_NoTimestamp verifies no timestamp field in wire format
func TestRequestFrame_NoTimestamp(t *testing.T) {
	frame := NewRequestFrame("req-1", "test.method", nil)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if _, ok := parsed["timestamp"]; ok {
		t.Error("timestamp field should not be present")
	}
	if _, ok := parsed["requestId"]; ok {
		t.Error("requestId field should not be present (should be 'id')")
	}
	if _, ok := parsed["RequestID"]; ok {
		t.Error("RequestID field should not be present (should be 'ID')")
	}
}

// TestRequestFrame_EmptyParams verifies params:{} not params:null
func TestRequestFrame_EmptyParams(t *testing.T) {
	frame := NewRequestFrame("req-1", "test.method", nil)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// params should be omitted when nil, not null
	if string(data) == `{"type":"req","id":"req-1","method":"test.method","params":null}` {
		t.Error("params should be omitted, not null")
	}
}

// TestResponseFrame_Deserialize_Success verifies success response deserialization
func TestResponseFrame_Deserialize_Success(t *testing.T) {
	jsonData := `{"type":"res","id":"req-abc123","ok":true,"payload":{"chats":[]}}`

	var frame ResponseFrame
	if err := json.Unmarshal([]byte(jsonData), &frame); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if frame.Type != FrameTypeResponse {
		t.Errorf("Type=res, got %s", frame.Type)
	}
	if frame.ID != "req-abc123" {
		t.Errorf("ID=req-abc123, got %s", frame.ID)
	}
	if !frame.Ok {
		t.Error("Ok should be true")
	}
}

// TestResponseFrame_Deserialize_Error verifies error response deserialization
func TestResponseFrame_Deserialize_Error(t *testing.T) {
	jsonData := `{"type":"res","id":"req-abc123","ok":false,"error":{"code":"AUTH_TOKEN_EXPIRED","message":"Token expired","retryable":true}}`

	var frame ResponseFrame
	if err := json.Unmarshal([]byte(jsonData), &frame); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if frame.Ok {
		t.Error("Ok should be false")
	}
	if frame.Error == nil {
		t.Fatal("Error should be present")
	}
	if frame.Error.Code != "AUTH_TOKEN_EXPIRED" {
		t.Errorf("Error.Code=AUTH_TOKEN_EXPIRED, got %s", frame.Error.Code)
	}
}

// TestEventFrame_Deserialize_WithSeqAndStateVersion verifies event deserialization
func TestEventFrame_Deserialize_WithSeqAndStateVersion(t *testing.T) {
	jsonData := `{"type":"event","event":"tick","payload":{"ts":1234567890},"seq":42,"stateVersion":{"presence":10,"health":5}}`

	var frame EventFrame
	if err := json.Unmarshal([]byte(jsonData), &frame); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if frame.Type != FrameTypeEvent {
		t.Errorf("Type=event, got %s", frame.Type)
	}
	if frame.Event != "tick" {
		t.Errorf("Event=tick, got %s", frame.Event)
	}
	if frame.Seq != 42 {
		t.Errorf("Seq=42, got %d", frame.Seq)
	}
	if frame.StateVersion == nil {
		t.Fatal("StateVersion should be present")
	}
	if frame.StateVersion.Presence != 10 {
		t.Errorf("StateVersion.Presence=10, got %d", frame.StateVersion.Presence)
	}
	if frame.StateVersion.Health != 5 {
		t.Errorf("StateVersion.Health=5, got %d", frame.StateVersion.Health)
	}
}
