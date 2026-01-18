package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// BuildCacheCleaner handles build tool cache cleanup.
type BuildCacheCleaner struct {
	verbose   bool
	dryRun    bool
	olderThan time.Duration
	toolTypes []BuildToolType
	basePaths []string
}

// BuildToolType represents different build tool types.
type BuildToolType string

const (
	BuildToolGradle BuildToolType = "gradle"
	BuildToolMaven  BuildToolType = "maven"
	BuildToolSBT    BuildToolType = "sbt"
)

// AvailableBuildTools returns all available build tool types.
func AvailableBuildTools() []BuildToolType {
	return []BuildToolType{
		BuildToolGradle,
		BuildToolMaven,
		BuildToolSBT,
	}
}

// NewBuildCacheCleaner creates build cache cleaner.
func NewBuildCacheCleaner(verbose, dryRun bool, olderThan string, excludes, basePaths []string) (*BuildCacheCleaner, error) {
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
		verbose:   verbose,
		dryRun:    dryRun,
		olderThan: duration,
		toolTypes: toolTypes,
		basePaths: normalizedPaths,
	}, nil
}

// Type returns operation type for build cache cleaner.
func (bcc *BuildCacheCleaner) Type() domain.OperationType {
	return domain.OperationTypeBuildCache
}

// IsAvailable checks if build cache cleaner is available.
func (bcc *BuildCacheCleaner) IsAvailable(ctx context.Context) bool {
	// Build cache cleaner is always available (uses file system operations)
	return true
}

// ValidateSettings validates build cache cleaner settings.
func (bcc *BuildCacheCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.BuildCache == nil {
		return nil // Settings are optional
	}

	// Validate tool types
	validToolTypes := map[BuildToolType]bool{
		BuildToolGradle: true,
		BuildToolMaven:  true,
		BuildToolSBT:    true,
	}

	for _, tool := range settings.BuildCache.ToolTypes {
		toolStr := BuildToolType(tool)
		if !validToolTypes[toolStr] {
			return fmt.Errorf("invalid tool type: %s (must be gradle, maven, or sbt)", tool)
		}
	}

	return nil
}

// Scan scans for build tool caches.
func (bcc *BuildCacheCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	// Get home directory
	homeDir, err := bcc.getHomeDir()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to get home directory: %w", err))
	}

	// Scan for each build tool type
	for _, toolType := range bcc.toolTypes {
		result := bcc.scanBuildTool(ctx, toolType, homeDir)
		if result.IsErr() {
			if bcc.verbose {
				fmt.Printf("Warning: failed to scan %s: %v\n", toolType, result.Error())
			}
			continue
		}

		items = append(items, result.Value()...)
	}

	return result.Ok(items)
}

// scanBuildTool scans cache for a specific build tool.
func (bcc *BuildCacheCleaner) scanBuildTool(ctx context.Context, toolType BuildToolType, homeDir string) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	switch toolType {
	case BuildToolGradle:
		// Gradle cache location: ~/.gradle/caches/*
		gradleCache := filepath.Join(homeDir, ".gradle", "caches")
		matches, err := filepath.Glob(filepath.Join(gradleCache, "*"))
		if err != nil {
			return result.Err[[]domain.ScanItem](fmt.Errorf("failed to find Gradle caches: %w", err))
		}

		for _, match := range matches {
			items = append(items, domain.ScanItem{
				Path:     match,
				Size:     bcc.getDirSize(match),
				Created:  bcc.getDirModTime(match),
				ScanType: domain.ScanTypeTemp,
			})

			if bcc.verbose {
				fmt.Printf("Found Gradle cache: %s\n", match)
			}
		}

	case BuildToolMaven:
		// Maven cache location: ~/.m2/repository
		mavenCache := filepath.Join(homeDir, ".m2", "repository")
		if info, err := os.Stat(mavenCache); err == nil && info.IsDir() {
			items = append(items, domain.ScanItem{
				Path:     mavenCache,
				Size:     bcc.getDirSize(mavenCache),
				Created:  bcc.getDirModTime(mavenCache),
				ScanType: domain.ScanTypeTemp,
			})

			if bcc.verbose {
				fmt.Printf("Found Maven repository: %s\n", mavenCache)
			}
		}

	case BuildToolSBT:
		// SBT cache location: ~/.ivy2/cache
		sbtCache := filepath.Join(homeDir, ".ivy2", "cache")
		if info, err := os.Stat(sbtCache); err == nil && info.IsDir() {
			items = append(items, domain.ScanItem{
				Path:     sbtCache,
				Size:     bcc.getDirSize(sbtCache),
				Created:  bcc.getDirModTime(sbtCache),
				ScanType: domain.ScanTypeTemp,
			})

			if bcc.verbose {
				fmt.Printf("Found SBT cache: %s\n", sbtCache)
			}
		}
	}

	return result.Ok(items)
}

