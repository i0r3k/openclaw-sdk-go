// Package types provides tests for error handling
package types

import (
	"errors"
	"testing"
)

// TestNewError tests the generic error constructor
func TestNewError(t *testing.T) {
	tests := []struct {
		name    string
		code    ErrorCode
		message string
		err     error
	}{
		{
			name:    "basic error",
			code:    ErrCodeUnknown,
			message: "something went wrong",
			err:     nil,
		},
		{
			name:    "error with cause",
			code:    ErrCodeConnection,
			message: "connection failed",
			err:     errors.New("underlying error"),
		},
		{
			name:    "all error codes",
			code:    ErrCodeProtocol,
			message: "protocol error",
			err:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code, tt.message, tt.err)

			if err.Error() != tt.message {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.message)
			}

			if err.Code() != tt.code {
				t.Errorf("Code() = %s, want %s", err.Code(), tt.code)
			}

			if unwrapped := err.Unwrap(); unwrapped != tt.err {
				t.Errorf("Unwrap() = %v, want %v", unwrapped, tt.err)
			}
		})
	}
}

// TestErrorConstructors tests all specific error type constructors
func TestErrorConstructors(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name        string
		constructor func(string, error) OpenClawError
		wantCode    ErrorCode
	}{
		{
			name:        "ConnectionError",
			constructor: NewConnectionError,
			wantCode:    ErrCodeConnection,
		},
		{
			name:        "AuthError",
			constructor: NewAuthError,
			wantCode:    ErrCodeAuth,
		},
		{
			name:        "TimeoutError",
			constructor: NewTimeoutError,
			wantCode:    ErrCodeTimeout,
		},
		{
			name:        "ProtocolError",
			constructor: NewProtocolError,
			wantCode:    ErrCodeProtocol,
		},
		{
			name:        "ValidationError",
			constructor: NewValidationError,
			wantCode:    ErrCodeValidation,
		},
		{
			name:        "TransportError",
			constructor: NewTransportError,
			wantCode:    ErrCodeTransport,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("test message", baseErr)

			if err.Error() != "test message" {
				t.Errorf("Error() = %q, want %q", err.Error(), "test message")
			}

			if err.Code() != tt.wantCode {
				t.Errorf("Code() = %s, want %s", err.Code(), tt.wantCode)
			}

			if unwrapped := err.Unwrap(); unwrapped != baseErr {
				t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
			}

			// Verify type matches expected error type
			switch tt.wantCode {
			case ErrCodeConnection:
				if _, ok := err.(*ConnectionError); !ok {
					t.Errorf("error type = %T, want *ConnectionError", err)
				}
			case ErrCodeAuth:
				if _, ok := err.(*AuthError); !ok {
					t.Errorf("error type = %T, want *AuthError", err)
				}
			case ErrCodeTimeout:
				if _, ok := err.(*TimeoutError); !ok {
					t.Errorf("error type = %T, want *TimeoutError", err)
				}
			case ErrCodeProtocol:
				if _, ok := err.(*ProtocolError); !ok {
					t.Errorf("error type = %T, want *ProtocolError", err)
				}
			case ErrCodeValidation:
				if _, ok := err.(*ValidationError); !ok {
					t.Errorf("error type = %T, want *ValidationError", err)
				}
			case ErrCodeTransport:
				if _, ok := err.(*TransportError); !ok {
					t.Errorf("error type = %T, want *TransportError", err)
				}
			}
		})
	}
}

// TestErrorConstructors_NilCause tests error constructors with nil cause
func TestErrorConstructors_NilCause(t *testing.T) {
	tests := []struct {
		name        string
		constructor func(string, error) OpenClawError
	}{
		{"ConnectionError", NewConnectionError},
		{"AuthError", NewAuthError},
		{"TimeoutError", NewTimeoutError},
		{"ProtocolError", NewProtocolError},
		{"ValidationError", NewValidationError},
		{"TransportError", NewTransportError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("test message", nil)

			if err.Unwrap() != nil {
				t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
			}
		})
	}
}

// TestIs tests the Is function for error code matching
func TestIs(t *testing.T) {
	baseErr := errors.New("base")

	tests := []struct {
		name     string
		err      error
		code     ErrorCode
		wantBool bool
	}{
		{
			name:     "matching connection error",
			err:      NewConnectionError("conn failed", baseErr),
			code:     ErrCodeConnection,
			wantBool: true,
		},
		{
			name:     "non-matching error code",
			err:      NewAuthError("auth failed", baseErr),
			code:     ErrCodeConnection,
			wantBool: false,
		},
		{
			name:     "nil error",
			err:      nil,
			code:     ErrCodeConnection,
			wantBool: false,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			code:     ErrCodeUnknown,
			wantBool: false,
		},
		{
			name:     "wrapped error - matches inner",
			err:      errors.New("wrapped: " + NewConnectionError("conn failed", baseErr).Error()),
			code:     ErrCodeConnection,
			wantBool: false,
		},
		{
			name:     "timeout error matches",
			err:      NewTimeoutError("timeout", nil),
			code:     ErrCodeTimeout,
			wantBool: true,
		},
		{
			name:     "protocol error matches",
			err:      NewProtocolError("invalid frame", nil),
			code:     ErrCodeProtocol,
			wantBool: true,
		},
		{
			name:     "validation error matches",
			err:      NewValidationError("invalid input", nil),
			code:     ErrCodeValidation,
			wantBool: true,
		},
		{
			name:     "transport error matches",
			err:      NewTransportError("send failed", nil),
			code:     ErrCodeTransport,
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Is(tt.err, tt.code)
			if got != tt.wantBool {
				t.Errorf("Is() = %v, want %v", got, tt.wantBool)
			}
		})
	}
}

