// Package connection provides connection types for OpenClaw SDK.
//
// This package contains connection protocol types matching TypeScript src/protocol/connection.ts.
package connection

// ============================================================================
// Connection Types
// ============================================================================

// PresenceEntry represents a presence entry in the snapshot.
type PresenceEntry struct {
	Node  string         `json:"node"`
	Extra map[string]any `json:"*"`
}

// Snapshot represents the server state snapshot received after authentication.
type Snapshot struct {
	Presence     map[string]any `json:"presence,omitempty"`
	Health       map[string]any `json:"health,omitempty"`
	StateVersion int64          `json:"stateVersion"`
	UptimeMs     int64          `json:"uptimeMs"`
	ConfigPath   *string        `json:"configPath,omitempty"`
	AuthMode     *string        `json:"authMode,omitempty"`
	Agents       map[string]any `json:"agents,omitempty"`
	Nodes        map[string]any `json:"nodes,omitempty"`
}

// ConnectParamsClient represents client information in ConnectParams.
type ConnectParamsClient struct {
	ID              string  `json:"id"`
	DisplayName     *string `json:"displayName,omitempty"`
	Version         string  `json:"version"`
	Platform        string  `json:"platform"`
	DeviceFamily    *string `json:"deviceFamily,omitempty"`
	ModelIdentifier *string `json:"modelIdentifier,omitempty"`
	Mode            string  `json:"mode"`
	InstanceID      *string `json:"instanceId,omitempty"`
}

// ConnectParamsDevice represents device authentication info.
type ConnectParamsDevice struct {
	ID        string `json:"id"`
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"`
	SignedAt  int64  `json:"signedAt"`
	Nonce     string `json:"nonce"`
}

// ConnectParamsAuth represents authentication info.
type ConnectParamsAuth struct {
	Token          *string `json:"token,omitempty"`
	BootstrapToken *string `json:"bootstrapToken,omitempty"`
	DeviceToken    *string `json:"deviceToken,omitempty"`
	Password       *string `json:"password,omitempty"`
}

// ConnectParams represents connection parameters sent to the server.
type ConnectParams struct {
	MinProtocol int                  `json:"minProtocol"`
	MaxProtocol int                  `json:"maxProtocol"`
	Client      ConnectParamsClient  `json:"client"`
	Caps        []string             `json:"caps,omitempty"`
	Commands    []string             `json:"commands,omitempty"`
	Permissions map[string]bool      `json:"permissions,omitempty"`
	PathEnv     *string              `json:"pathEnv,omitempty"`
	Role        *string              `json:"role,omitempty"`
	Scopes      []string             `json:"scopes,omitempty"`
	Device      *ConnectParamsDevice `json:"device,omitempty"`
	Auth        *ConnectParamsAuth   `json:"auth,omitempty"`
	Locale      *string              `json:"locale,omitempty"`
	UserAgent   *string              `json:"userAgent,omitempty"`
}

// HelloOkServer represents server info in HelloOk.
type HelloOkServer struct {
	Version string `json:"version"`
	ConnID  string `json:"connId"`
}

// HelloOkFeatures represents server capabilities.
type HelloOkFeatures struct {
	Methods []string `json:"methods"`
	Events  []string `json:"events"`
}

// HelloOkAuth represents authentication info from server.
type HelloOkAuth struct {
	DeviceToken string   `json:"deviceToken"`
	Role        string   `json:"role"`
	Scopes      []string `json:"scopes"`
	IssuedAtMs  *int64   `json:"issuedAtMs,omitempty"`
}

// Policy represents server policy settings.
type Policy struct {
	MaxPayload       int64 `json:"maxPayload"`
	MaxBufferedBytes int64 `json:"maxBufferedBytes"`
	TickIntervalMs   int64 `json:"tickIntervalMs"`
}

// HelloOk represents the successful connection response from the server.
type HelloOk struct {
	Type          string          `json:"type"`
	Protocol      int             `json:"protocol"`
	Server        HelloOkServer   `json:"server"`
	Features      HelloOkFeatures `json:"features"`
	Snapshot      Snapshot        `json:"snapshot"`
	CanvasHostUrl *string         `json:"canvasHostUrl,omitempty"`
	Auth          *HelloOkAuth    `json:"auth,omitempty"`
	Policy        Policy          `json:"policy"`
}

// ============================================================================
// Default Values
// ============================================================================

// DefaultPolicy returns the default server policy.
func DefaultPolicy() Policy {
	return Policy{
		MaxPayload:       1048576, // 1MB
		MaxBufferedBytes: 65536,   // 64KB
		TickIntervalMs:   30000,   // 30 seconds
	}
}

// ProtocolVersionRange represents supported protocol version range.
type ProtocolVersionRange struct {
	Min int
	Max int
}

// DefaultProtocolVersionRange returns the default protocol version range.
func DefaultProtocolVersionRange() ProtocolVersionRange {
	return ProtocolVersionRange{Min: 3, Max: 3}
}
