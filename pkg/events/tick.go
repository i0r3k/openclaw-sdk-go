// Package events provides event handling utilities for the OpenClaw SDK.
//
// This package provides:
//   - TickMonitor: Connection heartbeat monitoring with timeout detection
//   - GapDetector: Message gap detection for ordered message streams
package events

import (
	"context"
	"sync"
	"time"
)

// TickMonitor monitors connection heartbeat.
// It tracks tick events and detects when expected ticks are not received within the timeout.
type TickMonitor struct {
	interval  time.Duration      // Interval between tick events
	timeout   time.Duration      // Timeout duration
	ticker    *time.Ticker       // Ticker for periodic ticks
	timer     *time.Timer        // Timer for timeout detection
	tickCh    chan time.Time     // Channel for tick events
	ctx       context.Context    // Context for lifecycle
	cancel    context.CancelFunc // Cancel function
	wg        sync.WaitGroup     // WaitGroup for goroutines
	mu        sync.RWMutex       // Mutex for thread-safety
	onTick    func(time.Time)    // Callback for tick events
	onTimeout func()             // Callback for timeout events
	running   bool               // Flag indicating if monitor is running
	stopped   chan struct{}      // Channel to signal stop
}

// ValidationError represents validation failure for TickMonitor configuration.
type ValidationError struct {
	Field   string // Field name that failed validation
	Message string // Validation error message
}

// Error returns the string representation of the validation error.
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NewTickMonitor creates a new tick monitor with the specified interval and timeout.
// Returns error if interval or timeout is zero or negative.
func NewTickMonitor(interval time.Duration, timeout time.Duration) (*TickMonitor, error) {
	if interval <= 0 {
		return nil, &ValidationError{Field: "interval", Message: "must be positive"}
	}
	if timeout <= 0 {
		return nil, &ValidationError{Field: "timeout", Message: "must be positive"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &TickMonitor{
		interval: interval,
		timeout:  timeout,
		ticker:   time.NewTicker(interval),
		timer:    time.NewTimer(timeout),
		tickCh:   make(chan time.Time, 1),
		ctx:      ctx,
		cancel:   cancel,
		stopped:  make(chan struct{}),
	}, nil
}

// SetOnTick sets the callback function to be called when a tick is received.
func (tm *TickMonitor) SetOnTick(f func(time.Time)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTick = f
}

// SetOnTimeout sets the callback function to be called when a timeout occurs.
func (tm *TickMonitor) SetOnTimeout(f func()) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTimeout = f
}

// Start begins the tick monitoring in a background goroutine.
func (tm *TickMonitor) Start() {
	tm.mu.Lock()
	if tm.running {
		tm.mu.Unlock()
		return
	}
	tm.running = true
	tm.mu.Unlock()

	tm.wg.Add(1)
	go tm.run()
}

// Stop stops the tick monitoring.
// It is idempotent - calling multiple times is safe.
func (tm *TickMonitor) Stop() {
	tm.mu.Lock()
	if !tm.running {
		tm.mu.Unlock()
		return
	}
	tm.running = false
	// Mark as stopped to prevent double channel close
	stopped := tm.stopped
	tickCh := tm.tickCh
	tm.stopped = nil
	tm.tickCh = nil
	tm.mu.Unlock()

	tm.cancel()

	// Stop timer and ticker
	if !tm.timer.Stop() {
		// Drain timer channel if it fired
		select {
		case <-tm.timer.C:
		default:
		}
	}
	tm.ticker.Stop()

	// Wait for goroutine to finish
	tm.wg.Wait()

	// Close channels only if they weren't already closed
	// (protected by nil check since Stop is idempotent)
	if stopped != nil {
		close(stopped)
	}
	if tickCh != nil {
		close(tickCh)
	}
}

// run is the main monitoring loop.
// It listens for tick events and timeout events.
func (tm *TickMonitor) run() {
	defer tm.wg.Done()

	var lastTick time.Time
	for {
		select {
		case <-tm.ctx.Done():
			return
		case tick := <-tm.ticker.C:
			lastTick = tick
			select {
			case tm.tickCh <- tick:
			default:
			}
			// Call callback with lock
			tm.mu.RLock()
			onTick := tm.onTick
			tm.mu.RUnlock()
			if onTick != nil {
				onTick(tick)
			}
		case <-tm.timer.C:
			// Timer fired - timeout occurred
			// Check if still running before resetting
			tm.mu.RLock()
			onTimeout := tm.onTimeout
			running := tm.running
			tm.mu.RUnlock()
			if onTimeout != nil && !lastTick.IsZero() {
				onTimeout()
			}
			// Reset timer for next timeout (only if still running)
			if running && !lastTick.IsZero() {
				tm.timer.Reset(tm.timeout)
			}
		}
	}
}

// TickChannel returns the channel for receiving tick events.
func (tm *TickMonitor) TickChannel() <-chan time.Time {
	return tm.tickCh
}

// IsRunning returns whether the tick monitor is currently running.
func (tm *TickMonitor) IsRunning() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.running
}
