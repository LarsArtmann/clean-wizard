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

// SystemCacheCleaner handles macOS system cache cleanup.
type SystemCacheCleaner struct {
	verbose    bool
	dryRun     bool
	cacheTypes  []SystemCacheType
	olderThan  time.Duration
}

// SystemCacheType represents different system cache types.
type SystemCacheType string

const (
	SystemCacheSpotlight   SystemCacheType = "spotlight"
	SystemCacheXcode       SystemCacheType = "xcode"
	SystemCacheCocoaPods  SystemCacheType = "cocoapods"
	SystemCacheHomebrew   SystemCacheType = "homebrew"
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
		verbose:   verbose,
		dryRun:    dryRun,
		cacheTypes: cacheTypes,
		olderThan: duration,
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
		SystemCacheSpotlight:  true,
		SystemCacheXcode:      true,
		SystemCacheCocoaPods: true,
		SystemCacheHomebrew:   true,
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
	homeDir, err := scc.getHomeDir()
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

// scanSystemCache scans cache for a specific system cache type.
func (scc *SystemCacheCleaner) scanSystemCache(ctx context.Context, cacheType SystemCacheType, homeDir string) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	switch cacheType {
	case SystemCacheSpotlight:
		// Spotlight metadata: ~/Library/Metadata/CoreSpotlight/SpotlightKnowledgeEvents
		spotlightPath := filepath.Join(homeDir, "Library", "Metadata", "CoreSpotlight")
		if info, err := os.Stat(spotlightPath); err == nil && info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(spotlightPath, "*"))
			if err != nil {
				return result.Err[[]domain.ScanItem](fmt.Errorf("failed to find Spotlight metadata: %w", err))
			}

			for _, match := range matches {
				items = append(items, domain.ScanItem{
					Path:     match,
					Size:     scc.getDirSize(match),
					Created:  scc.getDirModTime(match),
					ScanType: domain.ScanTypeTemp,
				})

				if scc.verbose {
					fmt.Printf("Found Spotlight metadata: %s\n", filepath.Base(match))
				}
			}
		}

	case SystemCacheXcode:
		// Xcode Derived Data: ~/Library/Developer/Xcode/DerivedData
		xcodeDerivedData := filepath.Join(homeDir, "Library", "Developer", "Xcode", "DerivedData")
		if info, err := os.Stat(xcodeDerivedData); err == nil && info.IsDir() {
			items = append(items, domain.ScanItem{
				Path:     xcodeDerivedData,
				Size:     scc.getDirSize(xcodeDerivedData),
				Created:  scc.getDirModTime(xcodeDerivedData),
				ScanType: domain.ScanTypeTemp,
			})

			if scc.verbose {
				fmt.Printf("Found Xcode DerivedData: %s\n", xcodeDerivedData)
			}
		}

	case SystemCacheCocoaPods:
		// CocoaPods cache: ~/Library/Caches/CocoaPods
		cocoaPodsCache := filepath.Join(homeDir, "Library", "Caches", "CocoaPods")
		if info, err := os.Stat(cocoaPodsCache); err == nil && info.IsDir() {
			items = append(items, domain.ScanItem{
				Path:     cocoaPodsCache,
				Size:     scc.getDirSize(cocoaPodsCache),
				Created:  scc.getDirModTime(cocoaPodsCache),
				ScanType: domain.ScanTypeTemp,
			})

			if scc.verbose {
				fmt.Printf("Found CocoaPods cache: %s\n", cocoaPodsCache)
			}
		}

	case SystemCacheHomebrew:
		// Homebrew cache: ~/Library/Caches/Homebrew
		homebrewCache := filepath.Join(homeDir, "Library", "Caches", "Homebrew")
		if info, err := os.Stat(homebrewCache); err == nil && info.IsDir() {
			items = append(items, domain.ScanItem{
				Path:     homebrewCache,
				Size:     scc.getDirSize(homebrewCache),
				Created:  scc.getDirModTime(homebrewCache),
				ScanType: domain.ScanTypeTemp,
			})

			if scc.verbose {
				fmt.Printf("Found Homebrew cache: %s\n", homebrewCache)
			}
		}
	}

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
	homeDir, err := scc.getHomeDir()
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

