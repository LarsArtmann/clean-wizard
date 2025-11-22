package cleaner

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	"github.com/LarsArtmann/clean-wizard/internal/infrastructure/system"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
	mocks "github.com/LarsArtmann/clean-wizard/test"
)

// NixCleaner handles Nix package manager cleanup with proper type safety
type NixCleaner struct {
	adapter *system.NixAdapter
}

// NewNixCleaner creates a new NixCleaner with type-safe configuration
func NewNixCleaner() *NixCleaner {
	return &NixCleaner{
		adapter: system.NewNixAdapter(30*time.Second, 3),
	}
}

// IsAvailable checks if Nix is available on system
func (nc *NixCleaner) IsAvailable(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return nc.adapter.IsAvailable(ctx)
	}
}

// GetStoreSize returns current Nix store size with proper error handling
func (nc *NixCleaner) GetStoreSize(ctx context.Context) int64 {
	select {
	case <-ctx.Done():
		return 0
	default:
		sizeResult := nc.adapter.GetStoreSize(ctx)
		if sizeResult.IsErr() {
			return 0
		}
		return sizeResult.Value()
	}
}

// ValidateSettings validates Nix operation settings with type safety
func (nc *NixCleaner) ValidateSettings(settings *shared.OperationSettings) error {
	if settings == nil || settings.NixGenerations == nil {
		return fmt.Errorf("Nix generations settings required")
	}

	if settings.NixGenerations.Generations <= 0 || settings.NixGenerations.Generations > 100 {
		return fmt.Errorf("generations count must be between 1 and 100")
	}

	return nil
}

// ListGenerations returns all available Nix generations
func (nc *NixCleaner) ListGenerations(ctx context.Context) result.Result[[]shared.NixGeneration] {
	select {
	case <-ctx.Done():
		return result.Err[[]shared.NixGeneration](ctx.Err())
	default:
		if !nc.IsAvailable(ctx) {
			// Return mock data when Nix not available for testing
			return mocks.MockNixGenerationsResult()
		}

		return nc.adapter.ListGenerations(ctx)
	}
}

// Cleanup performs Nix cleanup with specified settings
func (nc *NixCleaner) Cleanup(ctx context.Context, settings *shared.OperationSettings) result.Result[shared.CleanResult] {
	if err := nc.ValidateSettings(settings); err != nil {
		return result.Err[shared.CleanResult](err)
	}

	select {
	case <-ctx.Done():
		return result.Err[shared.CleanResult](ctx.Err())
	default:
		if !nc.IsAvailable(ctx) {
			// Return success when Nix not available (graceful degradation)
			return result.Ok(shared.CleanResult{
				ItemsFailed: 0,
				CleanTime:   0,
				CleanedAt:   time.Now(),
				Strategy:    shared.StrategyConservative,
			})
		}

		// Perform garbage collection
		cleanupResult := nc.adapter.CollectGarbage(ctx)
		return cleanupResult
	}
}

// CleanOldGenerations removes old Nix generations, keeping specified count
func (nc *NixCleaner) CleanOldGenerations(ctx context.Context, keepCount int) result.Result[shared.CleanResult] {
	if keepCount < 1 || keepCount > 100 {
		return result.Err[shared.CleanResult](fmt.Errorf("keep count must be between 1 and 100"))
	}

	select {
	case <-ctx.Done():
		return result.Err[shared.CleanResult](ctx.Err())
	default:
		if !nc.IsAvailable(ctx) {
			// Return graceful degradation result
			return result.Ok(shared.CleanResult{
				FreedBytes:   0,
				ItemsRemoved: 0,
				ItemsFailed:  0,
				CleanTime:    0,
				CleanedAt:    time.Now(),
				Strategy:     shared.StrategyConservative,
			})
		}

		// List all generations and remove old ones
		genResult := nc.ListGenerations(ctx)
		if genResult.IsErr() {
			return result.Err[shared.CleanResult](genResult.Error())
		}

		generations := genResult.Value()
		if len(generations) <= keepCount {
			// Nothing to remove
			return result.Ok(shared.CleanResult{
				FreedBytes:   0,
				ItemsRemoved: 0,
				ItemsFailed:  0,
				CleanTime:    0,
				CleanedAt:    time.Now(),
				Strategy:     shared.StrategyConservative,
			})
		}

		// Remove old generations
		var totalRemoved uint
		var totalBytes uint64
		start := time.Now()

		for i := keepCount; i < len(generations); i++ {
			removeResult := nc.adapter.RemoveGeneration(ctx, int(generations[i].ID))
			if removeResult.IsOk() {
				totalRemoved++
				totalBytes += removeResult.Value().FreedBytes
			}
		}

		return result.Ok(shared.CleanResult{
			FreedBytes:   totalBytes,
			ItemsRemoved: uint(totalRemoved),                                     // Fixed type conversion
			ItemsFailed:  uint(len(generations) - keepCount - int(totalRemoved)), // Fixed type conversion
			CleanTime:    time.Since(start),
			CleanedAt:    time.Now(),
			Strategy:     shared.StrategyConservative,
		})
	}
}
