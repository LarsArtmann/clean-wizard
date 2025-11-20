package di

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_NewContainer(t *testing.T) {
	ctx := context.Background()

	container := NewContainer(ctx)
	require.NotNil(t, container)

	// Test basic dependencies
	assert.NotNil(t, container.GetConfig())
	assert.NotNil(t, container.GetLogger())
	assert.NotNil(t, container.GetCleaner())
	assert.NotNil(t, container.GetValidationMiddleware())

	// Test default config
	config := container.GetConfig()
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, domain.SafetyLevelEnabled, config.SafetyLevel)
	assert.Equal(t, 50, config.MaxDiskUsage)
	assert.Contains(t, config.Protected, "/System")
	assert.Len(t, config.Profiles, 1)
	assert.Contains(t, config.Profiles, "daily")
}

func TestContainer_UpdateConfig(t *testing.T) {
	ctx := context.Background()
	container := NewContainer(ctx)

	originalConfig := container.GetConfig()
	assert.Equal(t, "1.0.0", originalConfig.Version)

	// Update config
	newConfig := &domain.Config{
		Version:      "2.0.0",
		SafetyLevel:  domain.SafetyLevelDisabled,
		MaxDiskUsage: 80,
		Protected:    []string{"/tmp"},
		Profiles:     map[string]*domain.Profile{},
	}

	container.UpdateConfig(newConfig)
	updatedConfig := container.GetConfig()
	assert.Equal(t, "2.0.0", updatedConfig.Version)
	assert.Equal(t, domain.SafetyLevelDisabled, updatedConfig.SafetyLevel)
	assert.Equal(t, 80, updatedConfig.MaxDiskUsage)
}

func TestContainer_GetCleaner(t *testing.T) {
	ctx := context.Background()
	container := NewContainer(ctx)

	cleaner := container.GetCleaner()
	require.NotNil(t, cleaner)

	// Test cleaner is available (may be false in CI environment)
	// We just ensure the method doesn't panic
	_ = cleaner.IsAvailable(ctx)

	// Test store size (may return mock value)
	storeSize := cleaner.GetStoreSize(ctx)
	assert.GreaterOrEqual(t, storeSize, int64(0))
}

func TestContainer_Shutdown(t *testing.T) {
	ctx := context.Background()
	container := NewContainer(ctx)

	err := container.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestContainer_DefaultProfiles(t *testing.T) {
	ctx := context.Background()
	container := NewContainer(ctx)

	profiles := container.GetConfig().Profiles
	require.NotNil(t, profiles)
	require.Len(t, profiles, 1)

	dailyProfile, exists := profiles["daily"]
	require.True(t, exists)
	require.NotNil(t, dailyProfile)

	assert.Equal(t, "Daily Cleanup", dailyProfile.Name)
	assert.Equal(t, "Daily system cleanup operations", dailyProfile.Description)
	assert.Equal(t, domain.StatusEnabled, dailyProfile.Status)
	require.Len(t, dailyProfile.Operations, 1)

	operation := dailyProfile.Operations[0]
	assert.Equal(t, "nix-generations", operation.Name)
	assert.Equal(t, "Clean Nix generations", operation.Description)
	assert.Equal(t, domain.RiskLow, operation.RiskLevel)
	assert.Equal(t, domain.StatusEnabled, operation.Status)
	require.NotNil(t, operation.Settings)
	require.NotNil(t, operation.Settings.NixGenerations)
	assert.Equal(t, 3, operation.Settings.NixGenerations.Generations)
	assert.Equal(t, domain.OptimizationLevelAggressive, operation.Settings.NixGenerations.Optimization)
}
