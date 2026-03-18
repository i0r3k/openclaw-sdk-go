package events

import (
	"sync"
	"testing"
	"time"
)

func TestNewTickMonitor_InvalidInterval(t *testing.T) {
	_, err := NewTickMonitor(0, time.Second)
	if err == nil {
		t.Error("expected error for zero interval")
	}

	_, err = NewTickMonitor(-1, time.Second)
	if err == nil {
		t.Error("expected error for negative interval")
	}
}

func TestNewTickMonitor_InvalidTimeout(t *testing.T) {
	_, err := NewTickMonitor(time.Second, 0)
	if err == nil {
		t.Error("expected error for zero timeout")
	}

	_, err = NewTickMonitor(time.Second, -1)
	if err == nil {
		t.Error("expected error for negative timeout")
	}
}

func TestTickMonitor_StartStop(t *testing.T) {
	monitor, err := NewTickMonitor(50*time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	monitor.Start()
	time.Sleep(20 * time.Millisecond)

	if !monitor.IsRunning() {
		t.Error("expected monitor to be running")
	}

	monitor.Stop()

	if monitor.IsRunning() {
		t.Error("expected monitor to be stopped")
	}
}

func TestTickMonitor_Stop_Idempotent(t *testing.T) {
	monitor, err := NewTickMonitor(50*time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	monitor.Start()
	time.Sleep(10 * time.Millisecond)

	// Should not panic
	monitor.Stop()
	monitor.Stop()
}

func TestTickMonitor_TickCallback(t *testing.T) {
	monitor, err := NewTickMonitor(20*time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var tickCount int
	var mu sync.Mutex

	monitor.SetOnTick(func(t time.Time) {
		mu.Lock()
		tickCount++
		mu.Unlock()
	})

	monitor.Start()
	time.Sleep(60 * time.Millisecond)
	monitor.Stop()

	mu.Lock()
	if tickCount == 0 {
		t.Error("expected at least one tick")
	}
	mu.Unlock()
}

func TestTickMonitor_TimeoutCallback(t *testing.T) {
	// Use shorter interval so we get a tick first, then timeout triggers
	// after no tick is received within the timeout period
	monitor, err := NewTickMonitor(30*time.Millisecond, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	timeoutCalled := false
	monitor.SetOnTimeout(func() {
		timeoutCalled = true
	})

	monitor.Start()
	// Wait for at least one tick (30ms) + timeout (50ms) - should trigger after ~80ms
	time.Sleep(100 * time.Millisecond)
	monitor.Stop()

	if !timeoutCalled {
		t.Error("expected timeout callback to be called")
	}
}

func TestTickMonitor_TickChannel(t *testing.T) {
	monitor, err := NewTickMonitor(20*time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	monitor.Start()

	select {
	case <-monitor.TickChannel():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for tick")
	}

	monitor.Stop()
}

func TestTickMonitor_ConcurrentCallbacks(t *testing.T) {
	monitor, err := NewTickMonitor(10*time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Set callbacks from different goroutines
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			monitor.SetOnTick(func(t time.Time) {})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			monitor.SetOnTimeout(func() {})
		}
	}()

	monitor.Start()
	time.Sleep(30 * time.Millisecond)
	monitor.Stop()

	wg.Wait()
}
