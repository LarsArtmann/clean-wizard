package di

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_CreatesContainerWithInjector(t *testing.T) {
	container, cleanup := New()
	defer cleanup()

	assert.NotNil(t, container)
	assert.NotNil(t, container.Injector())
}

func TestNew_CleanupCallsShutdown(t *testing.T) {
	container, cleanup := New()
	_ = container

	cleanup()

	// After shutdown, the injector should still exist but services are shut down.
	// Calling Shutdown again is safe (no-op).
	assert.NotPanics(t, func() {
		_ = container.Shutdown()
	})
}

func TestRegisterAllServices_RegistersConfig(t *testing.T) {
	container, cleanup := New()
	defer cleanup()

	cfg := &domain.Config{}
	settings := RunSettings{Verbose: true, DryRun: false}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	resolved, err := Config(container.Injector())
	require.NoError(t, err)
	assert.Same(t, cfg, resolved)
}

func TestRegisterAllServices_RegistersSettings(t *testing.T) {
	container, cleanup := New()
	defer cleanup()

	cfg := &domain.Config{}
	settings := RunSettings{Verbose: true, DryRun: true}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	resolved, err := Settings(container.Injector())
	require.NoError(t, err)
	assert.Equal(t, settings, resolved)
}

func TestRegisterAllServices_RegistersCleanerRegistry(t *testing.T) {
	container, cleanup := New()
	defer cleanup()

	cfg := &domain.Config{}
	settings := RunSettings{Verbose: false, DryRun: true}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	registry, err := CleanerRegistry(container.Injector())
	require.NoError(t, err)
	assert.NotNil(t, registry)

	// Verify registry contains expected cleaners
	names := registry.Names()
	assert.NotEmpty(t, names)

	// Check a specific cleaner is registered
	c, ok := registry.Get(cleaner.CleanerCargo)
	assert.True(t, ok)
	assert.NotNil(t, c)
}

func TestOverrideRegistry_ReplacesRegistry(t *testing.T) {
	container, cleanup := New()
	defer cleanup()

	cfg := &domain.Config{}
	settings := RunSettings{}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	// Create a custom registry with just one cleaner
	mockRegistry := cleaner.NewRegistry()
	mockRegistry.Register("test-only", cleaner.NewCargoCleaner(false, false))

	OverrideRegistry(container.Injector(), mockRegistry)

	resolved, err := CleanerRegistry(container.Injector())
	require.NoError(t, err)
	assert.Same(t, mockRegistry, resolved)
}

func TestOverrideSettings_ReplacesSettings(t *testing.T) {
	container, cleanup := New()
	defer cleanup()

	cfg := &domain.Config{}
	originalSettings := RunSettings{Verbose: false, DryRun: false}

	err := RegisterAllServices(container.Injector(), cfg, originalSettings)
	require.NoError(t, err)

	newSettings := RunSettings{Verbose: true, DryRun: true}
	OverrideSettings(container.Injector(), newSettings)

	resolved, err := Settings(container.Injector())
	require.NoError(t, err)
	assert.Equal(t, newSettings, resolved)
}

func TestConfig_ReturnsErrorWhenNotRegistered(t *testing.T) {
	injector := do.New()

	_, err := Config(injector)
	assert.Error(t, err)
}

func TestCleanerRegistry_ReturnsErrorWhenNotRegistered(t *testing.T) {
	injector := do.New()

	_, err := CleanerRegistry(injector)
	assert.Error(t, err)
}
