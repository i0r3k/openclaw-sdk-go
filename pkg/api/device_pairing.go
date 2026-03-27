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
	raw, err := api.request(ctx, "device.pair.list", protocol.DevicePairListParams{})
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
	_, err := api.request(ctx, "device.pair.approve", params)
	return err
}

// Reject rejects a device pairing.
func (api *DevicePairingAPI) Reject(ctx context.Context, params protocol.DevicePairRejectParams) error {
	_, err := api.request(ctx, "device.pair.reject", params)
	return err
}

// Remove removes a device pairing.
func (api *DevicePairingAPI) Remove(ctx context.Context, params protocol.DevicePairRejectParams) error {
	_, err := api.request(ctx, "device.pair.remove", params)
	return err
}

// TokenRotate rotates a device token.
func (api *DevicePairingAPI) TokenRotate(ctx context.Context, params protocol.DeviceTokenRotateParams) error {
	_, err := api.request(ctx, "device.token.rotate", params)
	return err
}

// TokenRevoke revokes a device token.
func (api *DevicePairingAPI) TokenRevoke(ctx context.Context, params protocol.DeviceTokenRevokeParams) error {
	_, err := api.request(ctx, "device.token.revoke", params)
	return err
}

// DevicePairingInfo represents information about a device pairing.
type DevicePairingInfo struct {
	PairingID string `json:"pairingId"`
	DeviceID  string `json:"deviceId"`
	Status    string `json:"status,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}
