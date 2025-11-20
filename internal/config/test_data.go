package config

import (
	"fmt"
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
			"daily": { // Normal key - the sanitizer should clean up the name field inside
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

// createStandardProfile creates a standard daily cleanup profile
func createStandardProfile() *domain.Profile {
	return &domain.Profile{
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
	}
}

// createBaseConfig creates a base configuration with standard settings
func createBaseConfig(version string, maxDiskUsage int, protected []string) *domain.Config {
	return &domain.Config{
		Version:      version,
		SafetyLevel:  domain.SafetyLevelEnabled,
		MaxDiskUsage: maxDiskUsage,
		Protected:    protected,
		Profiles: map[string]*domain.Profile{
			"daily": createStandardProfile(),
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
			expectedChanges:  []string{"version", "profiles.daily.name", "profiles.daily.description", "profiles.daily.operations[0].name", "profiles.daily.operations[0].description"},
			expectedWarnings: 0,
		},
		{
			name: "max disk usage clamping",
			config: createBaseConfig("1.0.0", 150, []string{"/System", "/Library"}),
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

// CreateValidationTestConfigs creates test configurations specifically for validation testing
// Eliminates duplication from validation_types_test.go
func CreateValidationTestConfigs() map[string]*domain.Config {
	return map[string]*domain.Config{
		"valid": createBaseConfig("1.0.0", 50, []string{"/System", "/Library", "/Applications"}),
		"invalid_high_disk": createBaseConfig("1.0.0", 150, []string{"/System"}), // Invalid: too high
	}
}

// ValidateNixGenerationsOperation validates nix-generations operation settings
// Eliminates duplication from bdd_nix_validation_test.go
func ValidateNixGenerationsOperation(cfg *domain.Config, operationName string) error {
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		return fmt.Errorf("'nix-cleanup' profile not found in config")
	}

	operationFound := false
	for _, op := range profile.Operations {
		if op.Name == "nix-generations" {
			if op.Settings == nil {
				return fmt.Errorf("nix-generations operation has nil Settings")
			}
			if op.Settings.NixGenerations == nil {
				return fmt.Errorf("nix-generations operation has nil NixGenerations")
			}
			operationFound = true
			break
		}
	}

	if !operationFound {
		return fmt.Errorf("'nix-generations' operation not found in nix-cleanup profile")
	}
	return nil
}

// SetNixGenerationsCount sets the generations count for nix-generations operation
// Eliminates duplication from bdd_nix_validation_test.go
func SetNixGenerationsCount(cfg *domain.Config, generations int) error {
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		return fmt.Errorf("'nix-cleanup' profile not found in config")
	}

	operationFound := false
	for i, op := range profile.Operations {
		if op.Name == "nix-generations" {
			if op.Settings == nil {
				return fmt.Errorf("nix-generations operation has nil Settings")
			}
			if op.Settings.NixGenerations == nil {
				return fmt.Errorf("nix-generations operation has nil NixGenerations")
			}
			profile.Operations[i].Settings.NixGenerations.Generations = generations
			operationFound = true
			break
		}
	}

	if !operationFound {
		return fmt.Errorf("'nix-generations' operation not found in nix-cleanup profile")
	}
	return nil
}

// SetNixGenerationsOptimization sets the optimization level for nix-generations operation
// Eliminates duplication from bdd_nix_validation_test.go
func SetNixGenerationsOptimization(cfg *domain.Config, optimizationLevel domain.OptimizationLevelType) error {
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		return fmt.Errorf("'nix-cleanup' profile not found in config")
	}

	operationFound := false
	for i, op := range profile.Operations {
		if op.Name == "nix-generations" {
			if op.Settings == nil {
				return fmt.Errorf("nix-generations operation has nil Settings")
			}
			if op.Settings.NixGenerations == nil {
				return fmt.Errorf("nix-generations operation has nil NixGenerations")
			}
			profile.Operations[i].Settings.NixGenerations.Optimization = optimizationLevel
			operationFound = true
			break
		}
	}

	if !operationFound {
		return fmt.Errorf("'nix-generations' operation not found in nix-cleanup profile")
	}
	return nil
}

// CreateSemverTestConfig creates a standard test configuration for semver validation testing
// Eliminates duplication from semver_validation_test.go
func CreateSemverTestConfig(version string) *domain.Config {
	return &domain.Config{
		Version: version,
		Profiles: map[string]*domain.Profile{
			"test": {
				Name:        "test",
				Description: "Test profile",
				Operations: []domain.CleanupOperation{{
					Name:        "test-op",
					Description: "Test operation",
					RiskLevel:   domain.RiskLow,
					Status:      domain.StatusEnabled,
					Settings:    &domain.OperationSettings{NixGenerations: &domain.NixGenerationsSettings{Generations: 5}},
				}},
				Status: domain.StatusEnabled,
			},
		},
		Protected: []string{"/System"},
	}
}

// CreateValidationTestConfig creates a standard test configuration for validation testing
// Eliminates duplication from validation_validator_test.go
func CreateValidationTestConfig(version string, maxDiskUsage int, protected []string) *domain.Config {
	return &domain.Config{
		Version:      version,
		SafetyLevel:  domain.SafetyLevelEnabled,
		MaxDiskUsage: maxDiskUsage,
		Protected:    protected,
		Profiles: map[string]*domain.Profile{
			"daily": createStandardProfile(),
		},
	}
}