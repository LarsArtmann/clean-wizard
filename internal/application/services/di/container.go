package di

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/infrastructure/cleaners"
	"github.com/LarsArtmann/clean-wizard/internal/application/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	"github.com/LarsArtmann/clean-wizard/internal/shared/utils/middleware"
	"github.com/rs/zerolog"
)

// Container holds all dependency injection configuration
// Implements simple DI pattern without external libraries
type Container struct {
	logger     zerolog.Logger
	config     *domain.Config
	nixCleaner domain.Cleaner
	validation *middleware.ValidationMiddleware
}

// NewContainer creates new dependency injection container
func NewContainer(ctx context.Context) *Container {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Ctx(ctx).Logger()
	config := config.GetDefaultConfig()
	nixCleaner := cleaners.NewNixCleaner(false, false)
	validation := middleware.NewValidationMiddleware()

	return &Container{
		logger:     logger,
		config:     config,
		nixCleaner: nixCleaner,
		validation: validation,
	}
}

// UpdateConfig updates configuration in container
func (c *Container) UpdateConfig(config *domain.Config) {
	c.config = config
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

// Shutdown gracefully shuts down of container
func (c *Container) Shutdown(ctx context.Context) error {
	c.logger.Info().Msg("Shutting down dependency injection container")
	// Perform cleanup operations here
	return nil
}
