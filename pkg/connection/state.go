// Package connection provides connection management components
package connection

import (
	"fmt"
	"sync"

	"github.com/i0r3k/openclaw-sdk-go/pkg/types"
)

// StateChangeEvent represents a state change event
type StateChangeEvent struct {
	From   types.ConnectionState
	To     types.ConnectionState
	Reason error
}

// ConnectionStateMachine manages connection state
type ConnectionStateMachine struct {
	state  types.ConnectionState
	mu     sync.RWMutex
	events chan StateChangeEvent
}

// NewConnectionStateMachine creates a new state machine
func NewConnectionStateMachine(initial types.ConnectionState) *ConnectionStateMachine {
	return &ConnectionStateMachine{
		state:  initial,
		events: make(chan StateChangeEvent, 10),
	}
}

// validTransitions defines valid state transitions using typed constants
var validTransitions = map[types.ConnectionState][]types.ConnectionState{
	types.StateDisconnected:   {types.StateConnecting},
	types.StateConnecting:     {types.StateConnected, types.StateDisconnected, types.StateFailed},
	types.StateConnected:      {types.StateAuthenticating, types.StateDisconnected, types.StateReconnecting, types.StateFailed},
	types.StateAuthenticating: {types.StateAuthenticated, types.StateFailed},
	types.StateAuthenticated:  {types.StateDisconnected, types.StateReconnecting},
	types.StateReconnecting:   {types.StateConnecting, types.StateFailed},
	types.StateFailed:         {types.StateDisconnected},
}

func (csm *ConnectionStateMachine) validTransition(from, to types.ConnectionState) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// Transition changes the state
func (csm *ConnectionStateMachine) Transition(to types.ConnectionState, reason error) error {
	csm.mu.Lock()
	from := csm.state
	if !csm.validTransition(from, to) {
		csm.mu.Unlock()
		return fmt.Errorf("invalid state transition from %s to %s", from, to)
	}
	csm.state = to
	csm.mu.Unlock()

	select {
	case csm.events <- StateChangeEvent{From: from, To: to, Reason: reason}:
	default:
		// Channel full - return error so caller knows event was dropped
		return fmt.Errorf("state change event dropped: %s -> %s", from, to)
	}
	return nil
}

// State returns the current state
func (csm *ConnectionStateMachine) State() types.ConnectionState {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.state
}

// Events returns the state change event channel
func (csm *ConnectionStateMachine) Events() <-chan StateChangeEvent {
	return csm.events
}
