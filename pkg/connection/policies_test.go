// Package connection provides tests for policy management.
package connection

import (
	"testing"
)

func TestPolicyManager_SetPolicies(t *testing.T) {
	pm := NewPolicyManager()

	policy := Policy{
		MaxPayload:       2097152,
		MaxBufferedBytes: 131072,
		TickIntervalMs:   15000,
	}

	pm.SetPolicies(policy)

	if !pm.HasPolicy() {
		t.Error("expected HasPolicy() to return true after SetPolicies")
	}

	if pm.GetMaxPayload() != 2097152 {
		t.Errorf("expected 2097152, got %d", pm.GetMaxPayload())
	}

	if pm.GetMaxBufferedBytes() != 131072 {
		t.Errorf("expected 131072, got %d", pm.GetMaxBufferedBytes())
	}

	if pm.GetTickIntervalMs() != 15000 {
		t.Errorf("expected 15000, got %d", pm.GetTickIntervalMs())
	}
}

func TestPolicyManager_DefaultPolicy(t *testing.T) {
	pm := NewPolicyManager()

	// Default policy should not be marked as "set"
	if pm.HasPolicy() {
		t.Error("expected HasPolicy() to return false for default policy")
	}

	// But getters should still return defaults
	if pm.GetMaxPayload() != 1048576 {
		t.Errorf("expected 1048576, got %d", pm.GetMaxPayload())
	}

	if pm.GetTickIntervalMs() != 30000 {
		t.Errorf("expected 30000, got %d", pm.GetTickIntervalMs())
	}
}

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy()

	if policy.MaxPayload != 1048576 {
		t.Errorf("expected 1048576, got %d", policy.MaxPayload)
	}

	if policy.MaxBufferedBytes != 65536 {
		t.Errorf("expected 65536, got %d", policy.MaxBufferedBytes)
	}

	if policy.TickIntervalMs != 30000 {
		t.Errorf("expected 30000, got %d", policy.TickIntervalMs)
	}
}

func TestDefaultProtocolVersionRange(t *testing.T) {
	rangeVal := DefaultProtocolVersionRange()

	if rangeVal.Min != 3 {
		t.Errorf("expected Min=3, got %d", rangeVal.Min)
	}

	if rangeVal.Max != 3 {
		t.Errorf("expected Max=3, got %d", rangeVal.Max)
	}
}
