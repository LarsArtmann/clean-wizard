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

// BuildCacheCleaner handles build tool cache cleanup.
type BuildCacheCleaner struct {
	CleanerBase
	olderThan time.Duration
	toolTypes []JVMBuildToolType
	basePaths []string
}

// JVMBuildToolType represents different JVM build tool types.
type JVMBuildToolType string

const (
	JVMBuildToolGradle JVMBuildToolType = "gradle"
	JVMBuildToolMaven  JVMBuildToolType = "maven"
	JVMBuildToolSBT    JVMBuildToolType = "sbt"
)

// AvailableBuildTools returns all available JVM build tool types.
func AvailableBuildTools() []JVMBuildToolType {
	return []JVMBuildToolType{
		JVMBuildToolGradle,
		JVMBuildToolMaven,
		JVMBuildToolSBT,
	}
}

// NewBuildCacheCleaner creates build cache cleaner.
func NewBuildCacheCleaner(
	verbose, dryRun bool,
	olderThan string,
	_, basePaths []string,
) (*BuildCacheCleaner, error) {
	// Parse older than duration
	duration, err := domain.ParseCustomDuration(olderThan)
	if err != nil {
		return nil, fmt.Errorf("invalid older_than duration: %w", err)
	}

	// Default tool types to all available
	toolTypes := AvailableBuildTools()

	// Normalize base paths
	normalizedPaths := make([]string, 0, len(basePaths))
	for _, path := range basePaths {
		normalizedPaths = append(normalizedPaths, filepath.Clean(path))
	}

	return &BuildCacheCleaner{
		CleanerBase: NewCleanerBase(verbose, dryRun),
		olderThan:   duration,
		toolTypes:   toolTypes,
		basePaths:   normalizedPaths,
	}, nil
}

// Type returns operation type for build cache cleaner.
func (bcc *BuildCacheCleaner) Type() domain.OperationType {
	return domain.OperationTypeBuildCache
}

// Name returns the cleaner name for result tracking.
func (bcc *BuildCacheCleaner) Name() string {
	return "buildcache"
}

// IsAvailable checks if build cache cleaner is available.
func (bcc *BuildCacheCleaner) IsAvailable(_ context.Context) bool {
	// Build cache cleaner is always available (uses file system operations)
	return true
}

// ValidateSettings validates build cache cleaner settings.
func (bcc *BuildCacheCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	return ValidateBuildCacheSettings(settings)
}

// Scan scans for build tool caches.
func (bcc *BuildCacheCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	return scanWithIterator(
		ctx,
		bcc.toolTypes,
		bcc.scanBuildTool,
		bcc.verbose,
	)
}

// getCachePath returns the cache directory path for a JVM build tool type.
func getCachePath(toolType JVMBuildToolType, homeDir string) string {
	switch toolType {
	case JVMBuildToolGradle:
		return filepath.Join(homeDir, ".gradle", "caches")
	case JVMBuildToolMaven:
		return filepath.Join(homeDir, ".m2", "repository")
	case JVMBuildToolSBT:
		return filepath.Join(homeDir, ".ivy2", "cache")
	}

	return ""
}

// scanBuildTool scans cache for a specific JVM build tool.
func (bcc *BuildCacheCleaner) scanBuildTool(
	_ context.Context,
	toolType JVMBuildToolType,
	homeDir string,
) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	switch toolType {
	case JVMBuildToolGradle:
		gradleCache := getCachePath(toolType, homeDir)
		scanResult := ScanPath(
			"",
			domain.ScanTypeTemp,
			"Gradle cache",
			bcc.verbose,
			"*",
			gradleCache,
		)
		items = append(items, scanResult.Items...)

	case JVMBuildToolMaven:
		mavenCache := getCachePath(toolType, homeDir)
		scanResult := ScanDirectory(mavenCache, domain.ScanTypeTemp, bcc.verbose)
		items = append(items, scanResult.Items...)

	case JVMBuildToolSBT:
		sbtCache := getCachePath(toolType, homeDir)
		scanResult := ScanDirectory(sbtCache, domain.ScanTypeTemp, bcc.verbose)
		items = append(items, scanResult.Items...)
	}

	return result.Ok(items)
}

