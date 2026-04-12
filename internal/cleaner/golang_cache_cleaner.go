package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// goCommandTimeout is the timeout for Go cache cleaning operations.
// This is longer than lint timeout because Go caches can be large.
const goCommandTimeout = 60 * time.Second

// GoCacheCleaner handles built-in Go cache cleaning operations.
type GoCacheCleaner struct {
	CleanerBase

	cacheType GoCacheType
	helper    *golangHelpers
}

// NewGoCacheCleaner creates a new GoCacheCleaner.
func NewGoCacheCleaner(cacheType GoCacheType, verbose, dryRun bool) *GoCacheCleaner {
	return &GoCacheCleaner{
		cacheType:   cacheType,
		CleanerBase: NewCleanerBase(verbose, dryRun),
		helper:      &golangHelpers{},
	}
}

// IsAvailable checks if the Go cache cleaner is available.
// It verifies that the Go command is installed and accessible.
func (gcc *GoCacheCleaner) IsAvailable(ctx context.Context) bool {
	_, err := gcc.helper.getGoEnv(ctx, "GOROOT")

	return err == nil
}

// Name returns the cleaner name for result tracking.
func (gcc *GoCacheCleaner) Name() string {
	return "golang"
}

// Scan scans for the configured cache type and returns scan items.
func (gcc *GoCacheCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	switch gcc.cacheType {
	case GoCacheGOCACHE:
		items = append(items, gcc.scanGoEnvCache(ctx, "GOCACHE")...)
	case GoCacheTestCache:
		// Test cache doesn't have a separate path - it's part of GOCACHE
		// We just mark it as scannable without a specific path
	case GoCacheModCache:
		items = append(items, gcc.scanGoEnvCache(ctx, "GOMODCACHE")...)
	case GoCacheBuildCache:
		items = append(items, gcc.scanGoBuildCache()...)
	case GoCacheNone, GoCacheLintCache:
		// No scan items for these cache types
	}

	return result.Ok(items)
}

// scanGoEnvCache scans a Go environment variable cache path.
func (gcc *GoCacheCleaner) scanGoEnvCache(ctx context.Context, envVar string) []domain.ScanItem {
	cachePath, err := gcc.helper.getGoEnv(ctx, envVar)
	if err != nil || cachePath == "" {
		return []domain.ScanItem{}
	}

	return []domain.ScanItem{
		{
			Path:     cachePath,
			Size:     GetDirSize(cachePath),
			Created:  GetDirModTime(cachePath),
			ScanType: domain.ScanTypeTemp,
		},
	}
}

// getGoBuildCacheLocations returns all potential go-build cache locations.
// This ensures comprehensive coverage across different platforms and configurations.
func (gcc *GoCacheCleaner) getGoBuildCacheLocations() []string {
	locations := []string{
		os.TempDir(), // Platform-specific temp (e.g., /private/var/folders/... on macOS)
		"/tmp",       // Standard Unix temp
	}

	// Add macOS-specific Library/Caches location
	if homeDir := gcc.helper.getHomeDir(); homeDir != "" {
		locations = append(locations, filepath.Join(homeDir, "Library", "Caches"))
	}

	return locations
}

// scanGoBuildCache scans go-build* folders in temp directories.
func (gcc *GoCacheCleaner) scanGoBuildCache() []domain.ScanItem {
	items := make([]domain.ScanItem, 0)
	buildCachePattern := "go-build*"
	seen := make(map[string]bool) // Prevent duplicates

	for _, tempDir := range gcc.getGoBuildCacheLocations() {
		matches, err := filepath.Glob(filepath.Join(tempDir, buildCachePattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			// Skip duplicates (same file found via different paths like symlinks)
			if seen[match] {
				continue
			}

			seen[match] = true

			items = append(items, domain.ScanItem{
				Path:     match,
				Size:     GetDirSize(match),
				Created:  GetDirModTime(match),
				ScanType: domain.ScanTypeTemp,
			})
		}
	}

	return items
}

// Clean cleans the specified cache type.
func (gcc *GoCacheCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	switch gcc.cacheType {
	case GoCacheGOCACHE:
		return gcc.cleanGoCache(ctx)
	case GoCacheTestCache:
		return gcc.cleanGoTestCache(ctx)
	case GoCacheModCache:
		return gcc.cleanGoModCache(ctx)
	case GoCacheBuildCache:
		return gcc.cleanGoBuildCache(ctx)
	case GoCacheNone:
		return result.Err[domain.CleanResult](
			errors.New("no cache type specified"),
		)
	case GoCacheLintCache:
		return result.Err[domain.CleanResult](
			errors.New("lint cache cleaning not yet implemented"),
		)
	default:
		return result.Err[domain.CleanResult](
			fmt.Errorf("unsupported cache type: %v", gcc.cacheType),
		)
	}
}

// executeGoCleanCommand executes the go clean command with timeout protection.
func (gcc *GoCacheCleaner) executeGoCleanCommand(
	ctx context.Context,
	cleanFlag string,
	successMessage string,
	sizeEstimate uint64,
) result.Result[domain.CleanResult] {
	// Create a timeout context to prevent hanging
	timeoutCtx, cancel := context.WithTimeout(ctx, goCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "go", "clean", "-"+cleanFlag)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a timeout error
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return result.Err[domain.CleanResult](
				fmt.Errorf(
					"go clean -%s timed out after %v (command may be hanging)",
					cleanFlag,
					goCommandTimeout,
				),
			)
		}

		return result.Err[domain.CleanResult](fmt.Errorf("go clean -%s failed: %w (output: %s)",
			cleanFlag, err, string(output)))
	}

	if gcc.verbose {
		fmt.Println(successMessage)
	}

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		domain.StrategyConservativeType,
		1, int64(sizeEstimate),
		domain.SizeEstimate{Known: sizeEstimate},
	))
}

