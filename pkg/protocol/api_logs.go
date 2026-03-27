// Package protocol provides API parameter types for OpenClaw SDK.
//
// This file contains Logs and ExecApprovals types migrated from TypeScript: src/protocol/api-params.ts
package protocol

// ============================================================================
// Logs Types
// ============================================================================

// LogsTailParams parameters for tailing logs.
type LogsTailParams struct {
	Lines int64 `json:"lines,omitempty"`
}

// LogsTailResult result of tailing logs.
type LogsTailResult struct {
	Logs []string `json:"logs"`
}

// ============================================================================
// ExecApprovals Types
// ============================================================================

// ExecApprovalsGetParams parameters for getting exec approvals.
type ExecApprovalsGetParams struct{}

// ExecApprovalsSetParams parameters for setting exec approvals.
type ExecApprovalsSetParams struct {
	Enabled bool `json:"enabled"`
}

// ExecApprovalsSnapshot represents exec approvals snapshot.
type ExecApprovalsSnapshot struct {
	Approvals []any `json:"approvals"`
}

// ExecApprovalsNodeGetParams parameters for getting node exec approvals.
type ExecApprovalsNodeGetParams struct {
	NodeID string `json:"nodeId"`
}

// ExecApprovalsNodeSetParams parameters for setting node exec approvals.
type ExecApprovalsNodeSetParams struct {
	NodeID   string `json:"nodeId"`
	Approved bool   `json:"approved"`
}

// ExecApprovalRequestParams parameters for requesting exec approval.
type ExecApprovalRequestParams struct {
	NodeID  string `json:"nodeId"`
	Command string `json:"command"`
}

// ExecApprovalWaitDecisionParams parameters for waiting for exec approval decision.
type ExecApprovalWaitDecisionParams struct {
	NodeID  string `json:"nodeId"`
	Timeout int64  `json:"timeout,omitempty"`
}

// ExecApprovalResolveParams parameters for resolving exec approval.
type ExecApprovalResolveParams struct {
	NodeID   string `json:"nodeId"`
	Approved bool   `json:"approved"`
}
