package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/application/config/factories"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TODO: CRITICAL - Continue migrating remaining functions to domain-specific factory files:
// - createWhitespacedConfig() -> sanitization_factory.go
// - GetStandardTestCases() -> validation_factory.go
// - SanitizationTestCase -> validation_factory.go
// - createWhitespacedConfigForSanitizer() -> sanitization_factory.go
// - createWhitespacedConfigWithOptions() -> sanitization_factory.go

// CommonTestConfiguration represents standard test configuration structure
type CommonTestConfiguration struct {
	name             string
	config           *domain.Config
	expectedChanges  []string
	expectedWarnings int
}

// SanitizationTestCase defines a single sanitization test case (wrapper for backward compatibility)
type SanitizationTestCase struct {
	name             string
	config           *domain.Config
	expectedChanges  []string
	expectedWarnings int
}

// createWhitespacedConfig creates a config with formatting issues for testing
func createWhitespacedConfig() *domain.Config {
	return createWhitespacedConfigWithOptions(false)
}

// createWhitespacedConfigForSanitizer creates a config with basic formatting issues for sanitizer testing
// Eliminates duplication from validation_sanitizer_test.go
func createWhitespacedConfigForSanitizer() *domain.Config {
	return createWhitespacedConfigWithOptions(true)
}

// createWhitespacedConfigWithOptions creates a config with configurable whitespace issues
func createWhitespacedConfigWithOptions(minimal bool) *domain.Config {
	baseConfig := &domain.Config{
		SafetyLevel:  domain.SafetyLevelEnabled,
		MaxDiskUsage: 50,
		LastClean:    time.Now(),
		Updated:      time.Now(),
	}

	if minimal {
		baseConfig.Version = "  1.0.0  "
		baseConfig.Protected = []string{"/System", "/Library"}
		baseConfig.Profiles = map[string]*domain.Profile{
			"daily": {
				Name:        "  daily  ",
				Description: "Daily cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   domain.RiskLow,
						Status:      domain.StatusEnabled,
					},
				},
				Status: domain.StatusEnabled,
			},
		}
	} else {
		baseConfig.Version = " 1.0.0 "                           // Extra spaces
		baseConfig.Protected = []string{"/System ", " /Library"} // Trailing/leading spaces
		baseConfig.Profiles = map[string]*domain.Profile{
			"daily": { // Normal key - sanitizer should clean up name field inside
				Name:        " daily cleanup ", // Extra spaces
				Description: " Daily cleanup ",
				Operations: []domain.CleanupOperation{
					{
						Name:        " nix-generations ",
						Description: " Clean Nix generations ",
						RiskLevel:   domain.RiskLow,
						Status:      domain.StatusEnabled,
					},
				},
				Status: domain.StatusEnabled,
			},
		}
	}

	return baseConfig
}

// GetStandardTestCases returns common test cases for validation and sanitization
func GetStandardTestCases() []CommonTestConfiguration {
	return []CommonTestConfiguration{
		{
			name:             "whitespace cleanup",
			config:           createWhitespacedConfig(),
			expectedChanges:  []string{"version", "profiles.daily.name", "profiles.daily.description", "profiles.daily.operations[0].name", "profiles.daily.operations[0].description"},
			expectedWarnings: 0,
		},
		{
			name:             "max disk usage clamping",
			config:           factories.CreateValidationTestConfigs()["invalid_high_disk"],
			expectedChanges:  []string{"max_disk_usage"},
			expectedWarnings: 1,
		},
	}
}

// GetSanitizationTestCasesCompat returns all sanitization test cases (wrapper for backward compatibility)
func GetSanitizationTestCasesCompat() []SanitizationTestCase {
	standardCases := GetStandardTestCases()
	result := make([]SanitizationTestCase, len(standardCases))
	for i, tc := range standardCases {
		result[i] = SanitizationTestCase(tc)
	}
	return result
}
