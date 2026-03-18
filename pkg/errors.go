package openclaw

import "github.com/i0r3k/openclaw-sdk-go/pkg/types"

// Re-export error types from pkg/types for convenience
type ErrorCode = types.ErrorCode
type OpenClawError = types.OpenClawError
type BaseError = types.BaseError
type ConnectionError = types.ConnectionError
type AuthError = types.AuthError
type TimeoutError = types.TimeoutError
type ProtocolError = types.ProtocolError
type ValidationError = types.ValidationError
type TransportError = types.TransportError

// Re-export error code constants
const (
	ErrCodeConnection   = types.ErrCodeConnection
	ErrCodeAuth        = types.ErrCodeAuth
	ErrCodeTimeout     = types.ErrCodeTimeout
	ErrCodeProtocol    = types.ErrCodeProtocol
	ErrCodeValidation  = types.ErrCodeValidation
	ErrCodeTransport   = types.ErrCodeTransport
	ErrCodeUnknown     = types.ErrCodeUnknown
)

// Re-export error constructors
var (
	NewError             = types.NewError
	NewConnectionError   = types.NewConnectionError
	NewAuthError         = types.NewAuthError
	NewTimeoutError      = types.NewTimeoutError
	NewProtocolError     = types.NewProtocolError
	NewValidationError   = types.NewValidationError
	NewTransportError    = types.NewTransportError
	Is                   = types.Is
	As                   = types.As
)

