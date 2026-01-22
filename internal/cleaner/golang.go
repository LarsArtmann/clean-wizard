package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// GoCleaner handles Go language cleanup.
type GoCleaner struct {
	verbose          bool
	dryRun           bool
	cleanCache       bool
	cleanTestCache   bool
	cleanModCache    bool
	cleanBuildCache  bool
	cleanLintCache   bool
}

// NewGoCleaner creates Go cleaner.
func NewGoCleaner(verbose, dryRun, cleanCache, cleanTestCache, cleanModCache, cleanBuildCache, cleanLintCache bool) *GoCleaner {
	return &GoCleaner{
		verbose:          verbose,
		dryRun:           dryRun,
		cleanCache:       cleanCache,
		cleanTestCache:   cleanTestCache,
		cleanModCache:    cleanModCache,
		cleanBuildCache:  cleanBuildCache,
		cleanLintCache:   cleanLintCache,
	}
}

// Type returns operation type for Go cleaner.
func (gc *GoCleaner) Type() domain.OperationType {
	return domain.OperationTypeGoPackages
}

// IsAvailable checks if Go is available.
func (gc *GoCleaner) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("go")
	return err == nil
}

// ValidateSettings validates Go cleaner settings.
func (gc *GoCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.GoPackages == nil {
		return nil // Settings are optional
	}

	// All Go settings are valid by default
	return nil
}

// Scan scans for Go caches.
func (gc *GoCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	// Get GOCACHE location
	cachePath, err := gc.getGoEnv(ctx, "GOCACHE")
	if err == nil && cachePath != "" {
		items = append(items, domain.ScanItem{
			Path:     cachePath,
			Size:     gc.getDirSize(cachePath),
			Created:  gc.getDirModTime(cachePath),
			ScanType: domain.ScanTypeTemp,
		})

		if gc.verbose {
			fmt.Printf("Found Go cache: %s\n", cachePath)
		}
	}

	// Get GOMODCACHE location
	modCachePath, err := gc.getGoEnv(ctx, "GOMODCACHE")
	if err == nil && modCachePath != "" {
		items = append(items, domain.ScanItem{
			Path:     modCachePath,
			Size:     gc.getDirSize(modCachePath),
			Created:  gc.getDirModTime(modCachePath),
			ScanType: domain.ScanTypeTemp,
		})

		if gc.verbose {
			fmt.Printf("Found Go module cache: %s\n", modCachePath)
		}
	}

	// Check for build cache folders
	buildCachePattern := "go-build*"
	tempDir := filepath.Join("/", "tmp", "build")
	if homeDir := gc.getHomeDir(); homeDir != "" {
		tempDir = filepath.Join(homeDir, "Library", "Caches")
	}

	matches, err := filepath.Glob(filepath.Join(tempDir, buildCachePattern))
	if err == nil {
		for _, match := range matches {
			items = append(items, domain.ScanItem{
				Path:     match,
				Size:     gc.getDirSize(match),
				Created:  gc.getDirModTime(match),
				ScanType: domain.ScanTypeTemp,
			})

			if gc.verbose {
				fmt.Printf("Found Go build cache: %s\n", match)
			}
		}
	}

	return result.Ok(items)
}

