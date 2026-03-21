// Package api provides Sessions API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// SessionsAPI provides access to Sessions API methods.
type SessionsAPI struct {
	request RequestFn
}

// NewSessionsAPI creates a new SessionsAPI instance.
func NewSessionsAPI(request RequestFn) *SessionsAPI {
	return &SessionsAPI{request: request}
}

// List returns all sessions.
func (api *SessionsAPI) List(ctx context.Context) (protocol.SessionsListResult, error) {
	raw, err := api.request(ctx, "sessions.list", protocol.SessionsListParams{})
	if err != nil {
		return protocol.SessionsListResult{}, err
	}
	var result protocol.SessionsListResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.SessionsListResult{}, err
	}
	return result, nil
}

// Preview returns a session preview.
func (api *SessionsAPI) Preview(ctx context.Context, params protocol.SessionsPreviewParams) (protocol.SessionsPreviewResult, error) {
	raw, err := api.request(ctx, "sessions.preview", params)
	if err != nil {
		return protocol.SessionsPreviewResult{}, err
	}
	var result protocol.SessionsPreviewResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.SessionsPreviewResult{}, err
	}
	return result, nil
}

// Resolve resolves a session.
func (api *SessionsAPI) Resolve(ctx context.Context, params protocol.SessionsResolveParams) error {
	_, err := api.request(ctx, "sessions.resolve", params)
	return err
}

// Patch patches a session.
func (api *SessionsAPI) Patch(ctx context.Context, params protocol.SessionsPatchParams) error {
	_, err := api.request(ctx, "sessions.patch", params)
	return err
}

// Reset resets a session.
func (api *SessionsAPI) Reset(ctx context.Context, params protocol.SessionsResetParams) error {
	_, err := api.request(ctx, "sessions.reset", params)
	return err
}

// Delete deletes a session.
func (api *SessionsAPI) Delete(ctx context.Context, params protocol.SessionsDeleteParams) error {
	_, err := api.request(ctx, "sessions.delete", params)
	return err
}

// Compact compacts sessions.
func (api *SessionsAPI) Compact(ctx context.Context) error {
	_, err := api.request(ctx, "sessions.compact", protocol.SessionsCompactParams{})
	return err
}

// Usage returns session usage information.
func (api *SessionsAPI) Usage(ctx context.Context) (protocol.SessionsUsageResult, error) {
	raw, err := api.request(ctx, "sessions.usage", protocol.SessionsUsageParams{})
	if err != nil {
		return protocol.SessionsUsageResult{}, err
	}
	var result protocol.SessionsUsageResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.SessionsUsageResult{}, err
	}
	return result, nil
}
