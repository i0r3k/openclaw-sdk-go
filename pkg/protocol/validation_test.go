package protocol

import (
	"encoding/json"
	"testing"
)

func TestValidator_ValidateRequestFrame(t *testing.T) {
	v := NewValidator()

	// valid frame
	err := v.ValidateRequestFrame(&RequestFrame{
		ID:     "123",
		Method: "test.method",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// missing ID
	err = v.ValidateRequestFrame(&RequestFrame{Method: "test"})
	if err == nil {
		t.Error("expected error for missing ID")
	}

	// missing Method
	err = v.ValidateRequestFrame(&RequestFrame{ID: "123"})
	if err == nil {
		t.Error("expected error for missing Method")
	}

	// invalid method format
	err = v.ValidateRequestFrame(&RequestFrame{ID: "123", Method: "invalid"})
	if err == nil {
		t.Error("expected error for invalid method format")
	}
}

func TestValidator_ValidateResponseFrame(t *testing.T) {
	v := NewValidator()

	// valid success frame
	err := v.ValidateResponseFrame(&ResponseFrame{
		ID: "123",
		Ok: true,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// valid error frame
	err = v.ValidateResponseFrame(&ResponseFrame{
		ID:    "123",
		Ok:    false,
		Error: &ErrorShape{Code: "ERR001", Message: "error"},
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// ok with error (invalid)
	err = v.ValidateResponseFrame(&ResponseFrame{
		ID:    "123",
		Ok:    true,
		Error: &ErrorShape{Code: "ERR001", Message: "error"},
	})
	if err == nil {
		t.Error("expected error for ok with error")
	}

	// not ok and no error and not progress (invalid)
	err = v.ValidateResponseFrame(&ResponseFrame{
		ID: "123",
		Ok: false,
	})
	if err == nil {
		t.Error("expected error for not ok without error and not progress")
	}

	// progress frame (valid - ok=true with progress)
	err = v.ValidateResponseFrame(&ResponseFrame{
		ID:       "123",
		Ok:       true,
		Progress: true,
	})
	if err != nil {
		t.Errorf("unexpected error for progress frame: %v", err)
	}
}

func TestValidator_ValidateEventFrame(t *testing.T) {
	v := NewValidator()

	// valid frame
	err := v.ValidateEventFrame(&EventFrame{Event: "test.event"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// nil frame
	err = v.ValidateEventFrame(nil)
	if err == nil {
		t.Error("expected error for nil frame")
	}

	// empty event
	err = v.ValidateEventFrame(&EventFrame{})
	if err == nil {
		t.Error("expected error for empty event")
	}
}

func TestValidator_ValidateMethodName(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name      string
		method    string
		wantValid bool
	}{
		// Valid method names
		{"valid simple", "test.method", true},
		{"valid with numbers", "test123.method456", true},
		{"valid with underscore", "test_method.method_name", true},
		{"valid underscore prefix", "_internal.method", true},
		{"valid multi dot", "ns.sub.method", true},
		{"valid deep nesting", "a.b.c.d", true},
		{"valid uppercase", "Test.Method", true},
		{"valid mixed case", "myAPI.v2Handler", true},
		{"valid single char parts", "a.b", true},
		{"valid 64 char parts", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.method", true}, // 63 a's

		// Invalid: missing namespace
		{"invalid single part", "method", false},
		{"invalid no dot", "methodonly", false},

		// Invalid: starts with number
		{"invalid starts with number", "123.method", false},
		{"invalid namespace starts with number", "123test.method", false},
		{"invalid second part starts with number", "test.123method", false},

		// Invalid: special characters
		{"invalid hyphen in namespace", "test-name.method", false},
		{"invalid hyphen in method", "test.method-name", false},
		{"invalid dot in part", "test.me/thod", false},
		{"invalid at sign", "test@method", false},
		{"invalid space", "test me.thod", false},

		// Invalid: empty parts
		{"invalid empty namespace", ".method", false},
		{"invalid empty method", "test.", false},
		{"invalid consecutive dots", "test..method", false},
		{"invalid leading dot", ".test.method", false},
		{"invalid trailing dot", "test.method.", false},

		// Invalid: empty string
		{"invalid empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequestFrame(&RequestFrame{
				ID:     "123",
				Method: tt.method,
			})
			gotValid := err == nil
			if gotValid != tt.wantValid {
				if tt.wantValid {
					t.Errorf("expected valid, got error: %v", err)
				} else {
					t.Errorf("expected invalid, got no error")
				}
			}
		})
	}
}

func TestValidator_ValidatePayloadSize(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name       string
		payloadLen int
		maxPayload int64
		wantErr    bool
	}{
		{"within limit", 100, 1048576, false},
		{"exceeds limit", 200, 100, true},
		{"large payload within limit", 1024 * 1024, 1048576 + 200, false}, // +200 for JSON overhead
		{"large payload exceeds limit", 1048577, 1048576, true},
		{"nil frame", 0, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var frame *RequestFrame
			if tt.payloadLen > 0 {
				// Create valid JSON payload with specified length
				payload := make([]byte, tt.payloadLen)
				for i := range payload {
					payload[i] = 'x'
				}
				params := make(map[string]any)
				params["data"] = string(payload)
				paramsBytes, _ := json.Marshal(params)

				frame = &RequestFrame{
					ID:     "test-id",
					Method: "test.method",
					Params: paramsBytes,
				}
			}
			err := v.ValidatePayloadSize(frame, tt.maxPayload)
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
