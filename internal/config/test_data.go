package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// CommonTestConfiguration represents standard test configuration structure
type CommonTestConfiguration struct {
	name             string
	config           *domain.Config
	expectedChanges  []string
	expectedWarnings int
}

// createWhitespacedConfig creates a config with formatting issues for testing
func createWhitespacedConfig() *domain.Config {
	return &domain.Config{
		Version:     " 1.0.0 ", // Extra spaces
		SafetyLevel: domain.SafetyLevelEnabled, // Will be tested with whitespace parsing
		MaxDiskUsage: 50,
		Protected:   []string{"/System ", " /Library"}, // Trailing/leading spaces
		Profiles: map[string]*domain.Profile{
			" daily": { // Leading space
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
		},
		LastClean: time.Now(),
		Updated:   time.Now(),
	}
}

// GetStandardTestCases returns common test cases for validation and sanitization
func GetStandardTestCases() []CommonTestConfiguration {
	return []CommonTestConfiguration{
		{
			name:             "whitespace cleanup",
			config:           createWhitespacedConfig(),
			expectedChanges:  []string{"version", "profiles.daily.name"},
			expectedWarnings: 0,
		},
		{
			name: "max disk usage clamping",
			config: &domain.Config{
				Version:      "1.0.0",
				SafetyLevel:  domain.SafetyLevelEnabled,
				MaxDiskUsage: 150, // Will be clamped to 95
				Protected:    []string{"/System", "/Library"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
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
				},
			},
			expectedChanges:  []string{"max_disk_usage"},
			expectedWarnings: 1,
		},
	}
}

// SanitizationTestCase defines a single sanitization test case (for backward compatibility)
type SanitizationTestCase struct {
	name             string
	config           *domain.Config
	expectedChanges  []string
	expectedWarnings int
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