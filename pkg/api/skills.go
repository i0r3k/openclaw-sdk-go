// Package api provides Skills API client for OpenClaw SDK.
package api

import (
	"context"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// SkillsAPI provides access to Skills API methods.
type SkillsAPI struct {
	request RequestFn
}

// NewSkillsAPI creates a new SkillsAPI instance.
func NewSkillsAPI(request RequestFn) *SkillsAPI {
	return &SkillsAPI{request: request}
}

// Status returns skills status.
func (api *SkillsAPI) Status(ctx context.Context, params protocol.SkillsStatusParams) (any, error) {
	return api.request(ctx, "skills.status", params)
}

// ToolsCatalog returns the tools catalog.
func (api *SkillsAPI) ToolsCatalog(ctx context.Context) (protocol.ToolsCatalogResult, error) {
	result, err := api.request(ctx, "skills.toolsCatalog", protocol.ToolsCatalogParams{})
	if err != nil {
		return protocol.ToolsCatalogResult{}, err
	}
	return result.(protocol.ToolsCatalogResult), nil
}

// Bins returns skills bins.
func (api *SkillsAPI) Bins(ctx context.Context) (protocol.SkillsBinsResult, error) {
	result, err := api.request(ctx, "skills.bins", protocol.SkillsBinsParams{})
	if err != nil {
		return protocol.SkillsBinsResult{}, err
	}
	return result.(protocol.SkillsBinsResult), nil
}

// Install installs a skill.
func (api *SkillsAPI) Install(ctx context.Context, params protocol.SkillsInstallParams) error {
	_, err := api.request(ctx, "skills.install", params)
	return err
}

// Update updates a skill.
func (api *SkillsAPI) Update(ctx context.Context, params protocol.SkillsUpdateParams) error {
	_, err := api.request(ctx, "skills.update", params)
	return err
}
