package openclaw

import "github.com/i0r3k/openclaw-sdk-go/pkg/types"

// Re-export shared types from pkg/types for convenience
type ConnectionState = types.ConnectionState
type EventType = types.EventType
type Event = types.Event
type EventHandler = types.EventHandler
type ReconnectConfig = types.ReconnectConfig

// Re-export constants
const (
	StateDisconnected      = types.StateDisconnected
	StateConnecting        = types.StateConnecting
	StateConnected         = types.StateConnected
	StateAuthenticating    = types.StateAuthenticating
	StateAuthenticated     = types.StateAuthenticated
	StateReconnecting      = types.StateReconnecting
	StateFailed            = types.StateFailed

	EventConnect      = types.EventConnect
	EventDisconnect   = types.EventDisconnect
	EventError        = types.EventError
	EventMessage      = types.EventMessage
	EventRequest      = types.EventRequest
	EventResponse     = types.EventResponse
	EventTick         = types.EventTick
	EventGap          = types.EventGap
	EventStateChange  = types.EventStateChange
)

// DefaultReconnectConfig returns sensible defaults
func DefaultReconnectConfig() ReconnectConfig {
	return types.DefaultReconnectConfig()
}

// ValidateReconnectConfig validates the reconnect configuration
func ValidateReconnectConfig(r ReconnectConfig) error {
	if r.InitialDelay > r.MaxDelay {
		return types.NewValidationError("InitialDelay must be <= MaxDelay", nil)
	}
	return nil
}
