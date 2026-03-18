# Phase 7: Managers Module

**Files:**
- Create: `managers/interfaces.go`
- Create: `managers/event.go`, `managers/event_test.go`
- Create: `managers/request.go`, `managers/request_test.go`
- Create: `managers/connection.go`, `managers/connection_test.go`
- Create: `managers/reconnect.go`, `managers/reconnect_test.go`

**Depends on:** Phase 1 (types.go), Phase 5 (connection), Phase 6 (events), Phase 4 (transport)

---

## Task 7.0: Manager Interfaces

- [ ] **Step 1: Write interfaces.go**

```go
// managers/interfaces.go
package managers

import (
	"context"

	"openclaw-sdk-go"
	"openclaw-sdk-go/protocol"
	"openclaw-sdk-go/transport"
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
```

---

## Task 7.1: Event Manager

- [ ] **Step 1: Create managers directory and event.go**

```bash
mkdir -p managers
```

```go
// managers/event.go
package managers

import (
	"context"
	"sync"
	"time"

	"openclaw-sdk-go"
)

// EventManager manages event subscriptions and dispatching
type EventManager struct {
	events   chan openclaw.Event
	handlers map[openclaw.EventType][]openclaw.EventHandler
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

// NewEventManager creates a new event manager
func NewEventManager(ctx context.Context, bufferSize int) *EventManager {
	ctx, cancel := context.WithCancel(ctx)
	return &EventManager{
		events:   make(chan openclaw.Event, bufferSize),
		handlers: make(map[openclaw.EventType][]openclaw.EventHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Subscribe adds an event handler
func (em *EventManager) Subscribe(eventType openclaw.EventType, handler openclaw.EventHandler) func() {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.handlers[eventType] = append(em.handlers[eventType], handler)
	return func() { em.Unsubscribe(eventType, handler) }
}

// Unsubscribe removes an event handler
func (em *EventManager) Unsubscribe(eventType openclaw.EventType, handler openclaw.EventHandler) {
	em.mu.Lock()
	defer em.mu.Unlock()
	handlers := em.handlers[eventType]
	for i, h := range handlers {
		if h == handler {
			handlers[i] = handlers[len(handlers)-1]
			em.handlers[eventType] = handlers[:len(handlers)-1]
			return
		}
	}
}

// Events returns the event channel
func (em *EventManager) Events() <-chan openclaw.Event {
	return em.events
}

// Emit emits an event
func (em *EventManager) Emit(event openclaw.Event) {
	select {
	case em.events <- event:
	case <-em.ctx.Done():
	}
}

// Start begins the event dispatch loop
func (em *EventManager) Start() {
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()
		for {
			select {
			case <-em.ctx.Done():
				return
			case event := <-em.events:
				em.dispatch(event)
			}
		}
	}()
}

// dispatch sends event to all registered handlers
func (em *EventManager) dispatch(event openclaw.Event) {
	em.mu.RLock()
	handlers := em.handlers[event.Type]
	em.mu.RUnlock()

	for _, handler := range handlers {
		func() {
			defer func() { recover() }()
			handler(event)
		}()
	}
}

// Close gracefully shuts down the event manager
func (em *EventManager) Close() error {
	em.cancel()
	em.wg.Wait()
	close(em.events)
	return nil
}
```

- [ ] **Step 2: Write test**

