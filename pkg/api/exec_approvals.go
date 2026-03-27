// Package api provides ExecApprovals API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// ExecApprovalsAPI provides access to ExecApprovals API methods.
type ExecApprovalsAPI struct {
	request RequestFn
}

// NewExecApprovalsAPI creates a new ExecApprovalsAPI instance.
func NewExecApprovalsAPI(request RequestFn) *ExecApprovalsAPI {
	return &ExecApprovalsAPI{request: request}
}

// Get returns exec approvals configuration.
func (api *ExecApprovalsAPI) Get(ctx context.Context) (protocol.ExecApprovalsSnapshot, error) {
	raw, err := api.request(ctx, "exec_approvals.get", protocol.ExecApprovalsGetParams{})
	if err != nil {
		return protocol.ExecApprovalsSnapshot{}, err
	}
	var result protocol.ExecApprovalsSnapshot
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.ExecApprovalsSnapshot{}, err
	}
	return result, nil
}

// Set sets exec approvals configuration.
func (api *ExecApprovalsAPI) Set(ctx context.Context, params protocol.ExecApprovalsSetParams) error {
	_, err := api.request(ctx, "exec_approvals.set", params)
	return err
}

// NodeGet gets exec approval for a node.
func (api *ExecApprovalsAPI) NodeGet(ctx context.Context, params protocol.ExecApprovalsNodeGetParams) error {
	_, err := api.request(ctx, "exec_approvals.node.get", params)
	return err
}

// NodeSet sets exec approval for a node.
func (api *ExecApprovalsAPI) NodeSet(ctx context.Context, params protocol.ExecApprovalsNodeSetParams) error {
	_, err := api.request(ctx, "exec_approvals.node.set", params)
	return err
}

// ApprovalRequest requests exec approval.
func (api *ExecApprovalsAPI) ApprovalRequest(ctx context.Context, params protocol.ExecApprovalRequestParams) error {
	_, err := api.request(ctx, "exec_approvals.approval.request", params)
	return err
}

// ApprovalWaitDecision waits for an approval decision.
func (api *ExecApprovalsAPI) ApprovalWaitDecision(ctx context.Context, params protocol.ExecApprovalWaitDecisionParams) error {
	_, err := api.request(ctx, "exec_approvals.approval.wait_decision", params)
	return err
}

// ApprovalResolve resolves an exec approval.
func (api *ExecApprovalsAPI) ApprovalResolve(ctx context.Context, params protocol.ExecApprovalResolveParams) error {
	_, err := api.request(ctx, "exec_approvals.approval.resolve", params)
	return err
}
