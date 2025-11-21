package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// TODO: CRITICAL - Continue migrating remaining functions to domain-specific factory files:
// - CreateSemverTestConfig() -> validation_factory.go
// - CreateIntegrationConfig() -> integration_factory.go
// - CreateValidationTestConfig() -> validation_factory.go

// CreateSemverTestConfig creates a standard test configuration for semver validation testing
// Original location: internal/application/config/test_data.go lines 200-220
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
				}},
				Status: shared.StatusEnabled,
			},
		},
		CurrentProfile: "test",
		LastClean:      time.Now(),
		Updated:        time.Now(),
	}
}
