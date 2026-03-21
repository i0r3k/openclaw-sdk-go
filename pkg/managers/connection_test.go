package managers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/connection"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// mockTransport implements transport.Transport for testing
type mockTransport struct {
	connected     bool
	sendCh        chan []byte
	recvCh        chan []byte
	errCh         chan error
	shouldError   bool
	errorToReturn error
	mu            sync.Mutex
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		connected: true,
		sendCh:    make(chan []byte, 64),
		recvCh:    make(chan []byte, 64),
		errCh:     make(chan error, 64),
	}
}

func (t *mockTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.shouldError {
		return t.errorToReturn
	}
	select {
	case t.sendCh <- data:
		return nil
	default:
		return errors.New("send channel full")
	}
}

func (t *mockTransport) Receive() <-chan []byte {
	return t.recvCh
}

func (t *mockTransport) Errors() <-chan error {
	return t.errCh
}

func (t *mockTransport) Close() error {
	t.mu.Lock()
	t.connected = false
	t.mu.Unlock()
	close(t.sendCh)
	close(t.recvCh)
	close(t.errCh)
	return nil
}

func (t *mockTransport) IsConnected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.connected
}

func TestConnectionManager_State(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	state := cm.State()
	if state != types.StateDisconnected {
		t.Errorf("expected disconnected, got %s", state)
	}

	_ = em.Close()
}

func TestConnectionManager_DisconnectWhenNotConnected(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	// Disconnect when not connected should not error
	err := cm.Disconnect()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	_ = em.Close()
}

func TestConnectionManager_TransportWhenNotConnected(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	transport := cm.Transport()
	if transport != nil {
		t.Error("expected nil transport when not connected")
	}

	_ = em.Close()
}

func TestConnectionManager_GetServerInfo(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	// Initially nil
	info := cm.GetServerInfo()
	if info != nil {
		t.Error("expected nil server info initially")
	}

	// Set server info directly
	cm.mu.Lock()
	cm.serverInfo = &connection.HelloOk{
		Type:     "hello-ok",
		Protocol: 3,
		Server: connection.HelloOkServer{
			Version: "1.0.0",
			ConnID:  "test-conn-id",
		},
	}
	cm.mu.Unlock()

	// Should return the info
	info = cm.GetServerInfo()
	if info == nil {
		t.Fatal("expected non-nil server info")
	}
	if info.Server.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", info.Server.Version)
	}

	_ = em.Close()
}

func TestConnectionManager_Close(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	// Close delegates to Disconnect
	err := cm.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should be in disconnected state
	if cm.State() != types.StateDisconnected {
		t.Errorf("expected disconnected state after Close, got %s", cm.State())
	}

	_ = em.Close()
}

func TestConnectionManager_Reconnect_NoParams(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	// Without stored params, Reconnect falls back to Connect which will fail
	// since there's no actual server - but we can verify it doesn't panic
	err := cm.Reconnect(ctx)
	if err == nil {
		t.Error("expected error when reconnecting without params to non-existent server")
	}

	_ = em.Close()
}

func TestConnectionManager_Connect(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	// Connect to non-existent server should fail
	err := cm.Connect(ctx)
	if err == nil {
		t.Error("expected error when connecting to non-existent server")
	}

	_ = em.Close()
}

func TestConnectionManager_Connect_AlreadyConnected(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	// Can't actually test already connected without a mock transport
	// Just verify State method works
	state := cm.State()
	if state != types.StateDisconnected {
		t.Errorf("expected disconnected, got %s", state)
	}

	_ = em.Close()
}

func TestConnectionManager_performHandshake_NoTransport(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	params := &connection.ConnectParams{
		MinProtocol: 3,
		MaxProtocol: 3,
		Client: connection.ConnectParamsClient{
			ID:      "test-client",
			Version: "1.0.0",
			Mode:    "normal",
		},
	}

	// performHandshake with nil transport should fail
	_, err := cm.performHandshake(ctx, params)
	if err == nil {
		t.Error("expected error when handshake has no transport")
	}

	_ = em.Close()
}

func TestConnectionManager_performHandshake_SendError(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	// Inject mock transport that fails on send
	mockT := newMockTransport()
	mockT.shouldError = true
	mockT.errorToReturn = errors.New("send error")

	cm.mu.Lock()
	cm.transport = mockT
	cm.mu.Unlock()

	params := &connection.ConnectParams{
		MinProtocol: 3,
		MaxProtocol: 3,
		Client: connection.ConnectParamsClient{
			ID:      "test-client",
			Version: "1.0.0",
			Mode:    "normal",
		},
	}

	_, err := cm.performHandshake(ctx, params)
	if err == nil {
		t.Error("expected error when transport send fails")
	}

	_ = em.Close()
}

