// Package api provides API namespace clients for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"
)

// RequestFn is the function signature for making API requests.
// Each API namespace client uses this to make requests to the server.
// Returns json.RawMessage (raw response bytes) for caller to unmarshal.
type RequestFn func(ctx context.Context, method string, params any) (json.RawMessage, error)
