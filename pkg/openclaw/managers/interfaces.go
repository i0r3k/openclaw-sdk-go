package managers

import (
	"context"

	openclaw "github.com/i0r3k/openclaw-sdk-go/pkg/openclaw"
	"github.com/i0r3k/openclaw-sdk-go/pkg/openclaw/protocol"
	"github.com/i0r3k/openclaw-sdk-go/pkg/openclaw/transport"
)

// EventEmitter is the interface for event emission
type EventEmitter interface {
	Emit(event openclaw.Event)
	Events() <-chan openclaw.Event
}

// EventManagerInterface defines the interface for event management
type EventManagerInterface interface {
	Subscribe(eventType openclaw.EventType, handler openclaw.EventHandler) func()
	Unsubscribe(eventType openclaw.EventType, handler openclaw.EventHandler)
	Events() <-chan openclaw.Event
	Emit(event openclaw.Event)
	Start()
	Close() error
}

// RequestManagerInterface defines the interface for request management
type RequestManagerInterface interface {
	SendRequest(ctx context.Context, req *protocol.RequestFrame, sendFunc func(*protocol.RequestFrame) error) (*protocol.ResponseFrame, error)
	HandleResponse(frame *protocol.ResponseFrame)
	Close() error
}

// ConnectionManagerInterface defines the interface for connection management
type ConnectionManagerInterface interface {
	Connect(ctx context.Context) error
	Disconnect() error
	State() openclaw.ConnectionState
	Transport() transport.Transport
	Close() error
}

// ReconnectManagerInterface defines the interface for reconnection management
type ReconnectManagerInterface interface {
	SetOnReconnect(f func() error)
	SetOnReconnectFailed(f func(err error))
	Start()
	Stop()
}
