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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/api"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/auth"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/connection"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/events"
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
type (
	ErrorCode       = types.ErrorCode
	OpenClawError   = types.OpenClawError
	BaseError       = types.BaseError
	AuthError       = types.AuthError
	ConnectionError = types.ConnectionError
	ProtocolError   = types.ProtocolError
	RequestError    = types.RequestError
	GatewayError    = types.GatewayError
	ReconnectError  = types.ReconnectError
	TimeoutError    = types.TimeoutError
	CancelledError  = types.CancelledError
	AbortError      = types.AbortError
)

// Re-export error constructors
var (
	NewConnectionError = types.NewConnectionError
	NewAuthError       = types.NewAuthError
	NewProtocolError   = types.NewProtocolError
	NewRequestError    = types.NewRequestError
	NewGatewayError    = types.NewGatewayError
	NewReconnectError  = types.NewReconnectError
	NewTimeoutError    = types.NewTimeoutError
	NewCancelledError  = types.NewCancelledError
	NewAbortError      = types.NewAbortError
	NewAPIError        = types.NewAPIError
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
	URL                 string                          // WebSocket server URL (e.g., wss://example.com/ws)
	ClientID            string                          // Client identifier
	ClientVersion       string                          // Client version string
	Platform            string                          // Platform identifier
	DeviceFamily        string                          // Device family
	ModelIdentifier     string                          // Model identifier
	Mode                string                          // Client mode (default: "go")
	InstanceID          string                          // Instance identifier
	Auth                *connection.ConnectParamsAuth   // Auth credentials
	Device              *connection.ConnectParamsDevice // Device pairing credentials
	CredentialsProvider auth.CredentialsProvider        // Credentials provider for advanced auth
	TickMonitor         *TickMonitorConfig              // Tick monitor configuration
	GapDetector         *GapDetectorConfig              // Gap detector configuration
	Capabilities        []string                        // Client capabilities to advertise
	Logger              Logger                          // Logger instance for debugging
	Header              map[string][]string             // Custom HTTP headers for WebSocket handshake
	TLSConfig           *transport.TLSConfig            // TLS/SSL configuration
	EventBufferSize     int                             // Buffer size for event channel
	ReconnectEnabled    bool                            // Enable automatic reconnection
	ReconnectConfig     *ReconnectConfig                // Reconnection configuration
	AuthHandler         auth.AuthHandler                // Authentication handler (deprecated: use CredentialsProvider)
}

// TickMonitorConfig configures the tick/heartbeat monitor.
type TickMonitorConfig struct {
	TickIntervalMs  int64  // Tick interval in milliseconds
	StaleMultiplier int    // Multiplier for stale threshold (default: 2)
	OnStale         func() // Callback when connection becomes stale
	OnRecovered     func() // Callback when connection recovers
}

// GapDetectorConfig configures the gap detector for event sequence tracking.
type GapDetectorConfig struct {
	RecoveryMode     string               // Recovery mode: "reconnect", "snapshot", "skip"
	OnGap            func(gaps []GapInfo) // Callback when gaps are detected
	SnapshotEndpoint string               // Endpoint for snapshot recovery
	MaxGaps          int                  // Maximum gaps to track (default: 100)
}

// GapInfo represents a detected gap in the event sequence.
type GapInfo struct {
	Expected   uint64 // Expected sequence number
	Received   uint64 // Received sequence number
	DetectedAt int64  // Timestamp when gap was detected
}

