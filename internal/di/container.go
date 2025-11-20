package di

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/middleware"
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
	nixCleaner := cleaner.NewNixCleaner(false, false)
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

// Shutdown gracefully shuts down of container
func (c *Container) Shutdown(ctx context.Context) error {
	c.logger.Info().Msg("Shutting down dependency injection container")
	// Perform cleanup operations here
	return nil
}
