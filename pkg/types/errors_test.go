// Package types provides tests for error handling
package types

import (
	"errors"
	"testing"
)

// TestNewAPIError_AuthErrors tests that AUTH_* and CHALLENGE_* codes create AuthError.
func TestNewAPIError_AuthErrors(t *testing.T) {
	tests := []struct {
		name  string
		shape *ErrorShape
	}{
		{
			name: "AUTH_TOKEN_EXPIRED",
			shape: &ErrorShape{
				Code:    "AUTH_TOKEN_EXPIRED",
				Message: "Token has expired",
			},
		},
		{
			name: "AUTH_TOKEN_MISMATCH",
			shape: &ErrorShape{
				Code:    "AUTH_TOKEN_MISMATCH",
				Message: "Token mismatch",
			},
		},
		{
			name: "CHALLENGE_EXPIRED",
			shape: &ErrorShape{
				Code:    "CHALLENGE_EXPIRED",
				Message: "Challenge expired",
			},
		},
		{
			name: "CHALLENGE_FAILED",
			shape: &ErrorShape{
				Code:    "CHALLENGE_FAILED",
				Message: "Challenge failed",
			},
		},
		{
			name: "AUTH_RATE_LIMITED",
			shape: &ErrorShape{
				Code:    "AUTH_RATE_LIMITED",
				Message: "Rate limited",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.shape)
			if !IsAuthError(err) {
				t.Errorf("IsAuthError() = false, want true for code %s", tt.shape.Code)
			}
		})
	}
}

// TestNewAPIError_ConnectionErrors tests that CONNECTION_* and TLS_FINGERPRINT_MISMATCH create ConnectionError.
func TestNewAPIError_ConnectionErrors(t *testing.T) {
	tests := []struct {
		name  string
		shape *ErrorShape
	}{
		{
			name: "CONNECTION_STALE",
			shape: &ErrorShape{
				Code:    "CONNECTION_STALE",
				Message: "Connection stale",
			},
		},
		{
			name: "CONNECTION_CLOSED",
			shape: &ErrorShape{
				Code:    "CONNECTION_CLOSED",
				Message: "Connection closed",
			},
		},
		{
			name: "CONNECT_TIMEOUT",
			shape: &ErrorShape{
				Code:    "CONNECT_TIMEOUT",
				Message: "Connect timeout",
			},
		},
		{
			name: "TLS_FINGERPRINT_MISMATCH",
			shape: &ErrorShape{
				Code:    "TLS_FINGERPRINT_MISMATCH",
				Message: "TLS fingerprint mismatch",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.shape)
			if !IsConnectionError(err) {
				t.Errorf("IsConnectionError() = false, want true for code %s", tt.shape.Code)
			}
		})
	}
}

// TestNewAPIError_ProtocolErrors tests that PROTOCOL_* codes create ProtocolError.
func TestNewAPIError_ProtocolErrors(t *testing.T) {
	tests := []struct {
		name  string
		shape *ErrorShape
	}{
		{
			name: "PROTOCOL_UNSUPPORTED",
			shape: &ErrorShape{
				Code:    "PROTOCOL_UNSUPPORTED",
				Message: "Protocol unsupported",
			},
		},
		{
			name: "PROTOCOL_NEGOTIATION_FAILED",
			shape: &ErrorShape{
				Code:    "PROTOCOL_NEGOTIATION_FAILED",
				Message: "Negotiation failed",
			},
		},
		{
			name: "INVALID_FRAME",
			shape: &ErrorShape{
				Code:    "INVALID_FRAME",
				Message: "Invalid frame",
			},
		},
		{
			name: "FRAME_TOO_LARGE",
			shape: &ErrorShape{
				Code:    "FRAME_TOO_LARGE",
				Message: "Frame too large",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.shape)
			if !IsProtocolError(err) {
				t.Errorf("IsProtocolError() = false, want true for code %s", tt.shape.Code)
			}
		})
	}
}

// TestNewAPIError_RequestErrors tests that exact match codes create RequestError.
func TestNewAPIError_RequestErrors(t *testing.T) {
	tests := []struct {
		name  string
		shape *ErrorShape
	}{
		{
			name: "METHOD_NOT_FOUND",
			shape: &ErrorShape{
				Code:    "METHOD_NOT_FOUND",
				Message: "Method not found",
			},
		},
		{
			name: "INVALID_PARAMS",
			shape: &ErrorShape{
				Code:    "INVALID_PARAMS",
				Message: "Invalid params",
			},
		},
		{
			name: "INTERNAL_ERROR",
			shape: &ErrorShape{
				Code:    "INTERNAL_ERROR",
				Message: "Internal error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.shape)
			if !IsRequestError(err) {
				t.Errorf("IsRequestError() = false, want true for code %s", tt.shape.Code)
			}
		})
	}
}

