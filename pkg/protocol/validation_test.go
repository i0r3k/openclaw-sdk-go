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
		Method:    "test.method",
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
				RequestID: "123",
				Method:    tt.method,
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
