package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// TempFilesCleaner handles temporary files cleanup with proper type safety.
type TempFilesCleaner struct {
	verbose   bool
	dryRun    bool
	olderThan time.Duration
	excludes  []string
	basePaths []string
}

// NewTempFilesCleaner creates temp files cleaner with proper configuration.
func NewTempFilesCleaner(verbose, dryRun bool, olderThan string, excludes, basePaths []string) (*TempFilesCleaner, error) {
	// Parse older than duration using custom duration parser (supports "7d", "24h", etc.)
	duration, err := domain.ParseCustomDuration(olderThan)
	if err != nil {
		return nil, fmt.Errorf("invalid older_than duration: %w", err)
	}

	// Normalize excludes
	normalizedExcludes := make([]string, 0, len(excludes))
	for _, exclude := range excludes {
		normalizedExcludes = append(normalizedExcludes, filepath.Clean(exclude))
	}

	// Normalize base paths
	normalizedPaths := make([]string, 0, len(basePaths))
	for _, path := range basePaths {
		normalizedPaths = append(normalizedPaths, filepath.Clean(path))
	}

	return &TempFilesCleaner{
		verbose:   verbose,
		dryRun:    dryRun,
		olderThan: duration,
		excludes:  normalizedExcludes,
		basePaths: normalizedPaths,
	}, nil
}

// Type returns operation type for temp files cleaner.
func (tfc *TempFilesCleaner) Type() domain.OperationType {
	return domain.OperationTypeTempFiles
}

// Name returns the cleaner name for result tracking.
func (tfc *TempFilesCleaner) Name() string {
	return "tempfiles"
}

// IsAvailable checks if temp files cleaner is available.
func (tfc *TempFilesCleaner) IsAvailable(ctx context.Context) bool {
	// Temp files cleaner is always available
	return true
}

// ValidateSettings validates temp files cleaner settings with type safety.
func (tfc *TempFilesCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.TempFiles == nil {
		return nil // Settings are optional
	}

	if settings.TempFiles.OlderThan == "" {
		return errors.New("older_than must be specified")
	}

	// Parse older than to validate it's a valid duration using custom parser
	_, err := domain.ParseCustomDuration(settings.TempFiles.OlderThan)
	if err != nil {
		return fmt.Errorf("invalid older_than duration: %w", err)
	}

	return nil
}

// Scan scans for temp files that can be cleaned.
func (tfc *TempFilesCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)
	cutoffTime := time.Now().Add(-tfc.olderThan)

	// Scan each base path
	for _, basePath := range tfc.basePaths {
		// Skip if path doesn't exist
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}

		// Walk the directory tree
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Skip files we can't access
				return nil
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Check if path is excluded
			if tfc.isExcluded(path) {
				return nil
			}

			// Check if file is older than cutoff
			if info.ModTime().Before(cutoffTime) {
				items = append(items, domain.ScanItem{
					Path:     path,
					Size:     info.Size(),
					Created:  info.ModTime(),
					ScanType: domain.ScanTypeTemp,
				})
			}

			return nil
		})

		if err != nil && tfc.verbose {
			fmt.Printf("Warning: error scanning %s: %v\n", basePath, err)
		}
	}

	return result.Ok(items)
}

// Clean removes old temp files with proper type safety.
func (tfc *TempFilesCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	// Get files to clean first
	scanResult := tfc.Scan(ctx)
	if scanResult.IsErr() {
		return conversions.ToCleanResultFromError(scanResult.Error())
	}

	items := scanResult.Value()

	if len(items) == 0 {
		// Nothing to clean
		cleanResult := conversions.NewCleanResult(domain.CleanStrategyType(domain.StrategyConservativeType), 0, 0)
		return result.Ok(cleanResult)
	}

	if tfc.dryRun {
		// Calculate total bytes that would be freed
		var totalBytes int64
		for _, item := range items {
			totalBytes += item.Size
		}

		cleanResult := conversions.NewCleanResult(domain.CleanStrategyType(domain.StrategyDryRunType), len(items), totalBytes)
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	for _, item := range items {
		err := os.Remove(item.Path)
		if err != nil {
			itemsFailed++
			if tfc.verbose {
				fmt.Printf("Warning: failed to remove %s: %v\n", item.Path, err)
			}
			continue
		}

		itemsRemoved++
		bytesFreed += item.Size
	}

	duration := time.Since(startTime)
	return result.Ok(conversions.NewCleanResultWithFailures(
		domain.CleanStrategyType(domain.StrategyAggressiveType),
		itemsRemoved,
		itemsFailed,
		bytesFreed,
		duration,
	))
}

// isExcluded checks if a path should be excluded from cleanup.
func (tfc *TempFilesCleaner) isExcluded(path string) bool {
	cleanPath := filepath.Clean(path)

	for _, exclude := range tfc.excludes {
		if strings.HasPrefix(cleanPath, exclude) {
			return true
		}
	}

	return false
}
