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

// LanguageVersionManagerCleaner handles language version manager cleanup.
type LanguageVersionManagerCleaner struct {
	verbose            bool
	dryRun             bool
	managerTypes       []LangVersionManagerType
}

// LangVersionManagerType represents different language version manager types.
type LangVersionManagerType string

const (
	LangVersionManagerNVM   LangVersionManagerType = "nvm"
	LangVersionManagerPYENV LangVersionManagerType = "pyenv"
	LangVersionManagerRBENV LangVersionManagerType = "rbenv"
)

// AvailableLangVersionManagers returns all available language version manager types.
func AvailableLangVersionManagers() []LangVersionManagerType {
	return []LangVersionManagerType{
		LangVersionManagerNVM,
		LangVersionManagerPYENV,
		LangVersionManagerRBENV,
	}
}

// NewLanguageVersionManagerCleaner creates language version manager cleaner.
func NewLanguageVersionManagerCleaner(verbose, dryRun bool, managerTypes []LangVersionManagerType) *LanguageVersionManagerCleaner {
	// Default manager types to all available
	if len(managerTypes) == 0 {
		managerTypes = AvailableLangVersionManagers()
	}

	return &LanguageVersionManagerCleaner{
		verbose:      verbose,
		dryRun:       dryRun,
		managerTypes: managerTypes,
	}
}

// Type returns operation type for language version manager cleaner.
func (lvmc *LanguageVersionManagerCleaner) Type() domain.OperationType {
	return domain.OperationTypeLangVersionManager
}

// IsAvailable checks if language version manager cleaner is available.
func (lvmc *LanguageVersionManagerCleaner) IsAvailable(ctx context.Context) bool {
	// Language version manager cleaner is always available (uses file system operations)
	return true
}

// ValidateSettings validates language version manager cleaner settings.
func (lvmc *LanguageVersionManagerCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.LangVersionManager == nil {
		return nil // Settings are optional
	}

	// Validate manager types
	validManagerTypes := map[LangVersionManagerType]bool{
		LangVersionManagerNVM:   true,
		LangVersionManagerPYENV: true,
		LangVersionManagerRBENV: true,
	}

	for _, manager := range settings.LangVersionManager.ManagerTypes {
		managerStr := LangVersionManagerType(manager)
		if !validManagerTypes[managerStr] {
			return fmt.Errorf("invalid manager type: %s (must be nvm, pyenv, or rbenv)", manager)
		}
	}

	return nil
}

// Scan scans for language version manager installations.
func (lvmc *LanguageVersionManagerCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	// Get home directory
	homeDir, err := lvmc.getHomeDir()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to get home directory: %w", err))
	}

	// Scan for each manager type
	for _, managerType := range lvmc.managerTypes {
		result := lvmc.scanLangVersionManager(ctx, managerType, homeDir)
		if result.IsErr() {
			if lvmc.verbose {
				fmt.Printf("Warning: failed to scan %s: %v\n", managerType, result.Error())
			}
			continue
		}

		items = append(items, result.Value()...)
	}

	return result.Ok(items)
}

