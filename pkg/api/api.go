// Package api provides API namespace clients for OpenClaw SDK.
//
// This package contains:
//   - Shared types and interfaces (shared.go)
//   - Chat API (chat.go)
//   - Agents API (chat.go)
//   - Sessions API (sessions.go)
//   - Config API (config.go)
//   - Cron API (cron.go)
//   - Nodes API (nodes.go)
//   - Skills API (skills.go)
//   - DevicePairing API (device_pairing.go)
package api

import (
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// Re-export protocol types for convenience
type (
	// Agent types
	AgentIdentityParams   = protocol.AgentIdentityParams
	AgentIdentityResult   = protocol.AgentIdentityResult
	AgentWaitParams       = protocol.AgentWaitParams
	AgentsCreateParams    = protocol.AgentsCreateParams
	AgentsCreateResult    = protocol.AgentsCreateResult
	AgentsUpdateParams    = protocol.AgentsUpdateParams
	AgentsUpdateResult    = protocol.AgentsUpdateResult
	AgentsDeleteParams    = protocol.AgentsDeleteParams
	AgentsDeleteResult    = protocol.AgentsDeleteResult
	AgentsFilesListParams = protocol.AgentsFilesListParams
	AgentsFilesListResult = protocol.AgentsFilesListResult
	AgentsFilesGetParams  = protocol.AgentsFilesGetParams
	AgentsFilesGetResult  = protocol.AgentsFilesGetResult
	AgentsFilesSetParams  = protocol.AgentsFilesSetParams
	AgentsFilesSetResult  = protocol.AgentsFilesSetResult
	AgentsListParams      = protocol.AgentsListParams
	AgentsListResult      = protocol.AgentsListResult

	// Node Pairing types
	NodePairRequestParams = protocol.NodePairRequestParams
	NodePairListParams    = protocol.NodePairListParams
	NodePairApproveParams = protocol.NodePairApproveParams
	NodePairRejectParams  = protocol.NodePairRejectParams
	NodePairVerifyParams  = protocol.NodePairVerifyParams

	// Device Pairing types
	DevicePairListParams    = protocol.DevicePairListParams
	DevicePairApproveParams = protocol.DevicePairApproveParams
	DevicePairRejectParams  = protocol.DevicePairRejectParams

	// Config types
	ConfigGetParams    = protocol.ConfigGetParams
	ConfigSetParams    = protocol.ConfigSetParams
	ConfigApplyParams  = protocol.ConfigApplyParams
	ConfigPatchParams  = protocol.ConfigPatchParams
	ConfigSchemaParams = protocol.ConfigSchemaParams

	// Wizard types
	WizardStartParams  = protocol.WizardStartParams
	WizardNextParams   = protocol.WizardNextParams
	WizardCancelParams = protocol.WizardCancelParams
	WizardStatusParams = protocol.WizardStatusParams

	// Skills types
	SkillsStatusParams  = protocol.SkillsStatusParams
	ToolsCatalogParams  = protocol.ToolsCatalogParams
	ToolsCatalogResult  = protocol.ToolsCatalogResult
	SkillsBinsParams    = protocol.SkillsBinsParams
	SkillsBinsResult    = protocol.SkillsBinsResult
	SkillsInstallParams = protocol.SkillsInstallParams
	SkillsUpdateParams  = protocol.SkillsUpdateParams

	// Cron types
	CronListParams   = protocol.CronListParams
	CronStatusParams = protocol.CronStatusParams
	CronAddParams    = protocol.CronAddParams
	CronUpdateParams = protocol.CronUpdateParams
	CronRemoveParams = protocol.CronRemoveParams
	CronRunParams    = protocol.CronRunParams
	CronRunsParams   = protocol.CronRunsParams

	// Sessions types
	SessionsListParams    = protocol.SessionsListParams
	SessionsPreviewParams = protocol.SessionsPreviewParams
	SessionsResolveParams = protocol.SessionsResolveParams
	SessionsPatchParams   = protocol.SessionsPatchParams
	SessionsResetParams   = protocol.SessionsResetParams
	SessionsDeleteParams  = protocol.SessionsDeleteParams
	SessionsCompactParams = protocol.SessionsCompactParams
	SessionsUsageParams   = protocol.SessionsUsageParams

	// Node types
	NodeListParams           = protocol.NodeListParams
	NodeInvokeParams         = protocol.NodeInvokeParams
	NodeEventParams          = protocol.NodeEventParams
	NodePendingDrainParams   = protocol.NodePendingDrainParams
	NodePendingDrainResult   = protocol.NodePendingDrainResult
	NodePendingEnqueueParams = protocol.NodePendingEnqueueParams
	NodePendingEnqueueResult = protocol.NodePendingEnqueueResult

	// Chat types
	ChatInjectParams  = protocol.ChatInjectParams
	ChatListParams    = protocol.ChatListParams
	ChatHistoryParams = protocol.ChatHistoryParams
	ChatDeleteParams  = protocol.ChatDeleteParams
	ChatTitleParams   = protocol.ChatTitleParams

	// Common types
	AgentSummary          = protocol.AgentSummary
	WizardStep            = protocol.WizardStep
	CronJob               = protocol.CronJob
	CronRunLogEntry       = protocol.CronRunLogEntry
	TtsVoicesResult       = protocol.TtsVoicesResult
	VoiceWakeStatusResult = protocol.VoiceWakeStatusResult
	BrowserListResult     = protocol.BrowserListResult
	DoctorCheckResult     = protocol.DoctorCheckResult
	SecretsListResult     = protocol.SecretsListResult
	ChatListResult        = protocol.ChatListResult
)
