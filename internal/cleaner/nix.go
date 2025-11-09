package cleaner

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// NixCleaner handles Nix package manager cleanup with proper type safety
type NixCleaner struct {
	adapter  *adapters.NixAdapter
	verbose  bool
	dryRun   bool
}

// NewNixCleaner creates Nix cleaner with proper configuration
func NewNixCleaner(verbose bool, dryRun bool) *NixCleaner {
	return &NixCleaner{
		adapter:  adapters.NewNixAdapter(0, 0),
		verbose:  verbose,
		dryRun:   dryRun,
	}
}

// ListGenerations lists Nix generations with proper type safety
func (nc *NixCleaner) ListGenerations(ctx context.Context) result.Result[[]domain.NixGeneration] {
	// Check availability first
	if !nc.adapter.IsAvailable(ctx) {
		// Return mock data for CI/testing - proper adapter pattern eliminates ghost system
		return result.MockSuccess([]domain.NixGeneration{
			{ID: 300, Path: "/nix/var/nix/profiles/default-300-link", Date: time.Now().Add(-24*time.Hour), Current: true},
			{ID: 299, Path: "/nix/var/nix/profiles/default-299-link", Date: time.Now().Add(-48*time.Hour), Current: false},
			{ID: 298, Path: "/nix/var/nix/profiles/default-298-link", Date: time.Now().Add(-72*time.Hour), Current: false},
			{ID: 297, Path: "/nix/var/nix/profiles/default-297-link", Date: time.Now().Add(-96*time.Hour), Current: false},
			{ID: 296, Path: "/nix/var/nix/profiles/default-296-link", Date: time.Now().Add(-120*time.Hour), Current: false},
		}, "Nix not available - using mock data")
	}

	// Only call adapter if available
	return nc.adapter.ListGenerations(ctx)
}

// CleanOldGenerations removes old Nix generations with proper type safety
func (nc *NixCleaner) CleanOldGenerations(ctx context.Context, keepCount int) result.Result[domain.CleanResult] {
	// Get generations first
	genResult := nc.ListGenerations(ctx)
	if genResult.IsErr() {
		return result.Err[domain.CleanResult](genResult.Error())
	}

	generations := genResult.Value()
	
	// Count and remove old generations
	toRemove := countOldGenerations(generations, keepCount)
	
	if nc.dryRun {
		return result.Ok(domain.CleanResult{
			ItemsRemoved: 0,
			FreedBytes:   estimateSpaceToFree(toRemove),
			ItemsFailed:  0,
			Strategy:     "DRY RUN",
			CleanTime:    time.Since(time.Now()),
		})
	}

	// Call adapter only if we have real Nix
	if nc.adapter.IsAvailable(ctx) {
		return nc.adapter.CollectGarbage(ctx)
	}

	// Mock result for CI
	return result.MockSuccess(domain.CleanResult{
		ItemsRemoved: len(toRemove),
		FreedBytes:   estimateSpaceToFree(toRemove),
		ItemsFailed:  0,
		Strategy:     "MOCK CLEANUP",
		CleanTime:    time.Since(time.Now()),
	}, "Mock cleanup - Nix not available")
}

// GetStoreSize gets Nix store size with proper type handling
func (nc *NixCleaner) GetStoreSize(ctx context.Context) result.Result[int64] {
	// Check availability first
	if !nc.adapter.IsAvailable(ctx) {
		// Return mock size for CI
		return result.MockSuccess(int64(1024*1024*1024*10), "Mock store size - Nix not available")
	}

	// Get result from adapter (returns CleanResult)
	storeResult := nc.adapter.GetStoreSize(ctx)
	if storeResult.IsErr() {
		return result.Err[int64](storeResult.Error())
	}

	// Extract size from CleanResult using centralized conversion
	return conversions.ExtractBytesFromCleanResult(storeResult)
}

// Helper functions
func countOldGenerations(generations []domain.NixGeneration, keepCount int) []domain.NixGeneration {
	var old []domain.NixGeneration
	currentCount := 0
	
	for _, gen := range generations {
		if gen.Current {
			currentCount++
		} else {
			old = append(old, gen)
		}
	}
	
	// Keep last 'keepCount' generations (including current)
	if len(old)+currentCount > keepCount {
		return old[:(len(old)+currentCount-keepCount)]
	}
	
	return []domain.NixGeneration{}
}

func estimateSpaceToFree(generations []domain.NixGeneration) int64 {
	// Estimate ~100MB per old generation
	return int64(len(generations)) * 100 * 1024 * 1024
}
