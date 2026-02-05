package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// SystemCacheCleaner handles macOS system cache cleanup.
type SystemCacheCleaner struct {
	verbose    bool
	dryRun     bool
	cacheTypes []SystemCacheType
	olderThan  time.Duration
}

// SystemCacheType represents different system cache types.
type SystemCacheType string

const (
	SystemCacheSpotlight SystemCacheType = "spotlight"
	SystemCacheXcode     SystemCacheType = "xcode"
	SystemCacheCocoaPods SystemCacheType = "cocoapods"
	SystemCacheHomebrew  SystemCacheType = "homebrew"
)

// AvailableSystemCacheTypes returns all available system cache types.
func AvailableSystemCacheTypes() []SystemCacheType {
	return []SystemCacheType{
		SystemCacheSpotlight,
		SystemCacheXcode,
		SystemCacheCocoaPods,
		SystemCacheHomebrew,
	}
}

// NewSystemCacheCleaner creates system cache cleaner.
func NewSystemCacheCleaner(verbose, dryRun bool, olderThan string) (*SystemCacheCleaner, error) {
	// Parse older than duration
	duration, err := domain.ParseCustomDuration(olderThan)
	if err != nil {
		return nil, fmt.Errorf("invalid older_than duration: %w", err)
	}

	// Default cache types to all available
	cacheTypes := AvailableSystemCacheTypes()

	return &SystemCacheCleaner{
		verbose:    verbose,
		dryRun:     dryRun,
		cacheTypes: cacheTypes,
		olderThan:  duration,
	}, nil
}

// Type returns operation type for system cache cleaner.
func (scc *SystemCacheCleaner) Type() domain.OperationType {
	return domain.OperationTypeSystemCache
}

// IsAvailable checks if system cache cleaner is available.
func (scc *SystemCacheCleaner) IsAvailable(ctx context.Context) bool {
	// System cache cleaner is only available on macOS
	return scc.isMacOS()
}

// ValidateSettings validates system cache cleaner settings.
func (scc *SystemCacheCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.SystemCache == nil {
		return nil // Settings are optional
	}

	// Validate cache types
	validCacheTypes := map[SystemCacheType]bool{
		SystemCacheSpotlight: true,
		SystemCacheXcode:     true,
		SystemCacheCocoaPods: true,
		SystemCacheHomebrew:  true,
	}

	for _, cacheType := range settings.SystemCache.CacheTypes {
		cacheStr := SystemCacheType(cacheType)
		if !validCacheTypes[cacheStr] {
			return fmt.Errorf("invalid cache type: %s (must be spotlight, xcode, cocoapods, or homebrew)", cacheType)
		}
	}

	return nil
}

// Scan scans for system caches.
func (scc *SystemCacheCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	if !scc.IsAvailable(ctx) {
		return result.Ok(items)
	}

	// Get home directory
	homeDir, err := GetHomeDir()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to get home directory: %w", err))
	}

	// Scan for each cache type
	for _, cacheType := range scc.cacheTypes {
		result := scc.scanSystemCache(ctx, cacheType, homeDir)
		if result.IsErr() {
			if scc.verbose {
				fmt.Printf("Warning: failed to scan %s: %v\n", cacheType, result.Error())
			}
			continue
		}

		items = append(items, result.Value()...)
	}

	return result.Ok(items)
}

// addScanItems appends scan items from a cache path to the items slice.
func (scc *SystemCacheCleaner) addScanItems(ctx context.Context, homeDir string, scanType domain.ScanType, pathComponents ...string) result.Result[[]domain.ScanItem] {
	path := filepath.Join(append([]string{homeDir}, pathComponents...)...)
	return scc.scanCachePath(ctx, path, scanType)
}

// cacheTypeConfig holds configuration for each system cache type.
type cacheTypeConfig struct {
	pathComponents []string
	displayName    string
	scanType        domain.ScanType
}

