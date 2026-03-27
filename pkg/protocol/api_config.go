// Package protocol provides API parameter types for OpenClaw SDK.
//
// This file contains Config types migrated from TypeScript: src/protocol/api-params.ts
package protocol

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

// ConfigSchemaLookupParams parameters for looking up a config schema.
type ConfigSchemaLookupParams struct {
	Key string `json:"key"`
}

// ConfigSchemaLookupResult result of looking up a config schema.
type ConfigSchemaLookupResult struct {
	Schema any `json:"schema"`
}
