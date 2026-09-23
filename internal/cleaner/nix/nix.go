package nix

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// Byte conversion constants.
const (
	bytesPerMB = 1024 * 1024
	bytesPerGB = 1024 * 1024 * 1024
)

const (
	// NixMockStoreSizeGB is the mock store size in GB for unavailable Nix.
	NixMockStoreSizeGB = 300
	// NixMaxGenerationsToKeep is the maximum allowed generations to keep.
	NixMaxGenerationsToKeep = 10
	// NixDryRunBytesPerGeneration is the estimated bytes freed per generation in dry-run mode.
	NixDryRunBytesPerGeneration = 50 * bytesPerMB

	// Mock generation IDs for testing when Nix is unavailable.
	mockGenerationIDCurrent = 300
	mockGenerationIDRecent1 = 299
	mockGenerationIDRecent2 = 298
	mockGenerationIDOlder1  = 297
	mockGenerationIDOlder2  = 296

	// Mock generation age offsets in hours.
	mockGenerationAgeCurrent = 24
	mockGenerationAgeRecent  = 48
	mockGenerationAgeOlder   = 72
	mockGenerationAgeOld     = 96
	mockGenerationAgeVeryOld = 120
)

// NixCleaner handles Nix package manager cleanup with proper type safety.
type NixCleaner struct {
	cleaner.CleanerBase

	adapter   *adapters.NixAdapter
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
		adapter:     adapters.NewNixAdapter(0, 0),
		CleanerBase: cleaner.NewCleanerBase(verbose, dryRun),
		keepCount:   kc,
	}
	nc.adapter.SetDryRun(dryRun) // Pass dry-run to adapter

	return nc
}

// Type returns the operation type for Nix cleaner.
func (nc *NixCleaner) Type() operations.OperationType {
	return operations.OperationTypeNixGenerations
}

// Name returns the unique identifier for this cleaner.
func (nc *NixCleaner) Name() string {
	return "nix"
}

// IsAvailable checks if Nix cleaner is available.
func (nc *NixCleaner) IsAvailable(ctx context.Context) bool {
	return nc.adapter.IsAvailable(ctx)
}

// Scan scans for Nix generations and returns them as scan items.
func (nc *NixCleaner) Scan(ctx context.Context) result.Result[[]types.ScanItem] {
	genResult := nc.ListGenerations(ctx)
	if genResult.IsErr() {
		return result.Err[[]types.ScanItem](genResult.Error())
	}

	generations := genResult.Value()
	items := make([]types.ScanItem, 0, len(generations))

	for _, gen := range generations {
		items = append(items, types.ScanItem{
			Path:     gen.Path,
			Size:     0, // Individual generation size is hard to determine
			Created:  gen.Date,
			ScanType: types.ScanTypeNixStore,
		})
	}

	return result.Ok(items)
}

// Clean implements the Cleaner interface.
// It removes old Nix generations, keeping the configured number of generations.
func (nc *NixCleaner) Clean(ctx context.Context) result.Result[types.CleanResult] {
	return nc.CleanOldGenerations(ctx, nc.keepCount)
}

// GetStoreSize gets Nix store size with type safety.
func (nc *NixCleaner) GetStoreSize(ctx context.Context) int64 {
	if !nc.adapter.IsAvailable(ctx) {
		return int64(NixMockStoreSizeGB * bytesPerGB)
	}

	storeSizeResult := nc.adapter.GetStoreSize(ctx)
	if storeSizeResult.IsErr() {
		return 0
	}

	return storeSizeResult.Value()
}

// ValidateSettings validates Nix cleaner settings with type safety.
func (nc *NixCleaner) ValidateSettings(settings *operations.OperationSettings) error {
	return cleaner.ValidateOptionalSettings(
		settings,
		func(s *operations.OperationSettings) *operations.NixGenerationsSettings { return s.NixGenerations },
		func(n *operations.NixGenerationsSettings) error {
			if n.Generations < 1 {
				return fmt.Errorf(
					"generations to keep must be at least 1, got: %d",
					n.Generations,
				)
			}

			if n.Generations > NixMaxGenerationsToKeep {
				return fmt.Errorf(
					"generations to keep must not exceed %d, got: %d",
					NixMaxGenerationsToKeep,
					n.Generations,
				)
			}

			return nil
		},
	)
}