// TestAs tests the As function for error type casting
func TestAs(t *testing.T) {
	baseErr := errors.New("base")

	tests := []struct {
		name      string
		err       error
		wantMatch bool
		wantCode  ErrorCode
	}{
		{
			name:      "connection error",
			err:       NewConnectionError("conn failed", baseErr),
			wantMatch: true,
			wantCode:  ErrCodeConnection,
		},
		{
			name:      "auth error",
			err:       NewAuthError("auth failed", baseErr),
			wantMatch: true,
			wantCode:  ErrCodeAuth,
		},
		{
			name:      "nil error",
			err:       nil,
			wantMatch: false,
		},
		{
			name:      "standard error",
			err:       errors.New("standard"),
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target OpenClawError
			got := As(tt.err, &target)

			if got != tt.wantMatch {
				t.Errorf("As() = %v, want %v", got, tt.wantMatch)
			}

			if tt.wantMatch && target.Code() != tt.wantCode {
				t.Errorf("after As(), target.Code() = %s, want %s", target.Code(), tt.wantCode)
			}

			if !tt.wantMatch && target != nil {
				t.Errorf("after failed As(), target should be nil, got %v", target)
			}
		})
	}
}

// TestErrorChaining tests error wrapping and unwrapping
func TestErrorChaining(t *testing.T) {
	// Create a chain: base -> validation -> protocol
	baseErr := errors.New("base error")
	validErr := NewValidationError("invalid input", baseErr)
	protoErr := NewProtocolError("frame error", validErr)

	t.Run("unwrap once", func(t *testing.T) {
		unwrapped := protoErr.Unwrap()
		if unwrapped == nil {
			t.Fatal("Unwrap() returned nil")
		}

		if unwrapped.Error() != "invalid input" {
			t.Errorf("Unwrap().Error() = %s, want 'invalid input'", unwrapped.Error())
		}
	})

	t.Run("unwrap twice using errors.Is", func(t *testing.T) {
		// errors.Is should find the base error
		if !errors.Is(protoErr, baseErr) {
			t.Error("errors.Is(protoErr, baseErr) = false, want true")
		}
	})

	t.Run("chain code preservation", func(t *testing.T) {
		if protoErr.Code() != ErrCodeProtocol {
			t.Errorf("Code() = %s, want %s", protoErr.Code(), ErrCodeProtocol)
		}
	})
}

// TestErrorCodeValues tests all error code constants
func TestErrorCodeValues(t *testing.T) {
	tests := []struct {
		code ErrorCode
		want string
	}{
		{ErrCodeConnection, "CONNECTION_ERROR"},
		{ErrCodeAuth, "AUTH_ERROR"},
		{ErrCodeTimeout, "TIMEOUT"},
		{ErrCodeProtocol, "PROTOCOL_ERROR"},
		{ErrCodeValidation, "VALIDATION_ERROR"},
		{ErrCodeTransport, "TRANSPORT_ERROR"},
		{ErrCodeUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.code) != tt.want {
				t.Errorf("ErrorCode = %s, want %s", tt.code, tt.want)
			}
		})
	}
}

// TestErrorInterfaces tests that error types implement expected interfaces
func TestErrorInterfaces(t *testing.T) {
	err := NewConnectionError("test", nil)

	t.Run("implements error interface", func(t *testing.T) {
		var _ error = err
		_ = err.Error()
	})

	t.Run("implements OpenClawError interface", func(t *testing.T) {
		var _ = err
		_ = err.Code()
		_ = err.Unwrap()
	})
}

// BenchmarkErrorCreation benchmarks error creation performance
func BenchmarkErrorCreation(b *testing.B) {
	baseErr := errors.New("base")

	b.Run("NewConnectionError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewConnectionError("connection failed", baseErr)
		}
	})

	b.Run("NewError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewError(ErrCodeConnection, "connection failed", baseErr)
		}
	})

	b.Run("Is", func(b *testing.B) {
		err := NewConnectionError("test", baseErr)
		for i := 0; i < b.N; i++ {
			_ = Is(err, ErrCodeConnection)
		}
	})
}
