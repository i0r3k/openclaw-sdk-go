// Package api provides Browser API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// BrowserAPI provides access to Browser API methods.
type BrowserAPI struct {
	request RequestFn
}

// NewBrowserAPI creates a new BrowserAPI instance.
func NewBrowserAPI(request RequestFn) *BrowserAPI {
	return &BrowserAPI{request: request}
}

// Open opens a browser tab.
func (api *BrowserAPI) Open(ctx context.Context, params protocol.BrowserOpenParams) (protocol.BrowserOpenResult, error) {
	raw, err := api.request(ctx, "browser.open", params)
	if err != nil {
		return protocol.BrowserOpenResult{}, err
	}
	var result protocol.BrowserOpenResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.BrowserOpenResult{}, err
	}
	return result, nil
}

// List lists browser tabs.
func (api *BrowserAPI) List(ctx context.Context) (protocol.BrowserListResult, error) {
	raw, err := api.request(ctx, "browser.list", protocol.BrowserListParams{})
	if err != nil {
		return protocol.BrowserListResult{}, err
	}
	var result protocol.BrowserListResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.BrowserListResult{}, err
	}
	return result, nil
}

// Screenshot takes a screenshot of a browser tab.
func (api *BrowserAPI) Screenshot(ctx context.Context, params protocol.BrowserScreenshotParams) (protocol.BrowserScreenshotResult, error) {
	raw, err := api.request(ctx, "browser.screenshot", params)
	if err != nil {
		return protocol.BrowserScreenshotResult{}, err
	}
	var result protocol.BrowserScreenshotResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.BrowserScreenshotResult{}, err
	}
	return result, nil
}

// Eval evaluates a script in a browser tab.
func (api *BrowserAPI) Eval(ctx context.Context, params protocol.BrowserEvalParams) (protocol.BrowserEvalResult, error) {
	raw, err := api.request(ctx, "browser.eval", params)
	if err != nil {
		return protocol.BrowserEvalResult{}, err
	}
	var result protocol.BrowserEvalResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.BrowserEvalResult{}, err
	}
	return result, nil
}