// ListGenerations lists Nix generations with proper type safety.
func (nc *NixCleaner) ListGenerations(ctx context.Context) result.Result[[]types.NixGeneration] {
	// Check availability first
	if !nc.adapter.IsAvailable(ctx) {
		// Return mock data for CI/testing - proper adapter pattern eliminates ghost system
		return result.MockSuccess([]types.NixGeneration{
			{
				ID:   mockGenerationIDCurrent,
				Path: "/nix/var/nix/profiles/default-300-link",
				Date: time.Now().
					Add(-mockGenerationAgeCurrent * time.Hour),
				Current: enums.GenerationStatusCurrent,
			},
			{
				ID:   mockGenerationIDRecent1,
				Path: "/nix/var/nix/profiles/default-299-link",
				Date: time.Now().
					Add(-mockGenerationAgeRecent * time.Hour),
				Current: enums.GenerationStatusHistorical,
			},
			{
				ID:   mockGenerationIDRecent2,
				Path: "/nix/var/nix/profiles/default-298-link",
				Date: time.Now().
					Add(-mockGenerationAgeOlder * time.Hour),
				Current: enums.GenerationStatusHistorical,
			},
			{
				ID:   mockGenerationIDOlder1,
				Path: "/nix/var/nix/profiles/default-297-link",
				Date: time.Now().
					Add(-mockGenerationAgeOld * time.Hour),
				Current: enums.GenerationStatusHistorical,
			},
			{
				ID:   mockGenerationIDOlder2,
				Path: "/nix/var/nix/profiles/default-296-link",
				Date: time.Now().
					Add(-mockGenerationAgeVeryOld * time.Hour),
				Current: enums.GenerationStatusHistorical,
			},
		}, "Nix not available - using mock data")
	}

	// Only call adapter if available
	return nc.adapter.ListGenerations(ctx)
}

// CleanOldGenerations removes old Nix generations using centralized conversions.
func (nc *NixCleaner) CleanOldGenerations(
	ctx context.Context,
	keepCount int,
) result.Result[types.CleanResult] {
	// Get generations first
	genResult := nc.ListGenerations(ctx)
	if genResult.IsErr() {
		return conversions.ToCleanResultFromError(genResult.Error())
	}

	generations := genResult.Value()

	// Count and remove old generations
	toRemove := countOldGenerations(generations, keepCount)

	if nc.GetDryRun() {
		// Use real store size to calculate average generation size
		storeSize := nc.GetStoreSize(ctx)
		// Calculate average generation size (avoid division by zero)
		var avgSize int64
		if len(generations) > 0 {
			avgSize = storeSize / int64(len(generations))
		}
		// Estimate bytes to free based on average generation size
		estimatedBytes := avgSize * int64(toRemove)
		cleanResult := conversions.NewCleanResult(
			enums.StrategyDryRunType,
			toRemove,
			estimatedBytes,
		)

		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	if !nc.GetDryRun() && toRemove > 0 {
		// Remove old generations individually to track what's cleaned
		results := make([]types.CleanResult, 0, toRemove)
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
		combinedResult.Strategy = enums.StrategyAggressiveType

		return result.Ok(combinedResult)
	}

	// Dry-run or no generations to remove - use centralized conversion
	estimatedBytes := int64(toRemove) * NixDryRunBytesPerGeneration
	cleanResult := conversions.NewCleanResult(
		enums.StrategyDryRunType,
		toRemove,
		estimatedBytes,
	)

	return result.Ok(cleanResult)
}

// countOldGenerations counts generations to remove (keeping current + N others).
func countOldGenerations(generations []types.NixGeneration, keepCount int) int {
	if len(generations) <= keepCount {
		return 0
	}

	return len(generations) - keepCount
}

// GetKeepCount returns the configured number of generations to keep.
func (nc *NixCleaner) GetKeepCount() int { return nc.keepCount }
