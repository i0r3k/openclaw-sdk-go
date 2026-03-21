package managers

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

func TestEventManager_Subscribe(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 20*time.Millisecond)

	var mu sync.Mutex
	handlerCalled := false

	em.Subscribe(types.EventConnect, func(e types.Event) {
		mu.Lock()
		handlerCalled = true
		mu.Unlock()
	})

	em.Start()
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})

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

	_ = em.Close()
}

func TestEventManager_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 20*time.Millisecond)

	handler := func(e types.Event) {}
	unsubscribe := em.Subscribe(types.EventConnect, handler)
	unsubscribe()

	em.Start()
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})
	time.Sleep(20 * time.Millisecond)

	_ = em.Close()
}

func TestEventManager_MultipleHandlers(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 20*time.Millisecond)

	var wg sync.WaitGroup
	handler1Called := false
	handler2Called := false

	wg.Add(2)
	em.Subscribe(types.EventConnect, func(e types.Event) {
		defer wg.Done()
		handler1Called = true
	})
	em.Subscribe(types.EventConnect, func(e types.Event) {
		defer wg.Done()
		handler2Called = true
	})

	em.Start()
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		t.Error("timeout waiting for handlers")
	case <-done:
	}

	if !handler1Called {
		t.Error("handler 1 was not called")
	}
	if !handler2Called {
		t.Error("handler 2 was not called")
	}

	_ = em.Close()
}

func TestEventManager_DifferentEventTypes(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 20*time.Millisecond)

	var mu sync.Mutex
	connectCalled := false
	disconnectCalled := false

	em.Subscribe(types.EventConnect, func(e types.Event) {
		mu.Lock()
		connectCalled = true
		mu.Unlock()
	})
	em.Subscribe(types.EventDisconnect, func(e types.Event) {
		mu.Lock()
		disconnectCalled = true
		mu.Unlock()
	})

	em.Start()
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})
	em.Emit(types.Event{Type: types.EventDisconnect, Timestamp: time.Now()})

	timeout := time.After(100 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		mu.Lock()
		for !connectCalled || !disconnectCalled {
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
		}
		mu.Unlock()
		close(done)
	}()

	select {
	case <-timeout:
		t.Error("timeout waiting for handlers")
	case <-done:
	}

	_ = em.Close()
}

func TestEventManager_PanicRecovery(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 20*time.Millisecond)

	var goodHandlerCalled atomic.Bool

	// Subscribe a handler that panics
	em.Subscribe(types.EventConnect, func(e types.Event) {
		panic("intentional panic")
	})
	// Subscribe a handler that should still be called
	em.Subscribe(types.EventConnect, func(e types.Event) {
		goodHandlerCalled.Store(true)
	})

	em.Start()
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})

	timeout := time.After(100 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		for !goodHandlerCalled.Load() {
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-timeout:
		t.Error("timeout waiting for good handler - panic may have crashed dispatch")
	case <-done:
	}

	_ = em.Close()
}

func TestEventManager_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 100, 20*time.Millisecond)

	var wg sync.WaitGroup
	em.Start()

	// Concurrent subscribe/unsubscribe
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler := func(e types.Event) {}
			unsubscribe := em.Subscribe(types.EventConnect, handler)
			// Unsubscribe immediately after subscribe
			time.Sleep(time.Microsecond)
			unsubscribe()
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Error("timeout in concurrent subscribe/unsubscribe")
	case <-done:
	}

	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})
	time.Sleep(50 * time.Millisecond)

	_ = em.Close()
}

func TestEventManager_CloseIdempotent(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 20*time.Millisecond)

	em.Start()

	// Close multiple times - should be idempotent
	err1 := em.Close()
	err2 := em.Close()
	err3 := em.Close()

	if err1 != nil {
		t.Errorf("first Close error: %v", err1)
	}
	if err2 != nil {
		t.Errorf("second Close error: %v", err2)
	}
	if err3 != nil {
		t.Errorf("third Close error: %v", err3)
	}
}

func TestEventManager_EmitNonBlocking(t *testing.T) {
	ctx := context.Background()
	// Small buffer to test non-blocking emit
	em := NewEventManager(ctx, 1, 20*time.Millisecond)
	em.Start()

	// Fill the channel
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})

	// This should not block even though channel is full
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})

	_ = em.Close()
}

func TestEventManager_NilHandler(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 20*time.Millisecond)

	// Subscribe with nil handler - should not panic
	unsubscribe := em.Subscribe(types.EventConnect, nil)
	em.Start()
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})
	time.Sleep(20 * time.Millisecond)

	// Unsubscribe should also not panic
	unsubscribe()

	_ = em.Close()
}

func TestEventManager_Events(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 20*time.Millisecond)

	// Events() should return the event channel
	ch := em.Events()
	if ch == nil {
		t.Error("expected non-nil channel from Events()")
	}

	_ = em.Close()
}

func TestEventManager_EventsAfterClose(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 20*time.Millisecond)
	em.Start()
	_ = em.Close()

	// Events() should still return a valid channel after close (may be closed)
	ch := em.Events()
	if ch == nil {
		t.Error("expected non-nil channel from Events() after close")
	}
}

func TestEventManager_Emit_BackpressureTimeout(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 1, 50*time.Millisecond)
	// Do NOT start the event manager - events will accumulate in the channel
	// This allows us to test backpressure without dispatcher consuming events

	// Fill the channel
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})

	// Time how long the blocked emit takes
	start := time.Now()
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()}) // Should wait ~50ms then drop
	elapsed := time.Since(start)

	// Should have waited approximately the timeout duration before returning
	if elapsed < 40*time.Millisecond {
		t.Errorf("expected emit to wait at least 40ms, got %v", elapsed)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("expected emit to wait at most 100ms, got %v", elapsed)
	}

	_ = em.Close()
}

func TestEventManager_Emit_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	em := NewEventManager(ctx, 1, 1*time.Second) // Long timeout
	em.Start()

	// Fill the channel
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})

	// Cancel context
	cancel()

	// Emit should return immediately due to context cancellation
	start := time.Now()
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})
	elapsed := time.Since(start)

	// Should not have waited for timeout
	if elapsed > 20*time.Millisecond {
		t.Errorf("expected emit to return immediately on context cancellation, got %v", elapsed)
	}

	_ = em.Close()
}

func TestEventManager_Emit_NormalPath(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 10, 50*time.Millisecond)
	em.Start()
	defer em.Close()

	// Normal emit should not block
	start := time.Now()
	for i := 0; i < 10; i++ {
		em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})
	}
	elapsed := time.Since(start)

	// Should complete quickly without any timeout waits
	if elapsed > 20*time.Millisecond {
		t.Errorf("expected normal emits to complete quickly, got %v", elapsed)
	}
}

func TestEventManager_Emit_ConcurrentWithDispatch(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx, 5, 50*time.Millisecond)
	em.Start()
	defer em.Close()

	var counter atomic.Int64
	em.Subscribe(types.EventConnect, func(e types.Event) {
		time.Sleep(10 * time.Millisecond) // Simulate slow handler
		counter.Add(1)
	})

	// Emit events concurrently while dispatcher is processing
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond) // Allow dispatch to complete

	if counter.Load() == 0 {
		t.Error("expected some events to be dispatched")
	}
}
