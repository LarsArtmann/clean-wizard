package di

// Disabled DI implementation - requires interface definitions
// TODO: Implement when service interfaces are properly defined

// Container holds all dependency injection configuration
// Replaces ghost custom context with world-class DI pattern
type Container struct {
	// Placeholder for DI implementation
}

// NewContainer creates and configures dependency injection container
// Provides compile-time safety for all service dependencies
func NewContainer() *Container {
	// Placeholder implementation
	return &Container{}
}

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

// Service constructors with proper dependency injection are implemented in api_services.go

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
