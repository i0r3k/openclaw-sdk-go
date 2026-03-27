// Package api provides Secrets API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"
)

// SecretsAPI provides access to Secrets API methods.
type SecretsAPI struct {
	request RequestFn
}

// NewSecretsAPI creates a new SecretsAPI instance.
func NewSecretsAPI(request RequestFn) *SecretsAPI {
	return &SecretsAPI{request: request}
}

// Reload reloads secrets from the secret store.
func (api *SecretsAPI) Reload(ctx context.Context) error {
	_, err := api.request(ctx, "secrets.reload", struct{}{})
	return err
}

// Resolve resolves a secret reference.
func (api *SecretsAPI) Resolve(ctx context.Context, params struct {
	Ref string `json:"ref"`
}) (map[string]any, error) {
	raw, err := api.request(ctx, "secrets.resolve", params)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}
