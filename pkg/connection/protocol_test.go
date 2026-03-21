// Package connection provides tests for protocol negotiation.
package connection

import (
	"context"
	"testing"
)

func TestProtocolNegotiator_DefaultRange(t *testing.T) {
	negotiator := NewProtocolNegotiator()

	rangeVal := negotiator.GetRange()
	if rangeVal.Min != 3 {
		t.Errorf("expected Min=3, got %d", rangeVal.Min)
	}
	if rangeVal.Max != 3 {
		t.Errorf("expected Max=3, got %d", rangeVal.Max)
	}
}

func TestProtocolNegotiator_CustomRange(t *testing.T) {
	negotiator := NewProtocolNegotiator(ProtocolVersionRange{Min: 1, Max: 5})

	rangeVal := negotiator.GetRange()
	if rangeVal.Min != 1 {
		t.Errorf("expected Min=1, got %d", rangeVal.Min)
	}
	if rangeVal.Max != 5 {
		t.Errorf("expected Max=5, got %d", rangeVal.Max)
	}
}

func TestProtocolNegotiator_Negotiate_Success(t *testing.T) {
	negotiator := NewProtocolNegotiator()

	helloOk := &HelloOk{
		Protocol: 3,
		Server: HelloOkServer{
			Version: "1.0",
			ConnID:  "conn-123",
		},
		Features: HelloOkFeatures{
			Methods: []string{"chat.send", "agent.create"},
			Events:  []string{"tick", "message"},
		},
		Snapshot: Snapshot{
			StateVersion: 1,
			UptimeMs:     1000,
		},
		Policy: Policy{
			MaxPayload:       1048576,
			MaxBufferedBytes: 65536,
			TickIntervalMs:   30000,
		},
	}

	result, err := negotiator.Negotiate(context.Background(), helloOk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Version != 3 {
		t.Errorf("expected Version=3, got %d", result.Version)
	}

	if !negotiator.IsNegotiated() {
		t.Error("expected IsNegotiated() to return true")
	}
}

func TestProtocolNegotiator_Negotiate_OutOfRange(t *testing.T) {
	negotiator := NewProtocolNegotiator(ProtocolVersionRange{Min: 1, Max: 2})

	helloOk := &HelloOk{
		Protocol: 3, // Out of range
	}

	_, err := negotiator.Negotiate(context.Background(), helloOk)
	if err == nil {
		t.Fatal("expected error for out-of-range protocol")
	}
}

func TestProtocolNegotiator_Reset(t *testing.T) {
	negotiator := NewProtocolNegotiator()

	helloOk := &HelloOk{
		Protocol: 3,
		Server: HelloOkServer{
			Version: "1.0",
			ConnID:  "conn-123",
		},
		Features: HelloOkFeatures{
			Methods: []string{"chat.send"},
			Events:  []string{"tick"},
		},
		Snapshot: Snapshot{
			StateVersion: 1,
			UptimeMs:     1000,
		},
		Policy: DefaultPolicy(),
	}

	_, err := negotiator.Negotiate(context.Background(), helloOk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !negotiator.IsNegotiated() {
		t.Error("expected IsNegotiated() to return true")
	}

	negotiator.Reset()

	if negotiator.IsNegotiated() {
		t.Error("expected IsNegotiated() to return false after Reset()")
	}
}

func TestProtocolNegotiator_GetSupportedVersions(t *testing.T) {
	negotiator := NewProtocolNegotiator(ProtocolVersionRange{Min: 1, Max: 3})

	versions := negotiator.GetSupportedVersions()
	if len(versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(versions))
	}

	if versions[0] != 1 || versions[1] != 2 || versions[2] != 3 {
		t.Errorf("unexpected versions: %v", versions)
	}
}

func TestProtocolNegotiator_IsVersionSupported(t *testing.T) {
	negotiator := NewProtocolNegotiator(ProtocolVersionRange{Min: 2, Max: 4})

	tests := []struct {
		version   int
		supported bool
	}{
		{1, false},
		{2, true},
		{3, true},
		{4, true},
		{5, false},
	}

	for _, tt := range tests {
		if negotiator.IsVersionSupported(tt.version) != tt.supported {
			t.Errorf("IsVersionSupported(%d) = %v, want %v", tt.version, !tt.supported, tt.supported)
		}
	}
}
