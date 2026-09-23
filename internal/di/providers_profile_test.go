package di

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveProfileOperationSettings(t *testing.T) {
	t.Parallel()

	dockerSettings := &domain.OperationSettings{
		Docker: &domain.DockerSettings{PruneMode: domain.DockerPruneVolumes},
	}

	cfg := &domain.Config{
		Profiles: map[string]*domain.Profile{
			"daily": {
				Operations: []domain.CleanupOperation{
					{Name: "docker", Settings: dockerSettings},
				},
			},
		},
	}

	t.Run("no selected profile yields nil settings", func(t *testing.T) {
		t.Parallel()

		injector := do.New()
		defer injector.Shutdown()

		do.ProvideValue(injector, cfg)

		settings, err := resolveProfileOperationSettings(injector, RunSettings{Verbose: false, DryRun: true})
		require.NoError(t, err)
		assert.Nil(t, settings)
	})

	t.Run("selected profile yields merged settings", func(t *testing.T) {
		t.Parallel()

		injector := do.New()
		defer injector.Shutdown()

		do.ProvideValue(injector, cfg)

		settings, err := resolveProfileOperationSettings(injector, RunSettings{Profile: "daily"})
		require.NoError(t, err)
		require.NotNil(t, settings)
		assert.Same(t, dockerSettings.Docker, settings.Docker)
	})

	t.Run("unknown profile yields nil settings", func(t *testing.T) {
		t.Parallel()

		injector := do.New()
		defer injector.Shutdown()

		do.ProvideValue(injector, cfg)

		settings, err := resolveProfileOperationSettings(injector, RunSettings{Profile: "weekly"})
		require.NoError(t, err)
		assert.Nil(t, settings)
	})

	t.Run("missing config registration is a rejection", func(t *testing.T) {
		t.Parallel()

		injector := do.New()
		defer injector.Shutdown()

		_, err := resolveProfileOperationSettings(injector, RunSettings{Profile: "daily"})
		require.Error(t, err)
	})
}

func TestRegisterAllServices_ProfileSettingsReachRegistry(t *testing.T) {
	t.Parallel()

	cfg := &domain.Config{
		Profiles: map[string]*domain.Profile{
			"daily": {
				Operations: []domain.CleanupOperation{
					{
						Name: "docker",
						Settings: &domain.OperationSettings{
							Docker: &domain.DockerSettings{PruneMode: domain.DockerPruneVolumes},
						},
					},
				},
			},
		},
	}

	t.Run("profile settings build a valid registry", func(t *testing.T) {
		t.Parallel()

		container, cleanup := New()
		defer cleanup()

		err := RegisterAllServices(container.Injector(), cfg, RunSettings{Profile: "daily"})
		require.NoError(t, err)

		registry, err := CleanerRegistry(container.Injector())
		require.NoError(t, err)

		names := registry.Names()
		assert.NotEmpty(t, names)
		assert.Contains(t, names, cleaner.CleanerDocker)
	})

	t.Run("unknown profile falls back to defaults", func(t *testing.T) {
		t.Parallel()

		container, cleanup := New()
		defer cleanup()

		err := RegisterAllServices(container.Injector(), cfg, RunSettings{Profile: "weekly"})
		require.NoError(t, err)

		registry, err := CleanerRegistry(container.Injector())
		require.NoError(t, err)
		assert.NotEmpty(t, registry.Names())
	})
}
