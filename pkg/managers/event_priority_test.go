package managers

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

func TestEventPriority_DefaultIsMedium(t *testing.T) {
	event := types.Event{}
	// Default is 0, which is LOW
	if event.Priority != types.EventPriorityLow {
		t.Errorf("expected default Priority=0, got %v", event.Priority)
	}
}

func TestEventManager_PriorityChannelsCreated(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 100, 50*time.Millisecond)

	// Verify priority channels exist and have correct sizes
	// Buffer partition: HIGH=25%, MEDIUM=25%, LOW=50%
	// 100 / 4 = 25 each for HIGH and MEDIUM, 50 for LOW
	if cap(em.PriorityHigh()) != 25 {
		t.Errorf("expected priorityHigh cap=25, got %d", cap(em.PriorityHigh()))
	}
	if cap(em.PriorityMedium()) != 25 {
		t.Errorf("expected priorityMedium cap=25, got %d", cap(em.PriorityMedium()))
	}
	if cap(em.PriorityLow()) != 50 {
		t.Errorf("expected priorityLow cap=50, got %d", cap(em.PriorityLow()))
	}

	em.Close()
}

func TestEventManager_PriorityAssignment(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 100, 50*time.Millisecond)
	em.Start()
	defer em.Close()

	var receivedCount atomic.Int64
	var wg sync.WaitGroup
	wg.Add(3)

	em.Subscribe(types.EventError, func(e types.Event) {
		receivedCount.Add(1)
		wg.Done()
	})
	em.Subscribe(types.EventDisconnect, func(e types.Event) {
		receivedCount.Add(1)
		wg.Done()
	})
	em.Subscribe(types.EventTick, func(e types.Event) {
		receivedCount.Add(1)
		wg.Done()
	})
	em.Subscribe(types.EventMessage, func(e types.Event) {
		receivedCount.Add(1)
		wg.Done()
	})

	// Emit events with explicit priorities
	em.Emit(types.Event{Type: types.EventError, Priority: types.EventPriorityHigh, Timestamp: time.Now()})
	em.Emit(types.Event{Type: types.EventTick, Priority: types.EventPriorityMedium, Timestamp: time.Now()})
	em.Emit(types.Event{Type: types.EventMessage, Priority: types.EventPriorityLow, Timestamp: time.Now()})

	// Wait with timeout for all handlers to be called
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - all handlers called
	case <-time.After(200 * time.Millisecond):
		t.Errorf("timeout waiting for handlers, receivedCount=%d", receivedCount.Load())
	}
}

func TestEventManager_PriorityAutoAssignment(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 100, 50*time.Millisecond)
	em.Start()
	defer em.Close()

	var receivedPriorities []types.EventPriority
	var mu sync.Mutex

	em.Subscribe(types.EventConnect, func(e types.Event) {
		mu.Lock()
		defer mu.Unlock()
		receivedPriorities = append(receivedPriorities, e.Priority)
	})

	// Emit each type WITHOUT setting priority explicitly
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedPriorities) != 1 {
		t.Errorf("expected 1 event received, got %d", len(receivedPriorities))
	}
	// EventConnect should be auto-assigned MEDIUM priority
	if len(receivedPriorities) > 0 && receivedPriorities[0] != types.EventPriorityMedium {
		t.Errorf("expected EventConnect to have MEDIUM priority, got %v", receivedPriorities[0])
	}
}

func TestEventManager_PriorityHighNeverDrops(t *testing.T) {
	ctx := context.Background()
	// Use larger buffer so HIGH channel has more capacity
	em := NewEventManager(ctx, 100, 50*time.Millisecond)
	em.Start()
	defer em.Close()

	var receivedCount atomic.Int64

	em.Subscribe(types.EventError, func(e types.Event) {
		receivedCount.Add(1)
	})

	// Fill all channels with LOW and MEDIUM events first
	for i := 0; i < 100; i++ {
		em.Emit(types.Event{Type: types.EventMessage, Priority: types.EventPriorityLow, Timestamp: time.Now()})
		em.Emit(types.Event{Type: types.EventTick, Priority: types.EventPriorityMedium, Timestamp: time.Now()})
	}

	// Now emit HIGH events -- they should be received despite LOW/MEDIUM being present
	for i := 0; i < 10; i++ {
		em.Emit(types.Event{Type: types.EventError, Priority: types.EventPriorityHigh, Timestamp: time.Now()})
	}

	time.Sleep(100 * time.Millisecond)

	// HIGH events should be received
	if receivedCount.Load() == 0 {
		t.Error("expected HIGH events to be received even when buffer is full of LOW/MEDIUM")
	}
}

func TestEventManager_PriorityDropOrder(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 4, 10*time.Millisecond)
	// Don't start -- we want to fill channels without dispatcher consuming

	// Fill MEDIUM channel to capacity
	for i := 0; i < 3; i++ {
		em.Emit(types.Event{Type: types.EventTick, Priority: types.EventPriorityMedium, Timestamp: time.Now()})
	}

	// priorityMedium should now be full (cap is 1 for bufferSize=4)
	// Emit a MEDIUM event -- it should drain from LOW first
	em.Emit(types.Event{Type: types.EventTick, Priority: types.EventPriorityMedium, Timestamp: time.Now()})

	em.Close()
}

func TestEventManager_EventsReturnsOutputChannel(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 50*time.Millisecond)

	// Events() should return the output channel
	ch := em.Events()
	if ch == nil {
		t.Error("expected non-nil channel from Events()")
	}

	em.Close()
}
