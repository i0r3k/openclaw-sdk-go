// Package connection provides policy management for OpenClaw SDK.
package connection

// PolicyManager manages server policies received in hello-ok response.
type PolicyManager struct {
	policy       Policy
	hasSetPolicy bool
}

// NewPolicyManager creates a new PolicyManager with default policy.
func NewPolicyManager() *PolicyManager {
	return &PolicyManager{
		policy:       DefaultPolicy(),
		hasSetPolicy: false,
	}
}

// SetPolicies stores policies from hello-ok response.
func (pm *PolicyManager) SetPolicies(policy Policy) {
	pm.policy = policy
	pm.hasSetPolicy = true
}

// HasPolicy checks if policy has been explicitly set.
func (pm *PolicyManager) HasPolicy() bool {
	return pm.hasSetPolicy
}

// GetPolicy returns the current policy.
func (pm *PolicyManager) GetPolicy() Policy {
	return pm.policy
}

// GetMaxPayload returns maximum payload size.
func (pm *PolicyManager) GetMaxPayload() int64 {
	return pm.policy.MaxPayload
}

// GetMaxBufferedBytes returns maximum buffered bytes.
func (pm *PolicyManager) GetMaxBufferedBytes() int64 {
	return pm.policy.MaxBufferedBytes
}

// GetTickIntervalMs returns tick interval in milliseconds.
func (pm *PolicyManager) GetTickIntervalMs() int64 {
	return pm.policy.TickIntervalMs
}