// cleanSystemCache cleans cache for a specific system cache type.
func (scc *SystemCacheCleaner) cleanSystemCache(ctx context.Context, cacheType SystemCacheType, homeDir string) result.Result[domain.CleanResult] {
	switch cacheType {
	case SystemCacheSpotlight:
		// Remove SpotlightKnowledgeEvents
		spotlightPath := filepath.Join(homeDir, "Library", "Metadata", "CoreSpotlight", "SpotlightKnowledgeEvents")
		if scc.dryRun {
			if scc.verbose {
				fmt.Printf("  [DRY RUN] Would remove: %s\n", spotlightPath)
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

		err := os.RemoveAll(spotlightPath)
		if err != nil && !os.IsNotExist(err) {
			return result.Err[domain.CleanResult](fmt.Errorf("failed to remove Spotlight metadata: %w", err))
		}

		if scc.verbose {
			fmt.Println("  ✓ Spotlight metadata cleaned")
		}

		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})

	case SystemCacheXcode:
		// Remove Xcode DerivedData
		xcodeDerivedData := filepath.Join(homeDir, "Library", "Developer", "Xcode", "DerivedData")
		if scc.dryRun {
			if scc.verbose {
				fmt.Printf("  [DRY RUN] Would remove: %s\n", xcodeDerivedData)
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

		err := os.RemoveAll(xcodeDerivedData)
		if err != nil && !os.IsNotExist(err) {
			return result.Err[domain.CleanResult](fmt.Errorf("failed to remove Xcode DerivedData: %w", err))
		}

		if scc.verbose {
			fmt.Println("  ✓ Xcode DerivedData cleaned")
		}

		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})

	case SystemCacheCocoaPods:
		// Remove CocoaPods cache
		cocoaPodsCache := filepath.Join(homeDir, "Library", "Caches", "CocoaPods")
		if scc.dryRun {
			if scc.verbose {
				fmt.Printf("  [DRY RUN] Would remove: %s\n", cocoaPodsCache)
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

		err := os.RemoveAll(cocoaPodsCache)
		if err != nil && !os.IsNotExist(err) {
			return result.Err[domain.CleanResult](fmt.Errorf("failed to remove CocoaPods cache: %w", err))
		}

		if scc.verbose {
			fmt.Println("  ✓ CocoaPods cache cleaned")
		}

		return result.Ok(domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: 1,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.StrategyConservative,
		})

	case SystemCacheHomebrew:
		// Remove Homebrew cache
		homebrewCache := filepath.Join(homeDir, "Library", "Caches", "Homebrew")
		if scc.dryRun {
			if scc.verbose {
				fmt.Printf("  [DRY RUN] Would remove: %s\n", homebrewCache)
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

		err := os.RemoveAll(homebrewCache)
		if err != nil && !os.IsNotExist(err) {
			return result.Err[domain.CleanResult](fmt.Errorf("failed to remove Homebrew cache: %w", err))
		}

		if scc.verbose {
			fmt.Println("  ✓ Homebrew cache cleaned")
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

	return result.Err[domain.CleanResult](fmt.Errorf("unknown system cache type: %s", cacheType))
}

// isMacOS checks if the system is macOS.
func (scc *SystemCacheCleaner) isMacOS() bool {
	// Simple check for macOS
	return os.Getenv("GOOS") == "darwin" || os.Getenv("OSTYPE") == "darwin"
}

// getHomeDir returns user's home directory.
func (scc *SystemCacheCleaner) getHomeDir() (string, error) {
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
func (scc *SystemCacheCleaner) getDirSize(path string) int64 {
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
func (scc *SystemCacheCleaner) getDirModTime(path string) time.Time {
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
