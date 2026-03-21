// Package types provides shared types for the OpenClaw SDK.
//
// This package contains error types following Go best practices:
//   - Custom error types with error code and cause chain
//   - Standard errors.Is() and errors.As() compatibility
//   - Typed errors for different failure categories
package types

import (
	"errors"
	"strings"
)

// ============================================================================
// Error Codes
// ============================================================================

// ErrorCode represents an error code string.
type ErrorCode string

// AuthErrorCode represents authentication error codes.
type AuthErrorCode string

// Authentication error codes from src/errors.ts
const (
	AuthErrChallengeExpired AuthErrorCode = "CHALLENGE_EXPIRED"
	AuthErrChallengeFailed  AuthErrorCode = "CHALLENGE_FAILED"
	AuthErrTokenExpired     AuthErrorCode = "AUTH_TOKEN_EXPIRED"
	AuthErrTokenMismatch    AuthErrorCode = "AUTH_TOKEN_MISMATCH"
	AuthErrRateLimited      AuthErrorCode = "AUTH_RATE_LIMITED"
	AuthErrDeviceRejected   AuthErrorCode = "AUTH_DEVICE_REJECTED"
	AuthErrPasswordInvalid  AuthErrorCode = "AUTH_PASSWORD_INVALID"
	AuthErrNotSupported     AuthErrorCode = "AUTH_NOT_SUPPORTED"
	AuthErrConfiguration    AuthErrorCode = "AUTH_CONFIGURATION_ERROR"
)

// ConnectionErrorCode represents connection error codes.
type ConnectionErrorCode string

// Connection error codes from src/errors.ts
const (
	ConnectionErrTLSFingerprint ConnectionErrorCode = "TLS_FINGERPRINT_MISMATCH"
	ConnectionErrStale          ConnectionErrorCode = "CONNECTION_STALE"
	ConnectionErrClosed         ConnectionErrorCode = "CONNECTION_CLOSED"
	ConnectionErrTimeout        ConnectionErrorCode = "CONNECT_TIMEOUT"
	ConnectionErrRefused        ConnectionErrorCode = "CONNECTION_REFUSED"
	ConnectionErrNetwork        ConnectionErrorCode = "NETWORK_ERROR"
	ConnectionErrProtocol       ConnectionErrorCode = "PROTOCOL_ERROR"
)

// ProtocolErrorCode represents protocol error codes.
type ProtocolErrorCode string

// Protocol error codes from src/errors.ts
const (
	ProtocolErrUnsupported       ProtocolErrorCode = "PROTOCOL_UNSUPPORTED"
	ProtocolErrNegotiationFailed ProtocolErrorCode = "PROTOCOL_NEGOTIATION_FAILED"
	ProtocolErrInvalidFrame      ProtocolErrorCode = "INVALID_FRAME"
	ProtocolErrFrameTooLarge     ProtocolErrorCode = "FRAME_TOO_LARGE"
)

// RequestErrorCode represents request error codes.
type RequestErrorCode string

// Request error codes from src/errors.ts
// NOTE: REQUEST_TIMEOUT, REQUEST_CANCELLED, REQUEST_ABORTED are defined but
// createErrorFromResponse routes them to GatewayError, not RequestError.
const (
	RequestErrMethodNotFound RequestErrorCode = "METHOD_NOT_FOUND"
	RequestErrInvalidParams  RequestErrorCode = "INVALID_PARAMS"
	RequestErrInternal       RequestErrorCode = "INTERNAL_ERROR"
)

// GatewayErrorCode represents gateway/business logic error codes.
type GatewayErrorCode string

// Gateway error codes from src/errors.ts
const (
	GatewayErrAgentNotFound    GatewayErrorCode = "AGENT_NOT_FOUND"
	GatewayErrAgentBusy        GatewayErrorCode = "AGENT_BUSY"
	GatewayErrNodeNotFound     GatewayErrorCode = "NODE_NOT_FOUND"
	GatewayErrNodeOffline      GatewayErrorCode = "NODE_OFFLINE"
	GatewayErrSessionNotFound  GatewayErrorCode = "SESSION_NOT_FOUND"
	GatewayErrSessionExpired   GatewayErrorCode = "SESSION_EXPIRED"
	GatewayErrPermissionDenied GatewayErrorCode = "PERMISSION_DENIED"
	GatewayErrQuotaExceeded    GatewayErrorCode = "QUOTA_EXCEEDED"
)

// ReconnectErrorCode represents reconnection error codes.
type ReconnectErrorCode string

// Reconnection error codes from src/errors.ts
const (
	ReconnectErrMaxAttempts    ReconnectErrorCode = "MAX_RECONNECT_ATTEMPTS"
	ReconnectErrMaxAuthRetries ReconnectErrorCode = "MAX_AUTH_RETRIES"
	ReconnectErrDisabled       ReconnectErrorCode = "RECONNECT_DISABLED"
)

