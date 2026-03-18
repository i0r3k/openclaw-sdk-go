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