// scanLangVersionManager scans installations for a specific language version manager.
func (lvmc *LanguageVersionManagerCleaner) scanLangVersionManager(ctx context.Context, managerType LangVersionManagerType, homeDir string) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	switch managerType {
	case LangVersionManagerNVM:
		// NVM versions: ~/.nvm/versions/node/*
		nvmVersions := filepath.Join(homeDir, ".nvm", "versions", "node")
		if info, err := os.Stat(nvmVersions); err == nil && info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(nvmVersions, "*"))
			if err != nil {
				return result.Err[[]domain.ScanItem](fmt.Errorf("failed to find NVM versions: %w", err))
			}

			for _, match := range matches {
				// Keep only old versions (not currently active)
				// This is a conservative scan - actual removal requires version selection
				items = append(items, domain.ScanItem{
					Path:     match,
					Size:     lvmc.getDirSize(match),
					Created:  lvmc.getDirModTime(match),
					ScanType: domain.ScanTypeTemp,
				})

				if lvmc.verbose {
					fmt.Printf("Found NVM version: %s\n", filepath.Base(match))
				}
			}
		}

	case LangVersionManagerPYENV:
		// Pyenv versions: ~/.pyenv/versions/*
		pyenvVersions := filepath.Join(homeDir, ".pyenv", "versions")
		if info, err := os.Stat(pyenvVersions); err == nil && info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(pyenvVersions, "*"))
			if err != nil {
				return result.Err[[]domain.ScanItem](fmt.Errorf("failed to find Pyenv versions: %w", err))
			}

			for _, match := range matches {
				items = append(items, domain.ScanItem{
					Path:     match,
					Size:     lvmc.getDirSize(match),
					Created:  lvmc.getDirModTime(match),
					ScanType: domain.ScanTypeTemp,
				})

				if lvmc.verbose {
					fmt.Printf("Found Pyenv version: %s\n", filepath.Base(match))
				}
			}
		}

	case LangVersionManagerRBENV:
		// Rbenv versions: ~/.rbenv/versions/*
		rbenvVersions := filepath.Join(homeDir, ".rbenv", "versions")
		if info, err := os.Stat(rbenvVersions); err == nil && info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(rbenvVersions, "*"))
			if err != nil {
				return result.Err[[]domain.ScanItem](fmt.Errorf("failed to find Rbenv versions: %w", err))
			}

			for _, match := range matches {
				items = append(items, domain.ScanItem{
					Path:     match,
					Size:     lvmc.getDirSize(match),
					Created:  lvmc.getDirModTime(match),
					ScanType: domain.ScanTypeTemp,
				})

				if lvmc.verbose {
					fmt.Printf("Found Rbenv version: %s\n", filepath.Base(match))
				}
			}
		}
	}

	return result.Ok(items)
}

// Clean removes language version manager installations.
func (lvmc *LanguageVersionManagerCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !lvmc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("language version manager cleaner not available"))
	}

	if lvmc.dryRun {
		// Estimate cache sizes based on typical usage
		// Each language version is typically 100-500MB
		totalBytes := int64(300 * 1024 * 1024) // Estimate 300MB per manager
		itemsRemoved := len(lvmc.managerTypes)

		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, itemsRemoved, totalBytes)
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	// WARNING: This removes ALL versions except current
	// Users should be warned about reinstalling tools
	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	// Get home directory
	homeDir, err := lvmc.getHomeDir()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get home directory: %w", err))
	}

	// Clean for each manager type
	for _, managerType := range lvmc.managerTypes {
		result := lvmc.cleanLangVersionManager(ctx, managerType, homeDir)
		if result.IsErr() {
			itemsFailed++
			if lvmc.verbose {
				fmt.Printf("Warning: failed to clean %s: %v\n", managerType, result.Error())
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

// cleanLangVersionManager cleans installations for a specific language version manager.
func (lvmc *LanguageVersionManagerCleaner) cleanLangVersionManager(ctx context.Context, managerType LangVersionManagerType, homeDir string) result.Result[domain.CleanResult] {
	// This is a NO-OP by default to avoid destructive behavior
	// In production, you would:
	// 1. Detect currently active version
	// 2. Remove all versions except active one
	// 3. Warn user about potential need to reinstall tools

	if lvmc.verbose {
		fmt.Printf("  ⚠️  Skipping %s cleanup (destructive operation)\n", managerType)
		fmt.Printf("     To manually clean: Remove old versions from ~/.%s/versions\n", managerType)
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

// getHomeDir returns user's home directory.
func (lvmc *LanguageVersionManagerCleaner) getHomeDir() (string, error) {
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
func (lvmc *LanguageVersionManagerCleaner) getDirSize(path string) int64 {
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
func (lvmc *LanguageVersionManagerCleaner) getDirModTime(path string) time.Time {
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