// ============================================================================
// Error Shape (Wire Protocol)
// ============================================================================

// ErrorShape represents the error structure from wire protocol.
// Note: This mirrors protocol.ErrorShape but uses 'any' for Details
// to support runtime-constructed errors (not just wire-deserialized ones).
// protocol.ErrorShape uses json.RawMessage for wire-level JSON handling.
type ErrorShape struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	Retryable    *bool  `json:"retryable,omitempty"`
	Details      any    `json:"details,omitempty"`
	RetryAfterMs *int64 `json:"retryAfterMs,omitempty"`
}

// ============================================================================
// Base Error
// ============================================================================

// OpenClawError is the base error interface for all OpenClaw SDK errors.
type OpenClawError interface {
	error
	Code() string
	Retryable() bool
	Details() any
	Unwrap() error
}

// BaseError is the base error struct implementing OpenClawError.
type BaseError struct {
	code      string
	message   string
	retryable bool
	details   any
	err       error
}

func (e *BaseError) Error() string   { return e.message }
func (e *BaseError) Code() string    { return e.code }
func (e *BaseError) Retryable() bool { return e.retryable }
func (e *BaseError) Details() any    { return e.details }
func (e *BaseError) Unwrap() error   { return e.err }

// ============================================================================
// Error Types
// ============================================================================

// AuthError represents an authentication error.
type AuthError struct {
	*BaseError
}

// NewAuthError creates a new authentication error.
func NewAuthError(code, message string, retryable bool, details any) *AuthError {
	return &AuthError{&BaseError{
		code:      string(code),
		message:   message,
		retryable: retryable,
		details:   details,
	}}
}

// ConnectionError represents a connection error.
type ConnectionError struct {
	*BaseError
}

// NewConnectionError creates a new connection error.
func NewConnectionError(code, message string, retryable bool, details any) *ConnectionError {
	return &ConnectionError{&BaseError{
		code:      string(code),
		message:   message,
		retryable: retryable,
		details:   details,
	}}
}

// ProtocolError represents a protocol error.
type ProtocolError struct {
	*BaseError
}

// NewProtocolError creates a new protocol error.
func NewProtocolError(code, message string, retryable bool, details any) *ProtocolError {
	return &ProtocolError{&BaseError{
		code:      string(code),
		message:   message,
		retryable: retryable,
		details:   details,
	}}
}

// RequestError represents a request error.
type RequestError struct {
	*BaseError
}

// NewRequestError creates a new request error.
func NewRequestError(code, message string, retryable bool, details any) *RequestError {
	return &RequestError{&BaseError{
		code:      string(code),
		message:   message,
		retryable: retryable,
		details:   details,
	}}
}

// GatewayError represents a gateway/business logic error.
type GatewayError struct {
	*BaseError
}

// NewGatewayError creates a new gateway error.
func NewGatewayError(code, message string, retryable bool, details any) *GatewayError {
	return &GatewayError{&BaseError{
		code:      string(code),
		message:   message,
		retryable: retryable,
		details:   details,
	}}
}

// ReconnectError represents a reconnection error.
type ReconnectError struct {
	*BaseError
}

// NewReconnectError creates a new reconnect error.
func NewReconnectError(code, message string, retryable bool, details any) *ReconnectError {
	return &ReconnectError{&BaseError{
		code:      string(code),
		message:   message,
		retryable: retryable,
		details:   details,
	}}
}

// TimeoutError represents a timeout error.
type TimeoutError struct {
	*RequestError
}

// NewTimeoutError creates a new timeout error.
func NewTimeoutError(message string, details any) *TimeoutError {
	return &TimeoutError{&RequestError{&BaseError{
		code:      "REQUEST_TIMEOUT",
		message:   message,
		retryable: true,
		details:   details,
	}}}
}

// CancelledError represents a cancelled request error.
type CancelledError struct {
	*RequestError
}

// NewCancelledError creates a new cancelled error.
func NewCancelledError(message string, details any) *CancelledError {
	return &CancelledError{&RequestError{&BaseError{
		code:      "REQUEST_CANCELLED",
		message:   message,
		retryable: false,
		details:   details,
	}}}
}

// AbortError represents an aborted request error.
type AbortError struct {
	*RequestError
}

// NewAbortError creates a new abort error.
func NewAbortError(message string, details any) *AbortError {
	return &AbortError{&RequestError{&BaseError{
		code:      "REQUEST_ABORTED",
		message:   message,
		retryable: false,
		details:   details,
	}}}
}

// ============================================================================
// Error Factory
// ============================================================================

