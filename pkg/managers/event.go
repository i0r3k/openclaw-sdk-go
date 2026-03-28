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
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// EventManager manages event subscriptions and dispatching.
// It provides a thread-safe pub/sub system for SDK events.
// Priority-based dispatch ensures high-priority events (errors, disconnects) are
// never dropped when the buffer is full (OBS-03).
type EventManager struct {
	// Priority input channels (OBS-03)
	priorityHigh   chan types.Event
	priorityMedium chan types.Event
	priorityLow    chan types.Event
	// Output channel (from dispatch loop)
	events chan types.Event

	handlers    map[types.EventType]map[uint64]types.EventHandler // Map of event type to handlers
	ctx         context.Context                                   // Context for lifecycle management
	cancel      context.CancelFunc                                // Cancel function for context
	mu          sync.RWMutex                                      // Mutex for thread-safe handler access
	wg          sync.WaitGroup                                    // WaitGroup for goroutines
	closed      bool                                              // Flag indicating if manager is closed
	closedMu    sync.Mutex                                        // Mutex for close flag
	logger      types.Logger                                      // Logger for error reporting
	nextID      uint64                                            // Next handler ID for unique keys
	emitTimeout time.Duration                                     // Timeout for Emit operations
	emitTimer   *time.Timer                                       // Reusable timer for Emit backpressure
}