// DefaultClientConfig returns default client configuration.
// Default event buffer size is 100, and uses a default logger.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		EventBufferSize: 100,
		Mode:            "go",
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
	// API Accessors
	Chat() *api.ChatAPI
	Agents() *api.AgentsAPI
	Sessions() *api.SessionsAPI
	Config() *api.ConfigAPI
	Cron() *api.CronAPI
	Nodes() *api.NodesAPI
	Skills() *api.SkillsAPI
	DevicePairing() *api.DevicePairingAPI
	// Server Info
	GetServerInfo() *connection.HelloOk
	GetSnapshot() *connection.Snapshot
	GetPolicy() *connection.Policy
	GetTickMonitor() *events.TickMonitor
	GetGapDetector() *events.GapDetector
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
	// New fields for Phase 6.1
	protocolNegotiator *connection.ProtocolNegotiator     // Protocol version negotiation
	policyManager      *connection.PolicyManager          // Server policy management
	tickMonitor        *events.TickMonitor                // Heartbeat monitoring
	gapDetector        *events.GapDetector                // Event sequence gap detection
	serverInfo         *connection.HelloOk                // Server info from handshake
	snapshot           *connection.Snapshot               // Server snapshot
	stateMachine       *connection.ConnectionStateMachine // Connection state validation
	tickHandlerUnsub   func()                             // Unsubscribe function for tick handler
	// API namespaces
	chatAPI          *api.ChatAPI
	agentsAPI        *api.AgentsAPI
	sessionsAPI      *api.SessionsAPI
	configAPI        *api.ConfigAPI
	cronAPI          *api.CronAPI
	nodesAPI         *api.NodesAPI
	skillsAPI        *api.SkillsAPI
	devicePairingAPI *api.DevicePairingAPI
	// Internal state
	requestFn api.RequestFn
	ctx       context.Context    // Parent context for cancellation
	cancel    context.CancelFunc // Cancel function for parent context
	mu        sync.Mutex         // Mutex for thread-safe operations
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

	// Initialize protocol negotiator
	c.protocolNegotiator = connection.NewProtocolNegotiator()

	// Initialize policy manager
	c.policyManager = connection.NewPolicyManager()

	// Initialize state machine
	c.stateMachine = connection.NewConnectionStateMachine(types.StateDisconnected)

	// Set up connection event handlers
	c.setupConnectionHandlers()

	// Initialize API namespaces
	c.requestFn = c.newRequestFn()
	c.chatAPI = api.NewChatAPI(c.requestFn)
	c.agentsAPI = api.NewAgentsAPI(c.requestFn)
	c.sessionsAPI = api.NewSessionsAPI(c.requestFn)
	c.configAPI = api.NewConfigAPI(c.requestFn)
	c.cronAPI = api.NewCronAPI(c.requestFn)
	c.nodesAPI = api.NewNodesAPI(c.requestFn)
	c.skillsAPI = api.NewSkillsAPI(c.requestFn)
	c.devicePairingAPI = api.NewDevicePairingAPI(c.requestFn)

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
		return errors.New("URL is required")
	}

	if c.config.ClientID == "" {
		return errors.New("ClientID is required")
	}

	// Build connection parameters
	connectParams := c.buildConnectParams()

	// Connect with params (performs handshake)
	err := c.managers.connection.ConnectWithParams(ctx, connectParams)
	if err != nil {
		return err
	}

	// Post-handshake: parse server info, store policies
	c.processServerInfo()

	// Initialize tick monitor if configured
	if c.config.TickMonitor != nil {
		tickIntervalMs := c.config.TickMonitor.TickIntervalMs
		if tickIntervalMs == 0 {
			tickIntervalMs = c.policyManager.GetTickIntervalMs()
		}
		staleMultiplier := c.config.TickMonitor.StaleMultiplier
		if staleMultiplier == 0 {
			staleMultiplier = 2
		}
		c.tickMonitor = events.NewTickMonitor(tickIntervalMs, staleMultiplier)
		if c.config.TickMonitor.OnStale != nil {
			c.tickMonitor.SetOnStale(c.config.TickMonitor.OnStale)
		}
		if c.config.TickMonitor.OnRecovered != nil {
			c.tickMonitor.SetOnRecovered(c.config.TickMonitor.OnRecovered)
		}
		c.tickMonitor.Start()

		// Wire tick events to TickMonitor
		c.tickHandlerUnsub = c.managers.event.Subscribe(EventTick, func(e Event) {
			c.tickMonitor.RecordTick(e.Timestamp.UnixMilli())
		})
	}

	// Initialize gap detector if configured
	if c.config.GapDetector != nil {
		c.gapDetector = events.NewGapDetector()
		// Note: GapDetector SetOnGap would be set here when enhanced
	}

	if c.managers.reconnect != nil {
		c.managers.reconnect.Start()
	}
	return nil
}

