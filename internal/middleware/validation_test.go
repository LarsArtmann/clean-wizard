package middleware

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

func TestValidationMiddleware(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewValidationMiddleware()

	t.Run("ValidScanRequest", func(t *testing.T) {
		t.Parallel()

		req := types.ScanRequest{
			Type:      types.ScanTypeNixStore,
			Recursive: enums.ScanModeRecursive,
			Limit:     100,
		}

		result := validator.ValidateScanRequest(ctx, req)
		assert.True(t, result.IsOk())
	})

	t.Run("InvalidScanRequest", func(t *testing.T) {
		t.Parallel()

		req := types.ScanRequest{
			Type:      "invalid",
			Recursive: enums.ScanModeRecursive,
			Limit:     -1,
		}

		result := validator.ValidateScanRequest(ctx, req)
		assert.True(t, result.IsErr())
		assert.Contains(t, result.Error().Error(), "invalid scan type")
	})

	t.Run("ValidCleanRequest", func(t *testing.T) {
		t.Parallel()

		req := types.CleanRequest{
			Items: []types.ScanItem{
				{Path: "/tmp/file", Size: 1024, Created: time.Now(), ScanType: types.ScanTypeTemp},
			},
			Strategy: enums.StrategyConservativeType,
		}

		result := validator.ValidateCleanRequest(ctx, req)
		assert.True(t, result.IsOk())
	})

	t.Run("InvalidCleanRequest", func(t *testing.T) {
		t.Parallel()

		req := types.CleanRequest{
			Items:    []types.ScanItem{},
			Strategy: enums.CleanStrategyType(999), // Invalid strategy value
		}

		result := validator.ValidateCleanRequest(ctx, req)
		assert.True(t, result.IsErr())
		assert.Contains(t, result.Error().Error(), "invalid strategy")
	})

	t.Run("ValidCleanerSettings", func(t *testing.T) {
		t.Parallel()

		cleaner := &mockCleaner{}
		settings := &operations.OperationSettings{
			NixGenerations: &operations.NixGenerationsSettings{Generations: 3},
		}

		result := validator.ValidateCleanerSettings(ctx, cleaner, settings)
		assert.True(t, result.IsOk())
	})

	t.Run("InvalidCleanerSettings", func(t *testing.T) {
		t.Parallel()

		cleaner := &mockCleaner{}
		settings := &operations.OperationSettings{
			NixGenerations: &operations.NixGenerationsSettings{Generations: -1},
		}

		result := validator.ValidateCleanerSettings(ctx, cleaner, settings)
		assert.True(t, result.IsErr())
		assert.Contains(t, result.Error().Error(), "must be at least 1")
	})
}

// mockCleaner implements types.OperationHandler for testing.
type mockCleaner struct{}

func (m *mockCleaner) Type() operations.OperationType {
	return operations.OperationTypeNixGenerations
}

func (m *mockCleaner) IsAvailable(ctx context.Context) bool {
	return true
}

func (m *mockCleaner) GetStoreSize(ctx context.Context) int64 {
	return 1000
}

func (m *mockCleaner) ValidateSettings(settings *operations.OperationSettings) error {
	if settings != nil && settings.NixGenerations != nil &&
		settings.NixGenerations.Generations < 1 {
		return fmt.Errorf(
			"Generations to keep must be at least 1, got: %d",
			settings.NixGenerations.Generations,
		)
	}

	return nil
}