// cleanGoCacheEnv cleans a Go cache specified by environment variable.
func (gcc *GoCacheCleaner) cleanGoCacheEnv(
	ctx context.Context,
	envVar string,
	cleanFlag string,
	successMessage string,
) result.Result[domain.CleanResult] {
	cachePath, err := gcc.helper.getGoEnv(ctx, envVar)
	if err != nil || cachePath == "" {
		return gcc.executeGoCleanCommand(ctx, cleanFlag, successMessage, 0)
	}

	var bytesFreed int64

	if !gcc.dryRun {
		var cleanupErr error

		bytesFreed, _, _ = CalculateBytesFreed(cachePath, func() error {
			// Execute the clean command
			timeoutCtx, cancel := context.WithTimeout(ctx, goCommandTimeout)
			defer cancel()

			cmd := exec.CommandContext(timeoutCtx, "go", "clean", "-"+cleanFlag)

			output, err := cmd.CombinedOutput()
			if err != nil {
				cleanupErr = err

				if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
					return fmt.Errorf(
						"go clean -%s timed out after %v",
						cleanFlag,
						goCommandTimeout,
					)
				}

				return fmt.Errorf(
					"go clean -%s failed: %w (output: %s)",
					cleanFlag,
					err,
					string(output),
				)
			}

			return nil
		}, gcc.verbose, "Cache")

		if cleanupErr != nil {
			return result.Err[domain.CleanResult](cleanupErr)
		}

		if gcc.verbose {
			fmt.Println(successMessage)
		}

		return result.Ok(conversions.NewCleanResultWithSizeEstimate(
			domain.StrategyConservativeType,
			1, bytesFreed,
			domain.SizeEstimate{Known: uint64(bytesFreed)},
		))
	}

	// Dry run - calculate size estimate and return
	bytesFreed = GetDirSize(cachePath)

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		domain.StrategyConservativeType,
		1, bytesFreed,
		domain.SizeEstimate{Known: uint64(bytesFreed)},
	))
}

// cleanGoCache cleans GOCACHE.
func (gcc *GoCacheCleaner) cleanGoCache(ctx context.Context) result.Result[domain.CleanResult] {
	return gcc.cleanGoCacheEnv(ctx, "GOCACHE", "cache", "  ✓ Go cache cleaned")
}

// cleanGoTestCache cleans GOTESTCACHE.
func (gcc *GoCacheCleaner) cleanGoTestCache(ctx context.Context) result.Result[domain.CleanResult] {
	return gcc.executeGoCleanCommand(ctx, "testcache", "  ✓ Go test cache cleaned", 0)
}

// cleanGoModCache cleans GOMODCACHE.
func (gcc *GoCacheCleaner) cleanGoModCache(ctx context.Context) result.Result[domain.CleanResult] {
	return gcc.cleanGoCacheEnv(ctx, "GOMODCACHE", "modcache", "  ✓ Go module cache cleaned")
}

// cleanGoBuildCache removes go-build* folders from all temp locations.
func (gcc *GoCacheCleaner) cleanGoBuildCache(
	_ context.Context,
) result.Result[domain.CleanResult] {
	buildCachePattern := "go-build*"
	seen := make(map[string]bool) // Prevent cleaning same path twice
	itemsRemoved := 0

	var totalSizeEstimate domain.SizeEstimate

	for _, tempDir := range gcc.getGoBuildCacheLocations() {
		matches, err := filepath.Glob(filepath.Join(tempDir, buildCachePattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			// Skip duplicates
			if seen[match] {
				continue
			}

			seen[match] = true

			// Calculate size before removal (always, for accurate dry-run estimates)
			bytesFreed := GetDirSize(match)
			totalSizeEstimate = domain.SizeEstimate{
				Known: totalSizeEstimate.Known + uint64(bytesFreed),
			}

			if gcc.dryRun {
				itemsRemoved++

				continue
			}

			err := os.RemoveAll(match)
			if err != nil {
				if gcc.verbose {
					fmt.Printf("Warning: failed to remove %s: %v\n", match, err)
				}

				continue
			}

			itemsRemoved++

			if gcc.verbose {
				fmt.Printf("  ✓ Removed build cache: %s\n", match)
			}
		}
	}

	if gcc.verbose && itemsRemoved > 0 {
		fmt.Println("  ✓ Go build cache cleaned")
	}

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		domain.StrategyConservativeType,
		itemsRemoved, int64(totalSizeEstimate.Value()),
		totalSizeEstimate,
	))
}
