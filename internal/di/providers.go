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
// BuildFlow's di.RegisterAllServices pattern. A selected profile that does
// not exist in the config is a Rejection at the composition root — it used
// to fall back to factory defaults silently, so a typo'd --profile produced
// a full run with no profile applied.
func RegisterAllServices(injector do.Injector, cfg *types.Config, settings RunSettings) error {
	do.ProvideValue(injector, cfg)
	do.ProvideValue(injector, settings)

	if settings.Profile != "" && !cfg.HasProfile(settings.Profile) {
		return errorfamily.NewRejection(
			"di.unknown_profile",
			"profile "+settings.Profile+" does not exist in the configuration",
		)
	}

	AdaptersPackage(injector)
	CleanerPackage(injector)

	// Cleaner providers resolve adapters.NixStore lazily, so the alias only
	// needs to exist before the first invocation, not before CleanerPackage.
	if err := do.As[*adapters.NixAdapter, adapters.NixStore](injector); err != nil {
		return errorfamily.WrapRejection(
			err,
			"di.alias_nix_store",
			"failed to alias the nix adapter to its consumer-facing interface",
		)
	}

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
//
// NOTE (domain semantics, do not "fix"): the (nil, nil) return is the
// deliberate "no profile selected" signal, not a missing-value error — nil
// settings is the factory's documented default input.
func resolveProfileOperationSettings(
	injector do.Injector,
	settings RunSettings,
) (*operations.OperationSettings, error) {
	if settings.Profile == "" {
		return nil, nil //nolint:nilnil // "no profile" is a valid state, not a missing value — nil settings means factory defaults
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
