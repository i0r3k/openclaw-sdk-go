// Package types provides shared types for the OpenClaw SDK.
//
// This package contains core types used throughout the SDK, including:
//   - ConnectionState: Connection lifecycle states
//   - EventType: Event types for pub/sub
//   - Event: Generic event structure
//   - ReconnectConfig: Reconnection behavior configuration
//
// These types are re-exported from the main openclaw package for convenience.
package types

import "time"

// ConnectionState represents the state of the connection.
// It follows the connection lifecycle: Disconnected -> Connecting -> Connected -> Authenticated.
// States like Reconnecting and Failed are transitional states.
type ConnectionState string

const (
	StateDisconnected   ConnectionState = "disconnected"
	StateConnecting     ConnectionState = "connecting"
	StateConnected      ConnectionState = "connected"
	StateAuthenticating ConnectionState = "authenticating"
	StateAuthenticated  ConnectionState = "authenticated"
	StateReconnecting   ConnectionState = "reconnecting"
	StateFailed         ConnectionState = "failed"
)

// EventType represents the type of event in the pub/sub system.
// Events flow through the system and can be subscribed to by handlers.
type EventType string

const (
	EventConnect     EventType = "connect"
	EventDisconnect  EventType = "disconnect"
	EventError       EventType = "error"
	EventMessage     EventType = "message"
	EventRequest     EventType = "request"
	EventResponse    EventType = "response"
	EventTick        EventType = "tick"
	EventGap         EventType = "gap"
	EventStateChange EventType = "stateChange"
)

// Event represents a generic event in the system.
// Events are emitted by various components and delivered to subscribers.
type Event struct {
	Type      EventType
	Payload   any
	Err       error
	Timestamp time.Time
}

// EventHandler is a function type that handles events.
// It receives events and processes them based on the event type.
type EventHandler func(Event)

// ReconnectConfig holds reconnection settings for automatic reconnection.
// Uses Fibonacci backoff for delay calculation:
//   - InitialDelay: starting delay
//   - MaxDelay: maximum delay cap
//   - BackoffMultiplier: multiplier for Fibonacci sequence (1.618 = golden ratio)
type ReconnectConfig struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
}

// DefaultReconnectConfig returns sensible defaults
// Note: InitialDelay must be <= MaxDelay for valid backoff
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxAttempts:       0, // 0 = infinite
		InitialDelay:      1 * time.Second,
		MaxDelay:          60 * time.Second,
		BackoffMultiplier: 1.618,
	}
}