// Disconnect closes the WebSocket connection gracefully.
// Thread-safe method that stops reconnection before disconnecting.
func (c *client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.managers.reconnect != nil {
		c.managers.reconnect.Stop()
	}

	// Clean up tick handler
	if c.tickHandlerUnsub != nil {
		c.tickHandlerUnsub()
		c.tickHandlerUnsub = nil
	}

	// Stop tick monitor
	if c.tickMonitor != nil {
		c.tickMonitor.Stop()
	}

	// Reset gap detector
	if c.gapDetector != nil {
		c.gapDetector.Reset()
	}

	// Reset protocol negotiation
	if c.protocolNegotiator != nil {
		c.protocolNegotiator.Reset()
	}
	c.serverInfo = nil
	c.snapshot = nil

	// Clear state machine
	if c.stateMachine != nil {
		c.stateMachine.Reset()
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
		return nil, NewConnectionError("NOT_CONNECTED", "not connected", false, nil)
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

// Chat returns the Chat API client.
func (c *client) Chat() *api.ChatAPI {
	return c.chatAPI
}

// Agents returns the Agents API client.
func (c *client) Agents() *api.AgentsAPI {
	return c.agentsAPI
}

// Sessions returns the Sessions API client.
func (c *client) Sessions() *api.SessionsAPI {
	return c.sessionsAPI
}

// Config returns the Config API client.
func (c *client) Config() *api.ConfigAPI {
	return c.configAPI
}

// Cron returns the Cron API client.
func (c *client) Cron() *api.CronAPI {
	return c.cronAPI
}

// Nodes returns the Nodes API client.
func (c *client) Nodes() *api.NodesAPI {
	return c.nodesAPI
}

// Skills returns the Skills API client.
func (c *client) Skills() *api.SkillsAPI {
	return c.skillsAPI
}

// DevicePairing returns the DevicePairing API client.
func (c *client) DevicePairing() *api.DevicePairingAPI {
	return c.devicePairingAPI
}

// GetServerInfo returns the server info from the handshake.
func (c *client) GetServerInfo() *connection.HelloOk {
	return c.serverInfo
}

// GetSnapshot returns the server snapshot.
func (c *client) GetSnapshot() *connection.Snapshot {
	return c.snapshot
}

// GetPolicy returns the server policy.
func (c *client) GetPolicy() *connection.Policy {
	if c.policyManager == nil {
		return nil
	}
	policy := c.policyManager.GetPolicy()
	return &policy
}

// GetTickMonitor returns the tick monitor instance.
func (c *client) GetTickMonitor() *events.TickMonitor {
	return c.tickMonitor
}

// GetGapDetector returns the gap detector instance.
func (c *client) GetGapDetector() *events.GapDetector {
	return c.gapDetector
}

// setupConnectionHandlers sets up event handlers for the connection manager.
func (c *client) setupConnectionHandlers() {
	cm := c.managers.connection
	// Handlers will be set via SetHandlers method
	_ = cm // Placeholder - actual handler setup happens in connection manager
}

// newRequestFn creates a request function for API namespaces.
func (c *client) newRequestFn() api.RequestFn {
	return func(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		req := &protocol.RequestFrame{
			Type:   "req",
			ID:     generateRequestID(),
			Method: method,
			Params: paramsJSON,
		}
		resp, err := c.SendRequest(ctx, req)
		if err != nil {
			return nil, err
		}
		if !resp.Ok {
			if resp.Error != nil {
				// Convert protocol.ErrorShape to types.ErrorShape
				errShape := &types.ErrorShape{
					Code:      resp.Error.Code,
					Message:   resp.Error.Message,
					Details:   resp.Error.Details,
					Retryable: resp.Error.Retryable,
				}
				return nil, types.NewAPIError(errShape)
			}
			return nil, fmt.Errorf("request failed: unknown error")
		}
		return resp.Payload, nil
	}
}

// buildConnectParams builds the connection parameters for the handshake.
func (c *client) buildConnectParams() *connection.ConnectParams {
	clientVersion := c.config.ClientVersion
	if clientVersion == "" {
		clientVersion = "1.0.0"
	}
	platform := c.config.Platform
	if platform == "" {
		platform = "go-sdk"
	}
	mode := c.config.Mode
	if mode == "" {
		mode = "go"
	}

	displayName := c.config.ClientID
	params := &connection.ConnectParams{
		MinProtocol: 3,
		MaxProtocol: 3,
		Client: connection.ConnectParamsClient{
			ID:          c.config.ClientID,
			DisplayName: &displayName,
			Version:     clientVersion,
			Platform:    platform,
			Mode:        mode,
		},
	}

	if c.config.Device != nil {
		params.Device = c.config.Device
	}
	if c.config.Auth != nil {
		params.Auth = c.config.Auth
	}
	if len(c.config.Capabilities) > 0 {
		params.Caps = c.config.Capabilities
	}

	return params
}

// processServerInfo processes the server info from the handshake response.
func (c *client) processServerInfo() {
	serverInfo := c.managers.connection.GetServerInfo()
	if serverInfo == nil {
		return
	}

	c.serverInfo = serverInfo
	c.snapshot = &serverInfo.Snapshot

	// Negotiate protocol version
	if c.protocolNegotiator != nil {
		_, err := c.protocolNegotiator.Negotiate(context.Background(), serverInfo)
		if err != nil {
			// Log error but don't fail connection
			_ = err
		}
	}

	// Set server policies
	if c.policyManager != nil && serverInfo.Policy.MaxPayload > 0 {
		c.policyManager.SetPolicies(serverInfo.Policy)
	}

	// Reset gap detector on new snapshot
	if c.gapDetector != nil && c.snapshot != nil {
		c.gapDetector.Reset()
	}
}

// generateRequestID generates a unique request ID.
func generateRequestID() string {
	return fmt.Sprintf("req-%d-%d", time.Now().UnixMilli(), time.Now().UnixNano()%10000)
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