// Clean removes Go caches.
func (gc *GoCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !gc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("Go not available"))
	}

	if gc.dryRun {
		// Estimate cache sizes based on typical usage
		totalBytes := int64(200 * 1024 * 1024) // Estimate 200MB
		itemsRemoved := 0

		if gc.cleanCache {
			itemsRemoved++
		}
		if gc.cleanTestCache {
			itemsRemoved++
		}
		if gc.cleanModCache {
			itemsRemoved++
		}
		if gc.cleanBuildCache {
			itemsRemoved++
		}
		if gc.cleanLintCache {
			itemsRemoved++
		}

		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, itemsRemoved, totalBytes)
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	// Clean GOCACHE
	if gc.cleanCache {
		result := gc.cleanGoCache(ctx)
		if result.IsErr() {
			itemsFailed++
			if gc.verbose {
				fmt.Printf("Warning: failed to clean Go cache: %v\n", result.Error())
			}
		} else {
			itemsRemoved++
			bytesFreed += int64(result.Value().FreedBytes)
		}
	}

	// Clean GOTESTCACHE
	if gc.cleanTestCache {
		result := gc.cleanGoTestCache(ctx)
		if result.IsErr() {
			itemsFailed++
			if gc.verbose {
				fmt.Printf("Warning: failed to clean Go test cache: %v\n", result.Error())
			}
		} else {
			itemsRemoved++
			bytesFreed += int64(result.Value().FreedBytes)
		}
	}

	// Clean GOMODCACHE
	if gc.cleanModCache {
		result := gc.cleanGoModCache(ctx)
		if result.IsErr() {
			itemsFailed++
			if gc.verbose {
				fmt.Printf("Warning: failed to clean Go module cache: %v\n", result.Error())
			}
		} else {
			itemsRemoved++
			bytesFreed += int64(result.Value().FreedBytes)
		}
	}

	// Clean build cache folders
	if gc.cleanBuildCache {
		result := gc.cleanGoBuildCache(ctx)
		if result.IsErr() {
			itemsFailed++
			if gc.verbose {
				fmt.Printf("Warning: failed to clean Go build cache: %v\n", result.Error())
			}
		} else {
			itemsRemoved++
			bytesFreed += int64(result.Value().FreedBytes)
		}
	}

	// Clean golangci-lint cache
	if gc.cleanLintCache {
		result := gc.cleanGolangciLintCache(ctx)
		if result.IsErr() {
			itemsFailed++
			if gc.verbose {
				fmt.Printf("Warning: failed to clean golangci-lint cache: %v\n", result.Error())
			}
		} else {
			itemsRemoved++
			bytesFreed += int64(result.Value().FreedBytes)
		}
	}

	duration := time.Since(startTime)
	cleanResult := domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  uint(itemsFailed),
		CleanTime:    duration,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(cleanResult)
}

