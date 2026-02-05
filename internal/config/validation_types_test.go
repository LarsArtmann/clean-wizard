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

// CreateTestConfig creates a test configuration with optional overrides.
// Default values:
//   - Version: "1.0.0"
//   - MaxDiskUsage: 50
//   - ProfileName: "daily"
//   - Protected: []string{"/System", "/Library"}
//
// Example usage:
//   // Default config
//   CreateTestConfig()
//   // Custom max disk usage
//   CreateTestConfig(WithMaxDiskUsage(75))
//   // Custom protected paths
//   CreateTestConfig(WithProtectedPaths([]string{"/System"}))
func CreateTestConfig(opts ...ConfigOption) *domain.Config {
	cfg := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     domain.SafeModeEnabled,
		MaxDiskUsage: 50,
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
						Enabled:     domain.ProfileStatusEnabled,
					},
				},
				Enabled: domain.ProfileStatusEnabled,
			},
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// ConfigOption is a function that modifies a test configuration.
type ConfigOption func(*domain.Config)

// WithVersion sets the version of the configuration.
func WithVersion(version string) ConfigOption {
	return func(c *domain.Config) {
		c.Version = version
	}
}

// WithMaxDiskUsage sets the max disk usage percentage.
func WithMaxDiskUsage(percent int) ConfigOption {
	return func(c *domain.Config) {
		c.MaxDiskUsage = percent
	}
}

// WithProtectedPaths sets the protected paths.
func WithProtectedPaths(paths []string) ConfigOption {
	return func(c *domain.Config) {
		c.Protected = paths
	}
}

// WithProfileName sets the profile name.
func WithProfileName(name string) ConfigOption {
	return func(c *domain.Config) {
		c.Profiles["daily"].Name = name
	}
}

// GetSanitizationTestCases returns all sanitization test cases.
func GetSanitizationTestCases() []TestSanitizationTestCase {
	return []TestSanitizationTestCase{
		{
			name:             "whitespace cleanup",
			config:           CreateTestConfig(WithVersion("  1.0.0  "), WithProfileName("  daily  ")),
			expectedChanges:  []string{"version", "profiles.daily.name"},
			expectedWarnings: 0,
		},
		{
			name:             "max disk usage clamping",
			config:           CreateTestConfig(WithMaxDiskUsage(150)),
			expectedChanges:  []string{"max_disk_usage"},
			expectedWarnings: 1,
		},
		{
			name:             "duplicate paths",
			config:           CreateTestConfig(WithProtectedPaths([]string{"/System", "/Library", "/System"})),
			expectedChanges:  []string{"profiles.daily.operations[0].settings"},
			expectedWarnings: 0,
		},
	}
}