// NewEventManager creates a new event manager with the specified buffer size.
// The emitTimeout controls how long Emit will wait when the channel is full before dropping the event.
// Buffer is partitioned: HIGH=25%, MEDIUM=25%, LOW=50% (OBS-03).
func NewEventManager(ctx context.Context, bufferSize int, emitTimeout time.Duration) *EventManager {
	ctx, cancel := context.WithCancel(ctx)
	logger, _ := types.FromContext(ctx)
	if logger == nil {
		logger = &types.NopLogger{}
	}

	// Buffer partition: HIGH=25%, MEDIUM=25%, LOW=50% (OBS-03)
	// Use at least 1 capacity for each priority channel
	highSize := bufferSize / 4
	if highSize < 1 {
		highSize = 1
	}
	medSize := bufferSize / 4
	if medSize < 1 {
		medSize = 1
	}
	lowSize := bufferSize / 2
	if lowSize < 1 {
		lowSize = 1
	}

	return &EventManager{
		priorityHigh:   make(chan types.Event, highSize),
		priorityMedium: make(chan types.Event, medSize),
		priorityLow:    make(chan types.Event, lowSize),
		events:         make(chan types.Event, bufferSize),
		handlers:       make(map[types.EventType]map[uint64]types.EventHandler),
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
		emitTimeout:    emitTimeout,
		emitTimer:      time.NewTimer(emitTimeout),
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

// PriorityHigh returns the HIGH priority channel (for testing).
func (em *EventManager) PriorityHigh() chan types.Event {
	return em.priorityHigh
}

// PriorityMedium returns the MEDIUM priority channel (for testing).
func (em *EventManager) PriorityMedium() chan types.Event {
	return em.priorityMedium
}

// PriorityLow returns the LOW priority channel (for testing).
func (em *EventManager) PriorityLow() chan types.Event {
	return em.priorityLow
}

// Emit emits an event to the event channel.
// It uses priority-based routing: HIGH priority events are never dropped when
// MEDIUM and LOW buffers are full (OBS-03).
func (em *EventManager) Emit(event types.Event) {
	// Auto-assign priority based on event type if not explicitly set (OBS-03, D-11 through D-13)
	if event.Priority == 0 {
		switch event.Type {
		case types.EventError, types.EventDisconnect, types.EventStateChange, types.EventGap:
			event.Priority = types.EventPriorityHigh
		case types.EventTick, types.EventResponse, types.EventConnect:
			event.Priority = types.EventPriorityMedium
		case types.EventMessage, types.EventRequest:
			event.Priority = types.EventPriorityLow
		default:
			event.Priority = types.EventPriorityMedium
		}
	}

	// Select target priority channel
	var priorityCh chan types.Event
	switch event.Priority {
	case types.EventPriorityHigh:
		priorityCh = em.priorityHigh
	case types.EventPriorityMedium:
		priorityCh = em.priorityMedium
	default:
		priorityCh = em.priorityLow
	}

	// Try non-blocking send to priority channel
	select {
	case priorityCh <- event:
		return
	default:
	}

	// Priority channel full -- try to drain lower priorities to make room
	if event.Priority > types.EventPriorityLow {
		em.drainLowerPriority(event.Priority)
		select {
		case priorityCh <- event:
			return
		default:
		}
	}

	// Still full -- log and drop
	em.logger.Warn("event dropped", "type", event.Type, "priority", event.Priority)
}

// drainLowerPriority drains one event from lower priority channels to make room (OBS-03).
// Drops LOW first, then MEDIUM. Never drains HIGH.
func (em *EventManager) drainLowerPriority(priority types.EventPriority) {
	switch priority {
	case types.EventPriorityHigh:
		// Drain one from MEDIUM if available
		select {
		case <-em.priorityMedium:
		default:
		}
	case types.EventPriorityMedium:
		// Drain one from LOW if available
		select {
		case <-em.priorityLow:
		default:
		}
	}
}

// Start begins the event dispatch loop in background goroutines.
// The dispatcher implements priority-based event dispatch: HIGH events are always
// processed first when available. A separate goroutine reads from the output
// channel and dispatches to handlers (OBS-03).
func (em *EventManager) Start() {
	em.wg.Add(1)
	go em.dispatcher() // Priority-based routing to em.events
	em.wg.Add(1)
	go em.dispatchLoop() // Reads from em.events, calls handlers
}

// dispatchLoop reads events from the output channel and dispatches to handlers.
func (em *EventManager) dispatchLoop() {
	defer em.wg.Done()
	for {
		select {
		case <-em.ctx.Done():
			return
		case event := <-em.events:
			em.dispatch(event)
		}
	}
}

// dispatcher implements priority-based event dispatch (OBS-03).
// HIGH events are always processed first when available.
// MEDIUM events are processed when HIGH is empty.
// LOW events are processed when HIGH and MEDIUM are both empty.
func (em *EventManager) dispatcher() {
	defer em.wg.Done()
	for {
		select {
		case <-em.ctx.Done():
			return
		case e := <-em.priorityHigh:
			// Always prefer HIGH -- blocking select on output
			select {
			case em.events <- e:
			case <-em.ctx.Done():
				return
			}
		case e := <-em.priorityMedium:
			// MEDIUM -- prefer HIGH if it arrives while we're sending
			select {
			case em.events <- e:
			case <-em.priorityHigh:
				// Got HIGH instead -- send MEDIUM later, process HIGH now
				select {
				case em.events <- e: // send MEDIUM
				case <-em.ctx.Done():
					return
				}
				em.events <- e // process HIGH
				continue
			case <-em.ctx.Done():
				return
			}
		case e := <-em.priorityLow:
			// LOW -- wait for HIGH or MEDIUM if they arrive while we're sending
			select {
			case em.events <- e:
			case <-em.priorityHigh:
				// Got HIGH -- send LOW later, process HIGH now
				select {
				case em.events <- e: // send LOW
				case <-em.ctx.Done():
					return
				}
				em.events <- e // process HIGH
				continue
			case <-em.priorityMedium:
				// Got MEDIUM -- send LOW later, process MEDIUM now
				select {
				case em.events <- e: // send LOW
				case <-em.ctx.Done():
					return
				}
				em.events <- e // process MEDIUM
				continue
			case <-em.ctx.Done():
				return
			}
		}
	}
}

// dispatch sends event to all registered handlers for the event type.
// Handlers are called in goroutines to prevent blocking.
// Panics in handlers are recovered and logged to continue processing other handlers.
func (em *EventManager) dispatch(event types.Event) {
	em.mu.RLock()
	// Copy handlers to avoid race: we must not hold lock while iterating handlers
	var handlers []types.EventHandler
	if handlerMap := em.handlers[event.Type]; handlerMap != nil {
		handlers = make([]types.EventHandler, 0, len(handlerMap))
		for _, handler := range handlerMap {
			handlers = append(handlers, handler)
		}
	}
	em.mu.RUnlock()

	for _, handler := range handlers {
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

	em.emitTimer.Stop()
	em.cancel()
	em.wg.Wait()

	// Close all channels (safe to close already-closed channel in Go)
	close(em.priorityHigh)
	close(em.priorityMedium)
	close(em.priorityLow)
	close(em.events)
	return nil
}
