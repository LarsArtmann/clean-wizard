package middleware

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestValidationMiddleware(t *testing.T) {
	ctx := context.Background()
	validator := NewValidationMiddleware()

	t.Run("ValidScanRequest", func(t *testing.T) {
		req := domain.ScanRequest{
			Type:      domain.ScanTypeNixStore,
			Recursive: true,
			Limit:     100,
		}

		result := validator.ValidateScanRequest(ctx, req)
		assert.True(t, result.IsOk())
	})

	t.Run("InvalidScanRequest", func(t *testing.T) {
		req := domain.ScanRequest{
			Type:      "invalid",
			Recursive: true,
			Limit:     -1,
		}

		result := validator.ValidateScanRequest(ctx, req)
		assert.True(t, result.IsErr())
		assert.Contains(t, result.Error().Error(), "Invalid scan type")
	})

	t.Run("ValidCleanRequest", func(t *testing.T) {
		req := domain.CleanRequest{
			Items:    []domain.ScanItem{{Path: "/tmp/file", Size: 1024, Created: time.Now(), ScanType: domain.ScanTypeTemp}},
			Strategy: domain.StrategyConservative,
		}

		result := validator.ValidateCleanRequest(ctx, req)
		assert.True(t, result.IsOk())
	})

	t.Run("InvalidCleanRequest", func(t *testing.T) {
		req := domain.CleanRequest{
			Items:    []domain.ScanItem{},
			Strategy: domain.CleanStrategy(999), // Invalid strategy value
		}

		result := validator.ValidateCleanRequest(ctx, req)
		assert.True(t, result.IsErr())
		assert.Contains(t, result.Error().Error(), "Invalid strategy")
	})

	t.Run("ValidCleanerSettings", func(t *testing.T) {
		cleaner := &mockCleaner{}
		settings := &domain.OperationSettings{
			NixGenerations: &domain.NixGenerationsSettings{Generations: 3},
		}

		result := validator.ValidateCleanerSettings(ctx, cleaner, settings)
		assert.True(t, result.IsOk())
	})

	t.Run("InvalidCleanerSettings", func(t *testing.T) {
		cleaner := &mockCleaner{}
		settings := &domain.OperationSettings{
			NixGenerations: &domain.NixGenerationsSettings{Generations: -1},
		}

		result := validator.ValidateCleanerSettings(ctx, cleaner, settings)
		assert.True(t, result.IsErr())
		assert.Contains(t, result.Error().Error(), "must be at least 1")
	})
}

// mockCleaner implements domain.Cleaner for testing
type mockCleaner struct{}

func (m *mockCleaner) IsAvailable(ctx context.Context) bool {
	return true
}

func (m *mockCleaner) GetStoreSize(ctx context.Context) int64 {
	return 1000
}

func (m *mockCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings != nil && settings.NixGenerations != nil && settings.NixGenerations.Generations < 1 {
		return fmt.Errorf("Generations to keep must be at least 1, got: %d", settings.NixGenerations.Generations)
	}
	return nil
}
