package connection

import (
	"testing"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

func TestConnectionStateMachine_Transition(t *testing.T) {
	csm := NewConnectionStateMachine(types.StateDisconnected)

	err := csm.Transition(types.StateConnecting, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if csm.State() != types.StateConnecting {
		t.Errorf("expected 'connecting', got '%s'", csm.State())
	}
}

func TestConnectionStateMachine_InvalidTransition(t *testing.T) {
	csm := NewConnectionStateMachine(types.StateDisconnected)

	err := csm.Transition(types.StateAuthenticated, nil)
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}

func TestConnectionStateMachine_StateChangeEvent(t *testing.T) {
	csm := NewConnectionStateMachine(types.StateDisconnected)

	err := csm.Transition(types.StateConnecting, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case event := <-csm.Events():
		if event.From != types.StateDisconnected {
			t.Errorf("expected from 'disconnected', got '%s'", event.From)
		}
		if event.To != types.StateConnecting {
			t.Errorf("expected to 'connecting', got '%s'", event.To)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for state change event")
	}
}
