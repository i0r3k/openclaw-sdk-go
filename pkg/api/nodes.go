// Package api provides Nodes API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// NodesAPI provides access to Nodes API methods.
type NodesAPI struct {
	request RequestFn
}

// NewNodesAPI creates a new NodesAPI instance.
func NewNodesAPI(request RequestFn) *NodesAPI {
	return &NodesAPI{request: request}
}

// List returns all nodes.
func (api *NodesAPI) List(ctx context.Context) ([]NodeInfo, error) {
	raw, err := api.request(ctx, "node.list", protocol.NodeListParams{})
	if err != nil {
		return nil, err
	}
	var result []NodeInfo
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Invoke invokes a method on a node.
func (api *NodesAPI) Invoke(ctx context.Context, params protocol.NodeInvokeParams) (json.RawMessage, error) {
	return api.request(ctx, "node.invoke", params)
}

// Event sends an event to a node.
func (api *NodesAPI) Event(ctx context.Context, params protocol.NodeEventParams) error {
	_, err := api.request(ctx, "node.event", params)
	return err
}

// PendingDrain drains pending items from a node.
func (api *NodesAPI) PendingDrain(ctx context.Context, params protocol.NodePendingDrainParams) (protocol.NodePendingDrainResult, error) {
	raw, err := api.request(ctx, "node.pending.drain", params)
	if err != nil {
		return protocol.NodePendingDrainResult{}, err
	}
	var result protocol.NodePendingDrainResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.NodePendingDrainResult{}, err
	}
	return result, nil
}

// PendingEnqueue enqueues an item on a node.
func (api *NodesAPI) PendingEnqueue(ctx context.Context, params protocol.NodePendingEnqueueParams) error {
	_, err := api.request(ctx, "node.pending.enqueue", params)
	return err
}

// Describe returns detailed information about a node.
func (api *NodesAPI) Describe(ctx context.Context, params protocol.NodeDescribeParams) (protocol.NodeDescribeResult, error) {
	raw, err := api.request(ctx, "node.describe", params)
	if err != nil {
		return protocol.NodeDescribeResult{}, err
	}
	var result protocol.NodeDescribeResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.NodeDescribeResult{}, err
	}
	return result, nil
}

// PendingPull pulls pending items from a node.
func (api *NodesAPI) PendingPull(ctx context.Context, params protocol.NodePendingPullParams) (protocol.NodePendingPullResult, error) {
	raw, err := api.request(ctx, "node.pending.pull", params)
	if err != nil {
		return protocol.NodePendingPullResult{}, err
	}
	var result protocol.NodePendingPullResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.NodePendingPullResult{}, err
	}
	return result, nil
}

// PendingAck acknowledges pending items on a node.
func (api *NodesAPI) PendingAck(ctx context.Context, params protocol.NodePendingAckParams) error {
	_, err := api.request(ctx, "node.pending.ack", params)
	return err
}

// Rename renames a node.
func (api *NodesAPI) Rename(ctx context.Context, params protocol.NodeRenameParams) error {
	_, err := api.request(ctx, "node.rename", params)
	return err
}

// InvokeResult returns the result of a node invoke.
func (api *NodesAPI) InvokeResult(ctx context.Context, params protocol.NodeInvokeResultParams) (json.RawMessage, error) {
	return api.request(ctx, "node.invoke.result", params)
}

// CanvasCapabilityRefresh refreshes canvas capability for a node.
func (api *NodesAPI) CanvasCapabilityRefresh(ctx context.Context, params protocol.NodeCanvasCapabilityRefreshParams) error {
	_, err := api.request(ctx, "node.canvas.capability.refresh", params)
	return err
}

// PairingList lists node pairings.
func (api *NodesAPI) PairingList(ctx context.Context, params protocol.NodePairListParams) ([]NodePairingInfo, error) {
	raw, err := api.request(ctx, "node.pair.list", params)
	if err != nil {
		return nil, err
	}
	var result []NodePairingInfo
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PairingApprove approves a node pairing.
func (api *NodesAPI) PairingApprove(ctx context.Context, params protocol.NodePairApproveParams) error {
	_, err := api.request(ctx, "node.pair.approve", params)
	return err
}

// PairingReject rejects a node pairing.
func (api *NodesAPI) PairingReject(ctx context.Context, params protocol.NodePairRejectParams) error {
	_, err := api.request(ctx, "node.pair.reject", params)
	return err
}

// PairingVerify verifies a node pairing.
func (api *NodesAPI) PairingVerify(ctx context.Context, params protocol.NodePairVerifyParams) error {
	_, err := api.request(ctx, "node.pair.verify", params)
	return err
}

// PairingRequest requests a node pairing.
func (api *NodesAPI) PairingRequest(ctx context.Context, params protocol.NodePairRequestParams) error {
	_, err := api.request(ctx, "node.pair.request", params)
	return err
}

// NodeInfo represents information about a node.
type NodeInfo struct {
	ID       string         `json:"id"`
	Status   string         `json:"status,omitempty"`
	LastSeen int64          `json:"lastSeen,omitempty"`
	Extra    map[string]any `json:"*"`
}

// NodePairingInfo represents information about a node pairing.
type NodePairingInfo struct {
	PairingID string `json:"pairingId"`
	NodeID    string `json:"nodeId"`
	Status    string `json:"status,omitempty"`
}
