package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// SystemCacheCleaner handles system cache cleanup for macOS and Linux.
type SystemCacheCleaner struct {
	verbose    bool
	dryRun     bool
	cacheTypes []domain.CacheType
	olderThan  time.Duration
}

// AvailableSystemCacheTypes returns all available system cache types for the current platform.
func AvailableSystemCacheTypes() []domain.CacheType {
	// Use runtime check at call time for accurate platform detection
	switch runtime.GOOS {
	case "darwin":
		return []domain.CacheType{
			domain.CacheTypeSpotlight,
			domain.CacheTypeXcode,
			domain.CacheTypeCocoapods,
			domain.CacheTypeHomebrew,
		}
	case "linux":
		return []domain.CacheType{
			domain.CacheTypeXdgCache,
			domain.CacheTypeThumbnails,
			domain.CacheTypeHomebrew, // Homebrew works on Linux too
			domain.CacheTypePip,
			domain.CacheTypeNpm,
			domain.CacheTypeYarn,
			domain.CacheTypeCcache,
		}
	default:
		return []domain.CacheType{}
	}
}

// NewSystemCacheCleaner creates system cache cleaner.
func NewSystemCacheCleaner(verbose, dryRun bool, olderThan string, cacheTypes []domain.CacheType) (*SystemCacheCleaner, error) {
	// Parse older than duration
	duration, err := domain.ParseCustomDuration(olderThan)
	if err != nil {
		return nil, fmt.Errorf("invalid older_than duration: %w", err)
	}

	// Default to all available cache types if none specified
	if len(cacheTypes) == 0 {
		cacheTypes = AvailableSystemCacheTypes()
	}

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

// Name returns the cleaner name for result tracking.
func (scc *SystemCacheCleaner) Name() string {
	return "systemcache"
}

// IsAvailable checks if system cache cleaner is available.
func (scc *SystemCacheCleaner) IsAvailable(ctx context.Context) bool {
	// System cache cleaner is available on macOS and Linux
	return scc.isMacOS() || scc.isLinux()
}

// ValidateSettings validates system cache cleaner settings.
func (scc *SystemCacheCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.SystemCache == nil {
		return nil // Settings are optional
	}

	// Create valid cache types map
	validCacheTypes := make(map[domain.CacheType]bool)
	for _, ct := range AvailableSystemCacheTypes() {
		validCacheTypes[ct] = true
	}

	// Validate each CacheType in settings
	for i, ct := range settings.SystemCache.CacheTypes {
		if !ct.IsValid() {
			return fmt.Errorf("invalid CacheType at index %d: %d is not a valid cache type", i, ct)
		}
		if !validCacheTypes[ct] {
			return fmt.Errorf("invalid default CacheType at index %d: %d not supported on current platform (valid types: %v)", i, ct, validCacheTypes)
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
	scanType       domain.ScanType
}

// systemCacheConfigs maps cache types to their configuration.
var systemCacheConfigs = map[domain.CacheType]cacheTypeConfig{
	// macOS-specific cache types
	domain.CacheTypeSpotlight: {
		pathComponents: []string{"Library", "Metadata", "CoreSpotlight", "SpotlightKnowledgeEvents"},
		displayName:    "Spotlight metadata",
		scanType:       domain.ScanTypeTemp,
	},
	domain.CacheTypeXcode: {
		pathComponents: []string{"Library", "Developer", "Xcode", "DerivedData"},
		displayName:    "Xcode DerivedData",
		scanType:       domain.ScanTypeTemp,
	},
	domain.CacheTypeCocoapods: {
		pathComponents: []string{"Library", "Caches", "CocoaPods"},
		displayName:    "CocoaPods cache",
		scanType:       domain.ScanTypeTemp,
	},
	domain.CacheTypeHomebrew: {
		pathComponents: []string{"Library", "Caches", "Homebrew"},
		displayName:    "Homebrew cache",
		scanType:       domain.ScanTypeTemp,
	},
	// Linux-specific cache types
	domain.CacheTypeXdgCache: {
		pathComponents: []string{".cache"},
		displayName:    "XDG cache",
		scanType:       domain.ScanTypeTemp,
	},
	domain.CacheTypeThumbnails: {
		pathComponents: []string{".cache", "thumbnails"},
		displayName:    "Thumbnail cache",
		scanType:       domain.ScanTypeTemp,
	},
}

// scanSystemCache scans cache for a specific system cache type.
func (scc *SystemCacheCleaner) scanSystemCache(ctx context.Context, cacheType domain.CacheType, homeDir string) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	config, exists := systemCacheConfigs[cacheType]
	if !exists {
		return result.Err[[]domain.ScanItem](fmt.Errorf("unknown system cache type: %s", cacheType.String()))
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
		return result.Err[domain.CleanResult](errors.New("not available on this platform (requires macOS or Linux)"))
	}

	if scc.dryRun {
		// Estimate cache sizes based on typical usage
		totalBytes := int64(1 * 1024 * 1024 * 1024) // Estimate 1GB total
		itemsRemoved := len(scc.cacheTypes)

		cleanResult := conversions.NewCleanResult(domain.CleanStrategyType(domain.StrategyDryRunType), itemsRemoved, totalBytes)
		// Set SizeEstimate properly
		cleanResult.SizeEstimate = domain.SizeEstimate{
			Known:  uint64(totalBytes),
			Status: domain.SizeEstimateStatusKnown,
		}
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

	var status domain.SizeEstimateStatusType
	if bytesFreed > 0 {
		status = domain.SizeEstimateStatusKnown
	} else {
		status = domain.SizeEstimateStatusUnknown
	}

	duration := time.Since(startTime)
	cleanResult := domain.CleanResult{
		SizeEstimate: domain.SizeEstimate{
			Known:  uint64(bytesFreed),
			Status: status,
		},
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  uint(itemsFailed),
		CleanTime:    duration,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	}

	return result.Ok(cleanResult)
}

// removeCachePath removes a cache directory and returns the appropriate result.
func (scc *SystemCacheCleaner) removeCachePath(path, successMessage string) result.Result[domain.CleanResult] {
	if scc.dryRun {
		// Estimate size for dry-run
		estimatedSize := GetDirSize(path)
		if scc.verbose {
			fmt.Printf("  [DRY RUN] Would remove: %s (%s)\n", path, format.Bytes(estimatedSize))
		}
		return result.Ok(domain.CleanResult{
			SizeEstimate: domain.SizeEstimate{
				Known:  uint64(estimatedSize),
				Status: domain.SizeEstimateStatusKnown,
			},
			FreedBytes:   uint64(estimatedSize),
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
		})
	}

	// Measure size before removal
	bytesFreed := GetDirSize(path)

	err := os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to remove %s: %w", path, err))
	}

	if scc.verbose {
		fmt.Printf("  ✓ %s (%s freed)\n", successMessage, format.Bytes(bytesFreed))
	}

	return result.Ok(domain.CleanResult{
		SizeEstimate: domain.SizeEstimate{
			Known:  uint64(bytesFreed),
			Status: domain.SizeEstimateStatusKnown,
		},
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
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
func (scc *SystemCacheCleaner) cleanSystemCache(ctx context.Context, cacheType domain.CacheType, homeDir string) result.Result[domain.CleanResult] {
	config, exists := systemCacheConfigs[cacheType]
	if !exists {
		return result.Err[domain.CleanResult](fmt.Errorf("unknown system cache type: %s", cacheType.String()))
	}

	path := filepath.Join(append([]string{homeDir}, config.pathComponents...)...)
	return scc.removeCachePath(path, config.displayName+" cleaned")
}

// isMacOS checks if the system is macOS.
func (scc *SystemCacheCleaner) isMacOS() bool {
	return runtime.GOOS == "darwin"
}

// isLinux checks if the system is Linux.
func (scc *SystemCacheCleaner) isLinux() bool {
	return runtime.GOOS == "linux"
}

// platform returns the current platform name for logging.
func (scc *SystemCacheCleaner) platform() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return runtime.GOOS
	}
}
