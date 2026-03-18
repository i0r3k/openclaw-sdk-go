package managers

import (
	"context"
	"sync"
	"testing"
	"time"

	openclaw "github.com/i0r3k/openclaw-sdk-go/pkg/openclaw"
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
