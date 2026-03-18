package openclaw

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	err := NewError(ErrCodeConnection, "test error", nil)
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got '%s'", err.Error())
	}
	if err.Code() != ErrCodeConnection {
		t.Errorf("expected CONNECTION_ERROR, got %s", err.Code())
	}
}

func TestIs(t *testing.T) {
	err := NewConnectionError("connection failed", nil)
	if !Is(err, ErrCodeConnection) {
		t.Error("expected Is to return true for matching code")
	}
	if Is(err, ErrCodeAuth) {
		t.Error("expected Is to return false for non-matching code")
	}
}

func TestAs(t *testing.T) {
	baseErr := NewError(ErrCodeConnection, "test", nil)
	var target OpenClawError
	if !As(baseErr, &target) {
		t.Error("expected As to return true")
	}
	if target.Code() != ErrCodeConnection {
		t.Error("expected to extract error with correct code")
	}
}

func TestErrorUnwrap(t *testing.T) {
	original := errors.New("original error")
	err := NewConnectionError("wrapped error", original)

	if !errors.Is(err, original) {
		t.Error("expected unwrap to return original error")
	}
}

func TestAllErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err     OpenClawError
		expected ErrorCode
	}{
		{"ConnectionError", NewConnectionError("test", nil), ErrCodeConnection},
		{"AuthError", NewAuthError("test", nil), ErrCodeAuth},
		{"TimeoutError", NewTimeoutError("test", nil), ErrCodeTimeout},
		{"ProtocolError", NewProtocolError("test", nil), ErrCodeProtocol},
		{"ValidationError", NewValidationError("test", nil), ErrCodeValidation},
		{"TransportError", NewTransportError("test", nil), ErrCodeTransport},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.err.Code())
			}
		})
	}
}
