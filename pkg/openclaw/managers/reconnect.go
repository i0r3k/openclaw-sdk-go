// Package managers provides high-level manager components for the OpenClaw SDK.
package managers

import (
	"context"
	"sync"
	"time"

	openclaw "github.com/i0r3k/openclaw-sdk-go/pkg/openclaw"
)

// ReconnectConfig holds reconnection configuration
// Uses openclaw.ReconnectConfig from Phase 1
type ReconnectConfig = openclaw.ReconnectConfig

// DefaultReconnectConfig returns default configuration
func DefaultReconnectConfig() *ReconnectConfig {
	cfg := openclaw.DefaultReconnectConfig()
	return &cfg
}

// ReconnectManager handles automatic reconnection
type ReconnectManager struct {
	config            *ReconnectConfig
	mu                sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	onReconnect       func() error
	onReconnectFailed func(err error)
	stopped           chan struct{}
	stoppedOnce       sync.Once
}

// NewReconnectManager creates a new reconnect manager
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

// SetOnReconnect sets the reconnect callback
func (rm *ReconnectManager) SetOnReconnect(f func() error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onReconnect = f
}

// SetOnReconnectFailed sets the reconnect failed callback
func (rm *ReconnectManager) SetOnReconnectFailed(f func(err error)) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.onReconnectFailed = f
}

// Start begins the reconnection loop
func (rm *ReconnectManager) Start() {
	rm.wg.Add(1)
	go rm.run()
}

func (rm *ReconnectManager) run() {
	defer rm.wg.Done()

	// Fibonacci backoff: track last two delays
	// fib(0) = InitialDelay, fib(1) = InitialDelay, fib(n) = fib(n-1) + fib(n-2)
	prevDelay := time.Duration(0)
	delay := rm.config.InitialDelay
	attempt := 0

	for {
		attempt++
		select {
		case <-rm.ctx.Done():
			return
		case <-rm.stopped:
			return
		case <-time.After(delay):
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

// Stop stops the reconnection attempts (idempotent)
func (rm *ReconnectManager) Stop() {
	rm.cancel()
	rm.stoppedOnce.Do(func() {
		close(rm.stopped)
	})
	rm.wg.Wait()
}

// Reset is a no-op for compatibility (attempts tracked locally in run())
func (rm *ReconnectManager) Reset() {
	// Attempts are tracked in run() loop, not persisted
}
