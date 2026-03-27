// Package api provides Chat API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// ChatAPI provides access to Chat API methods.
type ChatAPI struct {
	request RequestFn
}

// NewChatAPI creates a new ChatAPI instance.
func NewChatAPI(request RequestFn) *ChatAPI {
	return &ChatAPI{request: request}
}

// Inject injects a message into a chat.
func (api *ChatAPI) Inject(ctx context.Context, params protocol.ChatInjectParams) error {
	_, err := api.request(ctx, "chat.inject", params)
	return err
}

// List returns all chats.
func (api *ChatAPI) List(ctx context.Context) (protocol.ChatListResult, error) {
	raw, err := api.request(ctx, "chat.list", protocol.ChatListParams{})
	if err != nil {
		return protocol.ChatListResult{}, err
	}
	var result protocol.ChatListResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.ChatListResult{}, err
	}
	return result, nil
}

// History returns chat history.
func (api *ChatAPI) History(ctx context.Context, params protocol.ChatHistoryParams) (protocol.ChatHistoryResult, error) {
	raw, err := api.request(ctx, "chat.history", params)
	if err != nil {
		return protocol.ChatHistoryResult{}, err
	}
	var result protocol.ChatHistoryResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.ChatHistoryResult{}, err
	}
	return result, nil
}

// Delete deletes a chat.
func (api *ChatAPI) Delete(ctx context.Context, params protocol.ChatDeleteParams) error {
	_, err := api.request(ctx, "chat.delete", params)
	return err
}

// Title returns the title of a chat.
func (api *ChatAPI) Title(ctx context.Context, params protocol.ChatTitleParams) (protocol.ChatTitleResult, error) {
	raw, err := api.request(ctx, "chat.title", params)
	if err != nil {
		return protocol.ChatTitleResult{}, err
	}
	var result protocol.ChatTitleResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.ChatTitleResult{}, err
	}
	return result, nil
}

// Abort aborts a chat.
func (api *ChatAPI) Abort(ctx context.Context, params protocol.ChatAbortParams) error {
	_, err := api.request(ctx, "chat.abort", params)
	return err
}

// Send sends a message to a chat.
func (api *ChatAPI) Send(ctx context.Context, params protocol.ChatSendParams) error {
	_, err := api.request(ctx, "chat.send", params)
	return err
}
