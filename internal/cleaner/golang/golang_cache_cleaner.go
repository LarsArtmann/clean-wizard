package golang

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// goCommandTimeout is the timeout for Go cache cleaning operations.
// This is longer than lint timeout because Go caches can be large.
const goCommandTimeout = 60 * time.Second

// GoCacheCleaner handles built-in Go cache cleaning operations.
type GoCacheCleaner struct {
	cleaner.CleanerBase

	cacheType	GoCacheType
	helper		*golangHelpers
}

// NewGoCacheCleaner creates a new GoCacheCleaner.
func NewGoCacheCleaner(cacheType GoCacheType, verbose, dryRun bool) *GoCacheCleaner {
	return &GoCacheCleaner{
		cacheType:	cacheType,
		CleanerBase:	cleaner.NewCleanerBase(verbose, dryRun),
		helper:		&golangHelpers{},
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
func (gcc *GoCacheCleaner) Scan(ctx context.Context) result.Result[[]types.ScanItem] {
	items := make([]types.ScanItem, 0)

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
func (gcc *GoCacheCleaner) scanGoEnvCache(ctx context.Context, envVar string) []types.ScanItem {
	cachePath, err := gcc.helper.getGoEnv(ctx, envVar)
	if err != nil || cachePath == "" {
		return []types.ScanItem{}
	}

	return []types.ScanItem{
		{
			Path:		cachePath,
			Size:		cleaner.GetDirSize(cachePath),
			Created:	cleaner.GetDirModTime(cachePath),
			ScanType:	types.ScanTypeTemp,
		},
	}
}

// getGoBuildCacheLocations returns all potential go-build cache locations.
// This ensures comprehensive coverage across different platforms and configurations.
func (gcc *GoCacheCleaner) getGoBuildCacheLocations() []string {
	seen := make(map[string]bool)
	locations := []string{}

	addUnique := func(path string) {
		if !seen[path] {
			seen[path] = true
			locations = append(locations, path)
		}
	}

	addUnique(os.TempDir())
	addUnique("/tmp")

	if homeDir := gcc.helper.getHomeDir(); homeDir != "" {
		addUnique(filepath.Join(homeDir, "Library", "Caches"))
	}

	return locations
}

// scanGoBuildCache scans go-build* folders in temp directories.
func (gcc *GoCacheCleaner) scanGoBuildCache() []types.ScanItem {
	items := make([]types.ScanItem, 0)
	buildCachePattern := "go-build*"
	seen := make(map[string]bool)	// Prevent duplicates

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

			items = append(items, types.ScanItem{
				Path:		match,
				Size:		cleaner.GetDirSize(match),
				Created:	cleaner.GetDirModTime(match),
				ScanType:	types.ScanTypeTemp,
			})
		}
	}

	return items
}

// Clean cleans the specified cache type.
func (gcc *GoCacheCleaner) Clean(ctx context.Context) result.Result[types.CleanResult] {
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
		return result.Err[types.CleanResult](ErrNoCacheTypeSpecified)
	case GoCacheLintCache:
		return result.Err[types.CleanResult](ErrLintCacheNotImplemented)
	default:
		//nolint:err113 // Dynamic: cacheType is included for debugging
		return result.Err[types.CleanResult](
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
) result.Result[types.CleanResult] {
	// Create a timeout context to prevent hanging
	timeoutCtx, cancel := context.WithTimeout(ctx, goCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "go", "clean", "-"+cleanFlag)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a timeout error
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return result.Err[types.CleanResult](
				fmt.Errorf(
					"go clean -%s timed out after %v for successMessage=%v, sizeEstimate=%v (command may be hanging)",
					cleanFlag,
					goCommandTimeout,
					successMessage,
					sizeEstimate,
				),
			)
		}

		return result.Err[types.CleanResult](fmt.Errorf("go clean -%s failed: %w (output: %s)",
			cleanFlag, err, string(output)))
	}

	if gcc.verbose {
		fmt.Println(successMessage)
	}

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		enums.StrategyConservativeType,
		1, int64(sizeEstimate),
		types.SizeEstimate{Known: sizeEstimate, Status: enums.SizeEstimateStatusKnown},
	))
}

// cleanGoCacheEnv cleans a Go cache specified by environment variable.
func (gcc *GoCacheCleaner) cleanGoCacheEnv(
	ctx context.Context,
	envVar string,
	cleanFlag string,
	successMessage string,
) result.Result[types.CleanResult] {
	cachePath, err := gcc.helper.getGoEnv(ctx, envVar)
	if err != nil || cachePath == "" {
		return gcc.executeGoCleanCommand(ctx, cleanFlag, successMessage, 0)
	}

	var bytesFreed int64

	if !gcc.dryRun {
		var cleanupErr error

		bytesFreed, _, _ = cleaner.CalculateBytesFreed(cachePath, func() error {
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
			return result.Err[types.CleanResult](
				fmt.Errorf(
					"envVar=%v, successMessage=%v, bytesFreed=%v: %w",
					envVar,
					successMessage,
					bytesFreed,
					cleanupErr,
				),
			)
		}

		if gcc.verbose {
			fmt.Println(successMessage)
		}

		return result.Ok(conversions.NewCleanResultWithSizeEstimate(
			enums.StrategyConservativeType,
			1, bytesFreed,
			types.SizeEstimate{Known: uint64(bytesFreed)},	//nolint:exhaustruct
		))
	}

	// Dry run - calculate size estimate and return
	bytesFreed = cleaner.GetDirSize(cachePath)

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		enums.StrategyConservativeType,
		1, bytesFreed,
		types.SizeEstimate{Known: uint64(bytesFreed)},	//nolint:exhaustruct
	))
}

// cleanGoCache cleans GOCACHE.
func (gcc *GoCacheCleaner) cleanGoCache(ctx context.Context) result.Result[types.CleanResult] {
	return gcc.cleanGoCacheEnv(ctx, "GOCACHE", "cache", "  ✓ Go cache cleaned")
}

// cleanGoTestCache cleans GOTESTCACHE.
func (gcc *GoCacheCleaner) cleanGoTestCache(ctx context.Context) result.Result[types.CleanResult] {
	return gcc.executeGoCleanCommand(ctx, "testcache", "  ✓ Go test cache cleaned", 0)
}

// cleanGoModCache cleans GOMODCACHE.
func (gcc *GoCacheCleaner) cleanGoModCache(ctx context.Context) result.Result[types.CleanResult] {
	return gcc.cleanGoCacheEnv(ctx, "GOMODCACHE", "modcache", "  ✓ Go module cache cleaned")
}

// cleanGoBuildCache removes go-build* folders from all temp locations.
func (gcc *GoCacheCleaner) cleanGoBuildCache(
	_ context.Context,
) result.Result[types.CleanResult] {
	buildCachePattern := "go-build*"
	seen := make(map[string]bool)	// Prevent cleaning same path twice
	itemsRemoved := 0

	var totalSizeEstimate types.SizeEstimate

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
			bytesFreed := cleaner.GetDirSize(match)
			totalSizeEstimate = types.SizeEstimate{	//nolint:exhaustruct
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
		enums.StrategyConservativeType,
		itemsRemoved, int64(totalSizeEstimate.Value()),
		totalSizeEstimate,
	))
}