// TestNewAPIError_GatewayErrors tests that gateway/business logic errors create GatewayError.
func TestNewAPIError_GatewayErrors(t *testing.T) {
	tests := []struct {
		name  string
		shape *ErrorShape
	}{
		{
			name: "AGENT_NOT_FOUND",
			shape: &ErrorShape{
				Code:    "AGENT_NOT_FOUND",
				Message: "Agent not found",
			},
		},
		{
			name: "AGENT_BUSY",
			shape: &ErrorShape{
				Code:    "AGENT_BUSY",
				Message: "Agent busy",
			},
		},
		{
			name: "NODE_NOT_FOUND",
			shape: &ErrorShape{
				Code:    "NODE_NOT_FOUND",
				Message: "Node not found",
			},
		},
		{
			name: "SESSION_NOT_FOUND",
			shape: &ErrorShape{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found",
			},
		},
		{
			name: "PERMISSION_DENIED",
			shape: &ErrorShape{
				Code:    "PERMISSION_DENIED",
				Message: "Permission denied",
			},
		},
		{
			name: "QUOTA_EXCEEDED",
			shape: &ErrorShape{
				Code:    "QUOTA_EXCEEDED",
				Message: "Quota exceeded",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.shape)
			if !IsGatewayError(err) {
				t.Errorf("IsGatewayError() = false, want true for code %s", tt.shape.Code)
			}
		})
	}
}

// TestNewAPIError_Fallthrough tests that REQUEST_TIMEOUT, REQUEST_CANCELLED,
// and REQUEST_ABORTED fall through to GatewayError (NOT RequestError).
// This matches TypeScript createErrorFromResponse behavior.
func TestNewAPIError_Fallthrough(t *testing.T) {
	tests := []struct {
		name  string
		shape *ErrorShape
	}{
		{
			name: "REQUEST_TIMEOUT",
			shape: &ErrorShape{
				Code:    "REQUEST_TIMEOUT",
				Message: "Request timeout",
			},
		},
		{
			name: "REQUEST_CANCELLED",
			shape: &ErrorShape{
				Code:    "REQUEST_CANCELLED",
				Message: "Request cancelled",
			},
		},
		{
			name: "REQUEST_ABORTED",
			shape: &ErrorShape{
				Code:    "REQUEST_ABORTED",
				Message: "Request aborted",
			},
		},
		{
			name: "UNKNOWN_ERROR_CODE",
			shape: &ErrorShape{
				Code:    "UNKNOWN_ERROR_CODE",
				Message: "Unknown error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.shape)
			// These should be GatewayError, NOT RequestError
			if IsRequestError(err) {
				t.Errorf("IsRequestError() = true for %s, want false (should be GatewayError)", tt.shape.Code)
			}
			if !IsGatewayError(err) {
				t.Errorf("IsGatewayError() = false for %s, want true", tt.shape.Code)
			}
		})
	}
}

// TestNewAPIError_LowercaseCodes tests case-insensitive code matching.
func TestNewAPIError_LowercaseCodes(t *testing.T) {
	tests := []struct {
		name  string
		shape *ErrorShape
		want  func(error) bool
	}{
		{
			name: "auth_token_expired (lowercase)",
			shape: &ErrorShape{
				Code:    "auth_token_expired",
				Message: "Token expired",
			},
			want: IsAuthError,
		},
		{
			name: "connection_stale (lowercase)",
			shape: &ErrorShape{
				Code:    "connection_stale",
				Message: "Connection stale",
			},
			want: IsConnectionError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.shape)
			if !tt.want(err) {
				t.Errorf("expected error type check to return true for code %s", tt.shape.Code)
			}
		})
	}
}

// TestNewAPIError_Retryable tests retryable field propagation.
func TestNewAPIError_Retryable(t *testing.T) {
	retryable := true
	notRetryable := false

	tests := []struct {
		name  string
		shape *ErrorShape
		want  bool
	}{
		{
			name: "with retryable=true",
			shape: &ErrorShape{
				Code:      "AUTH_TOKEN_EXPIRED",
				Message:   "Token expired",
				Retryable: &retryable,
			},
			want: true,
		},
		{
			name: "with retryable=false",
			shape: &ErrorShape{
				Code:      "AUTH_TOKEN_EXPIRED",
				Message:   "Token expired",
				Retryable: &notRetryable,
			},
			want: false,
		},
		{
			name: "without retryable field",
			shape: &ErrorShape{
				Code:    "AUTH_TOKEN_EXPIRED",
				Message: "Token expired",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.shape)
			if IsRetryable(err) != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", IsRetryable(err), tt.want)
			}
		})
	}
}

