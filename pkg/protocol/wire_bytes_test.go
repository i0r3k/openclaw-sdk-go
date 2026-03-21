package protocol

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestSendRequest_WireBytes_NoGatewayWrapper verifies no GatewayFrame wrapper
func TestSendRequest_WireBytes_NoGatewayWrapper(t *testing.T) {
	frame := NewRequestFrame("req-123", "test.method", nil)
	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Must NOT have gateway wrapper
	if bytes.Contains(data, []byte("\"type\":\"gateway\"")) {
		t.Error("found gateway wrapper - should be flat format")
	}
	if bytes.Contains(data, []byte("\"payload\":")) {
		t.Error("found payload wrapper - should be flat format")
	}
}

// TestSendRequest_WireBytes_FieldNames verifies correct field names
func TestSendRequest_WireBytes_FieldNames(t *testing.T) {
	frame := NewRequestFrame("req-abc123", "chat.list", json.RawMessage(`{}`))
	frameBytes, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Must have correct field names
	if !bytes.Contains(frameBytes, []byte("\"type\":\"req\"")) {
		t.Error("missing \"type\":\"req\"")
	}
	if !bytes.Contains(frameBytes, []byte("\"id\":\"req-abc123\"")) {
		t.Error("missing \"id\":\"req-abc123\"")
	}
	if !bytes.Contains(frameBytes, []byte("\"method\":\"chat.list\"")) {
		t.Error("missing \"method\":\"chat.list\"")
	}
	if !bytes.Contains(frameBytes, []byte("\"params\":{}")) {
		t.Error("missing \"params\":{}")
	}

	// Must NOT have old field names
	if bytes.Contains(frameBytes, []byte("\"requestId\"")) {
		t.Error("found old field \"requestId\" - should be \"id\"")
	}
	if bytes.Contains(frameBytes, []byte("\"RequestID\"")) {
		t.Error("found old field \"RequestID\" - should be \"id\"")
	}
	if bytes.Contains(frameBytes, []byte("\"Method\"")) {
		t.Error("found old field \"Method\" - should be \"method\"")
	}
	if bytes.Contains(frameBytes, []byte("\"Params\"")) {
		t.Error("found old field \"Params\" - should be \"params\"")
	}
}

// TestSendRequest_WireBytes_NoTimestamp verifies no timestamp field
func TestSendRequest_WireBytes_NoTimestamp(t *testing.T) {
	frame := NewRequestFrame("req-1", "test.method", nil)
	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Must NOT have timestamp
	if bytes.Contains(data, []byte("\"timestamp\"")) {
		t.Error("found timestamp field - should not be present")
	}
}

// TestSendRequest_WireBytes_EmptyParams verifies params:{} not params:null
func TestSendRequest_WireBytes_EmptyParams(t *testing.T) {
	// Use nil params - should be omitted, not null
	frame := NewRequestFrame("req-1", "test.method", nil)
	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// params should be omitted when nil (omitempty)
	if bytes.Contains(data, []byte("\"params\":null")) {
		t.Error("params should be omitted, not null")
	}
}

