package systemcache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

const (
	// pathComponentLibrary is the Library directory component for macOS paths.
	pathComponentLibrary	= "Library"
	// pathComponentDotCache is the .cache directory component for Linux paths.
	pathComponentDotCache	= ".cache"
)

// SystemCacheCleaner handles system cache cleanup for macOS and Linux.
type SystemCacheCleaner struct {
	cleaner.CleanerBase

	cacheTypes	[]enums.CacheType
	olderThan	time.Duration
}

// AvailableSystemCacheTypes returns all available system cache types for the current platform.
func AvailableSystemCacheTypes() []enums.CacheType {
	// Use runtime check at call time for accurate platform detection
	switch runtime.GOOS {
	case "darwin":
		return []enums.CacheType{
			enums.CacheTypeSpotlight,
			enums.CacheTypeXcode,
			enums.CacheTypeCocoapods,
			enums.CacheTypeHomebrew,
			enums.CacheTypePuppeteer,
			enums.CacheTypeTerraform,
			enums.CacheTypeGradleWrapper,
			enums.CacheTypeKonan,
			enums.CacheTypeRustup,
			enums.CacheTypeGopls,
			enums.CacheTypeGoimports,
			enums.CacheTypeJetBrains,
			enums.CacheTypeBunCache,
			enums.CacheTypePlaywright,
			enums.CacheTypeMozilla,
			enums.CacheTypeNixCache,
			enums.CacheTypeZig,
			enums.CacheTypeUv,
			enums.CacheTypeTinygo,
		}
	case "linux":
		return []enums.CacheType{
			enums.CacheTypeXdgCache,
			enums.CacheTypeThumbnails,
			enums.CacheTypePip,
			enums.CacheTypeNpm,
			enums.CacheTypeYarn,
			enums.CacheTypeCcache,
			enums.CacheTypePuppeteer,
			enums.CacheTypeTerraform,
			enums.CacheTypeGradleWrapper,
			enums.CacheTypeKonan,
			enums.CacheTypeRustup,
			enums.CacheTypeGopls,
			enums.CacheTypeGoimports,
			enums.CacheTypeJetBrains,
			enums.CacheTypeBunCache,
			enums.CacheTypePlaywright,
			enums.CacheTypeMozilla,
			enums.CacheTypeNixCache,
			enums.CacheTypeZig,
			enums.CacheTypeUv,
			enums.CacheTypeTinygo,
			enums.CacheTypeMesaShader,
			enums.CacheTypeComgr,
		}
	default:
		return []enums.CacheType{}
	}
}

// NewSystemCacheCleaner creates system cache cleaner.
func NewSystemCacheCleaner(
	verbose, dryRun bool, olderThan string, cacheTypes []enums.CacheType,
) (*SystemCacheCleaner, error) {
	// Parse older than duration
	duration, err := operations.ParseCustomDuration(olderThan)
	if err != nil {
		return nil, fmt.Errorf("invalid older_than duration for olderThan=%v: %w", olderThan, err)
	}

	// Default to all available cache types if none specified
	if len(cacheTypes) == 0 {
		cacheTypes = AvailableSystemCacheTypes()
	}

	return &SystemCacheCleaner{
		CleanerBase:	cleaner.NewCleanerBase(verbose, dryRun),
		cacheTypes:	cacheTypes,
		olderThan:	duration,
	}, nil
}

// Type returns operation type for system cache cleaner.
func (scc *SystemCacheCleaner) Type() operations.OperationType {
	return operations.OperationTypeSystemCache
}

// Name returns the cleaner name for result tracking.
func (scc *SystemCacheCleaner) Name() string {
	return cleaner.CleanerSystemCache
}

// IsAvailable checks if system cache cleaner is available.
func (scc *SystemCacheCleaner) IsAvailable(_ context.Context) bool {
	// System cache cleaner is available on macOS and Linux
	return scc.isMacOS() || scc.isLinux()
}

// ValidateSettings validates system cache cleaner settings.
func (scc *SystemCacheCleaner) ValidateSettings(settings *operations.OperationSettings) error {
	return cleaner.ValidateOptionalSettings(
		settings,
		func(s *operations.OperationSettings) *operations.SystemCacheSettings { return s.SystemCache },
		validateSystemCacheSettings,
	)
}

// validateSystemCacheSettings validates a non-nil SystemCacheSettings struct.
func validateSystemCacheSettings(sc *operations.SystemCacheSettings) error {
	// Create valid cache types map
	validCacheTypes := make(map[enums.CacheType]bool)
	for _, ct := range AvailableSystemCacheTypes() {
		validCacheTypes[ct] = true
	}

	// Validate each CacheType in settings
	for i, ct := range sc.CacheTypes {
		if !ct.IsValid() {
			return fmt.Errorf("invalid CacheType at index %d: %d is not a valid cache type", i, ct)
		}

		if !validCacheTypes[ct] {
			return fmt.Errorf(
				"invalid default CacheType at index %d: %d not supported on current platform (valid types: %v)",
				i,
				ct,
				validCacheTypes,
			)
		}
	}

	return nil
}

