// Package protocol provides API parameter types for OpenClaw SDK.
//
// This file contains Agent, Node Pairing, and Device Pairing types
// migrated from TypeScript: src/protocol/api-params.ts
package protocol

// ============================================================================
// Agent Types
// ============================================================================

// AgentIdentityParams parameters for agent identity verification.
type AgentIdentityParams struct {
	AgentID string `json:"agentId"`
}

// AgentIdentityResult result of agent identity verification.
type AgentIdentityResult struct {
	ID      string        `json:"id"`
	Summary *AgentSummary `json:"summary,omitempty"`
}

// AgentWaitParams parameters for waiting on agent.
type AgentWaitParams struct {
	AgentID   string `json:"agentId"`
	TimeoutMs int64  `json:"timeoutMs,omitempty"`
}

// AgentsFileEntry represents a file entry for agent file operations.
type AgentsFileEntry struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// AgentsCreateParams parameters for creating an agent.
// Updated to match TS v2.0.0 (commit 87ef46f).
type AgentsCreateParams struct {
	Name      string  `json:"name"`
	Workspace string  `json:"workspace"`
	Emoji     *string `json:"emoji,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
}

// AgentsCreateResult result of creating an agent.
// Updated to match TS v2.0.0 (commit 87ef46f).
type AgentsCreateResult struct {
	Ok        bool   `json:"ok"`
	AgentID   string `json:"agentId"`
	Name      string `json:"name"`
	Workspace string `json:"workspace"`
}

// AgentsUpdateParams parameters for updating an agent.
// Updated to match TS v2.0.0 (commit 87ef46f).
type AgentsUpdateParams struct {
	AgentID   string  `json:"agentId"`
	Name      *string `json:"name,omitempty"`
	Workspace *string `json:"workspace,omitempty"`
	Model     *string `json:"model,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
}

// AgentsUpdateResult result of updating an agent.
type AgentsUpdateResult struct {
	Ok      bool   `json:"ok"`
	AgentID string `json:"agentId"`
}

// AgentsDeleteParams parameters for deleting an agent.
type AgentsDeleteParams struct {
	AgentID string `json:"agentId"`
}

// AgentsDeleteResult result of deleting an agent.
type AgentsDeleteResult struct {
	AgentID string `json:"agentId"`
}

// AgentsFilesListParams parameters for listing agent files.
type AgentsFilesListParams struct {
	AgentID string `json:"agentId"`
}

// AgentsFilesListResult result of listing agent files.
type AgentsFilesListResult struct {
	Files []string `json:"files"`
}

// AgentsFilesGetParams parameters for getting an agent file.
type AgentsFilesGetParams struct {
	AgentID string `json:"agentId"`
	Path    string `json:"path"`
}

// AgentsFilesGetResult result of getting an agent file.
type AgentsFilesGetResult struct {
	Content string `json:"content"`
}

// AgentsFilesSetParams parameters for setting an agent file.
type AgentsFilesSetParams struct {
	AgentID string `json:"agentId"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

// AgentsFilesSetResult result of setting an agent file.
type AgentsFilesSetResult struct{}

// AgentsListParams parameters for listing agents.
type AgentsListParams struct{}

// AgentsListResult result of listing agents.
type AgentsListResult struct {
	Agents []AgentSummary `json:"agents"`
}

// ============================================================================
// Node Pairing Types
// ============================================================================

// NodePairRequestParams parameters for requesting node pairing.
type NodePairRequestParams struct {
	NodeID string `json:"nodeId"`
	TtlSec int64  `json:"ttlSec,omitempty"`
}

// NodePairListParams parameters for listing node pairings.
type NodePairListParams struct {
	NodeID string `json:"nodeId"`
}

// NodePairApproveParams parameters for approving node pairing.
type NodePairApproveParams struct {
	NodeID    string `json:"nodeId"`
	PairingID string `json:"pairingId"`
}

// NodePairRejectParams parameters for rejecting node pairing.
type NodePairRejectParams struct {
	NodeID    string `json:"nodeId"`
	PairingID string `json:"pairingId"`
}

// NodePairVerifyParams parameters for verifying node pairing.
type NodePairVerifyParams struct {
	NodeID    string `json:"nodeId"`
	PairingID string `json:"pairingId"`
	Code      string `json:"code"`
}

// ============================================================================
// Device Pairing Types
// ============================================================================

// DevicePairListParams parameters for listing device pairings.
type DevicePairListParams struct{}

// DevicePairApproveParams parameters for approving device pairing.
type DevicePairApproveParams struct {
	NodeID    string `json:"nodeId"`
	RequestID string `json:"requestId"`
}

// DevicePairRejectParams parameters for rejecting device pairing.
type DevicePairRejectParams struct {
	PairingID string `json:"pairingId"`
}

// DeviceTokenRotateParams parameters for rotating device token.
type DeviceTokenRotateParams struct {
	DeviceID string `json:"deviceId"`
}

// DeviceTokenRevokeParams parameters for revoking device token.
type DeviceTokenRevokeParams struct {
	DeviceID string `json:"deviceId"`
}

// NodeDescribeParams parameters for describing a node.
type NodeDescribeParams struct {
	NodeID string `json:"nodeId"`
}

// NodeDescribeResult result of describing a node.
type NodeDescribeResult struct {
	NodeID string `json:"nodeId"`
	Status string `json:"status"`
}

// NodePendingPullParams parameters for pulling pending node items.
type NodePendingPullParams struct {
	NodeID string `json:"nodeId"`
}

// NodePendingPullResult result of pulling pending node items.
type NodePendingPullResult struct {
	Items []any `json:"items"`
}

// NodePendingAckParams parameters for acknowledging pending node items.
type NodePendingAckParams struct {
	NodeID string `json:"nodeId"`
	ItemID string `json:"itemId"`
}

// NodeRenameParams parameters for renaming a node.
type NodeRenameParams struct {
	NodeID string `json:"nodeId"`
	Name   string `json:"name"`
}

// NodeCanvasCapabilityRefreshParams parameters for refreshing canvas capability.
type NodeCanvasCapabilityRefreshParams struct {
	NodeID string `json:"nodeId"`
}
