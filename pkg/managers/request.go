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

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// RequestManager manages pending requests and their responses.
// It correlates outgoing requests with incoming responses using request IDs.
type RequestManager struct {
	pending  map[string]chan *protocol.ResponseFrame // Map of request ID to response channel
	timeouts map[string]context.CancelFunc           // Map of request ID to timeout cancel function
	mu       sync.Mutex                              // Mutex for thread-safe access
	ctx      context.Context                         // Context for lifecycle management
	cancel   context.CancelFunc                      // Cancel function for context
}

// NewRequestManager creates a new request manager.
func NewRequestManager(ctx context.Context) *RequestManager {
	ctx, cancel := context.WithCancel(ctx)
	return &RequestManager{
		pending:  make(map[string]chan *protocol.ResponseFrame),
		timeouts: make(map[string]context.CancelFunc),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// SendRequest sends a request and waits for a response.
// It registers the request ID, sends the request via sendFunc, and waits for response.
// Returns the response or an error if the request times out or is cancelled.
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

// HandleResponse handles an incoming response frame.
// It correlates the response with the pending request using RequestID.
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

// Close cleans up all pending requests.
// It cancels the context, closes all pending channels, and clears timeout functions.
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
