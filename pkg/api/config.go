// Package api provides Config API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// ConfigAPI provides access to Config API methods.
type ConfigAPI struct {
	request RequestFn
}

// NewConfigAPI creates a new ConfigAPI instance.
func NewConfigAPI(request RequestFn) *ConfigAPI {
	return &ConfigAPI{request: request}
}

// Get returns config value(s).
func (api *ConfigAPI) Get(ctx context.Context, params protocol.ConfigGetParams) (json.RawMessage, error) {
	return api.request(ctx, "config.get", params)
}

// Set sets a config value.
func (api *ConfigAPI) Set(ctx context.Context, params protocol.ConfigSetParams) error {
	_, err := api.request(ctx, "config.set", params)
	return err
}

// Apply applies config changes.
func (api *ConfigAPI) Apply(ctx context.Context) error {
	_, err := api.request(ctx, "config.apply", protocol.ConfigApplyParams{})
	return err
}

// Patch patches config values.
func (api *ConfigAPI) Patch(ctx context.Context, params protocol.ConfigPatchParams) error {
	_, err := api.request(ctx, "config.patch", params)
	return err
}

// Schema returns config schema.
func (api *ConfigAPI) Schema(ctx context.Context, params protocol.ConfigSchemaParams) (protocol.ConfigSchemaResponse, error) {
	raw, err := api.request(ctx, "config.schema", params)
	if err != nil {
		return protocol.ConfigSchemaResponse{}, err
	}
	var result protocol.ConfigSchemaResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.ConfigSchemaResponse{}, err
	}
	return result, nil
}

// SchemaLookup looks up schema for a specific config key.
func (api *ConfigAPI) SchemaLookup(ctx context.Context, params protocol.ConfigSchemaLookupParams) (protocol.ConfigSchemaLookupResult, error) {
	raw, err := api.request(ctx, "config.schema.lookup", params)
	if err != nil {
		return protocol.ConfigSchemaLookupResult{}, err
	}
	var result protocol.ConfigSchemaLookupResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.ConfigSchemaLookupResult{}, err
	}
	return result, nil
}