// Scan scans for system caches.
func (scc *SystemCacheCleaner) Scan(ctx context.Context) result.Result[[]types.ScanItem] {
	items := make([]types.ScanItem, 0)

	if !scc.IsAvailable(ctx) {
		return result.Ok(items)
	}

	// Get home directory
	homeDir, err := cleaner.GetHomeDir()
	if err != nil {
		return result.Err[[]types.ScanItem](fmt.Errorf("failed to get home directory: %w", err))
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

// cacheTypeConfig holds configuration for each system cache type.
type cacheTypeConfig struct {
	pathComponents	[]string
	displayName	string
	scanType	types.ScanType
}

// systemCacheConfigs maps cache types to their configuration.
var systemCacheConfigs = map[enums.CacheType]cacheTypeConfig{	//nolint:gochecknoglobals
	// macOS-specific cache types
	enums.CacheTypeSpotlight: {
		pathComponents: []string{
			pathComponentLibrary,
			"Metadata",
			"CoreSpotlight",
			"SpotlightKnowledgeEvents",
		},
		displayName:	"Spotlight metadata",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeXcode: {
		pathComponents:	[]string{pathComponentLibrary, "Developer", "Xcode", "DerivedData"},
		displayName:	"Xcode DerivedData",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeCocoapods: {
		pathComponents:	[]string{pathComponentLibrary, "Caches", "CocoaPods"},
		displayName:	"CocoaPods cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeHomebrew: {
		pathComponents:	[]string{pathComponentLibrary, "Caches", "Homebrew"},
		displayName:	"Homebrew cache",
		scanType:	types.ScanTypeTemp,
	},
	// Linux-specific cache types
	enums.CacheTypeXdgCache: {
		pathComponents:	[]string{pathComponentDotCache},
		displayName:	"XDG cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeThumbnails: {
		pathComponents:	[]string{pathComponentDotCache, "thumbnails"},
		displayName:	"Thumbnail cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypePip: {
		pathComponents:	[]string{pathComponentDotCache, "pip"},
		displayName:	"Pip cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeNpm: {
		pathComponents:	[]string{pathComponentDotCache, "npm"},
		displayName:	"NPM cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeYarn: {
		pathComponents:	[]string{pathComponentDotCache, "yarn"},
		displayName:	"Yarn cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeCcache: {
		pathComponents:	[]string{pathComponentDotCache, "ccache"},
		displayName:	"Ccache",
		scanType:	types.ScanTypeTemp,
	},
	// Cross-platform cache types
	enums.CacheTypePuppeteer: {
		pathComponents:	[]string{pathComponentDotCache, "puppeteer"},
		displayName:	"Puppeteer browser cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeTerraform: {
		pathComponents:	[]string{".terraform.d", "plugin-cache"},
		displayName:	"Terraform plugin cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeGradleWrapper: {
		pathComponents:	[]string{".gradle", "wrapper"},
		displayName:	"Gradle wrapper distributions",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeKonan: {
		pathComponents:	[]string{".konan", "dependencies"},
		displayName:	"Kotlin/Native toolchain dependencies",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeRustup: {
		pathComponents:	[]string{".rustup", "toolchains"},
		displayName:	"Rust toolchain cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeGopls: {
		pathComponents:	[]string{pathComponentDotCache, "gopls"},
		displayName:	"gopls language server cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeGoimports: {
		pathComponents:	[]string{pathComponentDotCache, "goimports"},
		displayName:	"goimports cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeJetBrains: {
		pathComponents:	[]string{pathComponentDotCache, "JetBrains"},
		displayName:	"JetBrains IDE cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeBunCache: {
		pathComponents:	[]string{pathComponentDotCache, "bun"},
		displayName:	"Bun cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypePlaywright: {
		pathComponents:	[]string{pathComponentDotCache, "ms-playwright"},
		displayName:	"Playwright browser cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeMozilla: {
		pathComponents:	[]string{pathComponentDotCache, "mozilla"},
		displayName:	"Mozilla/Firefox cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeNixCache: {
		pathComponents:	[]string{pathComponentDotCache, "nix"},
		displayName:	"Nix evaluator/substituter cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeZig: {
		pathComponents:	[]string{pathComponentDotCache, "zig"},
		displayName:	"Zig compiler cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeUv: {
		pathComponents:	[]string{pathComponentDotCache, "uv"},
		displayName:	"uv Python package manager cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeTinygo: {
		pathComponents:	[]string{pathComponentDotCache, "tinygo"},
		displayName:	"TinyGo compiler cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeMesaShader: {
		pathComponents:	[]string{pathComponentDotCache, "mesa_shader_cache"},
		displayName:	"Mesa shader cache",
		scanType:	types.ScanTypeTemp,
	},
	enums.CacheTypeComgr: {
		pathComponents:	[]string{pathComponentDotCache, "comgr"},
		displayName:	"AMD GPU compiler cache",
		scanType:	types.ScanTypeTemp,
	},
}

// scanSystemCache scans cache for a specific system cache type.
func (scc *SystemCacheCleaner) scanSystemCache(
	ctx context.Context,
	cacheType enums.CacheType,
	homeDir string,
) result.Result[[]types.ScanItem] {
	config, exists := systemCacheConfigs[cacheType]
	if !exists {
		return result.Err[[]types.ScanItem](
			fmt.Errorf(
				"unknown system cache type %s for homeDir=%v: %w",
				cacheType.String(),
				homeDir,
				errors.New("cache type not found"),
			),
		)
	}

	return scc.scanCachePathWithConfig(ctx, homeDir, config)
}

// Clean removes system caches.
func (scc *SystemCacheCleaner) Clean(ctx context.Context) result.Result[types.CleanResult] {
	if !scc.IsAvailable(ctx) {
		return result.Err[types.CleanResult](
			cleaner.NewNotAvailableError("systemcache", "requires macOS or Linux"),
		)
	}

	if scc.dryRun {
		// Scan actual cache directories to get real sizes
		scanResult := scc.Scan(ctx)

		var (
			totalBytes	int64
			itemsRemoved	int
		)

		if scanResult.IsOk() {
			items := scanResult.Value()

			itemsRemoved = len(items)
			for _, item := range items {
				totalBytes += item.Size
			}
		} else {
			// Fallback to counting cache types if scan fails
			itemsRemoved = len(scc.cacheTypes)
		}

		cleanResult := conversions.NewCleanResult(
			enums.StrategyDryRunType,
			itemsRemoved,
			totalBytes,
		)
		cleanResult.SizeEstimate = types.SizeEstimate{
			Known:	uint64(totalBytes),
			Status:	enums.SizeEstimateStatusKnown,
		}

		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	counters := cleaner.NewCleanCounters()

	// Get home directory
	homeDir, err := cleaner.GetHomeDir()
	if err != nil {
		return result.Err[types.CleanResult](
			fmt.Errorf("failed to get home directory (itemsFailed=%v): %w", counters.ItemsFailed, err),
		)
	}

	// Clean for each cache type
	for _, cacheType := range scc.cacheTypes {
		result := scc.cleanSystemCache(ctx, cacheType, homeDir)
		if result.IsErr() {
			counters.RecordFailure(scc.verbose, cacheType, result.Error())

			continue
		}

		cleanResult := result.Value()
		counters.RecordSuccess(int64(cleanResult.FreedBytes))
	}

	var status enums.SizeEstimateStatusType
	if counters.BytesFreed > 0 {
		status = enums.SizeEstimateStatusKnown
	} else {
		status = enums.SizeEstimateStatusUnknown
	}

	return result.Ok(conversions.NewCleanResultWithTimingAndSize(
		enums.StrategyConservativeType,
		counters.ItemsRemoved, counters.ItemsFailed, counters.BytesFreed, counters.Duration(),
		types.SizeEstimate{Known: uint64(counters.BytesFreed), Status: status},
	))
}

// removeCachePath removes a cache directory and returns the appropriate result.
func (scc *SystemCacheCleaner) removeCachePath(
	path, successMessage string,
) result.Result[types.CleanResult] {
	if scc.dryRun {
		// Estimate size for dry-run
		estimatedSize := cleaner.GetDirSize(path)
		if scc.verbose {
			fmt.Printf("  [DRY RUN] Would remove: %s (%s)\n", path, format.Bytes(estimatedSize))
		}

		return result.Ok(conversions.NewCleanResultWithSizeEstimate(
			enums.StrategyConservativeType,
			1,
			estimatedSize,
			types.SizeEstimate{
				Known:	uint64(estimatedSize),
				Status:	enums.SizeEstimateStatusKnown,
			},
		))
	}

	// Measure size before removal
	bytesFreed := cleaner.GetDirSize(path)

	err := os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		return result.Err[types.CleanResult](
			fmt.Errorf("failed to remove %s (successMessage=%v): %w", path, successMessage, err),
		)
	}

	if scc.verbose {
		fmt.Printf("  ✓ %s (%s freed)\n", successMessage, format.Bytes(bytesFreed))
	}

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		enums.StrategyConservativeType,
		1, bytesFreed,
		types.SizeEstimate{Known: uint64(bytesFreed), Status: enums.SizeEstimateStatusKnown},
	))
}

// scanCachePathWithConfig scans a cache directory using configuration and returns scan items.
func (scc *SystemCacheCleaner) scanCachePathWithConfig(
	_ context.Context,
	homeDir string,
	config cacheTypeConfig,
) result.Result[[]types.ScanItem] {
	scanResult := cleaner.ScanPath(
		homeDir,
		config.scanType,
		config.displayName,
		scc.verbose,
		"",
		config.pathComponents...,
	)

	return result.Ok(scanResult.Items)
}

// cleanSystemCache cleans cache for a specific system cache type.
func (scc *SystemCacheCleaner) cleanSystemCache(
	_ context.Context,
	cacheType enums.CacheType,
	homeDir string,
) result.Result[types.CleanResult] {
	config, exists := systemCacheConfigs[cacheType]
	if !exists {
		return result.Err[types.CleanResult](
			fmt.Errorf("unknown system cache type: %s", cacheType.String()),
		)
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
