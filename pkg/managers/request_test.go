package managers

import (
	"context"
	"testing"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

func TestRequestManager_SendAndReceive(t *testing.T) {
	ctx := context.Background()
	rm := NewRequestManager(ctx)

	req := &protocol.RequestFrame{
		RequestID: "test-123",
		Method:    "test",
		Timestamp: time.Now(),
	}

	resp := &protocol.ResponseFrame{
		RequestID: "test-123",
		Success:   true,
		Timestamp: time.Now(),
	}

	// Send response after a small delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		rm.HandleResponse(resp)
	}()

	got, err := rm.SendRequest(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.RequestID != "test-123" {
		t.Errorf("expected 'test-123', got '%s'", got.RequestID)
	}

	_ = rm.Close()
}

func TestRequestManager_ContextCancellation(t *testing.T) {
	ctx := context.Background()
	rm := NewRequestManager(ctx)

	req := &protocol.RequestFrame{
		RequestID: "test-cancel",
		Method:    "test",
		Timestamp: time.Now(),
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := rm.SendRequest(ctx, req, nil)
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	_ = rm.Close()
}
