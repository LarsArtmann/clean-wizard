package di

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/cargo"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_CreatesContainerWithInjector(t *testing.T) {
	t.Parallel()

	container, cleanup := New()
	defer cleanup()

	assert.NotNil(t, container)
	assert.NotNil(t, container.Injector())
}

func TestNew_CleanupCallsShutdown(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	container, cleanup := New()
	defer cleanup()

	cfg := &types.Config{}
	settings := RunSettings{Verbose: true, DryRun: false}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	resolved, err := Config(container.Injector())
	require.NoError(t, err)
	assert.Same(t, cfg, resolved)
}

func TestRegisterAllServices_RegistersSettings(t *testing.T) {
	t.Parallel()

	container, cleanup := New()
	defer cleanup()

	cfg := &types.Config{}
	settings := RunSettings{Verbose: true, DryRun: true}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	resolved, err := Settings(container.Injector())
	require.NoError(t, err)
	assert.Equal(t, settings, resolved)
}

func TestRegisterAllServices_RegistersCleanerRegistry(t *testing.T) {
	t.Parallel()

	container, cleanup := New()
	defer cleanup()

	cfg := &types.Config{}
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
	t.Parallel()

	container, cleanup := New()
	defer cleanup()

	cfg := &types.Config{}
	settings := RunSettings{}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	// Create a custom registry with just one cleaner
	mockRegistry := cleaner.NewRegistry()
	mockRegistry.Register("test-only", cargo.NewCargoCleaner(false, false))

	OverrideRegistry(container.Injector(), mockRegistry)

	resolved, err := CleanerRegistry(container.Injector())
	require.NoError(t, err)
	assert.Same(t, mockRegistry, resolved)
}

func TestOverrideSettings_ReplacesSettings(t *testing.T) {
	t.Parallel()

	container, cleanup := New()
	defer cleanup()

	cfg := &types.Config{}
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
	t.Parallel()

	injector := do.New()

	_, err := Config(injector)
	assert.Error(t, err)
}

func TestCleanerRegistry_ReturnsErrorWhenNotRegistered(t *testing.T) {
	t.Parallel()

	injector := do.New()

	_, err := CleanerRegistry(injector)
	assert.Error(t, err)
}

func TestRegisterAllServices_ResolvesSingleCleanerFromContainer(t *testing.T) {
	t.Parallel()

	container, cleanup := New()
	defer cleanup()

	cfg := &types.Config{}
	settings := RunSettings{Verbose: true, DryRun: true}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	c, err := Cleaner(container.Injector(), cleaner.CleanerCargo)
	require.NoError(t, err)
	assert.Equal(t, cleaner.CleanerCargo, c.Name())

	_, err = Cleaner(container.Injector(), "does-not-exist")
	assert.Error(t, err)
}

func TestRegisterAllServices_AdapterInterfaceAliases(t *testing.T) {
	t.Parallel()

	container, cleanup := New()
	defer cleanup()

	cfg := &types.Config{}
	settings := RunSettings{Verbose: false, DryRun: true}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	store, err := NixStore(container.Injector())
	require.NoError(t, err)
	assert.Implements(t, (*adapters.NixStore)(nil), store)
	assert.IsType(t, &adapters.NixAdapter{}, store)
}

// fakeNixStore is a container-injected test double for adapters.NixStore. It
// records the calls the nix cleaner makes, so tests can assert the cleaner
// consumes the store resolved from the DI container.
type fakeNixStore struct {
	dryRunCalls  []bool
	availableAsk int
}

func (f *fakeNixStore) SetDryRun(dryRun bool) { f.dryRunCalls = append(f.dryRunCalls, dryRun) }

func (f *fakeNixStore) IsAvailable(context.Context) bool {
	f.availableAsk++

	return true
}

func (f *fakeNixStore) GetStoreSize(context.Context) result.Result[int64] {
	return result.Ok[int64](0)
}

func (f *fakeNixStore) ListGenerations(context.Context) result.Result[[]types.NixGeneration] {
	return result.Ok[[]types.NixGeneration](nil)
}

func (f *fakeNixStore) RemoveGeneration(context.Context, types.NixGenerationID) result.Result[types.CleanResult] {
	return result.Ok[types.CleanResult](types.CleanResult{})
}

func (f *fakeNixStore) CollectGarbage(context.Context) result.Result[types.CleanResult] {
	return result.Ok[types.CleanResult](types.CleanResult{})
}

func TestNixCleaner_UsesContainerInjectedStore(t *testing.T) {
	t.Parallel()

	container, cleanup := New()
	defer cleanup()

	cfg := &types.Config{}
	settings := RunSettings{Verbose: false, DryRun: true}

	err := RegisterAllServices(container.Injector(), cfg, settings)
	require.NoError(t, err)

	injectedStore := &fakeNixStore{}
	do.Override(container.Injector(), func(i do.Injector) (adapters.NixStore, error) {
		return injectedStore, nil
	})

	nixCleaner, err := Cleaner(container.Injector(), cleaner.CleanerNix)
	require.NoError(t, err)
	require.NotNil(t, nixCleaner)
	assert.Equal(t, cleaner.CleanerNix, nixCleaner.Name())

	// Construction hands the store to the cleaner via SetDryRun with the run
	// flag, proving the container-injected double is the store in use.
	assert.Equal(t, []bool{true}, injectedStore.dryRunCalls)

	// Availability checks delegate to the injected store, not a real adapter.
	assert.True(t, nixCleaner.IsAvailable(t.Context()))
	assert.Equal(t, 1, injectedStore.availableAsk)
}
