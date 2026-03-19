// Package openclaw provides the OpenClaw WebSocket SDK for Go.
//
// This is the main entry point for the OpenClaw SDK. It provides a client
// for connecting to WebSocket servers with support for:
//
//   - Connection management with automatic state transitions
//   - Request/response pattern with correlation
//   - Event subscription and dispatching
//   - Automatic reconnection with Fibonacci backoff
//   - TLS/SSL support
//   - Custom authentication handlers
//
// Basic usage:
//
//	client, err := openclaw.NewClient(
//	    openclaw.WithURL("wss://example.com/ws"),
//	    openclaw.WithAuthHandler(handler),
//	)
//	if err != nil {
//	    // handle error
//	}
//	defer client.Close()
//
//	err = client.Connect(ctx)
//	// use client to send requests
package openclaw

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/auth"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/managers"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/transport"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// Re-export types from pkg/types for convenience
type ConnectionState = types.ConnectionState
type EventType = types.EventType
type Event = types.Event
type EventHandler = types.EventHandler
type ReconnectConfig = types.ReconnectConfig

// Re-export state constants
const (
	StateDisconnected   = types.StateDisconnected
	StateConnecting     = types.StateConnecting
	StateConnected      = types.StateConnected
	StateAuthenticating = types.StateAuthenticating
	StateAuthenticated  = types.StateAuthenticated
	StateReconnecting   = types.StateReconnecting
	StateFailed         = types.StateFailed
)

// Re-export event constants
const (
	EventConnect     = types.EventConnect
	EventDisconnect  = types.EventDisconnect
	EventError       = types.EventError
	EventMessage     = types.EventMessage
	EventRequest     = types.EventRequest
	EventResponse    = types.EventResponse
	EventTick        = types.EventTick
	EventGap         = types.EventGap
	EventStateChange = types.EventStateChange
)

// Re-export DefaultReconnectConfig function
var DefaultReconnectConfig = types.DefaultReconnectConfig

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
	ErrCodeConnection = types.ErrCodeConnection
	ErrCodeAuth       = types.ErrCodeAuth
	ErrCodeTimeout    = types.ErrCodeTimeout
	ErrCodeProtocol   = types.ErrCodeProtocol
	ErrCodeValidation = types.ErrCodeValidation
	ErrCodeTransport  = types.ErrCodeTransport
	ErrCodeUnknown    = types.ErrCodeUnknown
)

// Re-export error constructors
var (
	NewError           = types.NewError
	NewConnectionError = types.NewConnectionError
	NewAuthError       = types.NewAuthError
	NewTimeoutError    = types.NewTimeoutError
	NewProtocolError   = types.NewProtocolError
	NewValidationError = types.NewValidationError
	NewTransportError  = types.NewTransportError
	Is                 = types.Is
	As                 = types.As
)

// Re-export logger types and functions from pkg/types for convenience
type Logger = types.Logger
type DefaultLogger = types.DefaultLogger
type NopLogger = types.NopLogger

// Re-export logger functions
var (
	NewDefaultLogger           = types.NewDefaultLogger
	NewDefaultLoggerWithWriter = types.NewDefaultLoggerWithWriter
	WithContext                = types.WithContext
	FromContext                = types.FromContext
)

// ClientConfig holds client configuration for creating an OpenClaw client.
// It contains all the settings needed to connect to a WebSocket server.
type ClientConfig struct {
	URL              string               // WebSocket server URL (e.g., wss://example.com/ws)
	AuthHandler      auth.AuthHandler     // Authentication handler
	ReconnectEnabled bool                 // Enable automatic reconnection
	ReconnectConfig  *ReconnectConfig     // Reconnection configuration
	Logger           Logger               // Logger instance for debugging
	Header           map[string][]string  // Custom HTTP headers for WebSocket handshake
	TLSConfig        *transport.TLSConfig // TLS/SSL configuration
	EventBufferSize  int                  // Buffer size for event channel
}

// DefaultClientConfig returns default client configuration.
// Default event buffer size is 100, and uses a default logger.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		EventBufferSize: 100,
		Logger:          NewDefaultLogger(),
	}
}

// ClientOption is a functional option for configuring the client.
// It modifies the ClientConfig and returns an error if the configuration is invalid.
type ClientOption func(*ClientConfig) error

// WithURL sets the WebSocket URL.
// Required option for establishing a connection.
func WithURL(url string) ClientOption {
	return func(c *ClientConfig) error {
		c.URL = url
		return nil
	}
}

// WithAuthHandler sets the auth handler for authentication.
func WithAuthHandler(handler auth.AuthHandler) ClientOption {
	return func(c *ClientConfig) error {
		c.AuthHandler = handler
		return nil
	}
}

// WithReconnect enables or disables automatic reconnection.
func WithReconnect(enabled bool) ClientOption {
	return func(c *ClientConfig) error {
		c.ReconnectEnabled = enabled
		return nil
	}
}

// WithReconnectConfig sets the reconnection configuration.
func WithReconnectConfig(config *ReconnectConfig) ClientOption {
	return func(c *ClientConfig) error {
		c.ReconnectConfig = config
		return nil
	}
}

// WithLogger sets the logger for debugging output.
func WithLogger(logger Logger) ClientOption {
	return func(c *ClientConfig) error {
		c.Logger = logger
		return nil
	}
}

// WithHeader sets custom HTTP headers for the WebSocket handshake.
func WithHeader(header map[string][]string) ClientOption {
	return func(c *ClientConfig) error {
		c.Header = header
		return nil
	}
}

