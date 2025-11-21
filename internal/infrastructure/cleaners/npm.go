package cleaner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// NpmCleaner handles npm package manager cleanup with proper type safety
type NpmCleaner struct {
	verbose bool
	dryRun  bool
}

// NewNpmCleaner creates npm cleaner with proper configuration
func NewNpmCleaner(verbose, dryRun bool) *NpmCleaner {
	return &NpmCleaner{
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// IsAvailable checks if npm cleaner is available
func (nc *NpmCleaner) IsAvailable(ctx context.Context) bool {
	return checkCommandExists(ctx, "npm")
}

// GetCacheSize estimates npm cache size
func (nc *NpmCleaner) GetCacheSize(ctx context.Context) int64 {
	if !nc.IsAvailable(ctx) {
		return 0 // npm not available
	}

	// Get cache size using npm
	cacheResult := runCommand(ctx, "npm", "cache", "verify")
	if cacheResult.IsErr() {
		return 0
	}

	// Parse npm cache verify output to extract size information
	output := cacheResult.Value()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Cache verified") && strings.Contains(line, "KB") {
			// Extract size from line like "Cache verified and compressed (1234 KB)"
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "KB" && i > 0 {
					var size float64
					if _, err := fmt.Sscanf(parts[i-1], "%f", &size); err == nil {
						return int64(size * 1024) // Convert KB to bytes
					}
				}
			}
		}
	}

	return 0
}

// GetStoreSize implements domain.Cleaner interface
func (nc *NpmCleaner) GetStoreSize(ctx context.Context) int64 {
	return nc.GetCacheSize(ctx)
}

// ValidateSettings validates npm cleaner settings with type safety
func (nc *NpmCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil {
		return nil // Settings are optional
	}

	// npm doesn't have specific settings yet, but validation ready for future
	return nil
}

// Cleanup performs npm cache cleanup
func (nc *NpmCleaner) Cleanup(ctx context.Context, settings *domain.OperationSettings) result.Result[domain.CleanResult] {
	startTime := time.Now()

	if !nc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("npm is not available"))
	}

	if nc.verbose {
		fmt.Println("📦 Cleaning npm cache...")
	}

	cleanupCmdResult := nc.runCleanupCommand(ctx)
	
	// Create final result
	cleanResult := domain.CleanResult{
		FreedBytes:   cleanupCmdResult.Value().FreedBytes,
		ItemsRemoved: cleanupCmdResult.Value().ItemsRemoved,
		ItemsFailed:  cleanupCmdResult.Value().ItemsFailed,
		CleanTime:    time.Since(startTime),
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative, // npm cache cleanup is conservative
	}

	return result.Ok(cleanResult)
}

// runCleanupCommand executes npm cache cleanup
func (nc *NpmCleaner) runCleanupCommand(ctx context.Context) result.Result[domain.CleanResult] {
	if nc.dryRun {
		return nc.simulateCleanupCommand()
	}

	// Execute actual npm cache clean
	cmdResult := runCommand(ctx, "npm", "cache", "clean", "--force")
	if cmdResult.IsErr() {
		return result.Err[domain.CleanResult](cmdResult.Error())
	}

	output := cmdResult.Value()
	
	// Parse npm cache clean output - npm typically doesn't give detailed stats
	// So we estimate based on typical cache sizes
	var itemsRemoved uint = 1 // Cache directory
	var bytesFreed uint64 = 100 * 1024 * 1024 // 100MB estimate

	// Try to extract any size information from output
	if strings.Contains(output, "deleted") || strings.Contains(output, "removed") {
		itemsRemoved = 1
		// Look for size information
		if strings.Contains(output, "MB") {
			var size float64
			if _, err := fmt.Sscanf(output, "%f", &size); err == nil {
				bytesFreed = uint64(size * 1024 * 1024)
			}
		}
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
func (nc *NpmCleaner) simulateCleanupCommand() result.Result[domain.CleanResult] {
	if nc.verbose {
		fmt.Println("🔍 DRY RUN: npm cache clean --force")
	}

	// Simulate typical npm cache cleanup results
	var itemsRemoved uint = 1 // Cache directory
	var bytesFreed uint64 = 150 * 1024 * 1024 // 150MB estimate for npm cache

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