// Package api provides DevicePairing API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// DevicePairingAPI provides access to Device Pairing API methods.
type DevicePairingAPI struct {
	request RequestFn
}

// NewDevicePairingAPI creates a new DevicePairingAPI instance.
func NewDevicePairingAPI(request RequestFn) *DevicePairingAPI {
	return &DevicePairingAPI{request: request}
}

// List returns all device pairings.
func (api *DevicePairingAPI) List(ctx context.Context) ([]DevicePairingInfo, error) {
	raw, err := api.request(ctx, "devicePairing.list", protocol.DevicePairListParams{})
	if err != nil {
		return nil, err
	}
	var result []DevicePairingInfo
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Approve approves a device pairing.
func (api *DevicePairingAPI) Approve(ctx context.Context, params protocol.DevicePairApproveParams) error {
	_, err := api.request(ctx, "devicePairing.approve", params)
	return err
}

// Reject rejects a device pairing.
func (api *DevicePairingAPI) Reject(ctx context.Context, params protocol.DevicePairRejectParams) error {
	_, err := api.request(ctx, "devicePairing.reject", params)
	return err
}

// DevicePairingInfo represents information about a device pairing.
type DevicePairingInfo struct {
	PairingID string `json:"pairingId"`
	DeviceID  string `json:"deviceId"`
	Status    string `json:"status,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}
