// Package protocol provides error types for wire protocol errors.
//
// This package defines error codes used in the wire protocol layer
// (as opposed to SDK-level errors defined in pkg/types/errors.go).
package protocol

// WireErrorCode represents error codes used in protocol-level errors.
type WireErrorCode string

// Protocol-level error codes from src/protocol/errors.ts
const (
	WireErrNotLinked    WireErrorCode = "NOT_LINKED"
	WireErrNotPaired    WireErrorCode = "NOT_PAIRED"
	WireErrAgentTimeout WireErrorCode = "AGENT_TIMEOUT"
	WireErrInvalidReq   WireErrorCode = "INVALID_REQUEST"
	WireErrUnavailable  WireErrorCode = "UNAVAILABLE"
)

// IsWireErrorCode checks if a string is a valid wire error code
func IsWireErrorCode(code string) bool {
	switch WireErrorCode(code) {
	case WireErrNotLinked, WireErrNotPaired, WireErrAgentTimeout, WireErrInvalidReq, WireErrUnavailable:
		return true
	}
	return false
}
