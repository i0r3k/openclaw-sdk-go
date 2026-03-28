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

import (
	"sync"
	"time"
)

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

// ReconnectConfig holds configuration for automatic reconnection.
//
// Uses Fibonacci backoff for delay calculation:
//   - InitialDelay: starting delay
//   - MaxDelay: maximum delay cap
//   - BackoffMultiplier: multiplier for Fibonacci sequence (1.618 = golden ratio)
//
// Retry budget (MaxRetries vs MaxAttempts):
//
//   - MaxRetries > 0: use MaxRetries as the limit (ignores MaxAttempts)
//   - MaxRetries == 0 AND MaxAttempts > 0: fall back to MaxAttempts
//   - MaxRetries == 0 AND MaxAttempts == 0: unlimited retries (backward compat)
//   - MaxRetries < 0: treated as 0 (same as unset)
//   - MaxAttempts < 0: treated as 0 (same as unset)
type ReconnectConfig struct {
	// MaxAttempts is the legacy retry budget field.
	// Deprecated: use MaxRetries instead. MaxAttempts=0 means infinite (default).
	MaxAttempts int

	// MaxRetries sets the maximum number of reconnection attempts (FOUND-02).
	//
	// Precedence rules:
	//   - MaxRetries > 0: use MaxRetries as the limit (ignores MaxAttempts)
	//   - MaxRetries == 0 AND MaxAttempts > 0: fall back to MaxAttempts
	//   - MaxRetries == 0 AND MaxAttempts == 0: unlimited retries (backward compat)
	//   - MaxRetries < 0: treated as 0 (same as unset)
	//   - MaxAttempts < 0: treated as 0 (same as unset)
	MaxRetries        int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
}

// DefaultReconnectConfig returns sensible defaults
// Note: InitialDelay must be <= MaxDelay for valid backoff
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxAttempts:       0,  // 0 = infinite (legacy, deprecated)
		MaxRetries:        10, // FOUND-02: sensible production default
		InitialDelay:      1 * time.Second,
		MaxDelay:          60 * time.Second,
		BackoffMultiplier: 1.618,
	}
}

// RequestRateLimiter controls request throughput (FOUND-01).
// Implementations must be safe for concurrent use by multiple goroutines.
type RequestRateLimiter interface {
	Allow() bool
}

// TokenBucketLimiter implements RequestRateLimiter using a token bucket algorithm.
// Rate is in tokens per second; burst is the maximum tokens that can accumulate.
// The limiter starts with a full bucket (burst tokens available immediately).
type TokenBucketLimiter struct {
	rate     float64
	burst    int
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

// NewTokenBucketLimiter creates a new TokenBucketLimiter with the given rate (tokens/sec) and burst.
func NewTokenBucketLimiter(rate float64, burst int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:     rate,
		burst:    burst,
		tokens:   float64(burst),
		lastTime: time.Now(),
	}
}

// Allow consumes one token if available. Returns true if allowed, false if rate limited.
// Thread-safe: safe to call from multiple goroutines.
func (l *TokenBucketLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now
	l.tokens += elapsed * l.rate
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}
