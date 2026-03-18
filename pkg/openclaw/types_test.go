package openclaw

import (
	"testing"
	"time"
)

func TestConnectionState(t *testing.T) {
	states := []ConnectionState{
		StateDisconnected,
		StateConnecting,
		StateConnected,
		StateAuthenticating,
		StateAuthenticated,
		StateReconnecting,
		StateFailed,
	}

	for _, s := range states {
		if s == "" {
			t.Error("state should not be empty")
		}
	}
}

func TestEventType(t *testing.T) {
	types := []EventType{
		EventConnect,
		EventDisconnect,
		EventError,
		EventMessage,
		EventRequest,
		EventResponse,
		EventTick,
		EventGap,
		EventStateChange,
	}

	for _, et := range types {
		if et == "" {
			t.Error("event type should not be empty")
		}
	}
}

func TestDefaultReconnectConfig(t *testing.T) {
	cfg := DefaultReconnectConfig()

	if cfg.MaxAttempts != 0 {
		t.Errorf("expected MaxAttempts=0 (infinite), got %d", cfg.MaxAttempts)
	}
	if cfg.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay=1s, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 60*time.Second {
		t.Errorf("expected MaxDelay=60s, got %v", cfg.MaxDelay)
	}
	if cfg.BackoffMultiplier != 1.618 {
		t.Errorf("expected BackoffMultiplier=1.618, got %f", cfg.BackoffMultiplier)
	}
}
