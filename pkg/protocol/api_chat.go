// Package protocol provides API parameter types for OpenClaw SDK.
//
// This file contains Chat types migrated from TypeScript: src/protocol/api-params.ts
package protocol

// ============================================================================
// Chat Extended Types
// ============================================================================

// ChatListParams parameters for listing chats.
type ChatListParams struct{}

// ChatHistoryParams parameters for getting chat history.
type ChatHistoryParams struct {
	ChatID string `json:"chatId"`
	Limit  int64  `json:"limit,omitempty"`
	Before string `json:"before,omitempty"`
}

// ChatHistoryResult result of getting chat history.
type ChatHistoryResult struct {
	Messages []any `json:"messages"`
}

// ChatDeleteParams parameters for deleting a chat.
type ChatDeleteParams struct {
	ChatID string `json:"chatId"`
}

// ChatDeleteResult result of deleting a chat.
type ChatDeleteResult struct{}

// ChatTitleParams parameters for getting chat title.
type ChatTitleParams struct {
	ChatID string `json:"chatId"`
}

// ChatTitleResult result of getting chat title.
type ChatTitleResult struct {
	Title string `json:"title"`
}

// ChatAbortParams parameters for aborting a chat.
type ChatAbortParams struct {
	ChatID string `json:"chatId"`
}

// ChatSendParams parameters for sending a chat message.
type ChatSendParams struct {
	ChatID string `json:"chatId"`
}
