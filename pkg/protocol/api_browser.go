// Package protocol provides API parameter types for OpenClaw SDK.
//
// This file contains Browser types migrated from TypeScript: src/protocol/api-params.ts
package protocol

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
