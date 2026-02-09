package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// goCommandTimeout is the timeout for Go cache cleaning operations.
// This is longer than lint timeout because Go caches can be large.
const goCommandTimeout = 60 * time.Second

// GoCacheCleaner handles built-in Go cache cleaning operations.
type GoCacheCleaner struct {
	cacheType GoCacheType
	verbose   bool
	dryRun    bool
	helper    *golangHelpers
}

// NewGoCacheCleaner creates a new GoCacheCleaner.
func NewGoCacheCleaner(cacheType GoCacheType, verbose, dryRun bool) *GoCacheCleaner {
	return &GoCacheCleaner{
		cacheType: cacheType,
		verbose:   verbose,
		dryRun:    dryRun,
		helper:    &golangHelpers{},
	}
}

// IsAvailable checks if the Go cache cleaner is available.
// It verifies that the Go command is installed and accessible.
func (gcc *GoCacheCleaner) IsAvailable(ctx context.Context) bool {
	_, err := gcc.helper.getGoEnv(ctx, "GOROOT")
	return err == nil
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
	default:
		return result.Err[domain.CleanResult](fmt.Errorf("unsupported cache type: %v", gcc.cacheType))
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
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return result.Err[domain.CleanResult](fmt.Errorf("go clean -%s timed out after %v (command may be hanging)", cleanFlag, goCommandTimeout))
		}
		return result.Err[domain.CleanResult](fmt.Errorf("go clean -%s failed: %w (output: %s)", cleanFlag, err, string(output)))
	}

	if gcc.verbose {
		fmt.Println(successMessage)
	}

	size := domain.SizeEstimate{Known: sizeEstimate}
	return result.Ok(domain.CleanResult{
		SizeEstimate: size,
		FreedBytes:   size.Value(), // Deprecated field
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
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

	var sizeEstimate uint64
	if !gcc.dryRun {
		sizeEstimate = uint64(GetDirSize(cachePath))
	}

	return gcc.executeGoCleanCommand(ctx, cleanFlag, successMessage, sizeEstimate)
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

// cleanGoBuildCache removes go-build* folders.
func (gcc *GoCacheCleaner) cleanGoBuildCache(ctx context.Context) result.Result[domain.CleanResult] {
	buildCachePattern := "go-build*"
	tempDir := "/tmp"
	if homeDir := gcc.helper.getHomeDir(); homeDir != "" {
		tempDir = homeDir + "/Library/Caches"
	}

	// Use shell globbing to find build cache folders
	matches, err := filepath.Glob(tempDir + "/" + buildCachePattern)
	if err != nil {
		return result.Ok(domain.CleanResult{
			SizeEstimate: domain.SizeEstimate{Known: 0},
			FreedBytes:   0,
			ItemsRemoved: 0,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
		})
	}

	itemsRemoved := 0
	var totalSizeEstimate domain.SizeEstimate
	for _, match := range matches {
		// Calculate size before removal
		if !gcc.dryRun {
			bytesFreed := GetDirSize(match)
			totalSizeEstimate = domain.SizeEstimate{Known: totalSizeEstimate.Known + uint64(bytesFreed)}
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

	if gcc.verbose && itemsRemoved > 0 {
		fmt.Println("  ✓ Go build cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		SizeEstimate: totalSizeEstimate,
		FreedBytes:   totalSizeEstimate.Value(), // Deprecated field
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
}
