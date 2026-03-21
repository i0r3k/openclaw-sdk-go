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
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/connection"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/transport"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// ClientConfig holds client configuration for ConnectionManager.
type ClientConfig struct {
	URL    string              // WebSocket server URL
	Header map[string][]string // Custom HTTP headers for WebSocket handshake
}

// NewConnectionManager creates a new connection manager with the given configuration.
func NewConnectionManager(ctx context.Context, config *ClientConfig, eventMgr *EventManager) *ConnectionManager {
	return &ConnectionManager{
		config:   config,
		state:    connection.NewConnectionStateMachine(types.StateDisconnected),
		eventMgr: eventMgr,
		ctx:      ctx,
	}
}

// ConnectionManager manages WebSocket connections.
// It handles connection lifecycle, state transitions, and transport management.
type ConnectionManager struct {
	config        *ClientConfig                      // Client configuration
	state         *connection.ConnectionStateMachine // Connection state machine
	transport     transport.Transport                // Underlying transport
	eventMgr      *EventManager                      // Event manager for emitting events
	ctx           context.Context                    // Context for lifecycle
	connectParams *connection.ConnectParams          // Connection parameters for handshake
	serverInfo    *connection.HelloOk                // Server info from handshake
	mu            sync.Mutex                         // Mutex for thread-safety
}

// Connect establishes a WebSocket connection to the configured URL.
// It transitions through states: Disconnected -> Connecting -> Connected.
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.transport != nil && cm.transport.IsConnected() {
		return types.NewConnectionError("CONNECTION_ALREADY_CONNECTED", "already connected", false, nil)
	}

	if err := cm.state.Transition(types.StateConnecting, nil); err != nil {
		return err
	}

	header := make(http.Header)
	if cm.config != nil && cm.config.Header != nil {
		for k, v := range cm.config.Header {
			header[k] = v
		}
	}

	t, err := transport.Dial(cm.config.URL, header, nil)
	if err != nil {
		if transitionErr := cm.state.Transition(types.StateFailed, err); transitionErr != nil {
			err = errors.Join(err, transitionErr)
		}
		return err
	}

	cm.transport = t
	t.Start()

	if err := cm.state.Transition(types.StateConnected, nil); err != nil {
		return err
	}

	if cm.eventMgr != nil {
		cm.eventMgr.Emit(types.Event{
			Type:      types.EventConnect,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// ConnectWithParams establishes a connection and performs the handshake.
func (cm *ConnectionManager) ConnectWithParams(ctx context.Context, params *connection.ConnectParams) error {
	// First connect
	if err := cm.Connect(ctx); err != nil {
		return err
	}

	// Store params for potential reconnect
	cm.mu.Lock()
	cm.connectParams = params
	cm.mu.Unlock()

	// Perform handshake - send connect request and wait for HelloOk
	// The HelloOk comes back as a response to our connect request
	helloOk, err := cm.performHandshake(ctx, params)
	if err != nil {
		_ = cm.Disconnect()
		return err
	}

	cm.mu.Lock()
	cm.serverInfo = helloOk
	cm.mu.Unlock()

	// Transition to ready state
	if err := cm.state.Transition(types.StateAuthenticated, nil); err != nil {
		return err
	}

	return nil
}

// performHandshake sends the connect request and waits for HelloOk response.
func (cm *ConnectionManager) performHandshake(ctx context.Context, params *connection.ConnectParams) (*connection.HelloOk, error) {
	// TODO: This is a simplified handshake - actual implementation
	// would send a request and wait for response via the transport
	// For now, return nil to indicate handshake is pending actual implementation
	return nil, nil
}

// GetServerInfo returns the server info from the handshake.
func (cm *ConnectionManager) GetServerInfo() *connection.HelloOk {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.serverInfo
}

// Disconnect closes the WebSocket connection.
// It transitions to the Disconnected state and emits a disconnect event.
func (cm *ConnectionManager) Disconnect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.transport == nil {
		return nil
	}

	err := cm.transport.Close()
	cm.transport = nil
	if transitionErr := cm.state.Transition(types.StateDisconnected, nil); transitionErr != nil {
		if err == nil {
			err = transitionErr
		}
	}

	if cm.eventMgr != nil {
		cm.eventMgr.Emit(types.Event{
			Type:      types.EventDisconnect,
			Timestamp: time.Now(),
		})
	}

	return err
}

// State returns the current connection state.
func (cm *ConnectionManager) State() types.ConnectionState {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.state == nil {
		return types.StateDisconnected
	}
	return cm.state.State()
}

// Transport returns the underlying transport for sending messages.
func (cm *ConnectionManager) Transport() transport.Transport {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.transport
}

// Close closes the connection manager.
// It delegates to Disconnect for cleanup.
func (cm *ConnectionManager) Close() error {
	return cm.Disconnect()
}
