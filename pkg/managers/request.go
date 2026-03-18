// Package managers provides high-level manager components for the OpenClaw SDK.
package managers

import (
	"context"
	"sync"

	"github.com/i0r3k/openclaw-sdk-go/pkg/protocol"
)

// RequestManager manages pending requests
type RequestManager struct {
	pending  map[string]chan *protocol.ResponseFrame
	timeouts map[string]context.CancelFunc
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewRequestManager creates a new request manager
func NewRequestManager(ctx context.Context) *RequestManager {
	ctx, cancel := context.WithCancel(ctx)
	return &RequestManager{
		pending:  make(map[string]chan *protocol.ResponseFrame),
		timeouts: make(map[string]context.CancelFunc),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// SendRequest sends a request and waits for response
func (rm *RequestManager) SendRequest(ctx context.Context, req *protocol.RequestFrame, sendFunc func(*protocol.RequestFrame) error) (*protocol.ResponseFrame, error) {
	respCh := make(chan *protocol.ResponseFrame, 1)

	rm.mu.Lock()
	rm.pending[req.RequestID] = respCh

	// Set up timeout cancellation if context has deadline
	if deadline, ok := ctx.Deadline(); ok {
		timeoutCtx, cancel := context.WithDeadline(ctx, deadline)
		rm.timeouts[req.RequestID] = cancel
		ctx = timeoutCtx
	}
	rm.mu.Unlock()

	cleanup := func() {
		rm.mu.Lock()
		delete(rm.pending, req.RequestID)
		if cancel, ok := rm.timeouts[req.RequestID]; ok {
			cancel()
			delete(rm.timeouts, req.RequestID)
		}
		rm.mu.Unlock()
		close(respCh)
	}
	defer cleanup()

	// Send the request via transport
	if sendFunc != nil {
		if err := sendFunc(req); err != nil {
			return nil, err
		}
	}

	select {
	case resp := <-respCh:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// HandleResponse handles an incoming response
func (rm *RequestManager) HandleResponse(frame *protocol.ResponseFrame) {
	rm.mu.Lock()
	ch, ok := rm.pending[frame.RequestID]
	rm.mu.Unlock()

	if ok && ch != nil {
		select {
		case ch <- frame:
		default:
		}
	}
}

// Close cleans up all pending requests (thread-safe)
func (rm *RequestManager) Close() error {
	rm.mu.Lock()
	// First cancel context while holding lock to prevent new operations
	rm.cancel()
	// Clear pending map and timeouts
	for id, ch := range rm.pending {
		close(ch)
		delete(rm.pending, id)
	}
	for _, cancel := range rm.timeouts {
		cancel()
	}
	rm.mu.Unlock()
	return nil
}
