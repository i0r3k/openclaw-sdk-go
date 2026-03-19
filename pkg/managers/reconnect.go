// Package managers provides high-level manager components for OpenClaw SDK.
//
// This package provides:
//   - EventManager: Pub/sub event management
//   - RequestManager: Pending request correlation
//   - ConnectionManager: WebSocket connection lifecycle
//   - ReconnectManager: Automatic reconnection with Fibonacci backoff
package managers

import (
	"context"
	"sync"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// ReconnectConfig holds reconnection configuration.
// Uses types.ReconnectConfig.
type ReconnectConfig = types.ReconnectConfig

// DefaultReconnectConfig returns default reconnection configuration.
// Default: max attempts = 0 (infinite), initial delay = 1s, max delay = 60s.
func DefaultReconnectConfig() *ReconnectConfig {
	cfg := types.DefaultReconnectConfig()
	return &cfg
}

// ReconnectManager handles automatic reconnection with Fibonacci backoff.
// It attempts to reconnect when the connection is lost, using exponential backoff.
type ReconnectManager struct {
	config            *ReconnectConfig   // Reconnection configuration
	mu                sync.Mutex         // Mutex for thread-safety
	ctx               context.Context    // Context for lifecycle
	cancel            context.CancelFunc // Cancel function
	wg                sync.WaitGroup     // WaitGroup for goroutines
	onReconnect       func() error       // Callback for reconnection attempts
	onReconnectFailed func(err error)    // Callback for reconnection failures
	stopped           chan struct{}      // Channel to signal stop
	stoppedOnce       sync.Once          // Ensure stopped channel is closed once
}

// NewReconnectManager creates a new reconnect manager with the given configuration.
// If config is nil, uses default configuration.
func NewReconnectManager(config *ReconnectConfig) *ReconnectManager {
	if config == nil {
		config = DefaultReconnectConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &ReconnectManager{
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
		stopped: make(chan struct{}),
	}
}

// SetOnReconnect sets the callback function to be called when attempting to reconnect.
func (rm *ReconnectManager) SetOnReconnect(f func() error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onReconnect = f
}

// SetOnReconnectFailed sets the callback function to be called when reconnection fails.
func (rm *ReconnectManager) SetOnReconnectFailed(f func(err error)) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onReconnectFailed = f
}

// Start begins the reconnection loop in a background goroutine.
// It uses Fibonacci backoff to calculate delay between attempts.
func (rm *ReconnectManager) Start() {
	rm.wg.Add(1)
	go rm.run()
}

// run is the main reconnection loop.
// It implements Fibonacci backoff: each delay is the sum of the previous two delays.
func (rm *ReconnectManager) run() {
	defer rm.wg.Done()

	// Fibonacci backoff: track last two delays
	// fib(0) = InitialDelay, fib(1) = InitialDelay, fib(n) = fib(n-1) + fib(n-2)
	prevDelay := time.Duration(0)
	delay := rm.config.InitialDelay
	attempt := 0

	for {
		attempt++
		// Use NewTimer instead of time.After to prevent memory leak
		timer := time.NewTimer(delay)
		select {
		case <-rm.ctx.Done():
			timer.Stop()
			return
		case <-rm.stopped:
			timer.Stop()
			return
		case <-timer.C:
			rm.mu.Lock()
			onReconnect := rm.onReconnect
			onReconnectFailed := rm.onReconnectFailed
			rm.mu.Unlock()

			if onReconnect == nil {
				// No callback set - stop reconnect loop to avoid infinite loop
				return
			}

			err := onReconnect()
			if err == nil {
				return
			}
			if onReconnectFailed != nil {
				onReconnectFailed(err)
			}

			if rm.config.MaxAttempts > 0 && attempt >= rm.config.MaxAttempts {
				return
			}

			// Fibonacci backoff: next delay = current + previous
			nextDelay := delay + prevDelay
			prevDelay = delay

			// Ensure minimum delay (don't go backwards)
			if nextDelay < delay {
				nextDelay = delay
			}

			// Apply max delay cap
			if nextDelay > rm.config.MaxDelay {
				nextDelay = rm.config.MaxDelay
			}

			delay = nextDelay
		}
	}
}

// Stop stops the reconnection attempts.
// It is idempotent - calling multiple times is safe.
func (rm *ReconnectManager) Stop() {
	rm.cancel()
	rm.stoppedOnce.Do(func() {
		close(rm.stopped)
	})
	rm.wg.Wait()
}

// Reset is a no-op for compatibility.
// Attempts are tracked locally in the run() loop.
func (rm *ReconnectManager) Reset() {
	// Attempts are tracked in run() loop, not persisted
}