// cleanGoCache cleans GOCACHE.
func (gc *GoCleaner) cleanGoCache(ctx context.Context) result.Result[domain.CleanResult] {
	// Get cache path first to calculate size
	cachePath, err := gc.getGoEnv(ctx, "GOCACHE")
	if err != nil || cachePath == "" {
		// If we can't get the path, still try to clean
		cmd := exec.CommandContext(ctx, "go", "clean", "-cache")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("go clean -cache failed: %w (output: %s)", err, string(output)))
		}
		if gc.verbose {
			fmt.Println("  ✓ Go cache cleaned")
		}
		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})
	}

	// Calculate size before cleaning
	bytesFreed := int64(0)
	if !gc.dryRun {
		bytesFreed = gc.getDirSize(cachePath)
	}

	cmd := exec.CommandContext(ctx, "go", "clean", "-cache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("go clean -cache failed: %w (output: %s)", err, string(output)))
	}

	if gc.verbose {
		fmt.Println("  ✓ Go cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
}

// cleanGoTestCache cleans GOTESTCACHE.
func (gc *GoCleaner) cleanGoTestCache(ctx context.Context) result.Result[domain.CleanResult] {
	// Get cache path first to calculate size
	cachePath, err := gc.getGoEnv(ctx, "GOCACHE")
	if err != nil || cachePath == "" {
		// If we can't get the path, still try to clean
		cmd := exec.CommandContext(ctx, "go", "clean", "-testcache")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("go clean -testcache failed: %w (output: %s)", err, string(output)))
		}
		if gc.verbose {
			fmt.Println("  ✓ Go test cache cleaned")
		}
		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})
	}

	// Calculate size before cleaning (test cache is part of GOCACHE)
	// Note: test cache doesn't have a separate path, it's managed internally by Go
	// We'll estimate it as 0 since we can't accurately measure test cache size separately
	bytesFreed := int64(0)

	cmd := exec.CommandContext(ctx, "go", "clean", "-testcache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("go clean -testcache failed: %w (output: %s)", err, string(output)))
	}

	if gc.verbose {
		fmt.Println("  ✓ Go test cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
}

// cleanGoModCache cleans GOMODCACHE.
func (gc *GoCleaner) cleanGoModCache(ctx context.Context) result.Result[domain.CleanResult] {
	// Get cache path first to calculate size
	cachePath, err := gc.getGoEnv(ctx, "GOMODCACHE")
	if err != nil || cachePath == "" {
		// If we can't get the path, still try to clean
		cmd := exec.CommandContext(ctx, "go", "clean", "-modcache")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("go clean -modcache failed: %w (output: %s)", err, string(output)))
		}
		if gc.verbose {
			fmt.Println("  ✓ Go module cache cleaned")
		}
		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})
	}

	// Calculate size before cleaning
	bytesFreed := int64(0)
	if !gc.dryRun {
		bytesFreed = gc.getDirSize(cachePath)
	}

	cmd := exec.CommandContext(ctx, "go", "clean", "-modcache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("go clean -modcache failed: %w (output: %s)", err, string(output)))
	}

	if gc.verbose {
		fmt.Println("  ✓ Go module cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
}

// cleanGoBuildCache removes go-build* folders.
func (gc *GoCleaner) cleanGoBuildCache(ctx context.Context) result.Result[domain.CleanResult] {
	buildCachePattern := "go-build*"
	tempDir := filepath.Join("/", "tmp")
	if homeDir := gc.getHomeDir(); homeDir != "" {
		tempDir = filepath.Join(homeDir, "Library", "Caches")
	}

	matches, err := filepath.Glob(filepath.Join(tempDir, buildCachePattern))
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to find build cache folders: %w", err))
	}

	itemsRemoved := 0
	bytesFreed := int64(0)
	for _, match := range matches {
		// Calculate size before removal
		if !gc.dryRun {
			bytesFreed += gc.getDirSize(match)
		}

		if gc.dryRun {
			itemsRemoved++
			continue
		}

		err := os.RemoveAll(match)
		if err != nil {
			if gc.verbose {
				fmt.Printf("Warning: failed to remove %s: %v\n", match, err)
			}
			continue
		}

		itemsRemoved++
		if gc.verbose {
			fmt.Printf("  ✓ Removed build cache: %s\n", match)
		}
	}

	if gc.verbose && itemsRemoved > 0 {
		fmt.Println("  ✓ Go build cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
}

// cleanGolangciLintCache cleans golangci-lint cache.
func (gc *GoCleaner) cleanGolangciLintCache(ctx context.Context) result.Result[domain.CleanResult] {
	_, err := exec.LookPath("golangci-lint")
	if err != nil {
		if gc.verbose {
			fmt.Println("  ⚠️  golangci-lint not found, skipping cache cleanup")
		}
		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: 0,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})
	}

	cmd := exec.CommandContext(ctx, "golangci-lint", "cache", "clean")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("golangci-lint cache clean failed: %w (output: %s)", err, string(output)))
	}

	if gc.verbose {
		fmt.Println("  ✓ golangci-lint cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   0,
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
}

// getGoEnv returns Go environment variable value.
func (gc *GoCleaner) getGoEnv(ctx context.Context, key string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "env", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Go env %s: %w", key, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getHomeDir returns user's home directory.
func (gc *GoCleaner) getHomeDir() string {
	// Try getting from HOME environment variable
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	// Fallback to user profile directory
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return userProfile
	}

	return ""
}

// getDirSize returns total size of directory recursively.
func (gc *GoCleaner) getDirSize(path string) int64 {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0
	}

	return size
}

// getDirModTime returns the most recent modification time in directory.
func (gc *GoCleaner) getDirModTime(path string) time.Time {
	var modTime time.Time

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.ModTime().After(modTime) {
			modTime = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return time.Time{}
	}

	return modTime
}
