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

// Pairing provides access to node pairing operations.
type NodesPairingAPI struct {
	request RequestFn
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

// Pairing returns a new NodesPairingAPI for pairing operations.
func (api *NodesAPI) Pairing() *NodesPairingAPI {
	return &NodesPairingAPI{request: api.request}
}

// List lists node pairings.
func (p *NodesPairingAPI) List(ctx context.Context, params protocol.NodePairListParams) ([]NodePairingInfo, error) {
	raw, err := p.request(ctx, "node.pairing.list", params)
	if err != nil {
		return nil, err
	}
	var result []NodePairingInfo
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Approve approves a node pairing.
func (p *NodesPairingAPI) Approve(ctx context.Context, params protocol.NodePairApproveParams) error {
	_, err := p.request(ctx, "node.pairing.approve", params)
	return err
}

// Reject rejects a node pairing.
func (p *NodesPairingAPI) Reject(ctx context.Context, params protocol.NodePairRejectParams) error {
	_, err := p.request(ctx, "node.pairing.reject", params)
	return err
}

// Verify verifies a node pairing.
func (p *NodesPairingAPI) Verify(ctx context.Context, params protocol.NodePairVerifyParams) error {
	_, err := p.request(ctx, "node.pairing.verify", params)
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
