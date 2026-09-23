package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
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

	AdaptersPackage(injector)
	CleanerPackage(injector)

	return nil
}

// CleanerPackage groups all cleaner-related provider registrations.
// Each cleaner is registered as its own named service (enabling per-cleaner
// resolution and config), and the registry aggregates them, matching
// BuildFlow's InfrastructurePackage / ApplicationPackage pattern.
var CleanerPackage = do.Package( //nolint:gochecknoglobals
	registerCleanerProviders,
	registerCleanerRegistryFromProviders,
)

// resolveProfileOperationSettings returns the merged OperationSettings of the
// selected profile. When no profile was selected (preset or interactive cleaner
// selection) it returns nil so cleaners fall back to their factory defaults.
func resolveProfileOperationSettings(
	injector do.Injector,
	settings RunSettings,
) (*operations.OperationSettings, error) {
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
