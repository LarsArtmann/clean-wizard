package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// TempFileCleaner handles temporary file cleanup with proper type safety
type TempFileCleaner struct {
	verbose bool
	dryRun  bool
}

// NewTempFileCleaner creates temp file cleaner with proper configuration
func NewTempFileCleaner(verbose, dryRun bool) *TempFileCleaner {
	return &TempFileCleaner{
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// IsAvailable checks if temp file cleaner is available (always true)
func (tfc *TempFileCleaner) IsAvailable(ctx context.Context) bool {
	// Temp file cleaner is always available - no external dependencies
	return true
}

// GetCacheSize estimates temporary files size
func (tfc *TempFileCleaner) GetCacheSize(ctx context.Context) int64 {
	var totalSize int64

	// Common temporary directories
	tempDirs := []string{
		os.TempDir(),
		filepath.Join(os.Getenv("HOME"), "tmp"),
		filepath.Join(os.Getenv("HOME"), ".tmp"),
	}

	for _, dir := range tempDirs {
		if _, err := os.Stat(dir); err == nil {
			size := tfc.calculateDirectorySize(dir)
			totalSize += size
		}
	}

	return totalSize
}

// GetStoreSize implements domain.Cleaner interface
func (tfc *TempFileCleaner) GetStoreSize(ctx context.Context) int64 {
	return tfc.GetCacheSize(ctx)
}

// ValidateSettings validates temp file cleaner settings with type safety
func (tfc *TempFileCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil {
		return nil // Settings are optional
	}

	// Temp file cleaner doesn't have specific settings yet, but validation ready for future
	return nil
}

// Cleanup performs temporary file cleanup
func (tfc *TempFileCleaner) Cleanup(ctx context.Context, settings *domain.OperationSettings) result.Result[domain.CleanResult] {
	startTime := time.Now()

	if !tfc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("temp file cleaner is not available"))
	}

	if tfc.verbose {
		fmt.Println("🗑️  Cleaning temporary files...")
	}

	cleanupCmdResult := tfc.runCleanupCommand(ctx)
	
	// Create final result
	cleanResult := domain.CleanResult{
		FreedBytes:   cleanupCmdResult.Value().FreedBytes,
		ItemsRemoved: cleanupCmdResult.Value().ItemsRemoved,
		ItemsFailed:  cleanupCmdResult.Value().ItemsFailed,
		CleanTime:    time.Since(startTime),
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative, // Temp file cleanup is conservative
	}

	return result.Ok(cleanResult)
}

// runCleanupCommand executes temporary file cleanup
func (tfc *TempFileCleaner) runCleanupCommand(ctx context.Context) result.Result[domain.CleanResult] {
	if tfc.dryRun {
		return tfc.simulateCleanupCommand()
	}

	// Common temporary directories to clean
	tempDirs := []string{
		os.TempDir(),
		filepath.Join(os.Getenv("HOME"), "tmp"),
		filepath.Join(os.Getenv("HOME"), ".tmp"),
	}

	var totalFreed uint64
	var totalRemoved uint
	var totalFailed uint

	for _, dir := range tempDirs {
		if _, err := os.Stat(dir); err != nil {
			continue // Directory doesn't exist, skip
		}

		if tfc.verbose {
			fmt.Printf("📁 Cleaning directory: %s\n", dir)
		}

		freed, removed, failed := tfc.cleanDirectory(ctx, dir)
		totalFreed += freed
		totalRemoved += removed
		totalFailed += failed
	}

	cleanupCmdResult := domain.CleanResult{
		FreedBytes:   totalFreed,
		ItemsRemoved: totalRemoved,
		ItemsFailed:  totalFailed,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(cleanupCmdResult)
}

// simulateCleanupCommand simulates cleanup for dry-run mode
func (tfc *TempFileCleaner) simulateCleanupCommand() result.Result[domain.CleanResult] {
	if tfc.verbose {
		fmt.Println("🔍 DRY RUN: Temporary file cleanup")
	}

	// Simulate typical temp file cleanup results
	var itemsRemoved uint = 5 // Multiple temp files and directories
	var bytesFreed uint64 = 250 * 1024 * 1024 // 250MB estimate

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

// cleanDirectory cleans a specific directory safely
func (tfc *TempFileCleaner) cleanDirectory(ctx context.Context, dir string) (uint64, uint, uint) {
	var totalSize uint64
	var removed uint
	var failed uint

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue with other files
		}

		// Skip root directory itself
		if path == dir {
			return nil
		}

		// Safety checks
		if tfc.shouldSkipFile(path, info) {
			if info.IsDir() {
				return filepath.SkipDir // Skip entire directory
			}
			return nil // Skip file
		}

		// Calculate file size
		if !info.IsDir() {
			totalSize += uint64(info.Size())
		}

		// Remove file or directory
		if tfc.dryRun {
			removed++
			return nil
		}

		if err := os.RemoveAll(path); err != nil {
			failed++
			if tfc.verbose {
				fmt.Printf("⚠️  Failed to remove %s: %v\n", path, err)
			}
		} else {
			removed++
			if tfc.verbose {
				fmt.Printf("🗑️  Removed: %s\n", path)
			}
		}

		return nil
	})

	if err != nil && tfc.verbose {
		fmt.Printf("⚠️  Error walking directory %s: %v\n", dir, err)
	}

	return totalSize, removed, failed
}

// shouldSkipFile implements safety checks for files to skip
func (tfc *TempFileCleaner) shouldSkipFile(path string, info os.FileInfo) bool {
	// Skip recently modified files (within last 24 hours)
	if time.Since(info.ModTime()) < 24*time.Hour {
		return true
	}

	// Skip files larger than 1GB (safety precaution)
	if !info.IsDir() && info.Size() > 1024*1024*1024 {
		return true
	}

	// Skip hidden files and directories
	if strings.HasPrefix(filepath.Base(path), ".") {
		return true
	}

	// Skip common system directories
	basePath := strings.ToLower(path)
	skipPaths := []string{
		"/var/folders",
		"/tmp/.x11-unix",
		"/tmp/.ICE-unix",
		"/tmp/.test-unix",
	}

	for _, skipPath := range skipPaths {
		if strings.Contains(basePath, strings.ToLower(skipPath)) {
			return true
		}
	}

	return false
}

// calculateDirectorySize calculates the size of a directory
func (tfc *TempFileCleaner) calculateDirectorySize(dir string) int64 {
	var totalSize int64

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Apply same safety checks as cleanup
		if tfc.shouldSkipFile(path, info) {
			return nil
		}

		totalSize += info.Size()
		return nil
	})

	return totalSize
}