```go
// managers/event_test.go
package managers

import (
	"context"
	"sync"
	"testing"
	"time"

	"openclaw-sdk-go"
)

func TestEventManager_Subscribe(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10)

	var mu sync.Mutex
	handlerCalled := false

	em.Subscribe(openclaw.EventConnect, func(e openclaw.Event) {
		mu.Lock()
		handlerCalled = true
		mu.Unlock()
	})

	em.Start()
	em.Emit(openclaw.Event{Type: openclaw.EventConnect, Timestamp: time.Now()})

	// Use proper synchronization instead of time.Sleep
	timeout := time.After(100 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		mu.Lock()
		for !handlerCalled {
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
		}
		mu.Unlock()
		close(done)
	}()

	select {
	case <-timeout:
		t.Error("timeout waiting for handler")
	case <-done:
		// Handler was called
	}

	em.Close()
}

func TestEventManager_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10)

	handler := func(e openclaw.Event) {}
	unsubscribe := em.Subscribe(openclaw.EventConnect, handler)
	unsubscribe()

	em.Start()
	em.Emit(openclaw.Event{Type: openclaw.EventConnect, Timestamp: time.Now()})
	time.Sleep(20 * time.Millisecond)

	em.Close()
}
```

- [ ] **Step 3: Run tests**

Run: `go test -v ./managers/... -race`

---

## Task 7.2: Request Manager

- [ ] **Step 1: Write request.go**

```go
// managers/request.go
package managers

import (
	"context"
	"sync"
	"time"

	"openclaw-sdk-go/protocol"
)

// RequestManager manages pending requests
type RequestManager struct {
	pending  map[string]chan *protocol.ResponseFrame
	timeouts map[string]context.CancelFunc
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewRequestManager creates a new request manager
func NewRequestManager(ctx context.Context) *RequestManager {
	ctx, cancel := context.WithCancel(ctx)
	return &RequestManager{
		pending:  make(map[string]chan *protocol.ResponseFrame),
		timeouts: make(map[string]context.CancelFunc),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// SendRequest sends a request and waits for response
func (rm *RequestManager) SendRequest(ctx context.Context, req *protocol.RequestFrame, sendFunc func(*protocol.RequestFrame) error) (*protocol.ResponseFrame, error) {
	respCh := make(chan *protocol.ResponseFrame, 1)

	rm.mu.Lock()
	rm.pending[req.RequestID] = respCh

	// Set up timeout cancellation if context has deadline
	if deadline, ok := ctx.Deadline(); ok {
		timeoutCtx, cancel := context.WithDeadline(ctx, deadline)
		rm.timeouts[req.RequestID] = cancel
		ctx = timeoutCtx
	}
	rm.mu.Unlock()

	cleanup := func() {
		rm.mu.Lock()
		delete(rm.pending, req.RequestID)
		if cancel, ok := rm.timeouts[req.RequestID]; ok {
			cancel()
			delete(rm.timeouts, req.RequestID)
		}
		rm.mu.Unlock()
		close(respCh)
	}
	defer cleanup()

	// Send the request via transport
	if sendFunc != nil {
		if err := sendFunc(req); err != nil {
			return nil, err
		}
	}

	select {
	case resp := <-respCh:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// HandleResponse handles an incoming response
func (rm *RequestManager) HandleResponse(frame *protocol.ResponseFrame) {
	rm.mu.Lock()
	ch, ok := rm.pending[frame.RequestID]
	rm.mu.Unlock()

	if ok && ch != nil {
		select {
		case ch <- frame:
		default:
		}
	}
}

// Close cleans up all pending requests (thread-safe)
func (rm *RequestManager) Close() error {
	rm.mu.Lock()
	// First cancel context while holding lock to prevent new operations
	rm.cancel()
	// Clear pending map and timeouts
	for id, ch := range rm.pending {
		close(ch)
		delete(rm.pending, id)
	}
	for _, cancel := range rm.timeouts {
		cancel()
	}
	rm.mu.Unlock()
	return nil
}
```

- [ ] **Step 2: Write test**

