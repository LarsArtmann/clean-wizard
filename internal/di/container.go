package di

import "github.com/samber/do"

// Container holds all dependency injection configuration
// Replaces ghost custom context with world-class DI pattern
type Container struct {
	container *do.Injector
}

// NewContainer creates and configures dependency injection container
// Provides compile-time safety for all service dependencies
func NewContainer() *Container {
	container := do.New()

	// Configure core services with proper dependency injection
	configureServices(container)

	return &Container{
		container: container,
	}
}

// Replaces manual service management with proper DI patterns
func configureServices(container *do.Injector) {
	// TODO: Implement service configuration when service types are defined
	// Configuration service - root of dependency tree
	// do.Provide(container, NewConfigService)

	// Domain services
	// do.Provide(container, NewValidationService)
	// do.Provide(container, NewSanitizationService)
	// do.Provide(container, NewCleanResultService)

	// API services (for future HTTP integration)
	// do.Provide(container, NewConfigAPIService)
	// do.Provide(container, NewValidationAPIService)
}

// TODO: Implement service constructors when service types are defined
/*
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

// NewValidationAPIService creates a validation API service with dependencies
func NewValidationAPIService(validationService *domain.ValidationService) *ValidationAPIService {
	return &ValidationAPIService{
		configService:     configService,
		validationService: validationService,
	}
}
*/

// MustInvoke is a convenience wrapper for do.MustInvoke
// Provides compile-time safety and error handling
func (c *Container) MustInvoke(service any) any {
	return do.MustInvoke[any](c.container)
}

// Invoke is a convenience wrapper for do.Invoke
// Provides error handling for dependency resolution
func (c *Container) Invoke(service any) (any, error) {
	return do.Invoke[any](c.container)
}
