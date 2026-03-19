// Package managers provides high-level manager components for OpenClaw SDK.
//
// This package provides:
//   - EventManager: Pub/sub event management
//   - RequestManager: Pending request correlation
//   - ConnectionManager: WebSocket connection lifecycle
//   - ReconnectManager: Automatic reconnection with Fibonacci backoff
package managers

import (
	"context"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/transport"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// EventEmitter is the interface for event emission.
// It defines the minimal interface for components that can emit events.
type EventEmitter interface {
	// Emit emits an event to subscribers.
	Emit(event types.Event)
	// Events returns the event channel.
	Events() <-chan types.Event
}

// EventManagerInterface defines the interface for event management.
// It provides pub/sub functionality for SDK events.
type EventManagerInterface interface {
	// Subscribe adds an event handler for the specified event type.
	// Returns an unsubscribe function.
	Subscribe(eventType types.EventType, handler types.EventHandler) func()
	// Unsubscribe removes an event handler.
	Unsubscribe(eventType types.EventType, handler types.EventHandler)
	// Events returns the event channel.
	Events() <-chan types.Event
	// Emit emits an event.
	Emit(event types.Event)
	// Start begins the event dispatch loop.
	Start()
	// Close shuts down the event manager.
	Close() error
}

// RequestManagerInterface defines the interface for request management.
// It handles request/response correlation.
type RequestManagerInterface interface {
	// SendRequest sends a request and waits for response.
	SendRequest(ctx context.Context, req *protocol.RequestFrame, sendFunc func(*protocol.RequestFrame) error) (*protocol.ResponseFrame, error)
	// HandleResponse handles an incoming response frame.
	HandleResponse(frame *protocol.ResponseFrame)
	// Close cleans up all pending requests.
	Close() error
}

// ConnectionManagerInterface defines the interface for connection management.
// It handles WebSocket connection lifecycle.
type ConnectionManagerInterface interface {
	// Connect establishes a connection to the server.
	Connect(ctx context.Context) error
	// Disconnect closes the connection.
	Disconnect() error
	// State returns the current connection state.
	State() types.ConnectionState
	// Transport returns the underlying transport.
	Transport() transport.Transport
	// Close closes the connection manager.
	Close() error
}

// ReconnectManagerInterface defines the interface for reconnection management.
// It handles automatic reconnection.
type ReconnectManagerInterface interface {
	// SetOnReconnect sets the callback for reconnection attempts.
	SetOnReconnect(f func() error)
	// SetOnReconnectFailed sets the callback for reconnection failures.
	SetOnReconnectFailed(f func(err error))
	// Start begins the reconnection loop.
	Start()
	// Stop stops the reconnection attempts.
	Stop()
}