// NewAPIError creates an appropriate error from an ErrorShape.
// This matches the TypeScript createErrorFromResponse logic.
func NewAPIError(shape *ErrorShape) error {
	code := strings.ToUpper(shape.Code)
	retryable := false
	if shape.Retryable != nil {
		retryable = *shape.Retryable
	}

	// 1. Auth errors: AUTH_* or CHALLENGE_* (prefix match)
	if strings.HasPrefix(code, "AUTH_") || strings.HasPrefix(code, "CHALLENGE_") {
		return NewAuthError(code, shape.Message, retryable, shape.Details)
	}

	// 2. Connection errors: CONNECTION_*, CONNECT_*, NETWORK_*, or TLS_FINGERPRINT_MISMATCH
	if strings.HasPrefix(code, "CONNECTION_") || strings.HasPrefix(code, "CONNECT_") ||
		strings.HasPrefix(code, "NETWORK_") || code == "TLS_FINGERPRINT_MISMATCH" || code == "PROTOCOL_ERROR" {
		return NewConnectionError(code, shape.Message, retryable, shape.Details)
	}

	// 3. Protocol errors: PROTOCOL_* (prefix match) or INVALID_FRAME, FRAME_TOO_LARGE
	if strings.HasPrefix(code, "PROTOCOL_") || code == "INVALID_FRAME" || code == "FRAME_TOO_LARGE" {
		return NewProtocolError(code, shape.Message, retryable, shape.Details)
	}

	// 4. Request errors: exact match only for METHOD_NOT_FOUND, INVALID_PARAMS, INTERNAL_ERROR
	// NOTE: REQUEST_TIMEOUT, REQUEST_CANCELLED, REQUEST_ABORTED fall through to GatewayError
	// This matches TypeScript createErrorFromResponse behavior
	if code == "METHOD_NOT_FOUND" || code == "INVALID_PARAMS" || code == "INTERNAL_ERROR" {
		return NewRequestError(code, shape.Message, retryable, shape.Details)
	}

	// 5. All other codes fall through to GatewayError (including REQUEST_TIMEOUT,
	// REQUEST_CANCELLED, REQUEST_ABORTED, and unknown codes)
	return NewGatewayError(code, shape.Message, retryable, shape.Details)
}

// ============================================================================
// Error Type Guards
// ============================================================================

// IsAuthError checks if the error is an AuthError.
func IsAuthError(err error) bool {
	var e *AuthError
	return errors.As(err, &e)
}

// IsConnectionError checks if the error is a ConnectionError.
func IsConnectionError(err error) bool {
	var e *ConnectionError
	return errors.As(err, &e)
}

// IsProtocolError checks if the error is a ProtocolError.
func IsProtocolError(err error) bool {
	var e *ProtocolError
	return errors.As(err, &e)
}

// IsRequestError checks if the error is a RequestError.
// TimeoutError, CancelledError, and AbortError embed RequestError,
// so errors.As already matches them through the embedding chain.
func IsRequestError(err error) bool {
	var e *RequestError
	return errors.As(err, &e)
}

// IsGatewayError checks if the error is a GatewayError.
func IsGatewayError(err error) bool {
	var e *GatewayError
	return errors.As(err, &e)
}

// IsReconnectError checks if the error is a ReconnectError.
func IsReconnectError(err error) bool {
	var e *ReconnectError
	return errors.As(err, &e)
}

// IsTimeoutError checks if the error is a TimeoutError.
func IsTimeoutError(err error) bool {
	var e *TimeoutError
	return errors.As(err, &e)
}

// IsCancelledError checks if the error is a CancelledError.
func IsCancelledError(err error) bool {
	var e *CancelledError
	return errors.As(err, &e)
}

// IsAbortError checks if the error is an AbortError.
func IsAbortError(err error) bool {
	var e *AbortError
	return errors.As(err, &e)
}

// IsRetryable checks if the error is retryable.
func IsRetryable(err error) bool {
	var e OpenClawError
	if errors.As(err, &e) {
		return e.Retryable()
	}
	return false
}

// ============================================================================
// Errors.Unwrap Support
// ============================================================================

func (e *AuthError) Unwrap() error       { return e.BaseError }
func (e *ConnectionError) Unwrap() error { return e.BaseError }
func (e *ProtocolError) Unwrap() error   { return e.BaseError }
func (e *RequestError) Unwrap() error    { return e.BaseError }
func (e *GatewayError) Unwrap() error    { return e.BaseError }
func (e *ReconnectError) Unwrap() error  { return e.BaseError }
func (e *TimeoutError) Unwrap() error    { return e.RequestError }
func (e *CancelledError) Unwrap() error  { return e.RequestError }
func (e *AbortError) Unwrap() error      { return e.RequestError }
