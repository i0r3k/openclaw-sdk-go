package events

import (
	"testing"
	"time"
)

func TestTickMonitor_StartStop(t *testing.T) {
	monitor := NewTickMonitor(50, 2)

	monitor.Start()
	if !monitor.IsRunning() {
		t.Error("expected monitor to be running")
	}

	monitor.Stop()
	if monitor.IsRunning() {
		t.Error("expected monitor to be stopped")
	}
}

func TestTickMonitor_RecordTick(t *testing.T) {
	monitor := NewTickMonitor(100, 2)
	monitor.Start()

	// Record a tick
	now := time.Now().UnixMilli()
	monitor.RecordTick(now)

	// Should not be stale immediately
	if monitor.IsStale() {
		t.Error("expected not stale immediately after tick")
	}

	monitor.Stop()
}

func TestTickMonitor_IsStale(t *testing.T) {
	// tickIntervalMs=10, staleMultiplier=2, so threshold=20ms
	monitor := NewTickMonitor(10, 2)
	monitor.Start()

	// Record a tick 100ms ago - definitely stale
	oldTime := time.Now().UnixMilli() - 100
	monitor.RecordTick(oldTime)

	// Should be stale
	if !monitor.IsStale() {
		t.Error("expected stale after threshold exceeded")
	}

	monitor.Stop()
}

func TestTickMonitor_NotStale(t *testing.T) {
	// tickIntervalMs=1000, staleMultiplier=2, so threshold=2000ms
	monitor := NewTickMonitor(1000, 2)
	monitor.Start()

	// Record a tick now
	now := time.Now().UnixMilli()
	monitor.RecordTick(now)

	// Should not be stale
	if monitor.IsStale() {
		t.Error("expected not stale immediately after tick")
	}

	monitor.Stop()
}

func TestTickMonitor_NotStaleWhenNotStarted(t *testing.T) {
	monitor := NewTickMonitor(10, 2)

	// Record a tick way in the past
	oldTime := time.Now().UnixMilli() - 10000
	monitor.RecordTick(oldTime)

	// Should NOT be stale because not started
	if monitor.IsStale() {
		t.Error("expected not stale when not started")
	}
}

func TestTickMonitor_CheckStale(t *testing.T) {
	monitor := NewTickMonitor(10, 2)
	monitor.Start()

	// Record a tick way in the past
	oldTime := time.Now().UnixMilli() - 100
	monitor.RecordTick(oldTime)

	// Check stale - should detect
	if !monitor.CheckStale() {
		t.Error("expected CheckStale to return true")
	}

	monitor.Stop()
}

func TestTickMonitor_OnStaleCallback(t *testing.T) {
	monitor := NewTickMonitor(10, 2)

	staleCalled := false
	monitor.SetOnStale(func() {
		staleCalled = true
	})

	monitor.Start()

	// Record old tick to trigger stale
	oldTime := time.Now().UnixMilli() - 100
	monitor.RecordTick(oldTime)

	// Check stale
	monitor.CheckStale()

	if !staleCalled {
		t.Error("expected stale callback to be called")
	}

	monitor.Stop()
}

func TestTickMonitor_OnRecoveredCallback(t *testing.T) {
	// tickIntervalMs=10, staleMultiplier=2, threshold=20ms
	// Use a short interval so test runs fast
	monitor := NewTickMonitor(10, 2)

	recoveredCalled := false
	monitor.SetOnRecovered(func() {
		recoveredCalled = true
	})

	monitor.Start()

	// Record old tick to make it stale (100ms ago > 20ms threshold)
	oldTime := time.Now().UnixMilli() - 100
	monitor.RecordTick(oldTime)
	monitor.CheckStale()

	// Now record a fresh tick - should recover (freshTime is now, well within threshold)
	freshTime := time.Now().UnixMilli()
	monitor.RecordTick(freshTime)

	if !recoveredCalled {
		t.Error("expected recovered callback to be called")
	}

	monitor.Stop()
}

func TestTickMonitor_GetTimeSinceLastTick(t *testing.T) {
	monitor := NewTickMonitor(1000, 2)
	monitor.Start()

	// Before any tick
	if monitor.GetTimeSinceLastTick() != 0 {
		t.Error("expected 0 before any tick")
	}

	// Record a tick
	now := time.Now().UnixMilli()
	monitor.RecordTick(now)

	// Should have some time since
	elapsed := monitor.GetTimeSinceLastTick()
	if elapsed < 0 || elapsed > 1000 {
		t.Errorf("unexpected time since last tick: %d", elapsed)
	}

	monitor.Stop()
}

func TestTickMonitor_GetStaleDuration(t *testing.T) {
	monitor := NewTickMonitor(10, 2)
	monitor.Start()

	// Before any tick
	if monitor.GetStaleDuration() != 0 {
		t.Error("expected 0 stale duration before any tick")
	}

	// Record old tick and check
	oldTime := time.Now().UnixMilli() - 100
	monitor.RecordTick(oldTime)
	monitor.CheckStale()

	// Should have some stale duration
	duration := monitor.GetStaleDuration()
	if duration < 50 {
		t.Errorf("expected stale duration > 50ms, got %d", duration)
	}

	monitor.Stop()
}

func TestTickMonitor_DefaultStaleMultiplier(t *testing.T) {
	monitor := NewTickMonitor(100, 0) // 0 should use default

	// Default is 2, so threshold should be 200ms
	now := time.Now().UnixMilli()
	monitor.RecordTick(now)

	// 150ms later - still not stale (threshold is 200ms)
	time.Sleep(150 * time.Millisecond)
	if monitor.IsStale() {
		t.Error("expected not stale within threshold")
	}

	monitor.Stop()
}
