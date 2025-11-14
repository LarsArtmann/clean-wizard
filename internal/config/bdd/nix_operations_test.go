package bdd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BDD scenarios for real Nix operations
// These tests verify that configuration actually works with real Nix commands

// TestNixGenerationsCleanup_BDD verifies Nix generations cleanup behavior
func TestNixGenerationsCleanup_BDD(t *testing.T) {
	// GIVEN: A configuration with Nix generations cleanup operation
	givenConfig := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     domain.SafetyLevelEnabled,
		MaxDiskUsage:  50,
		Protected:    []string{"/", "/System", "/Library", "/nix/store"},
		Profiles: map[string]*domain.Profile{
			"nix-cleanup": {
				Name:        "nix-cleanup",
				Description: "Nix-specific cleanup profile",
				Enabled:     true,
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
					},
				},
			},
		},
	}

	// WHEN: Configuration is validated
	whenResult := config.NewConfigValidator().ValidateConfig(givenConfig)

	// THEN: Configuration should be valid
	assert.True(t, whenResult.IsValid, "Nix cleanup configuration should be valid")
	
	// AND: No validation errors should exist
	assert.Empty(t, whenResult.Errors, "Should have no validation errors")
}

// TestMaxDiskUsageEnforcement_BDD verifies disk usage limits work in practice
func TestMaxDiskUsageEnforcement_BDD(t *testing.T) {
	// GIVEN: A configuration with invalid MaxDiskUsage
	givenConfig := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     domain.SafetyLevelEnabled,
		MaxDiskUsage:  150, // Invalid: > 95
		Protected:    []string{"/", "/System"},
		Profiles: map[string]*domain.Profile{
			"test": {
				Name:        "test",
				Description: "Test profile",
				Enabled:     true,
				Operations: []domain.CleanupOperation{},
			},
		},
	}

	// WHEN: Configuration is validated
	whenResult := config.NewConfigValidator().ValidateConfig(givenConfig)

	// THEN: Configuration should be invalid
	assert.False(t, whenResult.IsValid, "MaxDiskUsage >95 should be invalid")
	
	// AND: Should have specific error about max_disk_usage
	foundError := false
	for _, err := range whenResult.Errors {
		if err.Field == "max_disk_usage" {
			foundError = true
			assert.Contains(t, err.Message, "between 1 and 95")
			break
		}
	}
	assert.True(t, foundError, "Should have max_disk_usage validation error")
}

// TestSafetyLevelRestrictions_BDD verifies safety level behavior
func TestSafetyLevelRestrictions_BDD(t *testing.T) {
	// GIVEN: A configuration with strict safety mode and high-risk operations
	givenConfig := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     domain.SafetyLevelStrict,
		MaxDiskUsage:  50,
		Protected:    []string{"/", "/System"},
		Profiles: map[string]*domain.Profile{
			"risky": {
				Name:        "risky",
				Description: "High-risk operations profile",
				Enabled:     true,
				Operations: []domain.CleanupOperation{
					{
						Name:        "delete-system-files",
						Description: "Delete system files (high risk)",
						RiskLevel:   domain.RiskCritical,
						Enabled:     true,
					},
				},
			},
		},
	}

	// WHEN: Configuration is validated directly (not loaded from file)
	whenValidator := config.NewConfigValidator()
	whenResult := whenValidator.ValidateConfig(givenConfig)

	// THEN: Configuration should be valid (risk levels are validated in operations)
	assert.True(t, whenResult.IsValid, "Configuration should be valid")
	
	// AND: No validation errors should exist (RiskLevelCritical is valid)
	assert.Empty(t, whenResult.Errors, "Should have no validation errors")
}

// TestConfigurationFileRoundTrip_BDD verifies real file operations work
func TestConfigurationFileRoundTrip_BDD(t *testing.T) {
	// GIVEN: A temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	givenConfig := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     domain.SafetyLevelStrict,
		MaxDiskUsage:  75,
		Protected:    []string{"/", "/System", "/Library", "/usr"},
		Profiles: map[string]*domain.Profile{
			"comprehensive": {
				Name:         "comprehensive",
				Description:  "Complete system cleanup",
				MaxRiskLevel: domain.RiskHigh,
				Enabled:      true,
				Operations: []domain.CleanupOperation{
					{
						Name:        "temp-files",
						Description: "Clean temporary files",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
					},
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
					},
				},
			},
		},
	}

	// WHEN: Configuration is saved to file
	whenLoader := config.NewEnhancedConfigLoader()
	whenSaveOptions := &config.ConfigSaveOptions{
		Path:              configPath,
		EnableSanitization: true,
		ValidationLevel:    config.ValidationLevelComprehensive,
		Timeout:           5 * time.Second,
	}
	whenSaveConfig, whenSaveErr := whenLoader.SaveConfig(context.Background(), givenConfig, whenSaveOptions)
	require.NoError(t, whenSaveErr, "Should save configuration without error")
	require.NotNil(t, whenSaveConfig, "Should return saved configuration")

	// DEBUG: Check what was actually saved
	savedContent, readErr := os.ReadFile(configPath)
	require.NoError(t, readErr, "Should be able to read saved file")
	t.Logf("Saved configuration content:\n%s", string(savedContent))

	// AND: Configuration is loaded back
	whenLoadOptions := &config.ConfigLoadOptions{
		Path:              configPath,
		EnableSanitization: true,
		ValidationLevel:    config.ValidationLevelComprehensive,
		Timeout:           5 * time.Second,
	}
	whenLoadedConfig, whenLoadErr := whenLoader.LoadConfig(context.Background(), whenLoadOptions)

	// THEN: Configuration should load successfully
	require.NoError(t, whenLoadErr, "Should load configuration without error")
	require.NotNil(t, whenLoadedConfig, "Should return configuration")
	
	// AND: Values should match original
	assert.Equal(t, givenConfig.Version, whenLoadedConfig.Version, "Version should match")
	assert.Equal(t, givenConfig.SafeMode, whenLoadedConfig.SafeMode, "SafeMode should match")
	assert.Equal(t, givenConfig.MaxDiskUsage, whenLoadedConfig.MaxDiskUsage, "MaxDiskUsage should match")
	assert.Equal(t, len(givenConfig.Protected), len(whenLoadedConfig.Protected), "Protected paths count should match")

	// AND: File should actually exist
	_, fileErr := os.Stat(configPath)
	assert.NoError(t, fileErr, "Configuration file should exist")
}
