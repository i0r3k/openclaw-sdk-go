package managers

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/i0r3k/openclaw-sdk-go/pkg/types"
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
