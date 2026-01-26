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

// cleanGoCache cleans GOCACHE.
func (gcc *GoCacheCleaner) cleanGoCache(ctx context.Context) result.Result[domain.CleanResult] {
	// Get cache path first to calculate size
	cachePath, err := gcc.helper.getGoEnv(ctx, "GOCACHE")
	if err != nil || cachePath == "" {
		// If we can't get the path, still try to clean
		cmd := exec.CommandContext(ctx, "go", "clean", "-cache")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("go clean -cache failed: %w (output: %s)", err, string(output)))
		}
		if gcc.verbose {
			fmt.Println("  ✓ Go cache cleaned")
		}
		return result.Ok(domain.CleanResult{
			SizeEstimate: domain.SizeEstimate{Known: 0},
			FreedBytes:   0, // Deprecated field
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})
	}

	// Calculate size before cleaning
	var sizeEstimate domain.SizeEstimate
	if !gcc.dryRun {
		bytesFreed := gcc.helper.getDirSize(cachePath)
		sizeEstimate = domain.SizeEstimate{Known: uint64(bytesFreed)}
	} else {
		sizeEstimate = domain.SizeEstimate{Known: 0}
	}

	cmd := exec.CommandContext(ctx, "go", "clean", "-cache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("go clean -cache failed: %w (output: %s)", err, string(output)))
	}

	if gcc.verbose {
		fmt.Println("  ✓ Go cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		SizeEstimate: sizeEstimate,
		FreedBytes:   sizeEstimate.Value(), // Deprecated field
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
}

// cleanGoTestCache cleans GOTESTCACHE.
func (gcc *GoCacheCleaner) cleanGoTestCache(ctx context.Context) result.Result[domain.CleanResult] {
	// Get cache path first to calculate size
	cachePath, err := gcc.helper.getGoEnv(ctx, "GOCACHE")
	if err != nil || cachePath == "" {
		// If we can't get the path, still try to clean
		cmd := exec.CommandContext(ctx, "go", "clean", "-testcache")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("go clean -testcache failed: %w (output: %s)", err, string(output)))
		}
		if gcc.verbose {
			fmt.Println("  ✓ Go test cache cleaned")
		}
		return result.Ok(domain.CleanResult{
			SizeEstimate: domain.SizeEstimate{Known: 0},
			FreedBytes:   0, // Deprecated field
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})
	}

	// Note: test cache doesn't have a separate path, it's managed internally by Go
	// We'll estimate it as 0 since we can't accurately measure test cache size separately
	cmd := exec.CommandContext(ctx, "go", "clean", "-testcache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("go clean -testcache failed: %w (output: %s)", err, string(output)))
	}

	if gcc.verbose {
		fmt.Println("  ✓ Go test cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		SizeEstimate: domain.SizeEstimate{Known: 0},
		FreedBytes:   0, // Deprecated field
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
}

// cleanGoModCache cleans GOMODCACHE.
func (gcc *GoCacheCleaner) cleanGoModCache(ctx context.Context) result.Result[domain.CleanResult] {
	// Get cache path first to calculate size
	cachePath, err := gcc.helper.getGoEnv(ctx, "GOMODCACHE")
	if err != nil || cachePath == "" {
		// If we can't get the path, still try to clean
		cmd := exec.CommandContext(ctx, "go", "clean", "-modcache")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("go clean -modcache failed: %w (output: %s)", err, string(output)))
		}
		if gcc.verbose {
			fmt.Println("  ✓ Go module cache cleaned")
		}
		return result.Ok(domain.CleanResult{
			SizeEstimate: domain.SizeEstimate{Known: 0},
			FreedBytes:   0, // Deprecated field
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})
	}

	// Calculate size before cleaning
	var sizeEstimate domain.SizeEstimate
	if !gcc.dryRun {
		bytesFreed := gcc.helper.getDirSize(cachePath)
		sizeEstimate = domain.SizeEstimate{Known: uint64(bytesFreed)}
	} else {
		sizeEstimate = domain.SizeEstimate{Known: 0}
	}

	cmd := exec.CommandContext(ctx, "go", "clean", "-modcache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("go clean -modcache failed: %w (output: %s)", err, string(output)))
	}

	if gcc.verbose {
		fmt.Println("  ✓ Go module cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		SizeEstimate: sizeEstimate,
		FreedBytes:   sizeEstimate.Value(), // Deprecated field
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
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
			Strategy:     domain.StrategyConservative,
		})
	}
	
	itemsRemoved := 0
	var totalSizeEstimate domain.SizeEstimate
	for _, match := range matches {
		// Calculate size before removal
		if !gcc.dryRun {
			bytesFreed := gcc.helper.getDirSize(match)
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
		Strategy:     domain.StrategyConservative,
	})
}