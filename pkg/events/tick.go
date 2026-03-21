// Package events provides event handling utilities for the OpenClaw SDK.
//
// This package provides:
//   - TickMonitor: Connection heartbeat monitoring with timeout detection
//   - GapDetector: Message gap detection for ordered message streams
package events

import (
	"sync"
	"time"
)

const defaultStaleMultiplier = 2

// TickMonitor monitors connection heartbeat.
// It tracks tick events and detects when expected ticks are not received within the timeout.
type TickMonitor struct {
	tickIntervalMs  int64          // Tick interval in milliseconds
	staleMultiplier int            // Multiplier for stale threshold
	lastTickTime    int64          // Last tick timestamp (in milliseconds)
	staleDetected   bool           // Whether stale state has been detected
	staleStartTime  *int64         // When stale state started
	started         bool           // Whether monitoring is started
	mu              sync.RWMutex   // Mutex for thread-safety
	onStale         func()         // Callback when connection becomes stale
	onRecovered     func()         // Callback when connection recovers
	done            chan struct{}  // Channel to signal background goroutine stop
	wg              sync.WaitGroup // WaitGroup for background goroutine
}

// NewTickMonitor creates a new tick monitor with the specified interval and timeout.
// The timeout is derived from tickIntervalMs * staleMultiplier.
func NewTickMonitor(tickIntervalMs int64, staleMultiplier int) *TickMonitor {
	if staleMultiplier <= 0 {
		staleMultiplier = defaultStaleMultiplier
	}
	return &TickMonitor{
		tickIntervalMs:  tickIntervalMs,
		staleMultiplier: staleMultiplier,
	}
}

// SetOnStale sets the callback function to be called when connection becomes stale.
func (tm *TickMonitor) SetOnStale(f func()) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onStale = f
}

// SetOnRecovered sets the callback function to be called when connection recovers.
func (tm *TickMonitor) SetOnRecovered(f func()) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onRecovered = f
}

// Start begins the tick monitoring with a background goroutine
// that periodically checks for staleness.
func (tm *TickMonitor) Start() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.started {
		return
	}
	tm.started = true
	tm.done = make(chan struct{})

	checkInterval := time.Duration(tm.tickIntervalMs) * time.Millisecond
	if checkInterval <= 0 {
		checkInterval = time.Second
	}

	done := tm.done
	tm.wg.Add(1)
	go func() {
		defer tm.wg.Done()
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				tm.CheckStale()
			}
		}
	}()
}

// Stop stops the tick monitoring and waits for the background goroutine to exit.
func (tm *TickMonitor) Stop() {
	tm.mu.Lock()
	if !tm.started {
		tm.mu.Unlock()
		return
	}
	tm.started = false
	done := tm.done
	tm.done = nil
	tm.mu.Unlock()

	close(done)
	tm.wg.Wait()
}

// RecordTick records an incoming tick with the given timestamp.
// Timestamp should be in milliseconds.
func (tm *TickMonitor) RecordTick(ts int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	wasStale := tm.staleDetected
	tm.lastTickTime = ts
	tm.staleStartTime = nil

	if wasStale {
		// Only emit recovered if connection is genuinely healthy
		if !tm.isStaleLocked() && tm.onRecovered != nil {
			tm.onRecovered()
		}
		tm.staleDetected = false
	}
}

// isStaleLocked checks if the connection is stale. Caller must hold the lock.
func (tm *TickMonitor) isStaleLocked() bool {
	if !tm.started {
		return false
	}
	if tm.lastTickTime == 0 {
		return false
	}
	now := time.Now().UnixMilli()
	threshold := tm.tickIntervalMs * int64(tm.staleMultiplier)
	return now-tm.lastTickTime > threshold
}

// IsStale returns true if no tick received within threshold.
func (tm *TickMonitor) IsStale() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.isStaleLocked()
}

// CheckStale checks staleness and fires stale event if newly detected.
func (tm *TickMonitor) CheckStale() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	stale := tm.isStaleLocked()
	if stale && !tm.staleDetected {
		tm.staleDetected = true
		threshold := tm.tickIntervalMs * int64(tm.staleMultiplier)
		startTime := tm.lastTickTime + threshold
		tm.staleStartTime = &startTime
		if tm.onStale != nil {
			tm.onStale()
		}
	}
	return stale
}

// GetTimeSinceLastTick returns milliseconds since last tick.
func (tm *TickMonitor) GetTimeSinceLastTick() int64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if !tm.started || tm.lastTickTime == 0 {
		return 0
	}
	return time.Now().UnixMilli() - tm.lastTickTime
}

// GetStaleDuration returns milliseconds in stale state, 0 if not stale.
func (tm *TickMonitor) GetStaleDuration() int64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if !tm.isStaleLocked() {
		return 0
	}
	if tm.staleStartTime == nil {
		threshold := tm.tickIntervalMs * int64(tm.staleMultiplier)
		startTime := tm.lastTickTime + threshold
		tm.staleStartTime = &startTime
	}
	return time.Now().UnixMilli() - *tm.staleStartTime
}

// IsRunning returns whether the tick monitor is currently running.
func (tm *TickMonitor) IsRunning() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.started
}
