// Package events provides event handling utilities for the OpenClaw SDK.
package events

import (
	"context"
	"sync"
	"time"
)

// TickMonitor monitors connection heartbeat
type TickMonitor struct {
	interval   time.Duration
	timeout    time.Duration
	ticker     *time.Ticker
	timer      *time.Timer
	tickCh     chan time.Time
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	onTick     func(time.Time)
	onTimeout  func()
	running    bool
	stopped    chan struct{}
}

// ValidationError represents validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NewTickMonitor creates a new tick monitor
// Returns error if interval or timeout is zero or negative
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

// SetOnTick sets the tick callback (thread-safe)
func (tm *TickMonitor) SetOnTick(f func(time.Time)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTick = f
}

// SetOnTimeout sets the timeout callback (thread-safe)
func (tm *TickMonitor) SetOnTimeout(f func()) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTimeout = f
}

// Start begins the tick monitoring
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

// Stop stops the tick monitoring (idempotent)
func (tm *TickMonitor) Stop() {
	tm.mu.Lock()
	if !tm.running {
		tm.mu.Unlock()
		return
	}
	tm.running = false
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

	// Close channel only after goroutine is done
	close(tm.stopped)
	close(tm.tickCh)
}

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
			tm.mu.RLock()
			onTimeout := tm.onTimeout
			tm.mu.RUnlock()
			if onTimeout != nil && !lastTick.IsZero() {
				onTimeout()
			}
			// Reset timer for next timeout
			if !lastTick.IsZero() {
				tm.timer.Reset(tm.timeout)
			}
		}
	}
}

// TickChannel returns the tick channel
func (tm *TickMonitor) TickChannel() <-chan time.Time {
	return tm.tickCh
}

// IsRunning returns whether the monitor is running
func (tm *TickMonitor) IsRunning() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.running
}
