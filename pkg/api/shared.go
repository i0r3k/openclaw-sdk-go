// Package api provides API namespace clients for OpenClaw SDK.
package api

import "context"

// RequestFn is the function signature for making API requests.
// Each API namespace client uses this to make requests to the server.
type RequestFn func(ctx context.Context, method string, params any) (any, error)
