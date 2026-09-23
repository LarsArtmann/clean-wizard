package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
)

// TestSanitizationTestCase defines a single sanitization test case.
type TestSanitizationTestCase struct {
	name             string
	config           *types.Config
	expectedChanges  []string
	expectedWarnings int
}

// CreateTestConfigurations creates test configurations for validation testing.
func CreateTestConfigurations() map[string]*types.Config {
	return map[string]*types.Config{
		"valid": {
			Version:      "1.0.0",
			SafeMode:     enums.SafeModeEnabled,
			MaxDiskUsage: 50,
			Protected:    []string{"/System", "/Library", "/Applications"},
			Profiles: map[string]*types.Profile{
				"daily": CreateDailyProfile(),
			},
			LastClean: time.Now(),
			Updated:   time.Now(),
		},
		"invalid_high_disk": {
			Version:      "1.0.0",
			SafeMode:     enums.SafeModeEnabled,
			MaxDiskUsage: 150, // Invalid: too high
			Protected:    []string{"/System"},
			Profiles: map[string]*types.Profile{
				"daily": CreateDailyProfile(),
			},
			LastClean: time.Now(),
			Updated:   time.Now(),
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
//
//	// Default config
//	CreateTestConfig()
//	// Custom max disk usage
//	CreateTestConfig(WithMaxDiskUsage(75))
//	// Custom protected paths
//	CreateTestConfig(WithProtectedPaths([]string{"/System"}))
func CreateTestConfig(opts ...ConfigOption) *types.Config {
	cfg := &types.Config{
		Version:      "1.0.0",
		SafeMode:     enums.SafeModeEnabled,
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Library"},
		Profiles: map[string]*types.Profile{
			"daily": CreateDailyProfile(),
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// ConfigOption is a function that modifies a test configuration.
type ConfigOption func(*types.Config)

// WithVersion sets the version of the configuration.
func WithVersion(version string) ConfigOption {
	return func(c *types.Config) {
		c.Version = version
	}
}

// WithMaxDiskUsage sets the max disk usage percentage.
func WithMaxDiskUsage(percent int) ConfigOption {
	return func(c *types.Config) {
		c.MaxDiskUsage = percent
	}
}

// WithProtectedPaths sets the protected paths.
func WithProtectedPaths(paths []string) ConfigOption {
	return func(c *types.Config) {
		c.Protected = paths
	}
}

// WithProfileName sets the profile name.
func WithProfileName(name string) ConfigOption {
	return func(c *types.Config) {
		c.Profiles["daily"] = CreateDailyProfile(WithDailyProfileName(name))
	}
}

// WithEmptyProfiles sets the Profiles map to empty.
func WithEmptyProfiles() ConfigOption {
	return func(c *types.Config) {
		c.Profiles = map[string]*types.Profile{}
	}
}

// GetSanitizationTestCases returns all sanitization test cases.
func GetSanitizationTestCases() []TestSanitizationTestCase {
	return []TestSanitizationTestCase{
		{
			name: "whitespace cleanup",
			config: CreateTestConfig(
				WithVersion("  1.0.0  "),
				WithProfileName("  daily  "),
			),
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
			name: "duplicate paths",
			config: CreateTestConfig(
				WithProtectedPaths([]string{"/System", "/Library", "/System"}),
			),
			expectedChanges:  []string{"profiles.daily.operations[0].settings"},
			expectedWarnings: 0,
		},
	}
}

// CreateDailyProfile creates a test daily profile with customizable options.
func CreateDailyProfile(opts ...DailyProfileOption) *types.Profile {
	profile := &types.Profile{
		Name:        "daily",
		Description: "Daily cleanup",
		Operations: []types.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean Nix generations",
				RiskLevel:   enums.RiskLevelLowType,
				Enabled:     enums.ProfileStatusEnabled,
			},
		},
		Enabled: enums.ProfileStatusEnabled,
	}

	for _, opt := range opts {
		opt(profile)
	}

	return profile
}

// CreateWeeklyProfile creates a weekly profile for deep cleanup operations.
func CreateWeeklyProfile() *types.Profile {
	return &types.Profile{
		Name:        "Weekly Deep Cleanup",
		Description: "Weekly deep cleanup operations",
		Operations: []types.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Deep Nix cleanup",
				RiskLevel:   enums.RiskLevelMediumType,
				Enabled:     enums.ProfileStatusEnabled,
				Settings: &operations.OperationSettings{
					NixGenerations: &operations.NixGenerationsSettings{
						Generations: 5,
						Optimize:    enums.OptimizationModeEnabled,
					},
				},
			},
		},
		Enabled: enums.ProfileStatusEnabled,
	}
}

// DailyProfileOption is a function that modifies a test profile.
type DailyProfileOption func(*types.Profile)

// WithDailyProfileName sets the profile name.
func WithDailyProfileName(name string) DailyProfileOption {
	return func(p *types.Profile) {
		p.Name = name
	}
}

// createHomebrewOperation creates a homebrew cleanup operation.
func createHomebrewOperation() types.CleanupOperation {
	return types.CleanupOperation{
		Name:        "homebrew-cleanup",
		Description: "Clean Homebrew",
		RiskLevel:   enums.RiskLevelLowType,
		Enabled:     enums.ProfileStatusEnabled,
		Settings: &operations.OperationSettings{
			Homebrew: &operations.HomebrewSettings{
				UnusedOnly: enums.HomebrewModeUnusedOnly,
				Prune:      "30d",
			},
		},
	}
}

// createNixGenerationsOperation creates a nix-generations cleanup operation with settings.
func createNixGenerationsOperation() types.CleanupOperation {
	return types.CleanupOperation{
		Name:        "nix-generations",
		Description: "Clean Nix generations",
		RiskLevel:   enums.RiskLevelLowType,
		Enabled:     enums.ProfileStatusEnabled,
		Settings: &operations.OperationSettings{
			NixGenerations: &operations.NixGenerationsSettings{
				Generations: 3,
				Optimize:    enums.OptimizationModeEnabled,
			},
		},
	}
}

// CreateBenchmarkConfig creates a configuration suitable for benchmarking.
// This includes a daily profile with multiple operations covering all operation types.
func CreateBenchmarkConfig() *types.Config {
	return &types.Config{
		Version:      "1.0.0",
		SafeMode:     enums.SafeModeEnabled,
		MaxDiskUsage: 75,
		Protected:    []string{"/System", "/Applications", "/Library", "/usr", "/etc", "/var"},
		Profiles: map[string]*types.Profile{
			"daily": {
				Name:        "Daily Cleanup",
				Description: "Daily system cleanup",
				Operations: []types.CleanupOperation{
					createNixGenerationsOperation(),
					{
						Name:        "temp-files",
						Description: "Clean temporary files",
						RiskLevel:   enums.RiskLevelMediumType,
						Enabled:     enums.ProfileStatusEnabled,
						Settings: &operations.OperationSettings{
							TempFiles: &operations.TempFilesSettings{
								OlderThan: "7d",
								Excludes:  []string{"/tmp/keep", "/var/tmp/preserve"},
							},
						},
					},
					createHomebrewOperation(),
					{
						Name:        "system-temp",
						Description: "Clean system temp",
						RiskLevel:   enums.RiskLevelMediumType,
						Enabled:     enums.ProfileStatusEnabled,
						Settings: &operations.OperationSettings{
							SystemTemp: &operations.SystemTempSettings{
								Paths:     []string{"/tmp", "/var/tmp", "/tmp/.font-unix"},
								OlderThan: "14d",
							},
						},
					},
				},
				Enabled: enums.ProfileStatusEnabled,
			},
			"weekly": CreateWeeklyProfile(),
		},
	}
}

// CreateIntegrationTestConfig creates a complex configuration for integration testing.
// This includes multiple profiles with various operations for testing the complete pipeline.
func CreateIntegrationTestConfig() *types.Config {
	return &types.Config{
		Version:      " 1.0.0  ",
		SafeMode:     enums.SafeModeEnabled,
		MaxDiskUsage: 85,
		Protected:    []string{"/System", "/Library", "/Applications", "/System"},
		Profiles: map[string]*types.Profile{
			"daily": {
				Name:        "  Daily Cleanup  ",
				Description: "Daily system cleanup operations",
				Operations: []types.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: " Clean Nix generations ",
						RiskLevel:   enums.RiskLevelLowType,
						Enabled:     enums.ProfileStatusEnabled,
						Settings: &operations.OperationSettings{
							NixGenerations: &operations.NixGenerationsSettings{
								Generations: 3,
								Optimize:    enums.OptimizationModeEnabled,
							},
						},
					},
					{
						Name:        "temp-files",
						Description: "Clean temporary files",
						RiskLevel:   enums.RiskLevelMediumType,
						Enabled:     enums.ProfileStatusEnabled,
						Settings: &operations.OperationSettings{
							TempFiles: &operations.TempFilesSettings{
								OlderThan: " 7d  ",
								Excludes:  []string{"/tmp/keep", "/var/tmp/preserve", "/tmp/keep"},
							},
						},
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean Homebrew",
						RiskLevel:   enums.RiskLevelLowType,
						Enabled:     enums.ProfileStatusEnabled,
						Settings: &operations.OperationSettings{
							Homebrew: &operations.HomebrewSettings{
								UnusedOnly: enums.HomebrewModeUnusedOnly,
								Prune:      " 30d  ",
							},
						},
					},
					{
						Name:        "system-temp",
						Description: "Clean system temp",
						RiskLevel:   enums.RiskLevelMediumType,
						Enabled:     enums.ProfileStatusEnabled,
						Settings: &operations.OperationSettings{
							SystemTemp: &operations.SystemTempSettings{
								Paths:     []string{"/tmp", "/var/tmp", " /tmp/extra ", "/tmp"},
								OlderThan: "14d",
							},
						},
					},
				},
				Enabled: enums.ProfileStatusEnabled,
			},
			"weekly": CreateWeeklyProfile(),
		},
	}
}