// Clean removes build tool caches.
func (bcc *BuildCacheCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !bcc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("build cache cleaner not available"))
	}

	if bcc.dryRun {
		// Estimate cache sizes based on typical usage
		totalBytes := int64(300 * 1024 * 1024) // Estimate 300MB per tool
		itemsRemoved := len(bcc.toolTypes)

		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, itemsRemoved, totalBytes)
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	// Get home directory
	homeDir, err := bcc.getHomeDir()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get home directory: %w", err))
	}

	// Clean for each build tool type
	for _, toolType := range bcc.toolTypes {
		result := bcc.cleanBuildTool(ctx, toolType, homeDir)
		if result.IsErr() {
			itemsFailed++
			if bcc.verbose {
				fmt.Printf("Warning: failed to clean %s: %v\n", toolType, result.Error())
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

// cleanBuildTool cleans cache for a specific build tool.
func (bcc *BuildCacheCleaner) cleanBuildTool(ctx context.Context, toolType BuildToolType, homeDir string) result.Result[domain.CleanResult] {
	switch toolType {
	case BuildToolGradle:
		// Remove ~/.gradle/caches/*
		gradleCache := filepath.Join(homeDir, ".gradle", "caches")
		matches, err := filepath.Glob(filepath.Join(gradleCache, "*"))
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("failed to find Gradle caches: %w", err))
		}

		itemsRemoved := 0
		for _, match := range matches {
			if bcc.dryRun {
				itemsRemoved++
				continue
			}

			err := os.RemoveAll(match)
			if err != nil {
				if bcc.verbose {
					fmt.Printf("Warning: failed to remove %s: %v\n", match, err)
				}
				continue
			}

			itemsRemoved++
			if bcc.verbose {
				fmt.Printf("  ✓ Removed Gradle cache: %s\n", filepath.Base(match))
			}
		}

		if bcc.verbose && itemsRemoved > 0 {
			fmt.Println("  ✓ Gradle caches cleaned")
		}

		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: uint(itemsRemoved),
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})

	case BuildToolMaven:
		// Remove partial files from ~/.m2/repository
		// (This is conservative - we don't remove the entire repository)
		mavenRepository := filepath.Join(homeDir, ".m2", "repository")
		matches, err := filepath.Glob(filepath.Join(mavenRepository, "**", "*.part"))
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("failed to find Maven partial files: %w", err))
		}

		itemsRemoved := 0
		for _, match := range matches {
			if bcc.dryRun {
				itemsRemoved++
				continue
			}

			err := os.Remove(match)
			if err != nil {
				if bcc.verbose {
					fmt.Printf("Warning: failed to remove %s: %v\n", match, err)
				}
				continue
			}

			itemsRemoved++
			if bcc.verbose {
				fmt.Printf("  ✓ Removed Maven partial file: %s\n", filepath.Base(match))
			}
		}

		if bcc.verbose && itemsRemoved > 0 {
			fmt.Println("  ✓ Maven partial files cleaned")
		}

		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: uint(itemsRemoved),
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})

	case BuildToolSBT:
		// Remove ~/.ivy2/cache/*
		sbtCache := filepath.Join(homeDir, ".ivy2", "cache")
		matches, err := filepath.Glob(filepath.Join(sbtCache, "*"))
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("failed to find SBT caches: %w", err))
		}

		itemsRemoved := 0
		for _, match := range matches {
			if bcc.dryRun {
				itemsRemoved++
				continue
			}

			err := os.RemoveAll(match)
			if err != nil {
				if bcc.verbose {
					fmt.Printf("Warning: failed to remove %s: %v\n", match, err)
				}
				continue
			}

			itemsRemoved++
			if bcc.verbose {
				fmt.Printf("  ✓ Removed SBT cache: %s\n", filepath.Base(match))
			}
		}

		if bcc.verbose && itemsRemoved > 0 {
			fmt.Println("  ✓ SBT caches cleaned")
		}

		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: uint(itemsRemoved),
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})
	}

	return result.Err[domain.CleanResult](fmt.Errorf("unknown build tool type: %s", toolType))
}

// getHomeDir returns user's home directory.
func (bcc *BuildCacheCleaner) getHomeDir() (string, error) {
	// Try using os/user package
	currentUser, err := user.Current()
	if err == nil {
		return currentUser.HomeDir, nil
	}

	// Fallback to HOME environment variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// Fallback to user profile directory
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return userProfile, nil
	}

	return "", fmt.Errorf("unable to determine home directory")
}

// getDirSize returns total size of directory recursively.
func (bcc *BuildCacheCleaner) getDirSize(path string) int64 {
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

// getDirModTime returns most recent modification time in directory.
func (bcc *BuildCacheCleaner) getDirModTime(path string) time.Time {
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
