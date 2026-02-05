package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// LanguageVersionManagerCleaner handles language version manager cleanup.
type LanguageVersionManagerCleaner struct {
	verbose      bool
	dryRun       bool
	managerTypes []LangVersionManagerType
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
	homeDir, err := getHomeDir()
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
		nvmVersions := filepath.Join(homeDir, ".nvm", "versions", "node")
		if info, err := os.Stat(nvmVersions); err == nil && info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(nvmVersions, "*"))
			if err != nil {
				return result.Err[[]domain.ScanItem](fmt.Errorf("failed to find NVM versions: %w", err))
			}

			for _, match := range matches {
				items = append(items, domain.ScanItem{
					Path:     match,
					Size:     getDirSize(match),
					Created:  getDirModTime(match),
					ScanType: domain.ScanTypeTemp,
				})

				if lvmc.verbose {
					fmt.Printf("Found NVM version: %s\n", filepath.Base(match))
				}
			}
		}

	case LangVersionManagerPYENV:
		pyenvVersions := filepath.Join(homeDir, ".pyenv", "versions")
		if info, err := os.Stat(pyenvVersions); err == nil && info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(pyenvVersions, "*"))
			if err != nil {
				return result.Err[[]domain.ScanItem](fmt.Errorf("failed to find Pyenv versions: %w", err))
			}

			for _, match := range matches {
				items = append(items, domain.ScanItem{
					Path:     match,
					Size:     getDirSize(match),
					Created:  getDirModTime(match),
					ScanType: domain.ScanTypeTemp,
				})

				if lvmc.verbose {
					fmt.Printf("Found Pyenv version: %s\n", filepath.Base(match))
				}
			}
		}

	case LangVersionManagerRBENV:
		rbenvVersions := filepath.Join(homeDir, ".rbenv", "versions")
		if info, err := os.Stat(rbenvVersions); err == nil && info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(rbenvVersions, "*"))
			if err != nil {
				return result.Err[[]domain.ScanItem](fmt.Errorf("failed to find Rbenv versions: %w", err))
			}

			for _, match := range matches {
				items = append(items, domain.ScanItem{
					Path:     match,
					Size:     getDirSize(match),
					Created:  getDirModTime(match),
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
	return cleanWithIterator[LangVersionManagerType](
		ctx,
		"language version manager cleaner",
		lvmc.IsAvailable,
		lvmc.managerTypes,
		lvmc.cleanLangVersionManager,
		lvmc.verbose,
		lvmc.dryRun,
	)
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


