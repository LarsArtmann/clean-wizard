package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// CleanStats tracks cleaning metrics.
type CleanStats struct {
	Removed    uint
	Failed     uint
	FreedBytes uint64
}

// GoCleaner handles Go language cleanup using type-safe cache flags.
type GoCleaner struct {
	CleanerBase
	caches   GoCacheType
	scanner  *GoScanner
	cleaners map[GoCacheType]interface {
		Clean(ctx context.Context) result.Result[domain.CleanResult]
	}
}

// NewGoCleaner creates Go cleaner with type-safe cache configuration.
func NewGoCleaner(verbose, dryRun bool, caches GoCacheType) (*GoCleaner, error) {
	if !caches.IsValid() {
		return nil, errors.New("at least one cache type must be specified")
	}

	return NewGoCleanerWithSettings(verbose, dryRun, caches), (error)(nil)
}

// NewGoCleanerWithSettings creates Go cleaner with type-safe cache configuration (panics on invalid caches).
// This is a convenience function for tests and backward compatibility.
func NewGoCleanerWithSettings(verbose, dryRun bool, caches GoCacheType) *GoCleaner {
	scanner := NewGoScanner(verbose)
	cleaners := make(map[GoCacheType]interface {
		Clean(ctx context.Context) result.Result[domain.CleanResult]
	})

	for _, cacheType := range []GoCacheType{GoCacheGOCACHE, GoCacheTestCache, GoCacheModCache, GoCacheBuildCache} {
		if caches.Has(cacheType) {
			cleaners[cacheType] = NewGoCacheCleaner(cacheType, verbose, dryRun)
		}
	}

	if caches.Has(GoCacheLintCache) {
		cleaners[GoCacheLintCache] = NewGolangciLintCacheCleaner(verbose, dryRun)
	}

	return &GoCleaner{
		CleanerBase: NewCleanerBase(verbose, dryRun),
		caches:      caches,
		scanner:     scanner,
		cleaners:    cleaners,
	}
}

// Type returns operation type.
func (gc *GoCleaner) Type() domain.OperationType {
	return domain.OperationTypeGoPackages
}

// Name returns the cleaner name for result tracking.
func (gc *GoCleaner) Name() string {
	return "go"
}

// IsAvailable checks if Go is available.
func (gc *GoCleaner) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("go")

	return err == nil
}

// ValidateSettings validates settings.
func (gc *GoCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	return settings.ValidateSettings(domain.OperationTypeGoPackages)
}

// Scan scans for Go caches.
func (gc *GoCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	return gc.scanner.Scan(ctx, gc.caches)
}

// Clean removes Go caches.
func (gc *GoCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !gc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](errors.New("go not available"))
	}

	if gc.dryRun {
		return gc.dryRunClean(ctx)
	}

	startTime := time.Now()
	stats := CleanStats{}

	for _, cacheType := range gc.caches.EnabledTypes() {
		cleaner, ok := gc.cleaners[cacheType]
		if !ok {
			gc.logWarning("no cleaner for cache type: %v", cacheType)

			continue
		}

		result := cleaner.Clean(ctx)
		gc.processCacheResult(result, &stats, cacheType.String())
	}

	duration := time.Since(startTime)

	return gc.buildCleanResult(stats, duration)
}

// dryRunClean performs dry-run estimation by scanning actual cache sizes.
func (gc *GoCleaner) dryRunClean(ctx context.Context) result.Result[domain.CleanResult] {
	// Scan actual cache directories to get real sizes
	scanResult := gc.scanner.Scan(ctx, gc.caches)

	var (
		totalBytes   uint64
		itemsRemoved int
	)

	if scanResult.IsOk() {
		items := scanResult.Value()

		itemsRemoved = len(items)
		for _, item := range items {
			totalBytes += uint64(item.Size)
		}
	} else {
		// Fallback to counting enabled cache types if scan fails
		itemsRemoved = gc.caches.Count()
	}

	cleanResult := conversions.NewCleanResult(
		domain.StrategyDryRunType,
		itemsRemoved,
		int64(totalBytes),
	)
	cleanResult.SizeEstimate = domain.SizeEstimate{Known: totalBytes}

	return result.Ok(cleanResult)
}

// processCacheResult handles cache cleaning result uniformly.
func (gc *GoCleaner) processCacheResult(
	r result.Result[domain.CleanResult],
	stats *CleanStats,
	cacheName string,
) {
	if r.IsErr() {
		stats.Failed++

		gc.logWarning("failed to clean %s: %v", cacheName, r.Error())
	} else if r.IsOk() && r.Value().ItemsRemoved > 0 {
		stats.Removed += r.Value().ItemsRemoved
		stats.FreedBytes += r.Value().SizeEstimate.Value()
	}
}

// buildCleanResult creates CleanResult from stats.
func (gc *GoCleaner) buildCleanResult(
	stats CleanStats,
	duration time.Duration,
) result.Result[domain.CleanResult] {
	// Create result with honest size estimate - set Status explicitly to avoid validation errors
	var status domain.SizeEstimateStatusType
	if stats.FreedBytes > 0 {
		status = domain.SizeEstimateStatusKnown
	} else {
		status = domain.SizeEstimateStatusUnknown
	}

	sizeEstimate := domain.SizeEstimate{
		Known:  stats.FreedBytes,
		Status: status,
	}

	// Note: conversions.NewCleanResult uses FreedBytes (deprecated), so we update SizeEstimate
	cleanResult := conversions.NewCleanResult(
		domain.StrategyConservativeType,
		int(stats.Removed),
		int64(stats.FreedBytes),
	)
	cleanResult.SizeEstimate = sizeEstimate
	cleanResult.CleanTime = duration
	cleanResult.CleanedAt = time.Now()

	return result.Ok(cleanResult)
}

// logWarning logs warning message if verbose.
func (gc *GoCleaner) logWarning(format string, args ...any) {
	if gc.verbose {
		fmt.Printf("Warning: "+format+"\n", args...)
	}
}
