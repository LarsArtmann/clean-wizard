package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSafetyLevelWithBackwardCompatibility(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(*viper.Viper)
		expected         domain.SafetyLevelType
		expectedExplicit bool
	}{
		{
			name: "string enabled",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", "enabled")
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: true,
		},
		{
			name: "string disabled",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", "disabled")
			},
			expected:         domain.SafetyLevelDisabled,
			expectedExplicit: true,
		},
		{
			name: "string strict",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", "strict")
			},
			expected:         domain.SafetyLevelStrict,
			expectedExplicit: true,
		},
		{
			name: "string paranoid",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", "paranoid")
			},
			expected:         domain.SafetyLevelParanoid,
			expectedExplicit: true,
		},
		{
			name: "string with whitespace",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", "  enabled  ")
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: true,
		},
		{
			name: "string uppercase",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", "ENABLED")
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: true,
		},
		{
			name: "numeric 0",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", 0)
			},
			expected:         domain.SafetyLevelDisabled,
			expectedExplicit: true,
		},
		{
			name: "numeric 1",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", 1)
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: true,
		},
		{
			name: "numeric 2",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", 2)
			},
			expected:         domain.SafetyLevelStrict,
			expectedExplicit: true,
		},
		{
			name: "numeric 3",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", 3)
			},
			expected:         domain.SafetyLevelParanoid,
			expectedExplicit: true,
		},
		{
			name: "float64 1.0",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", 1.0)
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: true,
		},
		{
			name: "float64 1.5 (invalid float)",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", 1.5)
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: true,
		},
		{
			name: "legacy safe_mode true",
			setup: func(v *viper.Viper) {
				v.Set("safe_mode", true)
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: true,
		},
		{
			name: "legacy safe_mode false",
			setup: func(v *viper.Viper) {
				v.Set("safe_mode", false)
			},
			expected:         domain.SafetyLevelDisabled,
			expectedExplicit: true,
		},
		{
			name: "safety_level takes precedence over safe_mode",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", "strict")
				v.Set("safe_mode", false)
			},
			expected:         domain.SafetyLevelStrict,
			expectedExplicit: true,
		},
		{
			name: "invalid string defaults to enabled",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", "invalid")
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: true,
		},
		{
			name: "invalid numeric defaults to enabled",
			setup: func(v *viper.Viper) {
				v.Set("safety_level", 999)
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: true,
		},
		{
			name: "no safety config defaults to enabled",
			setup: func(v *viper.Viper) {
				// Don't set any safety-related keys
			},
			expected:         domain.SafetyLevelEnabled,
			expectedExplicit: false, // Only case where IsExplicit is false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			tt.setup(v)
			result := domain.ParseSafetyConfig(v)
			assert.Equal(t, tt.expected, result.ToSafetyLevel())
			assert.Equal(t, tt.expectedExplicit, result.IsExplicit, "IsExplicit should match expectation")
		})
	}
}

func TestLoadWithContext_SafetyLevelBackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".clean-wizard.yaml")

	tests := []struct {
		name           string
		configContent  string
		expectedSafety domain.SafetyLevelType
	}{
		{
			name: "new safety_level string format",
			configContent: `
version: "1.0.0"
safety_level: "strict"
max_disk_usage_percent: 50
protected: ["/System"]
profiles:
  test:
    name: "test"
    description: "Test profile"
    enabled: true
    operations:
      - name: "temp-files"
        description: "Clean temp files"
        risk_level: "LOW"
        enabled: true
`,
			expectedSafety: domain.SafetyLevelStrict,
		},
		{
			name: "new safety_level numeric format",
			configContent: `
version: "1.0.0"
safety_level: 2
max_disk_usage_percent: 50
protected: ["/System"]
profiles:
  test:
    name: "test"
    description: "Test profile"
    enabled: true
    operations:
      - name: "temp-files"
        description: "Clean temp files"
        risk_level: "LOW"
        enabled: true
`,
			expectedSafety: domain.SafetyLevelStrict,
		},
		{
			name: "legacy safe_mode format",
			configContent: `
version: "1.0.0"
safe_mode: false
max_disk_usage_percent: 50
protected: ["/System"]
profiles:
  test:
    name: "test"
    description: "Test profile"
    enabled: true
    operations:
      - name: "temp-files"
        description: "Clean temp files"
        risk_level: "LOW"
        enabled: true
`,
			expectedSafety: domain.SafetyLevelDisabled,
		},
		{
			name: "safety_level takes precedence",
			configContent: `
version: "1.0.0"
safety_level: "paranoid"
safe_mode: false
max_disk_usage_percent: 50
protected: ["/System"]
profiles:
  test:
    name: "test"
    description: "Test profile"
    enabled: true
    operations:
      - name: "temp-files"
        description: "Clean temp files"
        risk_level: "LOW"
        enabled: true
`,
			expectedSafety: domain.SafetyLevelParanoid,
		},
		{
			name: "default when no safety config",
			configContent: `
version: "1.0.0"
max_disk_usage_percent: 50
protected: ["/System"]
profiles:
  test:
    name: "test"
    description: "Test profile"
    enabled: true
    operations:
      - name: "temp-files"
        description: "Clean temp files"
        risk_level: "LOW"
        enabled: true
`,
			expectedSafety: domain.SafetyLevelEnabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write config file
			err := os.WriteFile(configPath, []byte(tt.configContent), 0o644)
			require.NoError(t, err)

			// Set config path environment variable
			t.Setenv("CONFIG_PATH", configPath)

			// Load config
			ctx := context.Background()
			config, err := LoadWithContext(ctx)
			require.NoError(t, err)
			require.NotNil(t, config)

			// Verify safety level
			assert.Equal(t, tt.expectedSafety, config.SafetyLevel)
		})
	}
}

func TestLoadWithContextAndPath_SafetyLevelBackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".clean-wizard.yaml")

	// Test config with safety_level
	configContent := `
version: "1.0.0"
safety_level: "paranoid"
max_disk_usage_percent: 50
protected: ["/System"]
profiles:
  test:
    name: "test"
    description: "Test profile"
    enabled: true
    operations:
      - name: "temp-files"
        description: "Clean temp files"
        risk_level: "LOW"
        enabled: true
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Load config with explicit path
	ctx := context.Background()
	config, err := LoadWithContextAndPath(ctx, configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify safety level
	assert.Equal(t, domain.SafetyLevelParanoid, config.SafetyLevel)
}
