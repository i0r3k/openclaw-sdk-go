// Package managers provides high-level manager components for the OpenClaw SDK.
package managers

import (
	"context"
	"sync"
	"unsafe"

	"github.com/i0r3k/openclaw-sdk-go/pkg/types"
)

// EventManager manages event subscriptions and dispatching
type EventManager struct {
	events   chan types.Event
	handlers map[types.EventType]map[uintptr]types.EventHandler
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

// NewEventManager creates a new event manager
func NewEventManager(ctx context.Context, bufferSize int) *EventManager {
	ctx, cancel := context.WithCancel(ctx)
	return &EventManager{
		events:   make(chan types.Event, bufferSize),
		handlers: make(map[types.EventType]map[uintptr]types.EventHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Subscribe adds an event handler
func (em *EventManager) Subscribe(eventType types.EventType, handler types.EventHandler) func() {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.handlers[eventType] == nil {
		em.handlers[eventType] = make(map[uintptr]types.EventHandler)
	}

	// Use pointer address as key - handlers are functions which have addresses
	key := uintptr(0)
	if handler != nil {
		// Get a unique identifier for the handler
		funcPtr := *(*uintptr)(unsafe.Pointer(&handler))
		key = funcPtr
	}

	em.handlers[eventType][key] = handler

	return func() { em.Unsubscribe(eventType, key) }
}

// Unsubscribe removes an event handler by key
func (em *EventManager) Unsubscribe(eventType types.EventType, key uintptr) {
	em.mu.Lock()
	defer em.mu.Unlock()
	if em.handlers[eventType] != nil {
		delete(em.handlers[eventType], key)
	}
}

// Events returns the event channel
func (em *EventManager) Events() <-chan types.Event {
	return em.events
}

// Emit emits an event
func (em *EventManager) Emit(event types.Event) {
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
						// In production, you might want to log this to a logger
						_ = r // Explicitly discard to avoid staticcheck warning
					}
				}()
				handler(event)
			}()
		}
	}
}

// Close gracefully shuts down the event manager
func (em *EventManager) Close() error {
	em.cancel()
	em.wg.Wait()
	close(em.events)
	return nil
}
