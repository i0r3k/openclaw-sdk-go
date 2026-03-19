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
	"sync"
	"sync/atomic"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// EventManager manages event subscriptions and dispatching.
// It provides a thread-safe pub/sub system for SDK events.
type EventManager struct {
	events   chan types.Event                                  // Channel for incoming events
	handlers map[types.EventType]map[uint64]types.EventHandler // Map of event type to handlers
	ctx      context.Context                                   // Context for lifecycle management
	cancel   context.CancelFunc                                // Cancel function for context
	mu       sync.RWMutex                                      // Mutex for thread-safe handler access
	wg       sync.WaitGroup                                    // WaitGroup for goroutines
	closed   bool                                              // Flag indicating if manager is closed
	closedMu sync.Mutex                                        // Mutex for close flag
	logger   types.Logger                                      // Logger for error reporting
	nextID   uint64                                            // Next handler ID for unique keys
}

// NewEventManager creates a new event manager with the specified buffer size.
func NewEventManager(ctx context.Context, bufferSize int) *EventManager {
	ctx, cancel := context.WithCancel(ctx)
	logger, _ := types.FromContext(ctx)
	if logger == nil {
		logger = &types.NopLogger{}
	}
	return &EventManager{
		events:   make(chan types.Event, bufferSize),
		handlers: make(map[types.EventType]map[uint64]types.EventHandler),
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
	}
}

// Subscribe adds an event handler for the specified event type.
// Returns an unsubscribe function that can be called to remove the handler.
func (em *EventManager) Subscribe(eventType types.EventType, handler types.EventHandler) func() {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.handlers[eventType] == nil {
		em.handlers[eventType] = make(map[uint64]types.EventHandler)
	}

	// Use atomic counter for unique handler IDs (atomic operation, no lock needed)
	key := atomic.AddUint64(&em.nextID, 1)

	em.handlers[eventType][key] = handler

	return func() { em.Unsubscribe(eventType, key) }
}

// Unsubscribe removes an event handler by event type and key.
func (em *EventManager) Unsubscribe(eventType types.EventType, key uint64) {
	em.mu.Lock()
	defer em.mu.Unlock()
	if em.handlers[eventType] != nil {
		delete(em.handlers[eventType], key)
	}
}

// Events returns the event channel for receiving all events.
func (em *EventManager) Events() <-chan types.Event {
	return em.events
}

// Emit emits an event to the event channel.
// Non-blocking: if the channel is full, it will return immediately.
func (em *EventManager) Emit(event types.Event) {
	select {
	case em.events <- event:
	case <-em.ctx.Done():
	}
}

// Start begins the event dispatch loop in a background goroutine.
// It listens for events and dispatches them to registered handlers.
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

// dispatch sends event to all registered handlers for the event type.
// Handlers are called in goroutines to prevent blocking.
// Panics in handlers are recovered and logged to continue processing other handlers.
func (em *EventManager) dispatch(event types.Event) {
	em.mu.RLock()
	handlerMap := em.handlers[event.Type]
	em.mu.RUnlock()

	for _, handler := range handlerMap {
		if handler != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Log panic but continue processing other handlers
						em.logger.Error("event handler panic recovered", "event", event.Type, "panic", r)
					}
				}()
				handler(event)
			}()
		}
	}
}

// Close gracefully shuts down the event manager.
// It is idempotent - calling multiple times is safe.
func (em *EventManager) Close() error {
	em.closedMu.Lock()
	if em.closed {
		em.closedMu.Unlock()
		return nil
	}
	em.closed = true
	em.closedMu.Unlock()

	em.cancel()
	em.wg.Wait()
	close(em.events)
	return nil
}