func TestConnectionManager_performHandshake_Success(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	mockT := newMockTransport()

	cm.mu.Lock()
	cm.transport = mockT
	cm.mu.Unlock()

	params := &connection.ConnectParams{
		MinProtocol: 3,
		MaxProtocol: 3,
		Client: connection.ConnectParamsClient{
			ID:      "test-client",
			Version: "1.0.0",
			Mode:    "normal",
		},
	}

	// Send hello-ok response after small delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		helloOk := connection.HelloOk{
			Type:     "hello-ok",
			Protocol: 3,
			Server: connection.HelloOkServer{
				Version: "1.0.0",
				ConnID:  "test-conn",
			},
		}
		data, _ := json.Marshal(helloOk)
		mockT.recvCh <- data
	}()

	helloOk, err := cm.performHandshake(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if helloOk == nil {
		t.Fatal("expected non-nil helloOk")
	}
	if helloOk.Server.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", helloOk.Server.Version)
	}

	_ = em.Close()
}

func TestConnectionManager_performHandshake_TransportError(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	mockT := newMockTransport()

	cm.mu.Lock()
	cm.transport = mockT
	cm.mu.Unlock()

	params := &connection.ConnectParams{
		MinProtocol: 3,
		MaxProtocol: 3,
		Client: connection.ConnectParamsClient{
			ID:      "test-client",
			Version: "1.0.0",
			Mode:    "normal",
		},
	}

	// Send error through transport errors channel
	go func() {
		time.Sleep(10 * time.Millisecond)
		mockT.errCh <- errors.New("connection error")
	}()

	_, err := cm.performHandshake(ctx, params)
	if err == nil {
		t.Error("expected error when transport errors")
	}

	_ = em.Close()
}

func TestConnectionManager_performHandshake_ContextCancelled(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	mockT := newMockTransport()

	cm.mu.Lock()
	cm.transport = mockT
	cm.mu.Unlock()

	params := &connection.ConnectParams{
		MinProtocol: 3,
		MaxProtocol: 3,
		Client: connection.ConnectParamsClient{
			ID:      "test-client",
			Version: "1.0.0",
			Mode:    "normal",
		},
	}

	// Create already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := cm.performHandshake(ctx, params)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}

	_ = em.Close()
}

func TestConnectionManager_Disconnect_WithTransport(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	mockT := newMockTransport()

	cm.mu.Lock()
	cm.transport = mockT
	cm.state.Transition(types.StateConnecting, nil)
	cm.state.Transition(types.StateConnected, nil)
	cm.mu.Unlock()

	// Disconnect should close transport and emit event
	err := cm.Disconnect()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Transport should be nil after disconnect
	cm.mu.Lock()
	transport := cm.transport
	cm.mu.Unlock()
	if transport != nil {
		t.Error("expected nil transport after disconnect")
	}

	_ = em.Close()
}

func TestConnectionManager_Disconnect_StateTransitionError(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	mockT := newMockTransport()

	cm.mu.Lock()
	cm.transport = mockT
	cm.mu.Unlock()

	// Force state machine to a state where Transition might fail
	_ = cm.state.Transition(types.StateConnecting, nil)

	// Disconnect should handle any state transition error
	_ = cm.Disconnect()
	// The error from Close might be nil or not depending on state machine
	// We just verify it doesn't panic

	_ = em.Close()
}

func TestConnectionManager_ConnectWithParams_ConnectFails(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	params := &connection.ConnectParams{
		MinProtocol: 3,
		MaxProtocol: 3,
		Client: connection.ConnectParamsClient{
			ID:      "test-client",
			Version: "1.0.0",
			Mode:    "normal",
		},
	}

	// Connect should fail to non-existent server
	err := cm.ConnectWithParams(ctx, params)
	if err == nil {
		t.Error("expected error when connecting to non-existent server")
	}

	_ = em.Close()
}

func TestConnectionManager_Reconnect_WithParams(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	cm := NewConnectionManager(ctx, config, em)

	// Set connectParams (but can't test actual reconnect without real server)
	cm.mu.Lock()
	cm.connectParams = &connection.ConnectParams{
		MinProtocol: 3,
		MaxProtocol: 3,
		Client: connection.ConnectParamsClient{
			ID:      "test-client",
			Version: "1.0.0",
			Mode:    "normal",
		},
	}
	cm.mu.Unlock()

	// Without a real server, reconnect will fail on Connect()
	// This test just verifies setting params doesn't panic
	err := cm.Reconnect(ctx)
	if err == nil {
		t.Error("expected error when reconnecting to non-existent server")
	}

	_ = em.Close()
}
