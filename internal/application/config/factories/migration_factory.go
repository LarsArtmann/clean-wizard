package factories

import (
	""
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
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
func CreateSemverTestConfig(version string) *config.Config {
	return &Config{
		Version: version,
		Profiles: map[string]*ConfigProfile{
			"test": {
				Name:        "test",
				Description: "Test profile",
				Operations: []ConfigCleanupOperation{{
					Name:        "test-op",
					Description: "Test operation",
					RiskLevel:   shared.RiskLow,
					Status:      shared.StatusEnabled,
					Settings:    &shared.OperationSettings{NixGenerations: &shared.NixGenerationsSettings{Generations: 5}},
				}},
				Status: shared.StatusEnabled,
			},
		},
		CurrentProfile: "test",
		LastClean:      time.Now(),
		Updated:        time.Now(),
	}
}

// TODO: Add remaining migration functions as needed
