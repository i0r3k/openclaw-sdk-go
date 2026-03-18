package openclaw

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/i0r3k/openclaw-sdk-go/pkg/auth"
	"github.com/i0r3k/openclaw-sdk-go/pkg/managers"
	"github.com/i0r3k/openclaw-sdk-go/pkg/protocol"
	"github.com/i0r3k/openclaw-sdk-go/pkg/transport"
)

// ClientConfig holds client configuration
type ClientConfig struct {
	URL              string
	AuthHandler      auth.AuthHandler
	ReconnectEnabled bool
	ReconnectConfig  *ReconnectConfig
	Logger           Logger
	Header           map[string][]string
	TLSConfig        *transport.TLSConfig
	EventBufferSize  int
}

// DefaultClientConfig returns default configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		EventBufferSize: 100,
		Logger:          NewDefaultLogger(),
	}
}

// ClientOption is a functional option
type ClientOption func(*ClientConfig) error

// WithURL sets the WebSocket URL
func WithURL(url string) ClientOption {
	return func(c *ClientConfig) error {
		c.URL = url
		return nil
	}
}

// WithAuthHandler sets the auth handler
func WithAuthHandler(handler auth.AuthHandler) ClientOption {
	return func(c *ClientConfig) error {
		c.AuthHandler = handler
		return nil
	}
}

// WithReconnect enables or disables reconnect
func WithReconnect(enabled bool) ClientOption {
	return func(c *ClientConfig) error {
		c.ReconnectEnabled = enabled
		return nil
	}
}

// WithReconnectConfig sets the reconnect configuration
func WithReconnectConfig(config *ReconnectConfig) ClientOption {
	return func(c *ClientConfig) error {
		c.ReconnectConfig = config
		return nil
	}
}

// WithLogger sets the logger
func WithLogger(logger Logger) ClientOption {
	return func(c *ClientConfig) error {
		c.Logger = logger
		return nil
	}
}

// WithHeader sets custom headers
func WithHeader(header map[string][]string) ClientOption {
	return func(c *ClientConfig) error {
		c.Header = header
		return nil
	}
}

// WithTLSConfig sets the TLS configuration
func WithTLSConfig(tlsConfig *transport.TLSConfig) ClientOption {
	return func(c *ClientConfig) error {
		c.TLSConfig = tlsConfig
		return nil
	}
}

// WithEventBufferSize sets the event buffer size
func WithEventBufferSize(size int) ClientOption {
	return func(c *ClientConfig) error {
		c.EventBufferSize = size
		return nil
	}
}

// OpenClawClient is the main client interface
type OpenClawClient interface {
	Connect(ctx context.Context) error
	Disconnect() error
	State() ConnectionState
	SendRequest(ctx context.Context, req *protocol.RequestFrame) (*protocol.ResponseFrame, error)
	Events() <-chan Event
	Subscribe(eventType EventType, handler EventHandler) func()
	Close() error
}

// client is the concrete implementation
type client struct {
	config   *ClientConfig
	state    ConnectionState
	managers struct {
		event      *managers.EventManager
		request    *managers.RequestManager
		connection *managers.ConnectionManager
		reconnect  *managers.ReconnectManager
	}
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

// NewClient creates a new OpenClaw client
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

// Connect establishes a connection (thread-safe)
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

// Disconnect closes the connection (thread-safe)
func (c *client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.managers.reconnect != nil {
		c.managers.reconnect.Stop()
	}
	return c.managers.connection.Disconnect()
}

// State returns the current connection state (thread-safe)
func (c *client) State() ConnectionState {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.managers.connection == nil {
		return StateDisconnected
	}
	return c.managers.connection.State()
}

// SendRequest sends a request (thread-safe)
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

// Events returns the event channel
func (c *client) Events() <-chan Event {
	return c.managers.event.Events()
}

// Subscribe subscribes to events
func (c *client) Subscribe(eventType EventType, handler EventHandler) func() {
	return c.managers.event.Subscribe(eventType, handler)
}

// Close closes the client (thread-safe)
func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cancel()

	if c.managers.event != nil {
		c.managers.event.Close()
	}
	if c.managers.request != nil {
		c.managers.request.Close()
	}
	if c.managers.connection != nil {
		c.managers.connection.Close()
	}
	if c.managers.reconnect != nil {
		c.managers.reconnect.Stop()
	}

	return nil
}
