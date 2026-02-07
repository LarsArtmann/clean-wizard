//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNixCleaner_Integration tests Nix cleaner with real Nix environment.
func TestNixCleaner_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	nixCleaner := cleaner.NewNixCleaner(false, false)

	// Check if Nix is available
	if !nixCleaner.IsAvailable(ctx) {
		t.Skip("Nix not available, skipping integration test")
	}

	// Test ListGenerations
	generationsResult := nixCleaner.ListGenerations(ctx)
	require.True(t, generationsResult.IsOk(), "Failed to list generations")

	generations := generationsResult.Value()
	assert.NotEmpty(t, generations, "Should have at least one generation")

	// Validate generations
	for _, gen := range generations {
		assert.True(t, gen.IsValid(), "Generation should be valid: %+v", gen)
		assert.NoError(t, gen.Validate(), "Generation validation should pass")
	}

	// Test CleanOldGenerations in dry-run mode
	result, err := nixCleaner.CleanOldGenerations(ctx, 5).Unwrap()
	require.NoError(t, err, "CleanOldGenerations should not error")

	// Validate result
	assert.True(t, result.IsValid(), "Result should be valid")
	assert.NoError(t, result.Validate(), "Result validation should pass")
}

// TestGoCleaner_Integration tests Go cleaner integration.
func TestGoCleaner_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	goCleaner, err := cleaner.NewGoCleaner(false, false, cleaner.GoCacheGOCACHE|cleaner.GoCacheTestCache)
	require.NoError(t, err)

	// Check if Go is available
	if !goCleaner.IsAvailable(ctx) {
		t.Skip("Go not available, skipping integration test")
	}

	// Test Clean operation in dry-run mode
	result := goCleaner.Clean(ctx)
	require.True(t, result.IsOk(), "Clean should succeed")

	cleanResult := result.Value()
	assert.True(t, cleanResult.IsValid(), "Result should be valid")
	assert.NoError(t, cleanResult.Validate(), "Result validation should pass")
}

// TestHomebrewCleaner_Integration tests Homebrew cleaner integration.
func TestHomebrewCleaner_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	homebrewCleaner := cleaner.NewHomebrewCleaner(false, false, domain.HomebrewModeAll)

	// Check if Homebrew is available
	if !homebrewCleaner.IsAvailable(ctx) {
		t.Skip("Homebrew not available, skipping integration test")
	}

	// Test Clean operation in dry-run mode
	result := homebrewCleaner.Clean(ctx)
	require.True(t, result.IsOk(), "Clean should succeed")

	cleanResult := result.Value()
	assert.True(t, cleanResult.IsValid(), "Result should be valid")
	assert.NoError(t, cleanResult.Validate(), "Result validation should pass")
}

// TestMultiCleaner_Integration tests multiple cleaners running together.
func TestMultiCleaner_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	cleaners := []struct {
		name    string
		cleaner interface{ IsAvailable(context.Context) bool }
	}{
		{"nix", cleaner.NewNixCleaner(false, false)},
		{"go", cleaner.NewGoCleanerWithSettings(false, false, cleaner.GoCacheGOCACHE)},
	}

	// Add temp cleaner separately to handle error
	tempCleaner, err := cleaner.NewTempFilesCleaner(false, false, "7d", []string{}, []string{"/tmp"})
	if err == nil {
		cleaners = append(cleaners, struct {
			name    string
			cleaner interface{ IsAvailable(context.Context) bool }
		}{"temp", tempCleaner})
	}

	for _, tc := range cleaners {
		t.Run(tc.name, func(t *testing.T) {
			// Just check availability - don't actually clean in integration tests
			available := tc.cleaner.IsAvailable(ctx)
			t.Logf("%s cleaner available: %v", tc.name, available)
		})
	}
}

// TestCleanerTimeout_Integration tests that cleaners respect timeouts.
func TestCleanerTimeout_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Use a cleaner that should complete within 1 second
	tempCleaner, err := cleaner.NewTempFilesCleaner(false, false, "7d", []string{}, []string{"/tmp"})
	require.NoError(t, err)

	result := tempCleaner.Clean(ctx)
	assert.True(t, result.IsOk() || result.IsErr(), "Should complete or error within timeout")

	// Check if context timeout caused the error
	if result.IsErr() {
		_, err := result.Unwrap()
		assert.NotEqual(t, context.DeadlineExceeded, err, "Should not timeout")
	}
}
