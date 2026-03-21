package managers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

func TestRequestManager_SendAndReceive(t *testing.T) {
	ctx := context.Background()
	rm := NewRequestManager(ctx)

	req := protocol.NewRequestFrame("test-123", "test", nil)

	resp := protocol.NewResponseFrameSuccess("test-123", json.RawMessage(`{}`))

	// Send response after a small delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		rm.HandleResponse(resp)
	}()

	got, err := rm.SendRequest(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "test-123" {
		t.Errorf("expected 'test-123', got '%s'", got.ID)
	}

	_ = rm.Close()
}

func TestRequestManager_ContextCancellation(t *testing.T) {
	ctx := context.Background()
	rm := NewRequestManager(ctx)

	req := protocol.NewRequestFrame("test-cancel", "test", nil)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := rm.SendRequest(ctx, req, nil)
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	_ = rm.Close()
}
