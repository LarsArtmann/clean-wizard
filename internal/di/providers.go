package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/samber/do/v2"
)

// RegisterAllServices wires every application service into the DI container.
// It registers the already-loaded config and run settings as eager values,
// then invokes the provider packages for lazy service creation.
//
// This is the single entry point for service registration, mirroring
// BuildFlow's di.RegisterAllServices pattern.
func RegisterAllServices(injector do.Injector, cfg *types.Config, settings RunSettings) error {
	do.ProvideValue(injector, cfg)
	do.ProvideValue(injector, settings)

	CleanerPackage(injector)

	return nil
}

// CleanerPackage groups all cleaner-related provider registrations.
// Using do.Package keeps the registration organized and composable,
// matching BuildFlow's InfrastructurePackage / ApplicationPackage pattern.
var CleanerPackage = do.Package( //nolint:gochecknoglobals
	registerCleanerRegistry,
)

// registerCleanerRegistry provides a *cleaner.Registry as a lazy singleton.
// The registry is created with the verbose/dryRun flags resolved from RunSettings
// and, when a CLI profile was selected, with the merged operation settings of that
// profile. This eliminates the former dual-registry pattern where cleaners were
// instantiated twice (once for discovery, once for execution).
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

		operationSettings, err := resolveProfileOperationSettings(i, settings)
		if err != nil {
			return nil, err
		}

		registry, err := cleaner.DefaultRegistryWithConfig(settings.Verbose, settings.DryRun, operationSettings)
		if err != nil {
			return nil, errorfamily.WrapRejection(err, "di.create_registry", "failed to create cleaner registry")
		}

		return registry, nil
	})
}

// resolveProfileOperationSettings returns the merged OperationSettings of the
// selected profile. When no profile was selected (preset or interactive cleaner
// selection) it returns nil so cleaners fall back to their factory defaults.
func resolveProfileOperationSettings(injector do.Injector, settings RunSettings) (*operations.OperationSettings, error) {
	if settings.Profile == "" {
		return nil, nil
	}

	cfg, err := do.Invoke[*types.Config](injector)
	if err != nil {
		return nil, errorfamily.WrapRejection(
			err,
			"di.resolve_config_for_registry",
			"failed to resolve Config for profile settings",
		)
	}

	return cfg.SettingsForProfile(settings.Profile), nil
}
