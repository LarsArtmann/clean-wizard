package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/samber/do/v2"
)

// Config resolves the application configuration from the DI container.
func Config(i do.Injector) (*domain.Config, error) {
	cfg, err := do.Invoke[*domain.Config](i)
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "di.resolve_config", "failed to resolve config service")
	}

	return cfg, nil
}

// Settings resolves the runtime settings (verbose, dryRun) from the DI container.
func Settings(i do.Injector) (RunSettings, error) {
	settings, err := do.Invoke[RunSettings](i)
	if err != nil {
		return RunSettings{}, errorfamily.WrapRejection(
			err,
			"di.resolve_settings",
			"failed to resolve run settings service",
		)
	}

	return settings, nil
}

// CleanerRegistry resolves the cleaner registry from the DI container.
// The registry is lazily created on first invocation with the registered RunSettings.
func CleanerRegistry(i do.Injector) (*cleaner.Registry, error) {
	registry, err := do.Invoke[*cleaner.Registry](i)
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "di.resolve_registry", "failed to resolve cleaner registry service")
	}

	return registry, nil
}
