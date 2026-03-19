// Package connection provides connection management components for OpenClaw SDK.
//
// This package provides:
//   - ConnectionStateMachine: State machine for managing connection lifecycle
//   - PolicyManager: Connection policy configuration
//   - ProtocolNegotiator: Protocol version negotiation
//   - TLS validation: Certificate and configuration validation
package connection

import (
	"context"
	"errors"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// ProtocolNegotiator handles protocol version negotiation.
// It compares client-supported versions with server-supported versions
// to find a mutually compatible protocol version.
type ProtocolNegotiator struct {
	supportedVersions []string      // Protocol versions supported by the client
	defaultTimeout    time.Duration // Default timeout for negotiation
}

// NewProtocolNegotiator creates a new protocol negotiator.
// If no supported versions are provided, defaults to "1.0".
func NewProtocolNegotiator(supportedVersions []string) *ProtocolNegotiator {
	if len(supportedVersions) == 0 {
		supportedVersions = []string{"1.0"}
	}
	return &ProtocolNegotiator{
		supportedVersions: supportedVersions,
		defaultTimeout:    5 * time.Second,
	}
}

// Negotiate performs protocol version negotiation with context support.
// It finds the first matching version from client and server supported versions.
// Returns the negotiated version or an error if no match is found or timeout occurs.
func (p *ProtocolNegotiator) Negotiate(ctx context.Context, serverVersions []string) (string, error) {
	// Create a timeout if context doesn't have one
	ctx, cancel := context.WithTimeout(ctx, p.defaultTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return "", types.NewProtocolError("protocol negotiation timeout", ctx.Err())
		default:
			// Check for matching versions
			for _, clientVer := range p.supportedVersions {
				for _, serverVer := range serverVersions {
					if clientVer == serverVer {
						return clientVer, nil
					}
				}
			}
			// No match found
			return "", types.NewProtocolError("no matching protocol version", nil)
		}
	}
}

// ErrNoMatchingProtocol is a sentinel error for protocol negotiation failures.
// Use errors.Is() to check for this specific error.
var ErrNoMatchingProtocol = errors.New("no matching protocol version")