// systemCacheConfigs maps cache types to their configuration.
var systemCacheConfigs = map[SystemCacheType]cacheTypeConfig{
	SystemCacheSpotlight: {
		pathComponents: []string{"Library", "Metadata", "CoreSpotlight", "SpotlightKnowledgeEvents"},
		displayName:    "Spotlight metadata",
		scanType:        domain.ScanTypeTemp,
	},
	SystemCacheXcode: {
		pathComponents: []string{"Library", "Developer", "Xcode", "DerivedData"},
		displayName:    "Xcode DerivedData",
		scanType:        domain.ScanTypeTemp,
	},
	SystemCacheCocoaPods: {
		pathComponents: []string{"Library", "Caches", "CocoaPods"},
		displayName:    "CocoaPods cache",
		scanType:        domain.ScanTypeTemp,
	},
	SystemCacheHomebrew: {
		pathComponents: []string{"Library", "Caches", "Homebrew"},
		displayName:    "Homebrew cache",
		scanType:        domain.ScanTypeTemp,
	},
}

// scanSystemCache scans cache for a specific system cache type.
func (scc *SystemCacheCleaner) scanSystemCache(ctx context.Context, cacheType SystemCacheType, homeDir string) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	config, exists := systemCacheConfigs[cacheType]
	if !exists {
		return result.Err[[]domain.ScanItem](fmt.Errorf("unknown system cache type: %s", cacheType))
	}

	scanResult := scc.scanCachePathWithConfig(ctx, homeDir, config)
	if scanResult.IsErr() {
		return result.Err[[]domain.ScanItem](scanResult.Error())
	}
	items = append(items, scanResult.Value()...)

	return result.Ok(items)
}

// Clean removes system caches.
func (scc *SystemCacheCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !scc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("not available on this platform (requires macOS)"))
	}

	if scc.dryRun {
		// Estimate cache sizes based on typical usage
		totalBytes := int64(1 * 1024 * 1024 * 1024) // Estimate 1GB total
		itemsRemoved := len(scc.cacheTypes)

		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, itemsRemoved, totalBytes)
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	// Get home directory
	homeDir, err := GetHomeDir()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get home directory: %w", err))
	}

	// Clean for each cache type
	for _, cacheType := range scc.cacheTypes {
		result := scc.cleanSystemCache(ctx, cacheType, homeDir)
		if result.IsErr() {
			itemsFailed++
			if scc.verbose {
				fmt.Printf("Warning: failed to clean %s: %v\n", cacheType, result.Error())
			}
			continue
		}

		cleanResult := result.Value()
		itemsRemoved++
		bytesFreed += int64(cleanResult.FreedBytes)
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

// removeCachePath removes a cache directory and returns the appropriate result.
func (scc *SystemCacheCleaner) removeCachePath(path, successMessage string) result.Result[domain.CleanResult] {
	if scc.dryRun {
		if scc.verbose {
			fmt.Printf("  [DRY RUN] Would remove: %s\n", path)
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

	err := os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to remove %s: %w", path, err))
	}

	if scc.verbose {
		fmt.Println("  ✓ " + successMessage)
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

// scanCachePath scans a cache directory and returns scan items.
func (scc *SystemCacheCleaner) scanCachePath(ctx context.Context, path string, scanType domain.ScanType) result.Result[[]domain.ScanItem] {
	scanResult := ScanDirectory(path, scanType, scc.verbose)
	return result.Ok(scanResult.Items)
}

// scanCachePathWithConfig scans a cache directory using configuration and returns scan items.
func (scc *SystemCacheCleaner) scanCachePathWithConfig(ctx context.Context, homeDir string, config cacheTypeConfig) result.Result[[]domain.ScanItem] {
	scanResult := ScanPath(homeDir, config.scanType, config.displayName, scc.verbose, "", config.pathComponents...)
	return result.Ok(scanResult.Items)
}

// cleanSystemCache cleans cache for a specific system cache type.
func (scc *SystemCacheCleaner) cleanSystemCache(ctx context.Context, cacheType SystemCacheType, homeDir string) result.Result[domain.CleanResult] {
	config, exists := systemCacheConfigs[cacheType]
	if !exists {
		return result.Err[domain.CleanResult](fmt.Errorf("unknown system cache type: %s", cacheType))
	}

	path := filepath.Join(append([]string{homeDir}, config.pathComponents...)...)
	return scc.removeCachePath(path, config.displayName+" cleaned")
}

// isMacOS checks if the system is macOS.
func (scc *SystemCacheCleaner) isMacOS() bool {
	// Simple check for macOS
	return os.Getenv("GOOS") == "darwin" || os.Getenv("OSTYPE") == "darwin"
}
