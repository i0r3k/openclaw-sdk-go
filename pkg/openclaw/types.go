package openclaw

import "time"

// ConnectionState represents the state of the connection
type ConnectionState string

const (
	StateDisconnected      ConnectionState = "disconnected"
	StateConnecting        ConnectionState = "connecting"
	StateConnected         ConnectionState = "connected"
	StateAuthenticating    ConnectionState = "authenticating"
	StateAuthenticated     ConnectionState = "authenticated"
	StateReconnecting      ConnectionState = "reconnecting"
	StateFailed            ConnectionState = "failed"
)

// EventType represents the type of event
type EventType string

const (
	EventConnect      EventType = "connect"
	EventDisconnect   EventType = "disconnect"
	EventError        EventType = "error"
	EventMessage      EventType = "message"
	EventRequest      EventType = "request"
	EventResponse     EventType = "response"
	EventTick         EventType = "tick"
	EventGap          EventType = "gap"
	EventStateChange  EventType = "stateChange"
)

// Event represents a generic event
type Event struct {
	Type      EventType
	Payload   interface{}
	Err       error
	Timestamp time.Time
}

// EventHandler is a function that handles events
type EventHandler func(Event)

// ReconnectConfig holds reconnection settings
type ReconnectConfig struct {
	MaxAttempts       int
	InitialDelay     time.Duration
	MaxDelay         time.Duration
	BackoffMultiplier float64
}

// DefaultReconnectConfig returns sensible defaults
// Note: InitialDelay must be <= MaxDelay for valid backoff
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxAttempts:       0, // 0 = infinite
		InitialDelay:     1 * time.Second,
		MaxDelay:         60 * time.Second,
		BackoffMultiplier: 1.618,
	}
}

// Validate validates the reconnect configuration
func (r ReconnectConfig) Validate() error {
	if r.InitialDelay > r.MaxDelay {
		return &ValidationError{&BaseError{ErrCodeValidation, "InitialDelay must be <= MaxDelay", nil}}
	}
	return nil
}
