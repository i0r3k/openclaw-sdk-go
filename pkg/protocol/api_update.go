// Package protocol provides API parameter types for OpenClaw SDK.
//
// This file contains Update, Poll, and ChatInject types migrated from TypeScript: src/protocol/api-params.ts
package protocol

// ============================================================================
// Poll / Update / ChatInject Types
// ============================================================================

// PollParams parameters for polling.
type PollParams struct{}

// UpdateRunParams parameters for running update.
type UpdateRunParams struct{}

// ChatInjectParams parameters for injecting chat.
// Updated to match TS SessionsSendParams alias (v2.0.0 breaking change).
type ChatInjectParams struct {
	Key            string  `json:"key"`
	Message        string  `json:"message"`
	Thinking       *string `json:"thinking,omitempty"`
	Attachments    []any   `json:"attachments,omitempty"`
	TimeoutMs      *int64  `json:"timeoutMs,omitempty"`
	IdempotencyKey *string `json:"idempotencyKey,omitempty"`
}

// ============================================================================
// Node Types
// ============================================================================

// NodeListParams parameters for listing nodes.
type NodeListParams struct{}

// NodeInvokeParams parameters for invoking a node.
type NodeInvokeParams struct {
	NodeID string `json:"nodeId"`
	Target string `json:"target"`
	Params any    `json:"params,omitempty"`
}

// NodeInvokeResultParams represents node invoke result parameters.
type NodeInvokeResultParams struct {
	Result any `json:"result"`
}

// NodeEventParams parameters for node event.
type NodeEventParams struct {
	NodeID  string `json:"nodeId"`
	Event   string `json:"event"`
	Payload any    `json:"payload,omitempty"`
}

// NodePendingDrainParams parameters for draining pending node items.
type NodePendingDrainParams struct {
	NodeID string `json:"nodeId"`
	Max    int64  `json:"max,omitempty"`
}

// NodePendingDrainResult result of draining pending node items.
type NodePendingDrainResult struct {
	Items []any `json:"items"`
}

// NodePendingEnqueueParams parameters for enqueueing a pending node item.
type NodePendingEnqueueParams struct {
	NodeID string `json:"nodeId"`
	Item   any    `json:"item"`
}

// NodePendingEnqueueResult result of enqueueing a pending node item.
type NodePendingEnqueueResult struct{}

// ============================================================================
// Update Types
// ============================================================================

// UpdateCheckParams parameters for checking for updates.
type UpdateCheckParams struct{}

// UpdateCheckResult result of checking for updates.
type UpdateCheckResult struct {
	Available bool   `json:"available"`
	Version   string `json:"version,omitempty"`
	Changelog string `json:"changelog,omitempty"`
}

// UpdateApplyParams parameters for applying an update.
type UpdateApplyParams struct {
	Version string `json:"version,omitempty"`
}

// UpdateApplyResult result of applying an update.
type UpdateApplyResult struct {
	Success bool   `json:"success"`
	Version string `json:"version,omitempty"`
}
