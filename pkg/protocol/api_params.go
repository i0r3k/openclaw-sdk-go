// Package protocol provides API parameter types for OpenClaw SDK.
//
// This file contains all XxxParams and XxxResult types from TypeScript:
// src/protocol/api-params.ts
package protocol

// ============================================================================
// Agent Types
// ============================================================================

// AgentIdentityParams parameters for agent identity verification.
type AgentIdentityParams struct {
	AgentID string `json:"agentId"`
}

// AgentIdentityResult result of agent identity verification.
type AgentIdentityResult struct {
	ID      string        `json:"id"`
	Summary *AgentSummary `json:"summary,omitempty"`
}

// AgentWaitParams parameters for waiting on agent.
type AgentWaitParams struct {
	AgentID   string `json:"agentId"`
	TimeoutMs int64  `json:"timeoutMs,omitempty"`
}

// AgentsFileEntry represents a file entry for agent file operations.
type AgentsFileEntry struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// AgentsCreateParams parameters for creating an agent.
type AgentsCreateParams struct {
	AgentID string            `json:"agentId"`
	Files   []AgentsFileEntry `json:"files"`
}

// AgentsCreateResult result of creating an agent.
type AgentsCreateResult struct {
	AgentID string `json:"agentId"`
}

// AgentsUpdateParams parameters for updating an agent.
type AgentsUpdateParams struct {
	AgentID string            `json:"agentId"`
	Files   []AgentsFileEntry `json:"files"`
}

// AgentsUpdateResult result of updating an agent.
type AgentsUpdateResult struct {
	AgentID string `json:"agentId"`
}

// AgentsDeleteParams parameters for deleting an agent.
type AgentsDeleteParams struct {
	AgentID string `json:"agentId"`
}

// AgentsDeleteResult result of deleting an agent.
type AgentsDeleteResult struct {
	AgentID string `json:"agentId"`
}

// AgentsFilesListParams parameters for listing agent files.
type AgentsFilesListParams struct {
	AgentID string `json:"agentId"`
}

// AgentsFilesListResult result of listing agent files.
type AgentsFilesListResult struct {
	Files []string `json:"files"`
}

// AgentsFilesGetParams parameters for getting an agent file.
type AgentsFilesGetParams struct {
	AgentID string `json:"agentId"`
	Path    string `json:"path"`
}

// AgentsFilesGetResult result of getting an agent file.
type AgentsFilesGetResult struct {
	Content string `json:"content"`
}

