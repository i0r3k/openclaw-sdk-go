// Package api provides Channels API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// ChannelsAPI provides access to Channels API methods.
type ChannelsAPI struct {
	request RequestFn
}

// NewChannelsAPI creates a new ChannelsAPI instance.
func NewChannelsAPI(request RequestFn) *ChannelsAPI {
	return &ChannelsAPI{request: request}
}

// Status returns the status of all channels.
func (api *ChannelsAPI) Status(ctx context.Context) (protocol.ChannelsStatusResult, error) {
	raw, err := api.request(ctx, "channels.status", protocol.ChannelsStatusParams{})
	if err != nil {
		return protocol.ChannelsStatusResult{}, err
	}
	var result protocol.ChannelsStatusResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.ChannelsStatusResult{}, err
	}
	return result, nil
}

// Logout logs out of a channel.
func (api *ChannelsAPI) Logout(ctx context.Context, params protocol.ChannelsLogoutParams) error {
	_, err := api.request(ctx, "channels.logout", params)
	return err
}

// TalkConfig returns talk configuration.
func (api *ChannelsAPI) TalkConfig(ctx context.Context) (protocol.TalkConfigResult, error) {
	raw, err := api.request(ctx, "talk.config", protocol.TalkConfigParams{})
	if err != nil {
		return protocol.TalkConfigResult{}, err
	}
	var result protocol.TalkConfigResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.TalkConfigResult{}, err
	}
	return result, nil
}

// TalkMode sets the talk mode.
func (api *ChannelsAPI) TalkMode(ctx context.Context, params protocol.TalkModeParams) error {
	_, err := api.request(ctx, "talk.mode", params)
	return err
}

// TalkSpeak speaks a message.
func (api *ChannelsAPI) TalkSpeak(ctx context.Context, params struct {
	Message  string `json:"message"`
	Language string `json:"language,omitempty"`
}) error {
	_, err := api.request(ctx, "talk.speak", params)
	return err
}

// TalkStart starts a talk session.
func (api *ChannelsAPI) TalkStart(ctx context.Context, params protocol.TalkStartParams) (protocol.TalkStartResult, error) {
	raw, err := api.request(ctx, "talk.start", params)
	if err != nil {
		return protocol.TalkStartResult{}, err
	}
	var result protocol.TalkStartResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.TalkStartResult{}, err
	}
	return result, nil
}

// TalkStop stops a talk session.
func (api *ChannelsAPI) TalkStop(ctx context.Context, params protocol.TalkStopParams) (protocol.TalkStopResult, error) {
	raw, err := api.request(ctx, "talk.stop", params)
	if err != nil {
		return protocol.TalkStopResult{}, err
	}
	var result protocol.TalkStopResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.TalkStopResult{}, err
	}
	return result, nil
}