// TestSendRequest_WireBytes_WithParams verifies complex params serialize correctly
func TestSendRequest_WireBytes_WithParams(t *testing.T) {
	params := json.RawMessage(`{"roomId":"room-123","limit":10,"filter":["a","b"]}`)
	frame := NewRequestFrame("req-complex", "chat.list", params)
	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Verify params content
	if !bytes.Contains(data, []byte("\"roomId\":\"room-123\"")) {
		t.Error("missing roomId in params")
	}
	if !bytes.Contains(data, []byte("\"limit\":10")) {
		t.Error("missing limit in params")
	}
	if !bytes.Contains(data, []byte("\"filter\"")) {
		t.Error("missing filter in params")
	}

	// Verify JSON is valid
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

// TestResponseFrame_WireBytes_Success verifies success response format
func TestResponseFrame_WireBytes_Success(t *testing.T) {
	payload := json.RawMessage(`{"chats":[]}`)
	frame := NewResponseFrameSuccess("req-123", payload)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed["type"] != "res" {
		t.Errorf("type=res, got %v", parsed["type"])
	}
	if parsed["id"] != "req-123" {
		t.Errorf("id=req-123, got %v", parsed["id"])
	}
	if parsed["ok"] != true {
		t.Errorf("ok=true, got %v", parsed["ok"])
	}
	if parsed["payload"] == nil {
		t.Error("payload should not be nil")
	}
}

// TestResponseFrame_WireBytes_Error verifies error response format
func TestResponseFrame_WireBytes_Error(t *testing.T) {
	errShape := &ErrorShape{
		Code:    "AUTH_TOKEN_EXPIRED",
		Message: "Token expired",
	}
	frame := NewResponseFrameError("req-123", errShape)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed["ok"] != false {
		t.Errorf("ok=false, got %v", parsed["ok"])
	}
	if parsed["error"] == nil {
		t.Fatal("error should not be nil")
	}
	errObj := parsed["error"].(map[string]any)
	if errObj["code"] != "AUTH_TOKEN_EXPIRED" {
		t.Errorf("error.code=AUTH_TOKEN_EXPIRED, got %v", errObj["code"])
	}
}

// TestEventFrame_WireBytes_WithSeqAndStateVersion verifies event format
func TestEventFrame_WireBytes_WithSeqAndStateVersion(t *testing.T) {
	payload := json.RawMessage(`{"ts":1234567890}`)
	sv := NewStateVersion(10, 5)
	frame := NewEventFrameWithStateVersion("tick", payload, 42, sv)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
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

// TestWireBytes_UTF8 verifies UTF-8 in all string fields
func TestWireBytes_UTF8(t *testing.T) {
	params := json.RawMessage(`{"name":"你好世界","emoji":"🚀"}`)
	frame := NewRequestFrame("req-uni", "test.方法", params)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Verify JSON is valid and contains UTF-8
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	p := parsed["params"].(map[string]any)
	if p["name"] != "你好世界" {
		t.Errorf("name=你好世界, got %v", p["name"])
	}
	if p["emoji"] != "🚀" {
		t.Errorf("emoji=🚀, got %v", p["emoji"])
	}
}

// TestWireBytes_LargePayload verifies large payloads don't corrupt
func TestWireBytes_LargePayload(t *testing.T) {
	// Create a large payload (100KB of printable characters)
	largeData := make([]byte, 100*1024)
	for i := range largeData {
		largeData[i] = byte('a' + (i % 26))
	}
	params := json.RawMessage(`{"data":"` + string(largeData) + `"}`)
	frame := NewRequestFrame("req-large", "upload.data", params)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Verify JSON is valid
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	p := parsed["params"].(map[string]any)
	if len(p["data"].(string)) != 100*1024 {
		t.Error("large payload corrupted")
	}
}

// TestWireBytes_SpecialChars verifies special characters are handled
func TestWireBytes_SpecialChars(t *testing.T) {
	params := json.RawMessage(`{"special":"\u0000\u001F\u2028\u2029","quotes":"\"\\\n\r\t"}`)
	frame := NewRequestFrame("req-special", "test.special", params)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Verify JSON is valid
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	p := parsed["params"].(map[string]any)
	if p["special"] != "\u0000\u001F\u2028\u2029" {
		t.Error("special chars corrupted")
	}
}

// TestWireBytes_EmptyID verifies empty ID handling
func TestWireBytes_EmptyID(t *testing.T) {
	frame := NewRequestFrame("", "test.method", nil)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Empty ID should still serialize correctly
	if !bytes.Contains(data, []byte("\"id\":\"\"")) {
		t.Error("empty id not serialized correctly")
	}
}

// TestWireBytes_WhitespaceInMethod verifies whitespace in method name
func TestWireBytes_WhitespaceInMethod(t *testing.T) {
	frame := NewRequestFrame("req-1", "room.with space", nil)

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Verify method with space serializes correctly
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed["method"] != "room.with space" {
		t.Errorf("method=room.with space, got %v", parsed["method"])
	}
}

// TestWireBytes_RoundTrip verifies serialization round-trip
func TestWireBytes_RoundTrip(t *testing.T) {
	// Create frame
	original := NewRequestFrame("req-roundtrip", "test.roundtrip", json.RawMessage(`{"key":"value"}`))

	// Serialize
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Deserialize
	var parsed RequestFrame
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	// Verify
	if parsed.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", parsed.ID, original.ID)
	}
	if parsed.Method != original.Method {
		t.Errorf("Method mismatch: got %s, want %s", parsed.Method, original.Method)
	}
	if string(parsed.Params) != string(original.Params) {
		t.Errorf("Params mismatch: got %s, want %s", string(parsed.Params), string(original.Params))
	}
}

// TestWireBytes_VerifyExactFormat verifies exact bytes match TypeScript output
func TestWireBytes_VerifyExactFormat(t *testing.T) {
	frame := NewRequestFrame("req-abc123", "chat.list", json.RawMessage(`{}`))

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Exact format from TypeScript
	expected := `{"type":"req","id":"req-abc123","method":"chat.list","params":{}}`

	if string(data) != expected {
		t.Errorf("wire format mismatch:\n  got:  %s\n  want: %s", string(data), expected)
	}
}

// TestWireBytes_VerifyTypeScriptCompatible verifies JSON is parseable by TS
func TestWireBytes_VerifyTypeScriptCompatible(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"request_empty", mustMarshal(NewRequestFrame("req-1", "ping", nil))},
		{"request_params", mustMarshal(NewRequestFrame("req-2", "chat.list", json.RawMessage(`{"roomId":"abc"}`)))},
		{"response_success", mustMarshal(NewResponseFrameSuccess("req-1", json.RawMessage(`{"ok":true}`)))},
		{"response_error", mustMarshal(NewResponseFrameError("req-1", NewErrorShape("ERR", "error")))},
		{"event_simple", mustMarshal(NewEventFrame("connected", nil))},
		{"event_with_seq", mustMarshal(NewEventFrameWithSeq("tick", nil, 42))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify valid JSON
			var parsed map[string]any
			if err := json.Unmarshal(tt.data, &parsed); err != nil {
				t.Fatalf("invalid JSON: %s", string(tt.data))
			}

			// Verify no control characters in string fields
			jsonStr := string(tt.data)
			if strings.Contains(jsonStr, "\x00") || strings.Contains(jsonStr, "\x01") {
				t.Error("found control characters in JSON")
			}
		})
	}
}

func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
