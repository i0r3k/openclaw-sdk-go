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

// TestTickMonitor_NoResetAfterStop verifies that timer is not reset after stop is called.
// This tests the fix for the race condition where Timer.Reset() could be called
// after Stop() has been initiated.
func TestTickMonitor_NoResetAfterStop(t *testing.T) {
	// Use very short timeout to trigger timer quickly
	monitor, err := NewTickMonitor(50*time.Millisecond, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var timeoutCount int
	var mu sync.Mutex

	monitor.SetOnTimeout(func() {
		mu.Lock()
		timeoutCount++
		mu.Unlock()
	})

	monitor.Start()

	// Wait for at least one timeout to occur
	time.Sleep(50 * time.Millisecond)

	// Stop the monitor
	monitor.Stop()

	// Record the timeout count at stop time
	mu.Lock()
	countAtStop := timeoutCount
	mu.Unlock()

	// Wait a bit more to see if additional timeouts occur after stop
	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	finalCount := timeoutCount
	mu.Unlock()

	// After stop, no more timeouts should occur (timer should not be reset)
	if finalCount > countAtStop+1 {
		t.Errorf("timeout called too many times after stop: got %d, want <= %d",
			finalCount, countAtStop+1)
	}
}

// TestTickMonitor_StopDuringTimeout simulates the race condition where Stop()
// is called while the timer is firing. This should not cause a panic.
func TestTickMonitor_StopDuringTimeout(t *testing.T) {
	// Create monitor with very short timeout to maximize chance of race
	monitor, err := NewTickMonitor(1*time.Millisecond, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	monitor.SetOnTimeout(func() {
		// During timeout callback, stop the monitor
		// This simulates the race condition
		monitor.Stop()
	})

	monitor.Start()

	// Wait for the race to occur
	time.Sleep(20 * time.Millisecond)

	// If we get here without panic, the test passes
	// Call Stop again to ensure it's idempotent
	monitor.Stop()
}

// TestTickMonitor_MultipleTimeoutCycles verifies that the timer correctly
// resets after multiple timeout cycles while the monitor is running.
func TestTickMonitor_MultipleTimeoutCycles(t *testing.T) {
	// Use interval > timeout to ensure ticks happen, then timeout triggers after no tick
	// interval=30ms means tick every 30ms, timeout=50ms means if no tick for 50ms, timeout fires
	monitor, err := NewTickMonitor(30*time.Millisecond, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var timeoutCount int
	var mu sync.Mutex

	monitor.SetOnTimeout(func() {
		mu.Lock()
		timeoutCount++
		mu.Unlock()
	})

	monitor.Start()

	// Wait for multiple timeout cycles (tick every 30ms, timeout after 50ms no tick)
	// After first tick at ~30ms, timeout fires at ~80ms (50ms after tick)
	// Then reset, next tick at ~60ms, timeout fires at ~110ms, etc.
	time.Sleep(200 * time.Millisecond)
	monitor.Stop()

	mu.Lock()
	count := timeoutCount
	mu.Unlock()

	// Should have multiple timeout callbacks
	if count < 2 {
		t.Errorf("expected at least 2 timeout callbacks, got %d", count)
	}
}

// TestTickMonitor_ConcurrentStopAndTick simulates concurrent Stop() and tick events
// to verify no race condition causes issues.
func TestTickMonitor_ConcurrentStopAndTick(t *testing.T) {
	monitor, err := NewTickMonitor(5*time.Millisecond, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	stopCalled := make(chan struct{})

	// Goroutine: start and then stop multiple times
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			monitor.Start()
			time.Sleep(1 * time.Millisecond)
			monitor.Stop()
		}
		close(stopCalled)
	}()

	// Goroutine: continuously set callbacks while monitor is running
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopCalled:
				return
			default:
				monitor.SetOnTick(func(t time.Time) {})
				monitor.SetOnTimeout(func() {})
			}
		}
	}()

	wg.Wait()
}

// TestTickMonitor_TimerResetRaceCondition specifically tests the scenario where
// timer fires and Reset is called while Stop() is being called from another goroutine.
func TestTickMonitor_TimerResetRaceCondition(t *testing.T) {
	// This test is designed to catch the race condition by running many iterations
	for i := 0; i < 50; i++ {
		monitor, err := NewTickMonitor(1*time.Millisecond, 1*time.Millisecond)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		monitor.SetOnTimeout(func() {
			// Give a small chance for Stop to interleave
			time.Sleep(1 * time.Millisecond)
		})

		monitor.Start()

		// Let it run for a bit to trigger the race
		time.Sleep(5 * time.Millisecond)

		// Stop should not panic even if timer fires at the same time
		monitor.Stop()
	}
}
