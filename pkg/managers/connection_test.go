package managers

import (
	"context"
	"testing"

	"github.com/i0r3k/openclaw-sdk-go/pkg/types"
)

func TestConnectionManager_State(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10)
	cm := NewConnectionManager(ctx, config, em)

	state := cm.State()
	if state != types.StateDisconnected {
		t.Errorf("expected disconnected, got %s", state)
	}

	_ = em.Close()
}

func TestConnectionManager_DisconnectWhenNotConnected(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10)
	cm := NewConnectionManager(ctx, config, em)

	// Disconnect when not connected should not error
	err := cm.Disconnect()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	_ = em.Close()
}

func TestConnectionManager_TransportWhenNotConnected(t *testing.T) {
	ctx := context.Background()
	config := &ClientConfig{URL: "ws://localhost:8080"}
	em := NewEventManager(ctx, 10)
	cm := NewConnectionManager(ctx, config, em)

	transport := cm.Transport()
	if transport != nil {
		t.Error("expected nil transport when not connected")
	}

	_ = em.Close()
}
