package di

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/middleware"
	"github.com/rs/zerolog"
)

// Container holds all dependency injection configuration
// Implements simple DI pattern without external libraries
type Container struct {
	logger          zerolog.Logger
	config          *domain.Config
	nixCleaner      domain.Cleaner
	validation      *middleware.ValidationMiddleware
}

// NewContainer creates new dependency injection container
func NewContainer(ctx context.Context) *Container {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	config := getDefaultConfig()
	nixCleaner := cleaner.NewNixCleaner(false, false)
	validation := middleware.NewValidationMiddleware()
	
	return &Container{
		logger:     logger,
		config:     config,
		nixCleaner: nixCleaner,
		validation: validation,
	}
}

// getDefaultConfig returns default configuration
func getDefaultConfig() *domain.Config {
	return &domain.Config{
		Version:      "1.0.0",
		SafeMode:     true,
		MaxDiskUsage:  50,
		Protected:    []string{"/System", "/Library", "/Applications"},
		Profiles:     getDefaultProfiles(),
	}
}

// getDefaultProfiles returns default configuration profiles
func getDefaultProfiles() map[string]*domain.Profile {
	return map[string]*domain.Profile{
		"daily": {
			Name:        "Daily Cleanup",
			Description: "Daily system cleanup operations",
			Operations: []domain.CleanupOperation{
				{
					Name:        "nix-generations",
					Description: "Clean Nix generations",
					RiskLevel:   domain.RiskLow,
					Enabled:     true,
					Settings: &domain.OperationSettings{
						NixGenerations: &domain.NixGenerationsSettings{
							Generations: 3,
							Optimize:    true,
						},
					},
				},
			},
			Enabled: true,
		},
	}
}

// GetLogger returns logger instance
func (c *Container) GetLogger() zerolog.Logger {
	return c.logger
}

// GetConfig returns configuration instance
func (c *Container) GetConfig() *domain.Config {
	return c.config
}

// GetCleaner returns Nix cleaner instance
func (c *Container) GetCleaner() domain.Cleaner {
	return c.nixCleaner
}

// GetValidationMiddleware returns validation middleware instance
func (c *Container) GetValidationMiddleware() *middleware.ValidationMiddleware {
	return c.validation
}

// UpdateConfig updates configuration in container
func (c *Container) UpdateConfig(config *domain.Config) {
	c.config = config
}

// Shutdown gracefully shuts down of container
func (c *Container) Shutdown(ctx context.Context) error {
	c.logger.Info().Msg("Shutting down dependency injection container")
	// Perform cleanup operations here
	return nil
}