// Package connection provides connection management components for OpenClaw SDK.
//
// This package provides:
//   - ConnectionStateMachine: State machine for managing connection lifecycle
//   - PolicyManager: Connection policy configuration
//   - ProtocolNegotiator: Protocol version negotiation
//   - TLS validation: Certificate and configuration validation
package connection

import (
	"fmt"
	"sync"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// StateChangeEvent represents a state change event in the connection lifecycle.
// It contains the previous state (From), new state (To), and optional reason (error).
type StateChangeEvent struct {
	From   types.ConnectionState // Previous connection state
	To     types.ConnectionState // New connection state
	Reason error                 // Optional error that caused the transition
}

// ConnectionStateMachine manages connection state transitions.
// It enforces valid state transitions and emits events on state changes.
type ConnectionStateMachine struct {
	state  types.ConnectionState // Current connection state
	mu     sync.RWMutex          // Protects state access
	events chan StateChangeEvent // Channel for state change events
}

// NewConnectionStateMachine creates a new state machine with the given initial state.
func NewConnectionStateMachine(initial types.ConnectionState) *ConnectionStateMachine {
	return &ConnectionStateMachine{
		state:  initial,
		events: make(chan StateChangeEvent, 10),
	}
}

// validTransitions defines valid state transitions using typed constants.
// Each key is a state, and the value is a slice of states it can transition to.
var validTransitions = map[types.ConnectionState][]types.ConnectionState{
	types.StateDisconnected:   {types.StateConnecting},
	types.StateConnecting:     {types.StateConnected, types.StateDisconnected, types.StateFailed},
	types.StateConnected:      {types.StateAuthenticating, types.StateDisconnected, types.StateReconnecting, types.StateFailed},
	types.StateAuthenticating: {types.StateAuthenticated, types.StateFailed},
	types.StateAuthenticated:  {types.StateDisconnected, types.StateReconnecting},
	types.StateReconnecting:   {types.StateConnecting, types.StateFailed},
	types.StateFailed:         {types.StateDisconnected},
}

// validTransition checks if a state transition is valid.
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

// Transition changes the state from the current state to the target state.
// Returns an error if the transition is invalid or if the event channel is full.
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

// State returns the current connection state (thread-safe).
func (csm *ConnectionStateMachine) State() types.ConnectionState {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.state
}

// Events returns the state change event channel.
// Callers can receive state change notifications from this channel.
func (csm *ConnectionStateMachine) Events() <-chan StateChangeEvent {
	return csm.events
}

// Reset resets the state machine to the disconnected state.
func (csm *ConnectionStateMachine) Reset() {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.state = types.StateDisconnected
}

// IsReady returns true if the connection state is Connected or Authenticated.
func (csm *ConnectionStateMachine) IsReady() bool {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.state == types.StateConnected || csm.state == types.StateAuthenticated
}
