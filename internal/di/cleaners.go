package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/factory"
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
func registerCleanerProviders(injector do.Injector) {
	for _, rc := range factory.CleanerRegistrations() {
		constructor := rc.Constructor

		do.ProvideNamed(injector, cleanerServiceName(rc.Name), func(i do.Injector) (cleaner.Cleaner, error) {
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

			return constructor(settings.Verbose, settings.DryRun, operationSettings)
		})
	}
}

// registerCleanerRegistryFromProviders provides the *cleaner.Registry by
// resolving every per-cleaner service from the container. Individual cleaner
// providers stay independently invokable while the registry preserves the
// canonical registration order.
func registerCleanerRegistryFromProviders(injector do.Injector) {
	do.Provide(injector, func(i do.Injector) (*cleaner.Registry, error) {
		registry := cleaner.NewRegistry()

		for _, rc := range factory.CleanerRegistrations() {
			c, err := do.InvokeNamed[cleaner.Cleaner](i, cleanerServiceName(rc.Name))
			if err != nil {
				return nil, errorfamily.WrapRejectionf(
					err,
					"di.resolve_cleaner_for_registry",
					"failed to resolve cleaner=%s from container",
					rc.Name,
				)
			}

			registry.Register(rc.Name, c)
		}

		return registry, nil
	})
}