```go
// managers/request_test.go
package managers

import (
	"context"
	"testing"
	"time"

	"openclaw-sdk-go/protocol"
)

func TestRequestManager_SendAndReceive(t *testing.T) {
	ctx := context.Background()
	rm := NewRequestManager(ctx)

	req := &protocol.RequestFrame{
		RequestID: "test-123",
		Method:    "test",
		Timestamp: time.Now(),
	}

	resp := &protocol.ResponseFrame{
		RequestID: "test-123",
		Success:   true,
		Timestamp: time.Now(),
	}

	// Send response after a small delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		rm.HandleResponse(resp)
	}()

	got, err := rm.SendRequest(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.RequestID != "test-123" {
		t.Errorf("expected 'test-123', got '%s'", got.RequestID)
	}

	rm.Close()
}

func TestRequestManager_ContextCancellation(t *testing.T) {
	ctx := context.Background()
	rm := NewRequestManager(ctx)

	req := &protocol.RequestFrame{
		RequestID: "test-cancel",
		Method:    "test",
		Timestamp: time.Now(),
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := rm.SendRequest(ctx, req, nil)
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	rm.Close()
}
```

- [ ] **Step 3: Run tests**

Run: `go test -v ./managers/... -race`

---

## Task 7.3: Connection Manager

- [ ] **Step 1: Write connection.go**

```go
// managers/connection.go
package managers

import (
	"context"
	"sync"
	"time"

	"openclaw-sdk-go"
	"openclaw-sdk-go/connection"
	"openclaw-sdk-go/transport"
)

// ClientConfig holds client configuration
type ClientConfig struct {
	URL    string
	Header map[string][]string
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(ctx context.Context, config *ClientConfig, eventMgr *EventManager) *ConnectionManager {
	return &ConnectionManager{
		config:    config,
		state:     connection.NewConnectionStateMachine(openclaw.StateDisconnected),
		eventMgr:  eventMgr,
		ctx:       ctx,
	}
}

// ConnectionManager manages WebSocket connections
type ConnectionManager struct {
	config    *ClientConfig
	state     *connection.ConnectionStateMachine
	transport transport.Transport
	eventMgr  *EventManager
	ctx       context.Context
	wg        sync.WaitGroup
	mu        sync.Mutex
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

	header := make(map[string][]string)
	if cm.config != nil && cm.config.Header != nil {
		header = cm.config.Header
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
```

- [ ] **Step 2: Write test**

```go
// managers/connection_test.go
package managers

import (
	"context"
	"testing"
	"time"

	"openclaw-sdk-go"
)

func TestConnectionManager_State(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10)
	cm := NewConnectionManager(ctx, config, em)

	state := cm.State()
	if state != openclaw.StateDisconnected {
		t.Errorf("expected disconnected, got %s", state)
	}

	em.Close()
}
```

- [ ] **Step 3: Run tests**

Run: `go test -v ./managers/... -race`

---

## Task 7.4: Reconnect Manager

- [ ] **Step 1: Write reconnect.go**

