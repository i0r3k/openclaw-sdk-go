// Package api provides System API client for OpenClaw SDK.
package api

import (
	"context"
	"encoding/json"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// SystemAPI provides access to System API methods.
type SystemAPI struct {
	request RequestFn
}

// NewSystemAPI creates a new SystemAPI instance.
func NewSystemAPI(request RequestFn) *SystemAPI {
	return &SystemAPI{request: request}
}

// Health returns system health status.
func (api *SystemAPI) Health(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.health", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Status returns system status.
func (api *SystemAPI) Status(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.status", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DoctorMemoryStatus returns memory doctor status.
func (api *SystemAPI) DoctorMemoryStatus(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.doctor.memory_status", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// LogsTail tails system logs.
func (api *SystemAPI) LogsTail(ctx context.Context, params protocol.LogsTailParams) (protocol.LogsTailResult, error) {
	raw, err := api.request(ctx, "system.logs.tail", params)
	if err != nil {
		return protocol.LogsTailResult{}, err
	}
	var result protocol.LogsTailResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.LogsTailResult{}, err
	}
	return result, nil
}

// UsageStatus returns system usage status.
func (api *SystemAPI) UsageStatus(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.usage.status", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UsageCost returns system usage cost.
func (api *SystemAPI) UsageCost(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.usage.cost", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ModelsList returns available models.
func (api *SystemAPI) ModelsList(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.models.list", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateRun runs a system update.
func (api *SystemAPI) UpdateRun(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.update.run", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// VoiceWakeGet gets voice wake configuration.
func (api *SystemAPI) VoiceWakeGet(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.voice_wake.get", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// VoiceWakeSet sets voice wake configuration.
func (api *SystemAPI) VoiceWakeSet(ctx context.Context, params struct {
	Sensitivity float64  `json:"sensitivity,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}) error {
	_, err := api.request(ctx, "system.voice_wake.set", params)
	return err
}

// GatewayIdentityGet gets gateway identity.
func (api *SystemAPI) GatewayIdentityGet(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.gateway.identity.get", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SystemPresence gets system presence.
func (api *SystemAPI) SystemPresence(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.presence", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SystemEvent sends a system event.
func (api *SystemAPI) SystemEvent(ctx context.Context, params struct {
	Event string `json:"event"`
	Data  any    `json:"data,omitempty"`
}) error {
	_, err := api.request(ctx, "system.event", params)
	return err
}

// LastHeartbeat returns last heartbeat info.
func (api *SystemAPI) LastHeartbeat(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.last_heartbeat", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetHeartbeats sets heartbeat configuration.
func (api *SystemAPI) SetHeartbeats(ctx context.Context, params struct {
	Enabled bool `json:"enabled"`
}) error {
	_, err := api.request(ctx, "system.set_heartbeats", params)
	return err
}

// Wake wakes the system.
func (api *SystemAPI) Wake(ctx context.Context) error {
	_, err := api.request(ctx, "system.wake", struct{}{})
	return err
}

// Agent returns agent info.
func (api *SystemAPI) Agent(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "system.agent", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Send sends a system message.
func (api *SystemAPI) Send(ctx context.Context, params struct {
	Message string `json:"message"`
}) error {
	_, err := api.request(ctx, "system.send", params)
	return err
}

// BrowserRequest makes a browser request.
func (api *SystemAPI) BrowserRequest(ctx context.Context, params struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}) (map[string]any, error) {
	raw, err := api.request(ctx, "system.browser.request", params)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// -------------------------------------------------------------------------
// TTS Methods
// -------------------------------------------------------------------------

// Speak speaks text using TTS.
func (api *SystemAPI) Speak(ctx context.Context, params protocol.TtsSpeakParams) (protocol.TtsSpeakResult, error) {
	raw, err := api.request(ctx, "tts.speak", params)
	if err != nil {
		return protocol.TtsSpeakResult{}, err
	}
	var result protocol.TtsSpeakResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.TtsSpeakResult{}, err
	}
	return result, nil
}

// Voices returns available TTS voices.
func (api *SystemAPI) Voices(ctx context.Context) (protocol.TtsVoicesResult, error) {
	raw, err := api.request(ctx, "tts.voices", protocol.TtsVoicesParams{})
	if err != nil {
		return protocol.TtsVoicesResult{}, err
	}
	var result protocol.TtsVoicesResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.TtsVoicesResult{}, err
	}
	return result, nil
}

// TtsStatus returns TTS status.
func (api *SystemAPI) TtsStatus(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "tts.status", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TtsProviders returns available TTS providers.
func (api *SystemAPI) TtsProviders(ctx context.Context) (map[string]any, error) {
	raw, err := api.request(ctx, "tts.providers", struct{}{})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TtsEnable enables TTS.
func (api *SystemAPI) TtsEnable(ctx context.Context) error {
	_, err := api.request(ctx, "tts.enable", struct{}{})
	return err
}

// TtsDisable disables TTS.
func (api *SystemAPI) TtsDisable(ctx context.Context) error {
	_, err := api.request(ctx, "tts.disable", struct{}{})
	return err
}

// TtsConvert converts text to audio.
func (api *SystemAPI) TtsConvert(ctx context.Context, params struct {
	Text   string `json:"text"`
	Voice  string `json:"voice,omitempty"`
	Format string `json:"format,omitempty"`
}) (map[string]any, error) {
	raw, err := api.request(ctx, "tts.convert", params)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TtsSetProvider sets the TTS provider.
func (api *SystemAPI) TtsSetProvider(ctx context.Context, params struct {
	Provider string `json:"provider"`
}) error {
	_, err := api.request(ctx, "tts.set_provider", params)
	return err
}

// -------------------------------------------------------------------------
// Wizard Methods
// -------------------------------------------------------------------------

// WizardNext proceeds to the next wizard step.
func (api *SystemAPI) WizardNext(ctx context.Context, params protocol.WizardNextParams) (protocol.WizardNextResult, error) {
	raw, err := api.request(ctx, "wizard.next", params)
	if err != nil {
		return protocol.WizardNextResult{}, err
	}
	var result protocol.WizardNextResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.WizardNextResult{}, err
	}
	return result, nil
}

// WizardCancel cancels a wizard.
func (api *SystemAPI) WizardCancel(ctx context.Context, params protocol.WizardCancelParams) error {
	_, err := api.request(ctx, "wizard.cancel", params)
	return err
}

// WizardStart starts a wizard.
func (api *SystemAPI) WizardStart(ctx context.Context, params protocol.WizardStartParams) (protocol.WizardStartResult, error) {
	raw, err := api.request(ctx, "wizard.start", params)
	if err != nil {
		return protocol.WizardStartResult{}, err
	}
	var result protocol.WizardStartResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.WizardStartResult{}, err
	}
	return result, nil
}

// WizardStatus returns the status of a wizard.
func (api *SystemAPI) WizardStatus(ctx context.Context, params protocol.WizardStatusParams) (protocol.WizardStatusResult, error) {
	raw, err := api.request(ctx, "wizard.status", params)
	if err != nil {
		return protocol.WizardStatusResult{}, err
	}
	var result protocol.WizardStatusResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return protocol.WizardStatusResult{}, err
	}
	return result, nil
}
