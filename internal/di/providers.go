package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/samber/do/v2"
)

// RegisterAllServices wires every application service into the DI container.
// It registers the already-loaded config and run settings as eager values,
// then invokes the provider packages for lazy service creation.
//
// This is the single entry point for service registration, mirroring
// BuildFlow's di.RegisterAllServices pattern.
func RegisterAllServices(injector do.Injector, cfg *domain.Config, settings RunSettings) error {
	do.ProvideValue(injector, cfg)
	do.ProvideValue(injector, settings)

	CleanerPackage(injector)

	return nil
}

// CleanerPackage groups all cleaner-related provider registrations.
// Using do.Package keeps the registration organized and composable,
// matching BuildFlow's InfrastructurePackage / ApplicationPackage pattern.
var CleanerPackage = do.Package(
	registerCleanerRegistry,
)

// registerCleanerRegistry provides a *cleaner.Registry as a lazy singleton.
// The registry is created with the verbose/dryRun flags resolved from RunSettings,
// eliminating the former dual-registry pattern where cleaners were instantiated
// twice (once for discovery, once for execution).
func registerCleanerRegistry(injector do.Injector) {
	do.Provide(injector, func(i do.Injector) (*cleaner.Registry, error) {
		settings, err := do.Invoke[RunSettings](i)
		if err != nil {
			return nil, errorfamily.WrapRejection(
				err,
				"di.resolve_settings_for_registry",
				"failed to resolve RunSettings for cleaner registry",
			)
		}

		registry, err := cleaner.DefaultRegistryWithConfig(settings.Verbose, settings.DryRun)
		if err != nil {
			return nil, errorfamily.WrapRejection(err, "di.create_registry", "failed to create cleaner registry")
		}

		return registry, nil
	})
}