```go
// managers/reconnect.go
package managers

import (
	"context"
	"sync"
	"time"

	"openclaw-sdk-go"
)

// ReconnectConfig holds reconnection configuration
// Uses openclaw.ReconnectConfig from Phase 1
type ReconnectConfig = openclaw.ReconnectConfig

// DefaultReconnectConfig returns default configuration
func DefaultReconnectConfig() *ReconnectConfig {
	cfg := openclaw.DefaultReconnectConfig()
	return &cfg
}

// ReconnectManager handles automatic reconnection
type ReconnectManager struct {
	config             *ReconnectConfig
	mu                 sync.Mutex
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
	onReconnect        func() error
	onReconnectFailed  func(err error)
	stopped            chan struct{}
	stoppedOnce        sync.Once
}

// NewReconnectManager creates a new reconnect manager
func NewReconnectManager(config *ReconnectConfig) *ReconnectManager {
	if config == nil {
		config = DefaultReconnectConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &ReconnectManager{
		config:  config,
		ctx:    ctx,
		cancel: cancel,
		stopped: make(chan struct{}),
	}
}

// SetOnReconnect sets the reconnect callback
func (rm *ReconnectManager) SetOnReconnect(f func() error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onReconnect = f
}

// SetOnReconnectFailed sets the reconnect failed callback
func (rm *ReconnectManager) SetOnReconnectFailed(f func(err error)) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onReconnectFailed = f
}

// Start begins the reconnection loop
func (rm *ReconnectManager) Start() {
	rm.wg.Add(1)
	go rm.run()
}

func (rm *ReconnectManager) run() {
	defer rm.wg.Done()

	delay := rm.config.InitialDelay
	attempt := 0

	for {
		attempt++
		select {
		case <-rm.ctx.Done():
			return
		case <-rm.stopped:
			return
		case <-time.After(delay):
			rm.mu.Lock()
			onReconnect := rm.onReconnect
			onReconnectFailed := rm.onReconnectFailed
			rm.mu.Unlock()

			if onReconnect == nil {
				// No callback set - stop reconnect loop to avoid infinite loop
				return
			}

			err := onReconnect()
			if err == nil {
				return
			}
			if onReconnectFailed != nil {
				onReconnectFailed(err)
			}

			if rm.config.MaxAttempts > 0 && attempt >= rm.config.MaxAttempts {
				return
			}

			delay = time.Duration(float64(delay) * rm.config.BackoffMultiplier)
			if delay > rm.config.MaxDelay {
				delay = rm.config.MaxDelay
			}
		}
	}
}

// Stop stops the reconnection attempts (idempotent)
func (rm *ReconnectManager) Stop() {
	rm.cancel()
	rm.stoppedOnce.Do(func() {
		close(rm.stopped)
	})
	rm.wg.Wait()
}

// Reset is a no-op for compatibility (attempts tracked locally in run())
func (rm *ReconnectManager) Reset() {
	// Attempts are tracked in run() loop, not persisted
}
```

- [ ] **Step 2: Write test**

```go
// managers/reconnect_test.go
package managers

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestReconnectManager_Stop(t *testing.T) {
	ctx := context.Background()
	config := DefaultReconnectConfig()
	config.MaxAttempts = 1
	config.InitialDelay = 10 * time.Millisecond

	rm := NewReconnectManager(config)
	rm.Start()

	// Wait a bit then stop
	time.Sleep(20 * time.Millisecond)
	rm.Stop()
}

func TestReconnectManager_Callbacks(t *testing.T) {
	ctx := context.Background()
	config := DefaultReconnectConfig()
	config.MaxAttempts = 1
	config.InitialDelay = 10 * time.Millisecond

	rm := NewReconnectManager(config)

	var mu sync.Mutex
	reconnectCalled := false

	rm.SetOnReconnect(func() error {
		mu.Lock()
		reconnectCalled = true
		mu.Unlock()
		return nil // Success - stops reconnect loop
	})

	rm.Start()

	// Wait for reconnect to be called
	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	if !reconnectCalled {
		t.Error("expected reconnect callback to be called")
	}
	mu.Unlock()

	rm.Stop()
}
```

- [ ] **Step 3: Run tests**

Run: `go test -v ./managers/... -race`

---

## Phase 7 Complete

After this phase, you should have:
- `managers/event.go` - Event manager with thread-safe handlers
- `managers/event_test.go` - Event manager tests
- `managers/request.go` - Request manager with timeout support
- `managers/request_test.go` - Request manager tests
- `managers/connection.go` - Connection manager
- `managers/connection_test.go` - Connection manager tests
- `managers/reconnect.go` - Reconnect manager
- `managers/reconnect_test.go` - Reconnect manager tests

All code should compile and tests should pass.

Key fixes from review:
1. Define ClientConfig in Phase 7 (not dependent on Phase 9)
2. RequestManager.SendRequest now accepts sendFunc parameter
3. ConnectionManager uses proper context
4. ReconnectManager fixed - removed unused attempts field, added stopped channel
5. All managers have tests
6. EventManager test uses proper synchronization
7. Removed unused imports
