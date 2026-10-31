package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/factory"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/samber/do/v2"
)

// cleanerServiceName is the DI service name for a single cleaner.
func cleanerServiceName(registryName string) string {
	return "cleaner." + registryName
}

// registerCleanerProviders registers every cleaner as its own named lazy
// singleton service ("cleaner.<name>"). Each provider resolves the run settings
// and profile operation settings itself, so a single cleaner can be invoked
// from the container without materializing the whole registry. Construction
// and per-cleaner settings resolution live in the factory package; this file
// only wires them into the container.
//
// The Nix cleaner is the one cleaner whose construction takes a
// container-managed dependency: it receives the adapters.NixStore service, so
// tests can swap the store via do.Override and future store implementations
// stay a registration change away.
func registerCleanerProviders(injector do.Injector) {
	for _, registration := range factory.CleanerRegistrations() {
		constructor := registration.Constructor

		do.ProvideNamed(injector, cleanerServiceName(registration.Name), func(i do.Injector) (cleaner.Cleaner, error) {
			settings, err := do.Invoke[RunSettings](i)
			if err != nil {
				return nil, errorfamily.WrapRejection(
					err,
					"di.resolve_settings_for_cleaner",
					"failed to resolve RunSettings for cleaner",
				)
			}

			operationSettings, err := resolveProfileOperationSettings(i, settings)
			if err != nil {
				return nil, err
			}

			if registration.Name == cleaner.CleanerNix {
				return buildNixCleaner(i, settings, operationSettings)
			}

			return constructor(settings.Verbose, settings.DryRun, operationSettings)
		})
	}
}

// buildNixCleaner constructs the nix cleaner with the container-resolved
// adapters.NixStore instead of the inline real adapter.
func buildNixCleaner( //nolint:ireturn // DI returns interfaces by design
	injector do.Injector,
	settings RunSettings,
	operationSettings *operations.OperationSettings,
) (cleaner.Cleaner, error) {
	store, err := do.Invoke[adapters.NixStore](injector)
	if err != nil {
		return nil, errorfamily.WrapRejection(
			err,
			"di.resolve_nix_store_for_cleaner",
			"failed to resolve NixStore for the nix cleaner",
		)
	}

	return factory.NixWithStore(settings.Verbose, settings.DryRun, store, operationSettings)
}

// registerCleanerRegistryFromProviders provides the *cleaner.Registry by
// resolving every per-cleaner service from the container. Individual cleaner
// providers stay independently invokable while the registry preserves the
// canonical registration order.
//
// NOTE (intentional, do not "fix"): this loop resolves a *distinct named
// service* per iteration (samber/do anti-pattern rule DO-5 targets repeated
// resolution of the *same* service). Named per-cleaner services are the point
// of the design: each stays independently invokable and the registry keeps
// canonical order.
func registerCleanerRegistryFromProviders(injector do.Injector) {
	do.Provide(injector, func(i do.Injector) (*cleaner.Registry, error) {
		registry := cleaner.NewRegistry()

		for _, registration := range factory.CleanerRegistrations() {
			cleanerInstance, err := do.InvokeNamed[cleaner.Cleaner](i, cleanerServiceName(registration.Name))
			if err != nil {
				return nil, errorfamily.WrapRejectionf(
					err,
					"di.resolve_cleaner_for_registry",
					"failed to resolve cleaner=%s from container",
					registration.Name,
				)
			}

			registry.Register(registration.Name, cleanerInstance)
		}

		return registry, nil
	})
}
