// Package connection provides protocol negotiation for OpenClaw SDK.
package connection

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// ProtocolNegotiator handles protocol version negotiation.
type ProtocolNegotiator struct {
	versionRange      ProtocolVersionRange
	negotiatedVersion *int
	defaultTimeout    time.Duration
}

// NegotiatedProtocol represents the result of protocol negotiation.
type NegotiatedProtocol struct {
	Version int
	Min     int
	Max     int
}

// NewProtocolNegotiator creates a new ProtocolNegotiator.
// If no versionRange is provided, defaults to {min: 3, max: 3}.
func NewProtocolNegotiator(versionRangeVal ...ProtocolVersionRange) *ProtocolNegotiator {
	p := &ProtocolNegotiator{
		versionRange:      DefaultProtocolVersionRange(),
		negotiatedVersion: nil,
		defaultTimeout:    5 * time.Second,
	}
	if len(versionRangeVal) > 0 {
		p.versionRange = versionRangeVal[0]
	}
	return p
}

// GetRange returns the protocol version versionRange.
func (p *ProtocolNegotiator) GetRange() ProtocolVersionRange {
	return p.versionRange
}

// GetNegotiatedVersion returns the negotiated version or nil if not yet negotiated.
func (p *ProtocolNegotiator) GetNegotiatedVersion() *int {
	return p.negotiatedVersion
}

// IsNegotiated returns true if negotiation has completed.
func (p *ProtocolNegotiator) IsNegotiated() bool {
	return p.negotiatedVersion != nil
}

// Negotiate performs protocol version negotiation with context support.
func (p *ProtocolNegotiator) Negotiate(ctx context.Context, helloOk *HelloOk) (*NegotiatedProtocol, error) {
	select {
	case <-ctx.Done():
		return nil, types.NewProtocolError("PROTOCOL_NEGOTIATION_TIMEOUT", "protocol negotiation timeout", false, ctx.Err())
	default:
	}

	serverVersion := helloOk.Protocol

	// Check if server version is within our supported versionRange
	if serverVersion < p.versionRange.Min || serverVersion > p.versionRange.Max {
		return nil, types.NewProtocolError(
			"PROTOCOL_NEGOTIATION_FAILED",
			"protocol version out of versionRange",
			false,
			errors.New("protocol version out of supported versionRange"),
		)
	}

	p.negotiatedVersion = &serverVersion

	return &NegotiatedProtocol{
		Version: serverVersion,
		Min:     p.versionRange.Min,
		Max:     p.versionRange.Max,
	}, nil
}

// Reset resets the negotiated version.
func (p *ProtocolNegotiator) Reset() {
	p.negotiatedVersion = nil
}

// NegotiateWithTimeout performs protocol negotiation with a timeout.
func (p *ProtocolNegotiator) NegotiateWithTimeout(helloOk *HelloOk, timeout time.Duration) (*NegotiatedProtocol, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return p.Negotiate(ctx, helloOk)
}

// GetSupportedVersions returns the list of supported protocol versions.
func (p *ProtocolNegotiator) GetSupportedVersions() []int {
	versions := make([]int, 0, p.versionRange.Max-p.versionRange.Min+1)
	for v := p.versionRange.Min; v <= p.versionRange.Max; v++ {
		versions = append(versions, v)
	}
	return versions
}

// IsVersionSupported checks if a version is within the supported versionRange.
func (p *ProtocolNegotiator) IsVersionSupported(version int) bool {
	return version >= p.versionRange.Min && version <= p.versionRange.Max
}

// NegotiateWithServerVersions performs negotiation given server versions list.
func (p *ProtocolNegotiator) NegotiateWithServerVersions(ctx context.Context, serverVersions []int) (string, error) {
	select {
	case <-ctx.Done():
		return "", types.NewProtocolError("PROTOCOL_NEGOTIATION_TIMEOUT", "protocol negotiation timeout", false, ctx.Err())
	default:
	}

	// Find first matching version
	for _, clientVer := range p.GetSupportedVersions() {
		if slices.Contains(serverVersions, clientVer) {
			return string(rune('0' + clientVer)), nil
		}
	}

	return "", types.NewProtocolError("PROTOCOL_NEGOTIATION_FAILED", "no matching protocol version", false, nil)
}
