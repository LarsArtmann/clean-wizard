package factories

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TODO: Migrate these functions to proper domain files:
// - CreateSemverTestConfig() -> validation_factory.go
// - CreateIntegrationConfig() -> integration_factory.go
// - CreateValidationTestConfig() -> validation_factory.go
// - ValidateNixGenerationsOperation() -> nix_factory.go
// - SetNixGenerationsCount() -> nix_factory.go
// - SetNixGenerationsOptimization() -> nix_factory.go
// - getNixGenerationsOperation() -> nix_factory.go

// CreateSemverTestConfig creates a standard test configuration for semver validation testing
// Original location: internal/config/test_data.go lines 200-220
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

// CreateIntegrationConfig creates dirty config for sanitization testing
// Original location: internal/config/test_data.go lines 342-390
func CreateIntegrationConfig() *domain.Config {
	cfg := CreateBenchmarkConfig()
	// Add integration test specific modifications
	cfg.MaxDiskUsage = 85
	cfg.Version = " 1.0.0  "                         // whitespace for sanitization testing
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
// Original location: internal/config/test_data.go lines 390-397
func CreateValidationTestConfig(version string, maxDiskUsage int, protected []string) *domain.Config {
	return CreateBaseConfig(version, maxDiskUsage, protected)
}

// ValidateNixGenerationsOperation validates nix-generations operation settings
// Original location: internal/config/test_data.go lines 150-170
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

// getNixGenerationsOperation returns nix-generations operation from nix-cleanup profile
// Original location: internal/config/test_data.go lines 175-195
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
// Original location: internal/config/test_data.go lines 200-210
func SetNixGenerationsCount(cfg *domain.Config, generations int) error {
	nixGen, err := getNixGenerationsOperation(cfg)
	if err != nil {
		return err
	}
	nixGen.Generations = generations
	return nil
}

// SetNixGenerationsOptimization sets the optimization level for nix-generations operation
// Original location: internal/config/test_data.go lines 215-225
func SetNixGenerationsOptimization(cfg *domain.Config, optimizationLevel domain.OptimizationLevelType) error {
	nixGen, err := getNixGenerationsOperation(cfg)
	if err != nil {
		return err
	}
	nixGen.Optimization = optimizationLevel
	return nil
}
