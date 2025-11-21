package cleaner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// PnpmCleaner handles pnpm package manager cleanup with proper type safety
type PnpmCleaner struct {
	verbose bool
	dryRun  bool
}

// NewPnpmCleaner creates pnpm cleaner with proper configuration
func NewPnpmCleaner(verbose, dryRun bool) *PnpmCleaner {
	return &PnpmCleaner{
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// IsAvailable checks if pnpm cleaner is available
func (pc *PnpmCleaner) IsAvailable(ctx context.Context) bool {
	return checkCommandExists(ctx, "pnpm")
}

// GetCacheSize estimates pnpm store size
func (pc *PnpmCleaner) GetCacheSize(ctx context.Context) int64 {
	if !pc.IsAvailable(ctx) {
		return 0 // pnpm not available
	}

	// Get store directory
	storeResult := runCommand(ctx, "pnpm", "store", "path")
	if storeResult.IsErr() {
		return 0
	}

	storeDir := strings.TrimSpace(storeResult.Value())
	sizeResult := runCommand(ctx, "du", "-sb", storeDir)
	if sizeResult.IsErr() {
		return 0
	}

	// Parse size output (format: "size\tpath")
	sizeStr := strings.Fields(sizeResult.Value())[0]
	var size int64
	if _, err := fmt.Sscanf(sizeStr, "%d", &size); err != nil {
		return 0
	}

	return size
}

// ValidateSettings validates pnpm cleaner settings with type safety
func (pc *PnpmCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil {
		return nil // Settings are optional
	}

	// pnpm doesn't have specific settings yet, but validation ready for future
	return nil
}

// Cleanup performs pnpm store cleanup
func (pc *PnpmCleaner) Cleanup(ctx context.Context, settings *domain.OperationSettings) result.Result[domain.CleanResult] {
	startTime := time.Now()

	if !pc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("pnpm is not available"))
	}

	if pc.verbose {
		fmt.Println("📦 Cleaning pnpm store...")
	}

	cleanupCmdResult := pc.runCleanupCommand(ctx)
	
	// Create final result
	cleanResult := domain.CleanResult{
		FreedBytes:   cleanupCmdResult.Value().FreedBytes,
		ItemsRemoved: cleanupCmdResult.Value().ItemsRemoved,
		ItemsFailed:  cleanupCmdResult.Value().ItemsFailed,
		CleanTime:    time.Since(startTime),
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative, // pnpm store cleanup is conservative
	}

	return result.Ok(cleanResult)
}

// runCleanupCommand executes pnpm store prune
func (pc *PnpmCleaner) runCleanupCommand(ctx context.Context) result.Result[domain.CleanResult] {
	if pc.dryRun {
		return pc.simulateCleanupCommand()
	}

	// Execute actual pnpm store prune
	cmdResult := runCommand(ctx, "pnpm", "store", "prune")
	if cmdResult.IsErr() {
		return result.Err[domain.CleanResult](cmdResult.Error())
	}

	output := cmdResult.Value()
	
	// Parse pnpm store prune output to extract cleanup information
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	var itemsRemoved uint
	var bytesFreed uint64
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "removed") || strings.Contains(line, "deleted") {
			itemsRemoved++
			// Try to extract size information if available
			if strings.Contains(line, "MB") {
				var size float64
				if _, err := fmt.Sscanf(line, "%f", &size); err == nil {
					bytesFreed += uint64(size * 1024 * 1024)
				}
			}
		}
	}

	// If no size info found, estimate typical pnpm store cleanup
	if bytesFreed == 0 && itemsRemoved > 0 {
		bytesFreed = 200 * 1024 * 1024 // 200MB estimate
	}

	cleanupCmdResult := domain.CleanResult{
		FreedBytes:   bytesFreed,
		ItemsRemoved: itemsRemoved,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(cleanupCmdResult)
}

// simulateCleanupCommand simulates cleanup for dry-run mode
func (pc *PnpmCleaner) simulateCleanupCommand() result.Result[domain.CleanResult] {
	if pc.verbose {
		fmt.Println("🔍 DRY RUN: pnpm store prune")
	}

	// Simulate typical pnpm store cleanup results
	var itemsRemoved uint = 1 // Store directory
	var bytesFreed uint64 = 200 * 1024 * 1024 // 200MB estimate

	simulateResult := domain.CleanResult{
		FreedBytes:   bytesFreed,
		ItemsRemoved: itemsRemoved,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyDryRun,
	}

	return result.Ok(simulateResult)
}