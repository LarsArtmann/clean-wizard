package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TestSanitizationTestCase defines a single sanitization test case
type TestSanitizationTestCase struct {
	name             string
	config           *domain.Config
	expectedChanges  []string
	expectedWarnings int
}

// TestValidationLevelTestCase defines validation level test cases
type TestValidationLevelTestCase struct {
	name        string
	config       *domain.Config
	level       ValidationLevel
	expectValid  bool
	expectErrors int
}

// CreateTestConfigurations creates test configurations for validation testing
func CreateTestConfigurations() map[string]*domain.Config {
	return map[string]*domain.Config{
		"valid": {
			Version:      "1.0.0",
			SafeMode:     true,
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
							Enabled:     true,
						},
					},
					Enabled: true,
				},
			},
			LastClean: time.Now(),
			Updated:   time.Now(),
		},
		"invalid_high_disk": {
			Version:      "1.0.0",
			SafeMode:     true,
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
							Enabled:     true,
						},
					},
					Enabled: true,
				},
			},
			LastClean: time.Now(),
			Updated:   time.Now(),
		},
	}
}

// createTestConfig creates a test config with common default values
func createTestConfig() *domain.Config {
	return &domain.Config{
		SafeMode:     true,
		MaxDiskUsage:  50,
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
						Enabled:     true,
					},
				},
				Enabled: true,
			},
		},
	}
}

// GetSanitizationTestCases returns all sanitization test cases
func GetSanitizationTestCases() []TestSanitizationTestCase {
	return []TestSanitizationTestCase{
		{
			name: "whitespace cleanup",
			config: func() *domain.Config {
				cfg := createTestConfig()
				cfg.Version = "  1.0.0  "
				cfg.Profiles["daily"].Name = "  daily  "
				return cfg
			}(),
			expectedChanges:  []string{"version", "profiles.daily.name"},
			expectedWarnings: 0,
		},
		{
			name: "max disk usage clamping",
			config: func() *domain.Config {
				cfg := createTestConfig()
				cfg.Version = "1.0.0"
				cfg.MaxDiskUsage = 150 // Will be clamped to 95
				return cfg
			}(),
			expectedChanges:  []string{"max_disk_usage"},
			expectedWarnings: 1,
		},
		{
			name: "duplicate paths",
			config: func() *domain.Config {
				cfg := createTestConfig()
				cfg.Version = "1.0.0"
				cfg.Protected = []string{"/System", "/Library", "/System"} // Duplicate /System
				return cfg
			}(),
			expectedChanges:  []string{"profiles.daily.operations[0].settings"}, // Settings sanitized
			expectedWarnings: 0,
		},
	}
}