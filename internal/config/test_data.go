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
		SafetyLevel: domain.SafetyLevelEnabled,
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
		baseConfig.Version = " 1.0.0 " // Extra spaces
		baseConfig.Protected = []string{"/System ", " /Library"} // Trailing/leading spaces
		baseConfig.Profiles = map[string]*domain.Profile{
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
		}
	}

	return baseConfig
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

// getNixGenerationsOperation returns the nix-generations operation from nix-cleanup profile
// Helper function to eliminate duplication between nix-generations helpers
func getNixGenerationsOperation(cfg *domain.Config) (*domain.NixGenerationsSettings, error) {
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		return nil, fmt.Errorf("'nix-cleanup' profile not found in config")
	}

	for _, op := range profile.Operations {
		if op.Name == "nix-generations" {
			if op.Settings == nil {
				return nil, fmt.Errorf("nix-generations operation has nil Settings")
			}
			if op.Settings.NixGenerations == nil {
				return nil, fmt.Errorf("nix-generations operation has nil NixGenerations")
			}
			return op.Settings.NixGenerations, nil
		}
	}

	return nil, fmt.Errorf("'nix-generations' operation not found in nix-cleanup profile")
}

// SetNixGenerationsCount sets the generations count for nix-generations operation
// Eliminates duplication from bdd_nix_validation_test.go
func SetNixGenerationsCount(cfg *domain.Config, generations int) error {
	nixGen, err := getNixGenerationsOperation(cfg)
	if err != nil {
		return err
	}
	nixGen.Generations = generations
	return nil
}

// SetNixGenerationsOptimization sets the optimization level for nix-generations operation
// Eliminates duplication from bdd_nix_validation_test.go
func SetNixGenerationsOptimization(cfg *domain.Config, optimizationLevel domain.OptimizationLevelType) error {
	nixGen, err := getNixGenerationsOperation(cfg)
	if err != nil {
		return err
	}
	nixGen.Optimization = optimizationLevel
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

// CreateDailyCleanupProfile creates a standard daily cleanup profile with all operations
func CreateDailyCleanupProfile() *domain.Profile {
	return &domain.Profile{
		Name:        "Daily Cleanup",
		Description: "Daily system cleanup",
		Operations: []domain.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean Nix generations",
				RiskLevel:   domain.RiskLow,
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					NixGenerations: &domain.NixGenerationsSettings{
						Generations:  3,
						Optimization: domain.OptimizationLevelConservative,
					},
				},
			},
			{
				Name:        "temp-files",
				Description: "Clean temporary files",
				RiskLevel:   domain.RiskMedium,
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					TempFiles: &domain.TempFilesSettings{
						OlderThan: "7d",
						Excludes:  []string{"/tmp/keep", "/var/tmp/preserve"},
					},
				},
			},
			{
				Name:        "homebrew-cleanup",
				Description: "Clean Homebrew",
				RiskLevel:   domain.RiskLow,
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					Homebrew: &domain.HomebrewSettings{
						FileSelectionStrategy: domain.FileSelectionStrategyUnusedOnly,
						Prune:                 "30d",
					},
				},
			},
			{
				Name:        "system-temp",
				Description: "Clean system temp",
				RiskLevel:   domain.RiskMedium,
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					SystemTemp: &domain.SystemTempSettings{
						Paths:     []string{"/tmp", "/var/tmp", "/tmp/.font-unix"},
						OlderThan: "14d",
					},
				},
			},
		},
		Status: domain.StatusEnabled,
	}
}

// CreateWeeklyCleanupProfile creates a standard weekly cleanup profile
func CreateWeeklyCleanupProfile() *domain.Profile {
	return &domain.Profile{
		Name:        "Weekly Cleanup",
		Description: "Weekly deep cleanup",
		Operations: []domain.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Deep Nix cleanup",
				RiskLevel:   domain.RiskMedium,
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					NixGenerations: &domain.NixGenerationsSettings{
						Generations:  5,
						Optimization: domain.OptimizationLevelConservative,
					},
				},
			},
		},
		Status: domain.StatusEnabled,
	}
}

// CreateBenchmarkConfig creates clean config for performance testing
// Eliminates duplication from validation_benchmark_test.go
func CreateBenchmarkConfig() *domain.Config {
	return &domain.Config{
		Version:      "1.0.0",
		SafetyLevel:  domain.SafetyLevelEnabled,
		MaxDiskUsage: 75,
		Protected:    []string{"/System", "/Applications", "/Library", "/usr", "/etc", "/var"},
		Profiles: map[string]*domain.Profile{
			"daily":  CreateDailyCleanupProfile(),
			"weekly": CreateWeeklyCleanupProfile(),
		},
	}
}

// CreateIntegrationConfig creates dirty config for sanitization testing  
// Eliminates duplication from integration_test.go
func CreateIntegrationConfig() *domain.Config {
	cfg := CreateBenchmarkConfig()
	// Add integration test specific modifications
	cfg.MaxDiskUsage = 85
	cfg.Version = " 1.0.0  " // whitespace for sanitization testing
	cfg.Protected = append(cfg.Protected, "/System") // duplicate for testing
	
	// Add whitespace to profile names for sanitization testing
	if daily, exists := cfg.Profiles["daily"]; exists {
		daily.Name = "  Daily Cleanup  "
		daily.Description = "Daily system cleanup operations"
		// Add whitespace to operation descriptions and duplicates
		for i := range daily.Operations {
			switch daily.Operations[i].Name {
			case "nix-generations":
				daily.Operations[i].Description = " Clean Nix generations "
			case "temp-files":
				daily.Operations[i].Settings.TempFiles.OlderThan = " 7d  "
				daily.Operations[i].Settings.TempFiles.Excludes = append(daily.Operations[i].Settings.TempFiles.Excludes, "/tmp/keep") // duplicate
			case "homebrew-cleanup":
				daily.Operations[i].Settings.Homebrew.Prune = " 30d  "
			case "system-temp":
				daily.Operations[i].Settings.SystemTemp.Paths = append(daily.Operations[i].Settings.SystemTemp.Paths, " /tmp/extra ", "/tmp") // whitespace and duplicate
			}
		}
	}
	
	if weekly, exists := cfg.Profiles["weekly"]; exists {
		weekly.Name = "Weekly Deep Cleanup"
		weekly.Description = "Weekly deep cleanup operations"
	}
	
	return cfg
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