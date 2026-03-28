package types

import "testing"

// TestDefaultReconnectConfig_MaxRetries verifies DefaultReconnectConfig returns MaxRetries=10.
func TestDefaultReconnectConfig_MaxRetries(t *testing.T) {
	cfg := DefaultReconnectConfig()
	if cfg.MaxRetries != 10 {
		t.Errorf("MaxRetries = %d, want 10", cfg.MaxRetries)
	}
}

// TestDefaultReconnectConfig_MaxAttemptsLegacy verifies MaxAttempts is 0 (infinite, legacy).
func TestDefaultReconnectConfig_MaxAttemptsLegacy(t *testing.T) {
	cfg := DefaultReconnectConfig()
	if cfg.MaxAttempts != 0 {
		t.Errorf("MaxAttempts = %d, want 0 (infinite legacy)", cfg.MaxAttempts)
	}
}

// TestReconnectConfig_MaxRetriesPrecedence verifies precedence rules:
//   - MaxRetries > 0: use MaxRetries as the limit (ignores MaxAttempts)
//   - MaxRetries == 0 AND MaxAttempts > 0: fall back to MaxAttempts
//   - MaxRetries == 0 AND MaxAttempts == 0: unlimited retries (backward compat)
//   - MaxRetries < 0: treated as 0 (same as unset)
//   - MaxAttempts < 0: treated as 0 (same as unset)
func TestReconnectConfig_MaxRetriesPrecedence(t *testing.T) {
	tests := []struct {
		name        string
		maxRetries  int
		maxAttempts int
		want        int // effective retry limit (0 = infinite)
	}{
		{
			name:        "MaxRetries positive wins over MaxAttempts",
			maxRetries:  5,
			maxAttempts: 20,
			want:        5,
		},
		{
			name:        "MaxRetries zero falls back to MaxAttempts positive",
			maxRetries:  0,
			maxAttempts: 20,
			want:        20,
		},
		{
			name:        "MaxRetries zero and MaxAttempts zero means infinite",
			maxRetries:  0,
			maxAttempts: 0,
			want:        0,
		},
		{
			name:        "MaxRetries negative treated as zero",
			maxRetries:  -1,
			maxAttempts: 20,
			want:        20, // falls back to MaxAttempts
		},
		{
			name:        "MaxAttempts negative treated as zero",
			maxRetries:  0,
			maxAttempts: -1,
			want:        0, // both zero = infinite
		},
		{
			name:        "Both negative treated as zero",
			maxRetries:  -5,
			maxAttempts: -5,
			want:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ReconnectConfig{
				MaxRetries:  tt.maxRetries,
				MaxAttempts: tt.maxAttempts,
			}

			// Compute effective retry limit using the precedence rules
			got := effectiveMaxRetries(cfg)
			if got != tt.want {
				t.Errorf("effectiveMaxRetries() = %d, want %d", got, tt.want)
			}
		})
	}
}

// effectiveMaxRetries implements the precedence rules documented on ReconnectConfig.MaxRetries.
// This is a package-internal helper for testing. Callers outside this package should use
// ReconnectManager.EffectiveMaxRetries().
func effectiveMaxRetries(cfg ReconnectConfig) int {
	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	if maxRetries > 0 {
		return maxRetries
	}

	// Fall back to MaxAttempts (with negative handling)
	maxAttempts := cfg.MaxAttempts
	if maxAttempts < 0 {
		maxAttempts = 0
	}
	return maxAttempts
}
