package cleaner

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

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

// Name returns the cleaner name for result tracking.
func (lvmc *LanguageVersionManagerCleaner) Name() string {
	return "langversion"
}

// IsAvailable checks if language version manager cleaner is available.
func (lvmc *LanguageVersionManagerCleaner) IsAvailable(ctx context.Context) bool {
	// Language version manager cleaner is always available (uses file system operations)
	return true
}

// ValidateSettings validates language version manager cleaner settings.
func (lvmc *LanguageVersionManagerCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	return ValidateLangVersionManagerSettings(settings)
}

// Scan scans for language version manager installations.
func (lvmc *LanguageVersionManagerCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	return scanWithIterator[LangVersionManagerType](
		ctx,
		lvmc.managerTypes,
		lvmc.scanLangVersionManager,
		lvmc.verbose,
	)
}

// scanVersionDir scans a versions directory and returns scan items.
func (lvmc *LanguageVersionManagerCleaner) scanVersionDir(versionsDir, managerName string) ([]domain.ScanItem, error) {
	scanResult := ScanVersionDirectory(context.Background(), versionsDir, managerName, lvmc.verbose)
	if scanResult.IsErr() {
		return nil, scanResult.Error()
	}
	return scanResult.Value(), nil
}

// scanLangVersionManager scans installations for a specific language version manager.
func (lvmc *LanguageVersionManagerCleaner) scanLangVersionManager(ctx context.Context, managerType LangVersionManagerType, homeDir string) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	switch managerType {
	case LangVersionManagerNVM:
		nvmVersions := filepath.Join(homeDir, ".nvm", "versions", "node")
		scannedItems, err := lvmc.scanVersionDir(nvmVersions, "NVM")
		if err != nil {
			return result.Err[[]domain.ScanItem](err)
		}
		items = append(items, scannedItems...)

	case LangVersionManagerPYENV:
		pyenvVersions := filepath.Join(homeDir, ".pyenv", "versions")
		scannedItems, err := lvmc.scanVersionDir(pyenvVersions, "Pyenv")
		if err != nil {
			return result.Err[[]domain.ScanItem](err)
		}
		items = append(items, scannedItems...)

	case LangVersionManagerRBENV:
		rbenvVersions := filepath.Join(homeDir, ".rbenv", "versions")
		scannedItems, err := lvmc.scanVersionDir(rbenvVersions, "Rbenv")
		if err != nil {
			return result.Err[[]domain.ScanItem](err)
		}
		items = append(items, scannedItems...)
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
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
}
