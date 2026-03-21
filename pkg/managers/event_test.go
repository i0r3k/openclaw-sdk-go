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
	em := NewEventManager(ctx, 10)

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
	em := NewEventManager(ctx, 10)

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
	em := NewEventManager(ctx, 10)

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
	em := NewEventManager(ctx, 10)

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
	em := NewEventManager(ctx, 10)

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
	em := NewEventManager(ctx, 100)

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
	em := NewEventManager(ctx, 10)

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
	em := NewEventManager(ctx, 1)
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
	em := NewEventManager(ctx, 10)

	// Subscribe with nil handler - should not panic
	unsubscribe := em.Subscribe(types.EventConnect, nil)
	em.Start()
	em.Emit(types.Event{Type: types.EventConnect, Timestamp: time.Now()})
	time.Sleep(20 * time.Millisecond)

	// Unsubscribe should also not panic
	unsubscribe()

	_ = em.Close()
}
