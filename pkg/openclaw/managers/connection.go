// Package managers provides high-level manager components for the OpenClaw SDK.
package managers

import (
	"context"
	"net/http"
	"sync"
	"time"

	openclaw "github.com/i0r3k/openclaw-sdk-go/pkg/openclaw"
	"github.com/i0r3k/openclaw-sdk-go/pkg/openclaw/connection"
	"github.com/i0r3k/openclaw-sdk-go/pkg/openclaw/transport"
)

// ClientConfig holds client configuration
type ClientConfig struct {
	URL    string
	Header map[string][]string
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(ctx context.Context, config *ClientConfig, eventMgr *EventManager) *ConnectionManager {
	return &ConnectionManager{
		config:   config,
		state:    connection.NewConnectionStateMachine(openclaw.StateDisconnected),
		eventMgr: eventMgr,
		ctx:      ctx,
	}
}

// ConnectionManager manages WebSocket connections
type ConnectionManager struct {
	config   *ClientConfig
	state    *connection.ConnectionStateMachine
	transport transport.Transport
	eventMgr *EventManager
	ctx      context.Context
	wg       sync.WaitGroup
	mu       sync.Mutex
}

// Connect establishes a connection
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.transport != nil && cm.transport.IsConnected() {
		return openclaw.NewConnectionError("already connected", nil)
	}

	if err := cm.state.Transition(openclaw.StateConnecting, nil); err != nil {
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
		cm.state.Transition(openclaw.StateFailed, err)
		return err
	}

	cm.transport = t
	t.Start()

	if err := cm.state.Transition(openclaw.StateConnected, nil); err != nil {
		return err
	}

	if cm.eventMgr != nil {
		cm.eventMgr.Emit(openclaw.Event{
			Type:      openclaw.EventConnect,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// Disconnect closes the connection
func (cm *ConnectionManager) Disconnect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.transport == nil {
		return nil
	}

	err := cm.transport.Close()
	cm.transport = nil
	cm.state.Transition(openclaw.StateDisconnected, nil)

	if cm.eventMgr != nil {
		cm.eventMgr.Emit(openclaw.Event{
			Type:      openclaw.EventDisconnect,
			Timestamp: time.Now(),
		})
	}

	return err
}

// State returns the current connection state
func (cm *ConnectionManager) State() openclaw.ConnectionState {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.state == nil {
		return openclaw.StateDisconnected
	}
	return cm.state.State()
}

// Transport returns the underlying transport
func (cm *ConnectionManager) Transport() transport.Transport {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.transport
}

// Close closes the connection manager
func (cm *ConnectionManager) Close() error {
	return cm.Disconnect()
}
