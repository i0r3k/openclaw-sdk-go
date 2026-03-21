// Package api provides Agents API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// AgentsAPI provides access to Agents API methods.
type AgentsAPI struct {
	request RequestFn
}

// NewAgentsAPI creates a new AgentsAPI instance.
func NewAgentsAPI(request RequestFn) *AgentsAPI {
	return &AgentsAPI{request: request}
}

// Files provides access to agent file operations.
type AgentsFilesAPI struct {
	request RequestFn
}

// Identity returns agent identity information.
func (api *AgentsAPI) Identity(ctx context.Context, params protocol.AgentIdentityParams) (protocol.AgentIdentityResult, error) {
	raw, err := api.request(ctx, "agents.identity", params)
	if err != nil {
		return protocol.AgentIdentityResult{}, err
	}
	var result protocol.AgentIdentityResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.AgentIdentityResult{}, err
	}
	return result, nil
}

// Wait waits for an agent to complete.
func (api *AgentsAPI) Wait(ctx context.Context, params protocol.AgentWaitParams) error {
	_, err := api.request(ctx, "agents.wait", params)
	return err
}

// Create creates a new agent.
func (api *AgentsAPI) Create(ctx context.Context, params protocol.AgentsCreateParams) (protocol.AgentsCreateResult, error) {
	raw, err := api.request(ctx, "agents.create", params)
	if err != nil {
		return protocol.AgentsCreateResult{}, err
	}
	var result protocol.AgentsCreateResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.AgentsCreateResult{}, err
	}
	return result, nil
}

// Update updates an existing agent.
func (api *AgentsAPI) Update(ctx context.Context, params protocol.AgentsUpdateParams) (protocol.AgentsUpdateResult, error) {
	raw, err := api.request(ctx, "agents.update", params)
	if err != nil {
		return protocol.AgentsUpdateResult{}, err
	}
	var result protocol.AgentsUpdateResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.AgentsUpdateResult{}, err
	}
	return result, nil
}

// Delete deletes an agent.
func (api *AgentsAPI) Delete(ctx context.Context, params protocol.AgentsDeleteParams) (protocol.AgentsDeleteResult, error) {
	raw, err := api.request(ctx, "agents.delete", params)
	if err != nil {
		return protocol.AgentsDeleteResult{}, err
	}
	var result protocol.AgentsDeleteResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.AgentsDeleteResult{}, err
	}
	return result, nil
}

// List returns all agents.
func (api *AgentsAPI) List(ctx context.Context) (protocol.AgentsListResult, error) {
	raw, err := api.request(ctx, "agents.list", protocol.AgentsListParams{})
	if err != nil {
		return protocol.AgentsListResult{}, err
	}
	var result protocol.AgentsListResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.AgentsListResult{}, err
	}
	return result, nil
}

// Files returns a new AgentsFilesAPI for file operations.
func (api *AgentsAPI) Files() *AgentsFilesAPI {
	return &AgentsFilesAPI{request: api.request}
}

// List lists agent files.
func (f *AgentsFilesAPI) List(ctx context.Context, params protocol.AgentsFilesListParams) (protocol.AgentsFilesListResult, error) {
	raw, err := f.request(ctx, "agents.files.list", params)
	if err != nil {
		return protocol.AgentsFilesListResult{}, err
	}
	var result protocol.AgentsFilesListResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.AgentsFilesListResult{}, err
	}
	return result, nil
}

// Get gets an agent file.
func (f *AgentsFilesAPI) Get(ctx context.Context, params protocol.AgentsFilesGetParams) (protocol.AgentsFilesGetResult, error) {
	raw, err := f.request(ctx, "agents.files.get", params)
	if err != nil {
		return protocol.AgentsFilesGetResult{}, err
	}
	var result protocol.AgentsFilesGetResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.AgentsFilesGetResult{}, err
	}
	return result, nil
}

// Set sets an agent file.
func (f *AgentsFilesAPI) Set(ctx context.Context, params protocol.AgentsFilesSetParams) (protocol.AgentsFilesSetResult, error) {
	raw, err := f.request(ctx, "agents.files.set", params)
	if err != nil {
		return protocol.AgentsFilesSetResult{}, err
	}
	var result protocol.AgentsFilesSetResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.AgentsFilesSetResult{}, err
	}
	return result, nil
}
