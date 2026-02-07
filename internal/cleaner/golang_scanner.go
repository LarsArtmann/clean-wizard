package cleaner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// GoScanner handles scanning for Go caches.
type GoScanner struct {
	verbose bool
	helper  *golangHelpers
}

// NewGoScanner creates a new GoScanner.
func NewGoScanner(verbose bool) *GoScanner {
	return &GoScanner{
		verbose: verbose,
		helper:  &golangHelpers{},
	}
}

// Scan scans for all enabled cache types.
func (gs *GoScanner) Scan(ctx context.Context, caches GoCacheType) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	// Scan GOCACHE
	if caches.Has(GoCacheGOCACHE) {
		items = append(items, gs.scanGoCache(ctx)...)
	}

	// Scan GOMODCACHE
	if caches.Has(GoCacheModCache) {
		items = append(items, gs.scanGoModCache(ctx)...)
	}

	// Scan build cache folders
	if caches.Has(GoCacheBuildCache) {
		items = append(items, gs.scanGoBuildCache()...)
	}

	// Scan lint cache
	if caches.Has(GoCacheLintCache) {
		items = append(items, gs.scanLintCache()...)
	}

	return result.Ok(items)
}

// scanGoCache scans GOCACHE.
func (gs *GoScanner) scanGoCache(ctx context.Context) []domain.ScanItem {
	return gs.scanGoEnvCache(ctx, "GOCACHE", "Go cache")
}

// scanGoModCache scans GOMODCACHE.
func (gs *GoScanner) scanGoModCache(ctx context.Context) []domain.ScanItem {
	return gs.scanGoEnvCache(ctx, "GOMODCACHE", "Go module cache")
}

// addScanItem creates a scan item for a cache directory and appends it to items.
func (gs *GoScanner) addScanItem(items []domain.ScanItem, path, cacheName string) []domain.ScanItem {
	items = append(items, domain.ScanItem{
		Path:     path,
		Size:     GetDirSize(path),
		Created:  GetDirModTime(path),
		ScanType: domain.ScanTypeTemp,
	})

	if gs.verbose {
		fmt.Printf("Found %s: %s\n", cacheName, path)
	}
	return items
}

// scanGoEnvCache scans a Go environment variable cache path.
func (gs *GoScanner) scanGoEnvCache(ctx context.Context, envVar, cacheName string) []domain.ScanItem {
	items := make([]domain.ScanItem, 0)
	cachePath, err := gs.helper.getGoEnv(ctx, envVar)
	if err == nil && cachePath != "" {
		items = gs.addScanItem(items, cachePath, cacheName)
	}
	return items
}

// scanGoBuildCache scans go-build* folders.
func (gs *GoScanner) scanGoBuildCache() []domain.ScanItem {
	items := make([]domain.ScanItem, 0)
	buildCachePattern := "go-build*"
	tempDir := filepath.Join("/", "tmp", "build")
	if homeDir := gs.helper.getHomeDir(); homeDir != "" {
		tempDir = filepath.Join(homeDir, "Library", "Caches")
	}

	matches, err := filepath.Glob(filepath.Join(tempDir, buildCachePattern))
	if err == nil {
		for _, match := range matches {
			items = gs.addScanItem(items, match, "Go build cache")
		}
	}
	return items
}

// scanLintCache scans for golangci-lint cache.
func (gs *GoScanner) scanLintCache() []domain.ScanItem {
	items := make([]domain.ScanItem, 0)
	cacheDir := gs.detectLintCacheDir()
	if cacheDir != "" {
		items = gs.addScanItem(items, cacheDir, "golangci-lint cache")
	}
	return items
}

// detectLintCacheDir finds golangci-lint cache directory.
func (gs *GoScanner) detectLintCacheDir() string {
	// Try XDG_CACHE_HOME first
	if xdgCache := gs.helper.getEnv("XDG_CACHE_HOME"); xdgCache != "" {
		cacheDir := filepath.Join(xdgCache, "golangci-lint")
		if gs.helper.pathExists(cacheDir) {
			return cacheDir
		}
	}

	// Fallback to ~/.cache
	if home := gs.helper.getEnv("HOME"); home != "" {
		cacheDir := filepath.Join(home, ".cache", "golangci-lint")
		if gs.helper.pathExists(cacheDir) {
			return cacheDir
		}
	}

	return ""
}
