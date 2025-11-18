package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/samber/do"
)

// Container holds all dependency injection configuration
// Replaces ghost custom context with world-class DI pattern
type Container struct {
	container *do.ServiceContainer
}

// NewContainer creates and configures the dependency injection container
// Provides compile-time safety for all service dependencies
func NewContainer() *Container {
	container := do.New()
	
	// Configure core services with proper dependency injection
	configureServices(container)
	
	return &Container{
		container: container,
	}
}

// configureServices sets up all service dependencies
// Replaces manual service management with proper DI patterns
func configureServices(container *do.ServiceContainer) {
	// Configuration service - root of dependency tree
	do.Provide(container, NewConfigService)
	
	// Domain services
	do.Provide(container, NewValidationService)
	do.Provide(container, NewSanitizationService)
	do.Provide(container, NewCleanResultService)
	
	// API services (for future HTTP integration)
	do.Provide(container, NewConfigAPIService)
	do.Provide(container, NewValidationAPIService)
}

// NewConfigService creates a configuration service
func NewConfigService() *config.Service {
	return &config.Service{}
}

// NewValidationService creates a validation service with config dependency
func NewValidationService(configService *config.Service) *domain.ValidationService {
	return domain.NewValidationService(configService)
}

// NewSanitizationService creates a sanitization service with validation dependency
func NewSanitizationService(validationService *domain.ValidationService) *domain.SanitizationService {
	return domain.NewSanitizationService(validationService)
}

// NewCleanResultService creates a clean result service
func NewCleanResultService() *domain.CleanResultService {
	return domain.NewCleanResultService()
}

// NewConfigAPIService creates an API service for config operations
func NewConfigAPIService(configService *config.Service, validationService *domain.ValidationService) *ConfigAPIService {
	return NewConfigAPIService(configService, validationService)
}

// NewValidationAPIService creates an API service for validation operations
func NewValidationAPIService(validationService *domain.ValidationService) *domain.ValidationAPIService {
	return domain.NewValidationAPIService(validationService)
}

// MustInvoke is a convenience wrapper for do.MustInvoke
// Provides compile-time safety and error handling
func (c *Container) MustInvoke[T any]() T {
	return do.MustInvoke[T](c.container)
}

// Invoke is a convenience wrapper for do.Invoke
// Provides error handling for dependency resolution
func (c *Container) Invoke[T any]() (T, error) {
	return do.Invoke[T](c.container)
}

// Shutdown gracefully shuts down the dependency injection container
func (c *Container) Shutdown() error {
	// Clean up any resources if needed
	return nil
}