// AgentsFilesSetParams parameters for setting an agent file.
type AgentsFilesSetParams struct {
	AgentID string `json:"agentId"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

// AgentsFilesSetResult result of setting an agent file.
type AgentsFilesSetResult struct{}

// AgentsListParams parameters for listing agents.
type AgentsListParams struct{}

// AgentsListResult result of listing agents.
type AgentsListResult struct {
	Agents []AgentSummary `json:"agents"`
}

// ============================================================================
// Node Pairing Types
// ============================================================================

// NodePairRequestParams parameters for requesting node pairing.
type NodePairRequestParams struct {
	NodeID string `json:"nodeId"`
	TtlSec int64  `json:"ttlSec,omitempty"`
}

// NodePairListParams parameters for listing node pairings.
type NodePairListParams struct {
	NodeID string `json:"nodeId"`
}

// NodePairApproveParams parameters for approving node pairing.
type NodePairApproveParams struct {
	NodeID    string `json:"nodeId"`
	PairingID string `json:"pairingId"`
}

// NodePairRejectParams parameters for rejecting node pairing.
type NodePairRejectParams struct {
	NodeID    string `json:"nodeId"`
	PairingID string `json:"pairingId"`
}

// NodePairVerifyParams parameters for verifying node pairing.
type NodePairVerifyParams struct {
	NodeID    string `json:"nodeId"`
	PairingID string `json:"pairingId"`
	Code      string `json:"code"`
}

// ============================================================================
// Device Pairing Types
// ============================================================================

// DevicePairListParams parameters for listing device pairings.
type DevicePairListParams struct{}

// DevicePairApproveParams parameters for approving device pairing.
type DevicePairApproveParams struct {
	PairingID string `json:"pairingId"`
}

// DevicePairRejectParams parameters for rejecting device pairing.
type DevicePairRejectParams struct {
	PairingID string `json:"pairingId"`
}

// ============================================================================
// Config Types
// ============================================================================

// ConfigGetParams parameters for getting config.
type ConfigGetParams struct {
	Key string `json:"key,omitempty"`
}

// ConfigSetParams parameters for setting config.
type ConfigSetParams struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// ConfigApplyParams parameters for applying config.
type ConfigApplyParams struct{}

// ConfigPatchParams parameters for patching config.
type ConfigPatchParams struct {
	Patches []ConfigPatchOp `json:"patches"`
}

// ConfigPatchOp represents a config patch operation.
type ConfigPatchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

// ConfigSchemaParams parameters for getting config schema.
type ConfigSchemaParams struct {
	Key string `json:"key,omitempty"`
}

// ConfigSchemaResponse response with config schema.
type ConfigSchemaResponse struct {
	Schema any `json:"schema"`
}

// ============================================================================
// Wizard Types
// ============================================================================

// WizardStartParams parameters for starting a wizard.
type WizardStartParams struct {
	WizardID string `json:"wizardId"`
	Input    any    `json:"input,omitempty"`
}

// WizardNextParams parameters for wizard next step.
type WizardNextParams struct {
	WizardID string `json:"wizardId"`
	Input    any    `json:"input,omitempty"`
}

// WizardCancelParams parameters for cancelling a wizard.
type WizardCancelParams struct {
	WizardID string `json:"wizardId"`
}

// WizardStatusParams parameters for wizard status.
type WizardStatusParams struct {
	WizardID string `json:"wizardId"`
}

// WizardNextResult result of wizard next step.
type WizardNextResult struct {
	Step     WizardStep `json:"step"`
	Complete bool       `json:"complete"`
}

// WizardStartResult result of wizard start.
type WizardStartResult = WizardNextResult

// WizardStatusResult result of wizard status.
type WizardStatusResult struct {
	CurrentStep WizardStep `json:"currentStep"`
	Complete    bool       `json:"complete"`
}

// ============================================================================
// Talk Types
// ============================================================================

// TalkConfigParams parameters for talk config.
type TalkConfigParams struct{}

// TalkConfigResult result of talk config.
type TalkConfigResult struct {
	Enabled bool           `json:"enabled"`
	Extra   map[string]any `json:"*"` // Allows additional properties
}

// TalkModeParams parameters for setting talk mode.
type TalkModeParams struct {
	Enabled bool `json:"enabled"`
}

// ============================================================================
// Channels Types
// ============================================================================

// ChannelsStatusParams parameters for channels status.
type ChannelsStatusParams struct{}

// ChannelsStatusResult result of channels status.
type ChannelsStatusResult struct {
	Channels []any `json:"channels"`
}

// ChannelsLogoutParams parameters for logging out of a channel.
type ChannelsLogoutParams struct {
	ChannelID string `json:"channelId"`
}

// ============================================================================
// WebLogin Types
// ============================================================================

// WebLoginStartParams parameters for starting web login.
type WebLoginStartParams struct {
	ReturnURL string `json:"returnUrl,omitempty"`
}

// WebLoginWaitParams parameters for waiting for web login.
type WebLoginWaitParams struct {
	Token     string `json:"token"`
	TimeoutMs int64  `json:"timeoutMs,omitempty"`
}

// WebLoginStartResult result of starting web login.
type WebLoginStartResult struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

// WebLoginWaitResult result of waiting for web login.
type WebLoginWaitResult struct {
	Success bool   `json:"success"`
	UserID  string `json:"userId,omitempty"`
}

// WebLoginCancelParams parameters for cancelling web login.
type WebLoginCancelParams struct {
	Token string `json:"token"`
}

// WebLoginCancelResult result of cancelling web login.
type WebLoginCancelResult struct{}

// ============================================================================
// Skills Types
// ============================================================================

// SkillsStatusParams parameters for skills status.
type SkillsStatusParams struct {
	SkillID string `json:"skillId,omitempty"`
}

// ToolsCatalogParams parameters for tools catalog.
type ToolsCatalogParams struct{}

// ToolsCatalogResult result of tools catalog.
type ToolsCatalogResult struct {
	Tools []any `json:"tools"`
}

// SkillsBinsParams parameters for skills bins.
type SkillsBinsParams struct{}

// SkillsBinsResult result of skills bins.
type SkillsBinsResult struct {
	Bins []any `json:"bins"`
}

// SkillsInstallParams parameters for installing a skill.
type SkillsInstallParams struct {
	SkillID string `json:"skillId"`
}

// SkillsUpdateParams parameters for updating a skill.
type SkillsUpdateParams struct {
	SkillID string `json:"skillId"`
}

// ============================================================================
// Cron Types
// ============================================================================

// CronListParams parameters for listing cron jobs.
type CronListParams struct{}

// CronStatusParams parameters for cron job status.
type CronStatusParams struct {
	JobID string `json:"jobId"`
}

// CronAddParams parameters for adding a cron job.
type CronAddParams struct {
	Cron   string `json:"cron"`
	Prompt string `json:"prompt"`
}

// CronUpdateParams parameters for updating a cron job.
type CronUpdateParams struct {
	JobID  string `json:"jobId"`
	Cron   string `json:"cron,omitempty"`
	Prompt string `json:"prompt,omitempty"`
}

// CronRemoveParams parameters for removing a cron job.
type CronRemoveParams struct {
	JobID string `json:"jobId"`
}

// CronRunParams parameters for running a cron job.
type CronRunParams struct {
	JobID string `json:"jobId"`
}

// CronRunsParams parameters for listing cron runs.
type CronRunsParams struct{}

// ============================================================================
// Logs Types
// ============================================================================

// LogsTailParams parameters for tailing logs.
type LogsTailParams struct {
	Lines int64 `json:"lines,omitempty"`
}

// LogsTailResult result of tailing logs.
type LogsTailResult struct {
	Logs []string `json:"logs"`
}

// ============================================================================
// ExecApprovals Types
// ============================================================================

// ExecApprovalsGetParams parameters for getting exec approvals.
type ExecApprovalsGetParams struct{}

// ExecApprovalsSetParams parameters for setting exec approvals.
type ExecApprovalsSetParams struct {
	Enabled bool `json:"enabled"`
}

// ExecApprovalsSnapshot represents exec approvals snapshot.
type ExecApprovalsSnapshot struct {
	Approvals []any `json:"approvals"`
}

// ============================================================================
// Sessions Types
// ============================================================================

// SessionsListParams parameters for listing sessions.
type SessionsListParams struct{}

// SessionsListResult result of listing sessions.
type SessionsListResult struct {
	Sessions []SessionInfo `json:"sessions"`
}

// SessionInfo represents a session in the list.
type SessionInfo struct {
	ID     string         `json:"id"`
	Status string         `json:"status,omitempty"`
	Extra  map[string]any `json:"*"` // Allows additional properties
}

// SessionsPreviewParams parameters for previewing a session.
type SessionsPreviewParams struct {
	SessionID string `json:"sessionId"`
}

// SessionsPreviewResult result of previewing a session.
type SessionsPreviewResult struct {
	Preview string `json:"preview"`
}

// SessionsResolveParams parameters for resolving a session.
type SessionsResolveParams struct {
	SessionID string `json:"sessionId"`
}

// SessionsPatchParams parameters for patching a session.
type SessionsPatchParams struct {
	SessionID string `json:"sessionId"`
	Patch     any    `json:"patch"`
}

// SessionsPatchResult result of patching a session.
type SessionsPatchResult struct{}

// SessionsResetParams parameters for resetting a session.
type SessionsResetParams struct {
	SessionID string `json:"sessionId"`
}

// SessionsDeleteParams parameters for deleting a session.
type SessionsDeleteParams struct {
	SessionID string `json:"sessionId"`
}

// SessionsCompactParams parameters for compacting sessions.
type SessionsCompactParams struct{}

// SessionsUsageParams parameters for session usage.
type SessionsUsageParams struct{}

// SessionsUsageResult result of session usage.
type SessionsUsageResult struct {
	Usage map[string]any `json:"usage"`
}

// ============================================================================
// Node Types
// ============================================================================

// NodeListParams parameters for listing nodes.
type NodeListParams struct{}

// NodeInvokeParams parameters for invoking a node.
type NodeInvokeParams struct {
	NodeID string `json:"nodeId"`
	Target string `json:"target"`
	Params any    `json:"params,omitempty"`
}

// NodeInvokeResultParams represents node invoke result parameters.
type NodeInvokeResultParams struct {
	Result any `json:"result"`
}

// NodeEventParams parameters for node event.
type NodeEventParams struct {
	NodeID  string `json:"nodeId"`
	Event   string `json:"event"`
	Payload any    `json:"payload,omitempty"`
}

// NodePendingDrainParams parameters for draining pending node items.
type NodePendingDrainParams struct {
	NodeID string `json:"nodeId"`
	Max    int64  `json:"max,omitempty"`
}

// NodePendingDrainResult result of draining pending node items.
type NodePendingDrainResult struct {
	Items []any `json:"items"`
}

// NodePendingEnqueueParams parameters for enqueueing a pending node item.
type NodePendingEnqueueParams struct {
	NodeID string `json:"nodeId"`
	Item   any    `json:"item"`
}

// NodePendingEnqueueResult result of enqueueing a pending node item.
type NodePendingEnqueueResult struct{}

// ============================================================================
// Poll / Update / ChatInject Types
// ============================================================================

// PollParams parameters for polling.
type PollParams struct{}

// UpdateRunParams parameters for running update.
type UpdateRunParams struct{}

// ChatInjectParams parameters for injecting chat.
type ChatInjectParams struct {
	ChatID  string `json:"chatId"`
	Message any    `json:"message"`
}

// ============================================================================
// TTS Types
// ============================================================================

// TtsSpeakParams parameters for TTS speak.
type TtsSpeakParams struct {
	Text     string  `json:"text"`
	Voice    string  `json:"voice,omitempty"`
	Language string  `json:"language,omitempty"`
	Speed    float64 `json:"speed,omitempty"`
}

// TtsSpeakResult result of TTS speak.
type TtsSpeakResult struct {
	AudioURL   string `json:"audioUrl,omitempty"`
	DurationMs int64  `json:"durationMs,omitempty"`
}

// TtsVoicesParams parameters for TTS voices.
type TtsVoicesParams struct{}

// ============================================================================
// Voice Wake Types
// ============================================================================

// VoiceWakeStartParams parameters for starting voice wake.
type VoiceWakeStartParams struct {
	Sensitivity float64  `json:"sensitivity,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// VoiceWakeStopParams parameters for stopping voice wake.
type VoiceWakeStopParams struct{}

// VoiceWakeStatusParams parameters for voice wake status.
type VoiceWakeStatusParams struct{}

// ============================================================================
// Browser Types
// ============================================================================

// BrowserOpenParams parameters for opening a browser.
type BrowserOpenParams struct {
	URL    string `json:"url"`
	NodeID string `json:"nodeId,omitempty"`
}

// BrowserOpenResult result of opening a browser.
type BrowserOpenResult struct {
	TabID string `json:"tabId"`
}

// BrowserCloseParams parameters for closing a browser tab.
type BrowserCloseParams struct {
	TabID string `json:"tabId"`
}

// BrowserCloseResult result of closing a browser tab.
type BrowserCloseResult struct{}

// BrowserListParams parameters for listing browser tabs.
type BrowserListParams struct{}

// BrowserScreenshotParams parameters for taking browser screenshot.
type BrowserScreenshotParams struct {
	TabID string `json:"tabId"`
}

// BrowserScreenshotResult result of taking browser screenshot.
type BrowserScreenshotResult struct {
	ImageURL string `json:"imageUrl"`
}

// BrowserEvalParams parameters for evaluating script in browser.
type BrowserEvalParams struct {
	TabID  string `json:"tabId"`
	Script string `json:"script"`
}

// BrowserEvalResult result of evaluating script in browser.
type BrowserEvalResult struct {
	Result any `json:"result"`
}

// ============================================================================
// Push Notification Types
// ============================================================================

// PushRegisterParams parameters for registering push notification.
type PushRegisterParams struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

// PushRegisterResult result of registering push notification.
type PushRegisterResult struct{}

// PushUnregisterParams parameters for unregistering push notification.
type PushUnregisterParams struct {
	Token string `json:"token"`
}

// PushUnregisterResult result of unregistering push notification.
type PushUnregisterResult struct{}

// PushSendParams parameters for sending push notification.
type PushSendParams struct {
	Target string         `json:"target"`
	Title  string         `json:"title"`
	Body   string         `json:"body"`
	Data   map[string]any `json:"data,omitempty"`
}

// PushSendResult result of sending push notification.
type PushSendResult struct{}

// ============================================================================
// Usage / Billing Types
// ============================================================================

// UsageSummaryParams parameters for usage summary.
type UsageSummaryParams struct {
	Period string `json:"period,omitempty"`
}

// UsageSummaryResult result of usage summary.
type UsageSummaryResult struct {
	TotalTokens *int64         `json:"totalTokens,omitempty"`
	TotalCost   *float64       `json:"totalCost,omitempty"`
	Period      string         `json:"period,omitempty"`
	Extra       map[string]any `json:"*"` // Allows additional properties
}

// UsageDetailsParams parameters for usage details.
type UsageDetailsParams struct {
	Period  string `json:"period,omitempty"`
	AgentID string `json:"agentId,omitempty"`
}

// UsageDetailsResult result of usage details.
type UsageDetailsResult struct {
	Entries []any `json:"entries"`
}

// ============================================================================
// Doctor / Diagnostics Types
// ============================================================================

// DoctorCheckParams parameters for doctor check.
type DoctorCheckParams struct{}

// DoctorFixParams parameters for doctor fix.
type DoctorFixParams struct {
	CheckName string `json:"checkName,omitempty"`
}

// DoctorFixResult result of doctor fix.
type DoctorFixResult struct {
	Fixed  []string `json:"fixed"`
	Failed []string `json:"failed"`
}

// ============================================================================
// Secrets Management Types
// ============================================================================

// SecretsListParams parameters for listing secrets.
type SecretsListParams struct{}

// SecretsGetParams parameters for getting a secret.
type SecretsGetParams struct {
	Key string `json:"key"`
}

// SecretsGetResult result of getting a secret.
type SecretsGetResult struct {
	Value string `json:"value"`
}

// SecretsSetParams parameters for setting a secret.
type SecretsSetParams struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SecretsSetResult result of setting a secret.
type SecretsSetResult struct{}

// SecretsDeleteParams parameters for deleting a secret.
type SecretsDeleteParams struct {
	Key string `json:"key"`
}

// SecretsDeleteResult result of deleting a secret.
type SecretsDeleteResult struct{}

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

// ============================================================================
// Talk Extended Types
// ============================================================================

// TalkStartParams parameters for starting a talk.
type TalkStartParams struct {
	Language string `json:"language,omitempty"`
}

// TalkStartResult result of starting a talk.
type TalkStartResult struct {
	SessionID string `json:"sessionId"`
}

// TalkStopParams parameters for stopping a talk.
type TalkStopParams struct {
	SessionID string `json:"sessionId"`
}

// TalkStopResult result of stopping a talk.
type TalkStopResult struct{}

// ============================================================================
// Update Types
// ============================================================================

// UpdateCheckParams parameters for checking for updates.
type UpdateCheckParams struct{}

// UpdateCheckResult result of checking for updates.
type UpdateCheckResult struct {
	Available bool   `json:"available"`
	Version   string `json:"version,omitempty"`
	Changelog string `json:"changelog,omitempty"`
}

// UpdateApplyParams parameters for applying an update.
type UpdateApplyParams struct {
	Version string `json:"version,omitempty"`
}

// UpdateApplyResult result of applying an update.
type UpdateApplyResult struct {
	Success bool   `json:"success"`
	Version string `json:"version,omitempty"`
}

// ============================================================================
// Diagnostics Types
// ============================================================================

// DiagnosticsSnapshotParams parameters for diagnostics snapshot.
type DiagnosticsSnapshotParams struct{}

// DiagnosticsSnapshotResult result of diagnostics snapshot.
type DiagnosticsSnapshotResult struct {
	Snapshot any `json:"snapshot"`
}
