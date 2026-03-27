// Package api provides Usage API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"
)

// UsageAPI provides access to Usage API methods.
type UsageAPI struct {
	request RequestFn
}

// NewUsageAPI creates a new UsageAPI instance.
func NewUsageAPI(request RequestFn) *UsageAPI {
	return &UsageAPI{request: request}
}

// Status returns usage status.
func (api *UsageAPI) Status(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "usage.status", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Cost returns usage cost information.
func (api *UsageAPI) Cost(ctx context.Context, params struct {
	Period string `json:"period,omitempty"`
}) (map[string]any, error) {
	raw, err := api.request(ctx, "usage.cost", params)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}
