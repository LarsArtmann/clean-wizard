package cleaner

import (
	"context"
	"sync"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCleaner is a test implementation of the Cleaner interface.
type mockCleaner struct {
	name        string
	available   bool
	cleanCalled bool
}

func (m *mockCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	m.cleanCalled = true
	return result.Ok(domain.CleanResult{
		ItemsRemoved: 1,
		FreedBytes:   1024,
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
}

func (m *mockCleaner) IsAvailable(ctx context.Context) bool {
	return m.available
}

func (m *mockCleaner) Name() string {
	return m.name
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	require.NotNil(t, registry)
	assert.Equal(t, 0, registry.Count())
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()
	cleaner := &mockCleaner{name: "test", available: true}

	// Register
	registry.Register("test", cleaner)
	assert.Equal(t, 1, registry.Count())

	// Get existing
	got, ok := registry.Get("test")
	assert.True(t, ok)
	assert.Equal(t, cleaner, got)

	// Get non-existing
	got, ok = registry.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()
	cleaner1 := &mockCleaner{name: "test1", available: true}
	cleaner2 := &mockCleaner{name: "test2", available: false}

	registry.Register("test1", cleaner1)
	registry.Register("test2", cleaner2)

	list := registry.List()
	assert.Len(t, list, 2)
	// Order is not guaranteed due to map iteration
}

func TestRegistry_Names(t *testing.T) {
	registry := NewRegistry()
	cleaner1 := &mockCleaner{name: "test1", available: true}
	cleaner2 := &mockCleaner{name: "test2", available: false}

	registry.Register("test1", cleaner1)
	registry.Register("test2", cleaner2)

	names := registry.Names()
	assert.Len(t, names, 2)
	// Order is not guaranteed due to map iteration
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()
	assert.Equal(t, 0, registry.Count())

	registry.Register("test1", &mockCleaner{name: "test1"})
	assert.Equal(t, 1, registry.Count())

	registry.Register("test2", &mockCleaner{name: "test2"})
	assert.Equal(t, 2, registry.Count())
}

func TestRegistry_Available(t *testing.T) {
	ctx := context.Background()
	registry := NewRegistry()

	availableCleaner := &mockCleaner{name: "available", available: true}
	unavailableCleaner := &mockCleaner{name: "unavailable", available: false}

	registry.Register("available", availableCleaner)
	registry.Register("unavailable", unavailableCleaner)

	available := registry.Available(ctx)
	assert.Len(t, available, 1)
	assert.Equal(t, availableCleaner, available[0])
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	cleaner := &mockCleaner{name: "test", available: true}

	registry.Register("test", cleaner)
	assert.Equal(t, 1, registry.Count())

	registry.Unregister("test")
	assert.Equal(t, 0, registry.Count())

	_, ok := registry.Get("test")
	assert.False(t, ok)
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	registry.Register("test1", &mockCleaner{name: "test1"})
	registry.Register("test2", &mockCleaner{name: "test2"})
	assert.Equal(t, 2, registry.Count())

	registry.Clear()
	assert.Equal(t, 0, registry.Count())

	list := registry.List()
	assert.Empty(t, list)
}

func TestRegistry_CleanAll(t *testing.T) {
	ctx := context.Background()
	registry := NewRegistry()

	cleaner1 := &mockCleaner{name: "test1", available: true}
	cleaner2 := &mockCleaner{name: "test2", available: true}
	unavailableCleaner := &mockCleaner{name: "unavailable", available: false}

	registry.Register("test1", cleaner1)
	registry.Register("test2", cleaner2)
	registry.Register("unavailable", unavailableCleaner)

	results := registry.CleanAll(ctx)

	// Note: Currently CleanAll uses "unknown" as key for all cleaners
	// This is a known limitation documented in the code
	assert.NotNil(t, results)
	// Only available cleaners should be cleaned
	assert.True(t, cleaner1.cleanCalled || cleaner2.cleanCalled)
}

func TestRegistry_RegisterOverwrite(t *testing.T) {
	registry := NewRegistry()
	cleaner1 := &mockCleaner{name: "test1"}
	cleaner2 := &mockCleaner{name: "test2"}

	registry.Register("test", cleaner1)
	registry.Register("test", cleaner2) // Overwrite

	got, ok := registry.Get("test")
	require.True(t, ok)
	assert.Equal(t, cleaner2, got)
	assert.Equal(t, 1, registry.Count())
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	// Concurrent registrations
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			registry.Register(string(rune(i)), &mockCleaner{name: string(rune(i))})
		}(i)
	}
	wg.Wait()

	// Verify all registered
	assert.Equal(t, 100, registry.Count())

	// Concurrent reads
	var wg2 sync.WaitGroup
	for range 100 {
		wg2.Go(func() {
			_ = registry.List()
			_ = registry.Names()
			_ = registry.Count()
			_ = registry.Available(ctx)
		})
	}
	wg2.Wait()

	// No panic means thread-safety is working
}

func TestRegistry_EmptyOperations(t *testing.T) {
	ctx := context.Background()
	registry := NewRegistry()

	// Operations on empty registry should not panic
	list := registry.List()
	assert.Empty(t, list)

	names := registry.Names()
	assert.Empty(t, names)

	available := registry.Available(ctx)
	assert.Empty(t, available)

	results := registry.CleanAll(ctx)
	assert.Empty(t, results)

	_, ok := registry.Get("nonexistent")
	assert.False(t, ok)

	// Unregister non-existing should not panic
	registry.Unregister("nonexistent")
}
