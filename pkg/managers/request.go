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

// RequestOptions contains options for a pending request.
type RequestOptions struct {
	Timeout    any       // Timeout for the request (time.Duration)
	OnProgress func(any) // Progress callback for intermediate updates
}

// pendingRequest holds state for an in-flight request.
type pendingRequest struct {
	responseCh chan *protocol.ResponseFrame
	onProgress func(any)
}

// RequestManager manages pending requests and their responses.
// It correlates outgoing requests with incoming responses using request IDs.
type RequestManager struct {
	pending  map[string]*pendingRequest    // Map of request ID to pending request
	timeouts map[string]context.CancelFunc // Map of request ID to timeout cancel function
	mu       sync.Mutex                    // Mutex for thread-safe access
	ctx      context.Context               // Context for lifecycle management
	cancel   context.CancelFunc            // Cancel function for context
}

// NewRequestManager creates a new request manager.
func NewRequestManager(ctx context.Context) *RequestManager {
	ctx, cancel := context.WithCancel(ctx)
	return &RequestManager{
		pending:  make(map[string]*pendingRequest),
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
	rm.pending[req.ID] = &pendingRequest{
		responseCh: respCh,
	}

	// Set up timeout cancellation if context has deadline
	if deadline, ok := ctx.Deadline(); ok {
		timeoutCtx, cancel := context.WithDeadline(ctx, deadline)
		rm.timeouts[req.ID] = cancel
		ctx = timeoutCtx
	}
	rm.mu.Unlock()

	cleanup := func() {
		rm.mu.Lock()
		delete(rm.pending, req.ID)
		if cancel, ok := rm.timeouts[req.ID]; ok {
			cancel()
			delete(rm.timeouts, req.ID)
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
	defer rm.mu.Unlock()

	req, ok := rm.pending[frame.ID]
	if !ok || req.responseCh == nil {
		return
	}

	select {
	case req.responseCh <- frame:
	default:
	}
}

// ResolveProgress delivers a progress update to the pending request.
func (rm *RequestManager) ResolveProgress(requestID string, payload any) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	req, ok := rm.pending[requestID]
	if !ok || req.onProgress == nil {
		return
	}

	req.onProgress(payload)
}

// AbortRequest aborts a pending request by ID.
func (rm *RequestManager) AbortRequest(requestID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	req, ok := rm.pending[requestID]
	if !ok {
		return
	}

	// Send a cancelled response
	cancelledResp := &protocol.ResponseFrame{
		Type:  protocol.FrameTypeResponse,
		ID:    requestID,
		Ok:    false,
		Error: &protocol.ErrorShape{Code: "REQUEST_CANCELLED", Message: "Request cancelled"},
	}

	select {
	case req.responseCh <- cancelledResp:
	default:
	}
}

// Clear cancels all pending requests.
func (rm *RequestManager) Clear() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for id, req := range rm.pending {
		close(req.responseCh)
		delete(rm.pending, id)
	}
	for _, cancel := range rm.timeouts {
		cancel()
	}
}

// Close cleans up all pending requests.
// It cancels the context, closes all pending channels, and clears timeout functions.
func (rm *RequestManager) Close() error {
	rm.mu.Lock()
	// First cancel context while holding lock to prevent new operations
	rm.cancel()
	// Clear pending map and timeouts
	for id, req := range rm.pending {
		close(req.responseCh)
		delete(rm.pending, id)
	}
	for _, cancel := range rm.timeouts {
		cancel()
	}
	rm.mu.Unlock()
	return nil
}
