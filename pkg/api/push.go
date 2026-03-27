// Package api provides Push API client for OpenClaw SDK.
package api

import (
	"context"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// PushAPI provides access to Push API methods.
type PushAPI struct {
	request RequestFn
}

// NewPushAPI creates a new PushAPI instance.
func NewPushAPI(request RequestFn) *PushAPI {
	return &PushAPI{request: request}
}

// Register registers a device for push notifications.
func (api *PushAPI) Register(ctx context.Context, params protocol.PushRegisterParams) (protocol.PushRegisterResult, error) {
	raw, err := api.request(ctx, "push.register", params)
	if err != nil {
		return protocol.PushRegisterResult{}, err
	}
	var result protocol.PushRegisterResult
	// Result is typically empty {}, just return empty struct
	_ = raw
	return result, nil
}

// Unregister unregisters a device from push notifications.
func (api *PushAPI) Unregister(ctx context.Context, params protocol.PushUnregisterParams) (protocol.PushUnregisterResult, error) {
	raw, err := api.request(ctx, "push.unregister", params)
	if err != nil {
		return protocol.PushUnregisterResult{}, err
	}
	var result protocol.PushUnregisterResult
	_ = raw
	return result, nil
}

// Send sends a push notification.
func (api *PushAPI) Send(ctx context.Context, params protocol.PushSendParams) (protocol.PushSendResult, error) {
	raw, err := api.request(ctx, "push.send", params)
	if err != nil {
		return protocol.PushSendResult{}, err
	}
	var result protocol.PushSendResult
	_ = raw
	return result, nil
}
