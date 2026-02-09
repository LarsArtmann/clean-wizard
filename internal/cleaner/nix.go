package cleaner

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// boolToGenerationStatus converts boolean to GenerationStatus enum.
func boolToGenerationStatus(b bool) domain.GenerationStatus {
	if b {
		return domain.GenerationStatusCurrent
	}
	return domain.GenerationStatusHistorical
}

// NixCleaner handles Nix package manager cleanup with proper type safety.
type NixCleaner struct {
	adapter   *adapters.NixAdapter
	verbose   bool
	dryRun    bool
	keepCount int
}

// NewNixCleaner creates Nix cleaner with proper configuration.
func NewNixCleaner(verbose, dryRun bool, keepCount ...int) *NixCleaner {
	// Default keep count is 5
	kc := 5
	if len(keepCount) > 0 {
		kc = keepCount[0]
	}

	nc := &NixCleaner{
		adapter:   adapters.NewNixAdapter(0, 0),
		verbose:   verbose,
		dryRun:    dryRun,
		keepCount: kc,
	}
	nc.adapter.SetDryRun(dryRun) // Pass dry-run to adapter
	return nc
}

// Type returns the operation type for Nix cleaner.
func (nc *NixCleaner) Type() domain.OperationType {
	return domain.OperationTypeNixGenerations
}

// IsAvailable checks if Nix cleaner is available.
func (nc *NixCleaner) IsAvailable(ctx context.Context) bool {
	return nc.adapter.IsAvailable(ctx)
}

// Clean implements the Cleaner interface.
// It removes old Nix generations, keeping the configured number of generations.
func (nc *NixCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	return nc.CleanOldGenerations(ctx, nc.keepCount)
}

// GetStoreSize gets Nix store size with type safety.
func (nc *NixCleaner) GetStoreSize(ctx context.Context) int64 {
	if !nc.adapter.IsAvailable(ctx) {
		return int64(300 * 1024 * 1024 * 1024) // Mock store size
	}

	storeSizeResult := nc.adapter.GetStoreSize(ctx)
	if storeSizeResult.IsErr() {
		return 0
	}
	return storeSizeResult.Value()
}

// ValidateSettings validates Nix cleaner settings with type safety.
func (nc *NixCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.NixGenerations == nil {
		return nil // Settings are optional
	}

	if settings.NixGenerations.Generations < 1 {
		return fmt.Errorf("Generations to keep must be at least 1, got: %d", settings.NixGenerations.Generations)
	}

	if settings.NixGenerations.Generations > 10 {
		return fmt.Errorf("Generations to keep must not exceed %d, got: %d", 10, settings.NixGenerations.Generations)
	}

	return nil
}

// ListGenerations lists Nix generations with proper type safety.
func (nc *NixCleaner) ListGenerations(ctx context.Context) result.Result[[]domain.NixGeneration] {
	// Check availability first
	if !nc.adapter.IsAvailable(ctx) {
		// Return mock data for CI/testing - proper adapter pattern eliminates ghost system
		return result.MockSuccess([]domain.NixGeneration{
			{ID: 300, Path: "/nix/var/nix/profiles/default-300-link", Date: time.Now().Add(-24 * time.Hour), Current: domain.GenerationStatusCurrent},
			{ID: 299, Path: "/nix/var/nix/profiles/default-299-link", Date: time.Now().Add(-48 * time.Hour), Current: domain.GenerationStatusHistorical},
			{ID: 298, Path: "/nix/var/nix/profiles/default-298-link", Date: time.Now().Add(-72 * time.Hour), Current: domain.GenerationStatusHistorical},
			{ID: 297, Path: "/nix/var/nix/profiles/default-297-link", Date: time.Now().Add(-96 * time.Hour), Current: domain.GenerationStatusHistorical},
			{ID: 296, Path: "/nix/var/nix/profiles/default-296-link", Date: time.Now().Add(-120 * time.Hour), Current: domain.GenerationStatusHistorical},
		}, "Nix not available - using mock data")
	}

	// Only call adapter if available
	return nc.adapter.ListGenerations(ctx)
}

// CleanOldGenerations removes old Nix generations using centralized conversions.
func (nc *NixCleaner) CleanOldGenerations(ctx context.Context, keepCount int) result.Result[domain.CleanResult] {
	// Get generations first
	genResult := nc.ListGenerations(ctx)
	if genResult.IsErr() {
		return conversions.ToCleanResultFromError(genResult.Error())
	}

	generations := genResult.Value()

	// Count and remove old generations
	toRemove := countOldGenerations(generations, keepCount)

	if nc.dryRun {
		// Use centralized conversion for dry-run
		estimatedBytes := int64(toRemove * 50 * 1024 * 1024) // Use 50MB per generation
		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, toRemove, estimatedBytes)
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	if !nc.dryRun && toRemove > 0 {
		// Remove old generations individually to track what's cleaned
		results := make([]domain.CleanResult, 0, toRemove)
		start := time.Now()

		for i := len(generations) - toRemove; i < len(generations); i++ {
			// Skip current generation
			if generations[i].Current.IsCurrent() {
				continue
			}

			// Remove this generation
			cleanResult := nc.adapter.RemoveGeneration(ctx, generations[i].ID)
			if cleanResult.IsErr() {
				return conversions.ToCleanResultFromError(cleanResult.Error())
			}

			results = append(results, cleanResult.Value())
		}

		// Run garbage collection to clean up references
		gcResult := nc.adapter.CollectGarbage(ctx)
		if gcResult.IsErr() {
			return conversions.ToCleanResultFromError(gcResult.Error())
		}

		results = append(results, gcResult.Value())

		// Combine all results using centralized function
		combinedResult := conversions.CombineCleanResults(results)
		combinedResult.CleanTime = time.Since(start)
		combinedResult.Strategy = domain.StrategyAggressive

		return result.Ok(combinedResult)
	}

	// Dry-run or no generations to remove - use centralized conversion
	estimatedBytes := int64(toRemove * 50 * 1024 * 1024) // Estimated
	cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, toRemove, estimatedBytes)
	return result.Ok(cleanResult)
}

// countOldGenerations counts generations to remove (keeping current + N others).
func countOldGenerations(generations []domain.NixGeneration, keepCount int) int {
	if len(generations) <= keepCount {
		return 0
	}
	return len(generations) - keepCount
}
