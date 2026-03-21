// Package api provides Cron API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// CronAPI provides access to Cron API methods.
type CronAPI struct {
	request RequestFn
}

// NewCronAPI creates a new CronAPI instance.
func NewCronAPI(request RequestFn) *CronAPI {
	return &CronAPI{request: request}
}

// List returns all cron jobs.
func (api *CronAPI) List(ctx context.Context) ([]protocol.CronJob, error) {
	raw, err := api.request(ctx, "cron.list", protocol.CronListParams{})
	if err != nil {
		return nil, err
	}
	var result []protocol.CronJob
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Status returns cron job status.
func (api *CronAPI) Status(ctx context.Context, params protocol.CronStatusParams) (protocol.CronJob, error) {
	raw, err := api.request(ctx, "cron.status", params)
	if err != nil {
		return protocol.CronJob{}, err
	}
	var result protocol.CronJob
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.CronJob{}, err
	}
	return result, nil
}

// Add adds a new cron job.
func (api *CronAPI) Add(ctx context.Context, params protocol.CronAddParams) error {
	_, err := api.request(ctx, "cron.add", params)
	return err
}

// Update updates a cron job.
func (api *CronAPI) Update(ctx context.Context, params protocol.CronUpdateParams) error {
	_, err := api.request(ctx, "cron.update", params)
	return err
}

// Remove removes a cron job.
func (api *CronAPI) Remove(ctx context.Context, params protocol.CronRemoveParams) error {
	_, err := api.request(ctx, "cron.remove", params)
	return err
}

// Run runs a cron job immediately.
func (api *CronAPI) Run(ctx context.Context, params protocol.CronRunParams) error {
	_, err := api.request(ctx, "cron.run", params)
	return err
}

// Runs returns cron run history.
func (api *CronAPI) Runs(ctx context.Context, params protocol.CronRunsParams) ([]protocol.CronRunLogEntry, error) {
	raw, err := api.request(ctx, "cron.runs", params)
	if err != nil {
		return nil, err
	}
	var result []protocol.CronRunLogEntry
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}
