// Package protocol provides API parameter and common types for OpenClaw SDK.
//
// This package contains Go implementations of TypeScript interfaces from:
// - src/protocol/api-common.ts (shared types)
// - src/protocol/api-params.ts (all XxxParams and XxxResult types)
package protocol

// ============================================================================
// Common API Types
// ============================================================================

// AgentSummary represents agent summary information returned from identity verification.
type AgentSummary = map[string]any

// WizardStep represents a step in a wizard flow.
type WizardStep struct {
	ID     string         `json:"id"`
	Prompt string         `json:"prompt,omitempty"`
	Extra  map[string]any `json:"*"` // Allows additional properties
}

// CronJob represents a scheduled cron job.
type CronJob struct {
	ID     string         `json:"id"`
	Cron   string         `json:"cron"`
	Prompt string         `json:"prompt"`
	Extra  map[string]any `json:"*"` // Allows additional properties
}

// CronRunLogEntry represents an entry in a cron run log.
type CronRunLogEntry struct {
	Timestamp int64          `json:"timestamp"`
	Extra     map[string]any `json:"*"` // Allows additional properties
}

// TtsVoicesResult represents TTS voices list result.
type TtsVoicesResult struct {
	Voices []TtsVoice `json:"voices"`
}

// TtsVoice represents a TTS voice option.
type TtsVoice struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Language string         `json:"language,omitempty"`
	Extra    map[string]any `json:"*"` // Allows additional properties
}

// VoiceWakeStatusResult represents voice wake status result.
type VoiceWakeStatusResult struct {
	Active      bool    `json:"active"`
	Sensitivity float64 `json:"sensitivity,omitempty"`
}

// BrowserListResult represents browser tabs list result.
type BrowserListResult struct {
	Tabs []BrowserTab `json:"tabs"`
}

// BrowserTab represents a browser tab.
type BrowserTab struct {
	TabID string `json:"tabId"`
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

// DoctorCheckResult represents doctor check result.
type DoctorCheckResult struct {
	Checks []DoctorCheck `json:"checks"`
}

// DoctorCheck represents a single doctor check item.
type DoctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// SecretsListResult represents secrets list result.
type SecretsListResult struct {
	Keys []string `json:"keys"`
}

// ChatListResult represents chat list result.
type ChatListResult struct {
	Chats []ChatInfo `json:"chats"`
}

// ChatInfo represents a chat in the list.
type ChatInfo struct {
	ChatID string         `json:"chatId"`
	Extra  map[string]any `json:"*"` // Allows additional properties
}
