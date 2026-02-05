package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TestSanitizationTestCase defines a single sanitization test case.
type TestSanitizationTestCase struct {
	name             string
	config           *domain.Config
	expectedChanges  []string
	expectedWarnings int
}

// TestValidationLevelTestCase defines validation level test cases.
type TestValidationLevelTestCase struct {
	name         string
	config       *domain.Config
	level        ValidationLevel
	expectValid  bool
	expectErrors int
}

// CreateTestConfigurations creates test configurations for validation testing.
func CreateTestConfigurations() map[string]*domain.Config {
	return map[string]*domain.Config{
		"valid": {
			Version:      "1.0.0",
			SafeMode:     domain.SafeModeEnabled,
			MaxDiskUsage: 50,
			Protected:    []string{"/System", "/Library", "/Applications"},
			Profiles: map[string]*domain.Profile{
				"daily": {
					Name:        "daily",
					Description: "Daily cleanup",
					Operations: []domain.CleanupOperation{
						{
							Name:        "nix-generations",
							Description: "Clean Nix generations",
							RiskLevel:   domain.RiskLow,
							Enabled:     domain.ProfileStatusEnabled,
						},
					},
					Enabled: domain.ProfileStatusEnabled,
				},
			},
			LastClean: time.Now(),
			Updated:   time.Now(),
		},
		"invalid_high_disk": {
			Version:      "1.0.0",
			SafeMode:     domain.SafeModeEnabled,
			MaxDiskUsage: 150, // Invalid: too high
			Protected:    []string{"/System"},
			Profiles: map[string]*domain.Profile{
				"daily": {
					Name:        "daily",
					Description: "Daily cleanup",
					Operations: []domain.CleanupOperation{
						{
							Name:        "nix-generations",
							Description: "Clean Nix generations",
							RiskLevel:   domain.RiskLow,
							Enabled:     domain.ProfileStatusEnabled,
						},
					},
					Enabled: domain.ProfileStatusEnabled,
				},
			},
			LastClean: time.Now(),
			Updated:   time.Now(),
		},
	}
}

// createTestConfig creates a test configuration with customizable fields.
func createTestConfig(version string, maxDiskUsage int, profileName string, protected []string) *domain.Config {
	return &domain.Config{
		Version:      version,
		SafeMode:     domain.SafeModeEnabled,
		MaxDiskUsage: maxDiskUsage,
		Protected:    protected,
		Profiles: map[string]*domain.Profile{
			"daily": {
				Name:        profileName,
				Description: "Daily cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   domain.RiskLow,
						Enabled:     domain.ProfileStatusEnabled,
					},
				},
				Enabled: domain.ProfileStatusEnabled,
			},
		},
	}
}

// GetSanitizationTestCases returns all sanitization test cases.
func GetSanitizationTestCases() []TestSanitizationTestCase {
	return []TestSanitizationTestCase{
		{
			name:             "whitespace cleanup",
			config:           createTestConfig("  1.0.0  ", 50, "  daily  ", []string{"/System", "/Library"}),
			expectedChanges:  []string{"version", "profiles.daily.name"},
			expectedWarnings: 0,
		},
		{
			name:             "max disk usage clamping",
			config:           createTestConfig("1.0.0", 150, "daily", []string{"/System", "/Library"}),
			expectedChanges:  []string{"max_disk_usage"},
			expectedWarnings: 1,
		},
		{
			name: "duplicate paths",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     domain.SafeModeEnabled,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/Library", "/System"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     domain.ProfileStatusEnabled,
							},
						},
						Enabled: domain.ProfileStatusEnabled,
					},
				},
			},
			expectedChanges:  []string{"profiles.daily.operations[0].settings"},
			expectedWarnings: 0,
		},
	}
}