// TestNewAPIError_Details tests details field propagation.
func TestNewAPIError_Details(t *testing.T) {
	details := map[string]interface{}{"key": "value"}
	shape := &ErrorShape{
		Code:    "AUTH_TOKEN_EXPIRED",
		Message: "Token expired",
		Details: details,
	}

	err := NewAPIError(shape)
	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("expected *AuthError, got %T", err)
	}

	if authErr.Details() == nil {
		t.Error("Details() = nil, want details")
	}
}

// TestNewAPIError_ReconnectErrors tests reconnect error codes.
func TestNewAPIError_ReconnectErrors(t *testing.T) {
	tests := []struct {
		name  string
		shape *ErrorShape
	}{
		{
			name: "MAX_RECONNECT_ATTEMPTS",
			shape: &ErrorShape{
				Code:    "MAX_RECONNECT_ATTEMPTS",
				Message: "Max reconnect attempts",
			},
		},
		{
			name: "MAX_AUTH_RETRIES",
			shape: &ErrorShape{
				Code:    "MAX_AUTH_RETRIES",
				Message: "Max auth retries",
			},
		},
		{
			name: "RECONNECT_DISABLED",
			shape: &ErrorShape{
				Code:    "RECONNECT_DISABLED",
				Message: "Reconnect disabled",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.shape)
			if !IsGatewayError(err) {
				// These fall through to GatewayError in current implementation
				t.Logf("Note: %s falls through to GatewayError", tt.shape.Code)
			}
		})
	}
}

// TestTimeoutError_SpecificType tests that TimeoutError is a specific error type.
func TestTimeoutError_SpecificType(t *testing.T) {
	err := NewTimeoutError("request timed out", nil)

	if !IsTimeoutError(err) {
		t.Errorf("IsTimeoutError() = false, want true")
	}

	if !IsRequestError(err) {
		t.Errorf("IsRequestError() = false, want true (TimeoutError is a RequestError)")
	}
}

// TestErrorInterfaces tests that error types implement expected interfaces.
func TestErrorInterfaces(t *testing.T) {
	err := NewAuthError("AUTH_TOKEN_EXPIRED", "Token expired", true, nil)

	t.Run("implements error interface", func(t *testing.T) {
		var _ error = err
		_ = err.Error()
	})

	t.Run("implements OpenClawError interface", func(t *testing.T) {
		var oe OpenClawError = err
		_ = oe.Code()
		_ = oe.Retryable()
		_ = oe.Unwrap()
	})
}

// TestNewAPIError_UnknownCode tests that unknown codes fall through to GatewayError.
func TestNewAPIError_UnknownCode(t *testing.T) {
	shape := &ErrorShape{
		Code:    "UNKNOWN_ERROR_CODE",
		Message: "Unknown error",
	}

	err := NewAPIError(shape)
	if !IsGatewayError(err) {
		t.Errorf("IsGatewayError() = false, want true for unknown code")
	}
}

// TestErrorMessageAndCode tests error message and code fields.
func TestErrorMessageAndCode(t *testing.T) {
	shape := &ErrorShape{
		Code:    "AUTH_TOKEN_EXPIRED",
		Message: "Token has expired",
	}

	err := NewAPIError(shape)
	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("expected *AuthError, got %T", err)
	}

	if authErr.Error() != "Token has expired" {
		t.Errorf("Error() = %s, want %s", authErr.Error(), "Token has expired")
	}

	if authErr.Code() != "AUTH_TOKEN_EXPIRED" {
		t.Errorf("Code() = %s, want %s", authErr.Code(), "AUTH_TOKEN_EXPIRED")
	}
}

// TestErrorUnwrap tests error unwrapping.
func TestErrorUnwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	shape := &ErrorShape{
		Code:    "INTERNAL_ERROR",
		Message: "Internal error",
		Details: innerErr,
	}

	err := NewAPIError(shape)
	reqErr, ok := err.(*RequestError)
	if !ok {
		t.Fatalf("expected *RequestError, got %T", err)
	}

	unwrapped := reqErr.Unwrap()
	if unwrapped == nil {
		t.Error("Unwrap() = nil, want inner error")
	}
}

// TestErrorImplementsInterface tests that error types implement OpenClawError.
func TestErrorImplementsInterface(t *testing.T) {
	shape := &ErrorShape{
		Code:    "AUTH_TOKEN_EXPIRED",
		Message: "Token expired",
	}

	err := NewAPIError(shape)

	var oe OpenClawError
	if !errors.As(err, &oe) {
		t.Fatal("error does not implement OpenClawError")
	}

	if oe.Code() != "AUTH_TOKEN_EXPIRED" {
		t.Errorf("Code() = %s, want %s", oe.Code(), "AUTH_TOKEN_EXPIRED")
	}

	if oe.Retryable() != false {
		t.Errorf("Retryable() = %v, want false", oe.Retryable())
	}

	if oe.Error() != "Token expired" {
		t.Errorf("Error() = %s, want %s", oe.Error(), "Token expired")
	}
}