// WithTLSConfig sets the TLS configuration for secure connections.
func WithTLSConfig(tlsConfig *transport.TLSConfig) ClientOption {
	return func(c *ClientConfig) error {
		c.TLSConfig = tlsConfig
		return nil
	}
}

// WithEventBufferSize sets the event buffer size.
// Default is 100.
func WithEventBufferSize(size int) ClientOption {
	return func(c *ClientConfig) error {
		c.EventBufferSize = size
		return nil
	}
}

// OpenClawClient is the main client interface for the OpenClaw SDK.
// It provides methods for connecting, sending requests, and managing events.
type OpenClawClient interface {
	// Connect establishes a WebSocket connection to the server.
	Connect(ctx context.Context) error
	// Disconnect closes the connection gracefully.
	Disconnect() error
	// State returns the current connection state.
	State() ConnectionState
	// SendRequest sends a request and waits for a response.
	SendRequest(ctx context.Context, req *protocol.RequestFrame) (*protocol.ResponseFrame, error)
	// Events returns the event channel for receiving events.
	Events() <-chan Event
	// Subscribe adds an event handler for the specified event type.
	// Returns an unsubscribe function.
	Subscribe(eventType EventType, handler EventHandler) func()
	// Close shuts down the client and releases all resources.
	Close() error
}

// client is the concrete implementation of OpenClawClient.
// It coordinates multiple managers for event handling, request/response, connection, and reconnection.
type client struct {
	config   *ClientConfig
	managers struct {
		event      *managers.EventManager      // Event pub/sub management
		request    *managers.RequestManager    // Request/response correlation
		connection *managers.ConnectionManager // WebSocket connection lifecycle
		reconnect  *managers.ReconnectManager  // Automatic reconnection
	}
	ctx    context.Context    // Parent context for cancellation
	cancel context.CancelFunc // Cancel function for parent context
	mu     sync.Mutex         // Mutex for thread-safe operations
}

// NewClient creates a new OpenClaw client with the given options.
// It initializes all managers and returns an error if configuration fails.
// The client must be closed after use to release resources.
func NewClient(opts ...ClientOption) (OpenClawClient, error) {
	cfg := DefaultClientConfig()
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &client{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize managers
	c.managers.event = managers.NewEventManager(ctx, cfg.EventBufferSize)
	c.managers.request = managers.NewRequestManager(ctx)
	c.managers.connection = managers.NewConnectionManager(ctx, &managers.ClientConfig{
		URL:    cfg.URL,
		Header: cfg.Header,
	}, c.managers.event)

	if cfg.ReconnectEnabled {
		reconnectConfig := cfg.ReconnectConfig
		if reconnectConfig == nil {
			defaultCfg := DefaultReconnectConfig()
			reconnectConfig = &defaultCfg
		}
		c.managers.reconnect = managers.NewReconnectManager(reconnectConfig)
		// Set up reconnect callbacks
		c.managers.reconnect.SetOnReconnect(func() error {
			return c.managers.connection.Connect(ctx)
		})
		c.managers.reconnect.SetOnReconnectFailed(func(err error) {
			c.managers.event.Emit(Event{
				Type:      EventError,
				Err:       err,
				Timestamp: time.Now(),
			})
		})
	}

	c.managers.event.Start()

	return c, nil
}

// Connect establishes a WebSocket connection to the server.
// Thread-safe method that validates URL and initiates connection.
func (c *client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.config.URL == "" {
		return NewValidationError("URL is required", nil)
	}

	err := c.managers.connection.Connect(ctx)
	if err == nil && c.managers.reconnect != nil {
		c.managers.reconnect.Start()
	}
	return err
}

// Disconnect closes the WebSocket connection gracefully.
// Thread-safe method that stops reconnection before disconnecting.
func (c *client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.managers.reconnect != nil {
		c.managers.reconnect.Stop()
	}
	return c.managers.connection.Disconnect()
}

// State returns the current connection state.
// Thread-safe method that returns the state from the connection manager.
func (c *client) State() ConnectionState {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.managers.connection == nil {
		return StateDisconnected
	}
	return c.managers.connection.State()
}

// SendRequest sends a request and waits for a response.
// Thread-safe method that serializes the request and sends it via the transport.
func (c *client) SendRequest(ctx context.Context, req *protocol.RequestFrame) (*protocol.ResponseFrame, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.managers.connection == nil {
		return nil, NewConnectionError("not connected", nil)
	}

	sendFunc := func(r *protocol.RequestFrame) error {
		data, err := json.Marshal(r)
		if err != nil {
			return err
		}
		return c.managers.connection.Transport().Send(data)
	}
	return c.managers.request.SendRequest(ctx, req, sendFunc)
}

// Events returns the event channel for receiving all events.
func (c *client) Events() <-chan Event {
	return c.managers.event.Events()
}

// Subscribe adds an event handler for the specified event type.
// Returns an unsubscribe function that can be called to remove the handler.
func (c *client) Subscribe(eventType EventType, handler EventHandler) func() {
	return c.managers.event.Subscribe(eventType, handler)
}

// Close shuts down the client and releases all resources.
// Thread-safe method that closes all managers in the correct order.
func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cancel()

	var closeErr error
	if c.managers.event != nil {
		if err := c.managers.event.Close(); err != nil {
			closeErr = err
		}
	}
	if c.managers.request != nil {
		if err := c.managers.request.Close(); err != nil {
			closeErr = err
		}
	}
	if c.managers.connection != nil {
		if err := c.managers.connection.Close(); err != nil {
			closeErr = err
		}
	}
	if c.managers.reconnect != nil {
		c.managers.reconnect.Stop()
	}

	return closeErr
}
