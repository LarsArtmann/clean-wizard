package cleaner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/infrastructure/system"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// HomebrewCleaner handles Homebrew package manager cleanup with proper type safety
type HomebrewCleaner struct {
	adapter *system.HTTPClient
	verbose bool
	dryRun  bool
}

// NewHomebrewCleaner creates Homebrew cleaner with proper configuration
func NewHomebrewCleaner(verbose, dryRun bool) *HomebrewCleaner {
	return &HomebrewCleaner{
		adapter: system.NewHTTPClient(),
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// IsAvailable checks if Homebrew cleaner is available
func (hc *HomebrewCleaner) IsAvailable(ctx context.Context) bool {
	// Check if brew command exists and is executable
	return checkCommandExists(ctx, "brew")
}

// GetCacheSize estimates Homebrew cache size
func (hc *HomebrewCleaner) GetCacheSize(ctx context.Context) int64 {
	if !hc.IsAvailable(ctx) {
		return 0 // Homebrew not available
	}

	// Get cache directory size
	cacheResult := runCommand(ctx, "brew", "--cache")
	if cacheResult.IsErr() {
		return 0
	}

	cacheDir := strings.TrimSpace(cacheResult.Value())
	sizeResult := runCommand(ctx, "du", "-sb", cacheDir)
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

// GetStoreSize implements domain.Cleaner interface
func (hc *HomebrewCleaner) GetStoreSize(ctx context.Context) int64 {
	return hc.GetCacheSize(ctx)
}

// ValidateSettings validates Homebrew cleaner settings with type safety
func (hc *HomebrewCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil {
		return nil // Settings are optional
	}

	// Homebrew doesn't have specific settings yet, but validation ready for future
	return nil
}

// Cleanup performs Homebrew cleanup with comprehensive operations
func (hc *HomebrewCleaner) Cleanup(ctx context.Context, settings *domain.OperationSettings) result.Result[domain.CleanResult] {
	startTime := time.Now()

	if !hc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("homebrew is not available"))
	}

	var totalFreed uint64
	var totalRemoved uint
	var totalFailed uint

	// Step 1: Remove outdated packages
	if hc.verbose {
		fmt.Println("🍺 Removing outdated Homebrew packages...")
	}
	autoremoveResult := hc.runCleanupCommand(ctx, "autoremove")
	if autoremoveResult.IsOk() {
		result := autoremoveResult.Value()
		totalFreed += result.FreedBytes
		totalRemoved += result.ItemsRemoved
		totalFailed += result.ItemsFailed
	} else {
		totalFailed++
		if hc.verbose {
			fmt.Printf("⚠️  Autoremove failed: %v\n", autoremoveResult.Error())
		}
	}

	// Step 2: Clean cache and prune old files
	if hc.verbose {
		fmt.Println("🧹 Cleaning Homebrew cache and pruning old files...")
	}
	cleanupResult := hc.runCleanupCommand(ctx, "cleanup", "--prune=all", "-s")
	if cleanupResult.IsOk() {
		result := cleanupResult.Value()
		totalFreed += result.FreedBytes
		totalRemoved += result.ItemsRemoved
		totalFailed += result.ItemsFailed
	} else {
		totalFailed++
		if hc.verbose {
			fmt.Printf("⚠️  Cleanup failed: %v\n", cleanupResult.Error())
		}
	}

	// Step 3: Clean up cask cache if available
	caskResult := hc.cleanupCaskCache(ctx)
	if caskResult.IsOk() {
		result := caskResult.Value()
		totalFreed += result.FreedBytes
		totalRemoved += result.ItemsRemoved
		totalFailed += result.ItemsFailed
	}

	// Create final result
	cleanResult := domain.CleanResult{
		FreedBytes:   totalFreed,
		ItemsRemoved: totalRemoved,
		ItemsFailed:  totalFailed,
		CleanTime:    time.Since(startTime),
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative, // Homebrew cleanup is conservative
	}

	return result.Ok(cleanResult)
}

// runCleanupCommand executes a Homebrew cleanup command
func (hc *HomebrewCleaner) runCleanupCommand(ctx context.Context, args ...string) result.Result[domain.CleanResult] {
	if hc.dryRun {
		return hc.simulateCleanupCommand(args...)
	}

	// Execute the actual command
	cmdResult := runCommand(ctx, "brew", args...)
	if cmdResult.IsErr() {
		return result.Err[domain.CleanResult](cmdResult.Error())
	}

	output := cmdResult.Value()
	
	// Parse Homebrew output to extract cleanup information
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

	cleanupResult := domain.CleanResult{
		FreedBytes:   bytesFreed,
		ItemsRemoved: itemsRemoved,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(cleanupResult)
}

// simulateCleanupCommand simulates cleanup for dry-run mode
func (hc *HomebrewCleaner) simulateCleanupCommand(args ...string) result.Result[domain.CleanResult] {
	if hc.verbose {
		fmt.Printf("🔍 DRY RUN: brew %s\n", strings.Join(args, " "))
	}

	// Simulate typical cleanup results
	var itemsRemoved uint = 1
	var bytesFreed uint64 = 50 * 1024 * 1024 // 50MB estimate

	command := strings.Join(args, " ")
	switch {
	case strings.Contains(command, "autoremove"):
		itemsRemoved = 2 // Typically removes 1-2 packages
		bytesFreed = 100 * 1024 * 1024 // 100MB estimate
	case strings.Contains(command, "cleanup"):
		itemsRemoved = 5 // Typically removes several cached files
		bytesFreed = 200 * 1024 * 1024 // 200MB estimate
	}

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

// cleanupCaskCache cleans up Homebrew Cask cache
func (hc *HomebrewCleaner) cleanupCaskCache(ctx context.Context) result.Result[domain.CleanResult] {
	// Check if Cask is available
	caskResult := runCommand(ctx, "brew", "--help")
	if caskResult.IsErr() {
		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: 0,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})
	}

	if hc.verbose {
		fmt.Println("🍎 Cleaning Homebrew Cask cache...")
	}

	return hc.runCleanupCommand(ctx, "cleanup", "--prune=all")
}