// Clean removes build tool caches.
func (bcc *BuildCacheCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	return cleanWithIterator(
		ctx,
		"build cache cleaner",
		bcc.IsAvailable,
		bcc.toolTypes,
		bcc.cleanBuildTool,
		bcc.verbose,
		bcc.dryRun,
		bcc.estimateBuildToolSize,
	)
}

// estimateBuildToolSize estimates the size of a build tool's cache for dry-run mode.
func (bcc *BuildCacheCleaner) estimateBuildToolSize(toolType JVMBuildToolType) int64 {
	homeDir, err := GetHomeDir()
	if err != nil {
		return DryRunBytesPerItem
	}

	cachePath := getCachePath(toolType, homeDir)
	if cachePath == "" {
		return DryRunBytesPerItem
	}

	size := GetDirSize(cachePath)
	if size > 0 {
		return size
	}

	return DryRunBytesPerItem
}

type CacheCleanerFunc func(ctx context.Context, toolType JVMBuildToolType, homeDir string) result.Result[domain.CleanResult]

type RemoveFunc func(path string) error

// printVerbose prints a verbose message if verbose mode is enabled.
func (bcc *BuildCacheCleaner) printVerbose(action, verboseMsg, baseName string) {
	if !bcc.verbose {
		return
	}

	fmt.Printf("  ✓ %s %s: %s\n", action, verboseMsg, baseName)
}

// genericClean handles common cleanup logic for both cache directories and partial files.
func (bcc *BuildCacheCleaner) genericClean(
	_ context.Context,
	toolName string,
	baseDir string,
	pattern string,
	verboseMsg string,
	removeFn RemoveFunc,
) result.Result[domain.CleanResult] {
	matches, err := filepath.Glob(filepath.Join(baseDir, pattern))
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to find %s: %w", toolName, err))
	}

	itemsRemoved := 0
	bytesFreed := int64(0)

	for _, match := range matches {
		if !bcc.dryRun {
			bytesFreed += GetDirSize(match)
		}

		if bcc.dryRun {
			itemsRemoved++

			bcc.printVerbose("Would remove", verboseMsg, filepath.Base(match))

			continue
		}

		err := removeFn(match)
		if err != nil {
			if bcc.verbose {
				fmt.Printf("Warning: failed to remove %s: %v\n", match, err)
			}

			continue
		}

		itemsRemoved++

		bcc.printVerbose("Removed", verboseMsg, filepath.Base(match))
	}

	if bcc.verbose && itemsRemoved > 0 {
		fmt.Printf("  ✓ %s cleaned\n", toolName)
	}

	return result.Ok(
		conversions.NewCleanResult(
			domain.StrategyConservativeType,
			itemsRemoved,
			bytesFreed,
		),
	)
}

// cleanBuildTool cleans cache for a specific JVM build tool.
func (bcc *BuildCacheCleaner) cleanBuildTool(
	ctx context.Context,
	toolType JVMBuildToolType,
	homeDir string,
) result.Result[domain.CleanResult] {
	switch toolType {
	case JVMBuildToolGradle:
		gradleCache := getCachePath(toolType, homeDir)

		return bcc.genericClean(ctx, "Gradle", gradleCache, "*", "Gradle cache", os.RemoveAll)

	case JVMBuildToolMaven:
		mavenRepository := getCachePath(toolType, homeDir)

		return bcc.genericClean(
			ctx,
			"Maven",
			mavenRepository,
			"**/*.part",
			"Maven partial file",
			os.Remove,
		)

	case JVMBuildToolSBT:
		sbtCache := getCachePath(toolType, homeDir)

		return bcc.genericClean(ctx, "SBT", sbtCache, "*", "SBT cache", os.RemoveAll)
	}

	return result.Err[domain.CleanResult](fmt.Errorf("unknown build tool type: %s", toolType))
}
