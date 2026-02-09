package domain

import (
	"fmt"
)

// DefaultSettings returns default settings for given operation type.
func DefaultSettings(opType OperationType) *OperationSettings {
	var settings *OperationSettings
	switch opType {
	case OperationTypeNixGenerations:
		settings = &OperationSettings{
			NixGenerations: &NixGenerationsSettings{
				Generations: 1,
				Optimize:    OptimizationModeDisabled,
				DryRun:      ExecutionModeNormal,
			},
		}
	case OperationTypeTempFiles:
		settings = &OperationSettings{
			TempFiles: &TempFilesSettings{
				OlderThan: "7d",
				Excludes:  []string{"/tmp/keep"},
			},
		}
	case OperationTypeHomebrew:
		settings = &OperationSettings{
			Homebrew: &HomebrewSettings{
				UnusedOnly: HomebrewModeUnusedOnly,
			},
		}
	case OperationTypeNodePackages:
		settings = &OperationSettings{
			NodePackages: &NodePackagesSettings{
				PackageManagers: []PackageManagerType{PackageManagerNpm, PackageManagerPnpm, PackageManagerYarn, PackageManagerBun},
			},
		}
	case OperationTypeGoPackages:
		settings = &OperationSettings{
			GoPackages: &GoPackagesSettings{
				CleanCache:      CacheCleanupEnabled,
				CleanTestCache:  CacheCleanupEnabled,
				CleanModCache:   CacheCleanupDisabled,
				CleanBuildCache: CacheCleanupEnabled,
				CleanLintCache:  CacheCleanupDisabled,
			},
		}
	case OperationTypeCargoPackages:
		settings = &OperationSettings{
			CargoPackages: &CargoPackagesSettings{
				Autoclean: CacheCleanupEnabled,
			},
		}
	case OperationTypeBuildCache:
		settings = &OperationSettings{
			BuildCache: &BuildCacheSettings{
				ToolTypes: []BuildToolType{BuildToolJava, BuildToolScala},
				OlderThan: "30d",
			},
		}
	case OperationTypeDocker:
		settings = &OperationSettings{
			Docker: &DockerSettings{
				PruneMode: DockerPruneAll,
			},
		}
	case OperationTypeSystemCache:
		settings = &OperationSettings{
			SystemCache: &SystemCacheSettings{
				CacheTypes: []CacheType{CacheTypeSpotlight, CacheTypeXcode, CacheTypeCocoapods, CacheTypeHomebrew},
				OlderThan:  "30d",
			},
		}
	case OperationTypeLangVersionManager:
		settings = &OperationSettings{
			LangVersionManager: &LangVersionManagerSettings{
				ManagerTypes: []VersionManagerType{VersionManagerNvm, VersionManagerPyenv, VersionManagerRbenv},
			},
		}
	case OperationTypeSystemTemp:
		settings = &OperationSettings{
			SystemTemp: &SystemTempSettings{
				Paths:     []string{"/tmp", "/var/tmp"},
				OlderThan: "30d",
			},
		}
	case OperationTypeProjectsManagementAutomation:
		settings = &OperationSettings{
			ProjectsManagementAutomation: &ProjectsManagementAutomationSettings{
				ClearCache: CacheCleanupEnabled,
			},
		}
	default:
		return &OperationSettings{} // Empty settings for custom types
	}

	// Validate all enum defaults
	if err := validateEnumDefaults(settings, opType); err != nil {
		panic(fmt.Sprintf("DefaultSettings validation failed for %s: %v", opType, err))
	}

	return settings
}

// validateEnumDefaults validates all enum values in default settings.
// This is called from DefaultSettings to ensure defaults are valid.
func validateEnumDefaults(settings *OperationSettings, opType OperationType) error {
	if settings == nil {
		return fmt.Errorf("nil settings for operation type: %s", opType)
	}

	// Validate NixGenerations
	if settings.NixGenerations != nil {
		if !settings.NixGenerations.Optimize.IsValid() {
			return fmt.Errorf("invalid default OptimizationMode in NixGenerations: %d", settings.NixGenerations.Optimize)
		}
		if !settings.NixGenerations.DryRun.IsValid() {
			return fmt.Errorf("invalid default ExecutionMode in NixGenerations: %d", settings.NixGenerations.DryRun)
		}
	}

	// Validate Homebrew
	if settings.Homebrew != nil {
		if !settings.Homebrew.UnusedOnly.IsValid() {
			return fmt.Errorf("invalid default HomebrewMode: %d", settings.Homebrew.UnusedOnly)
		}
	}

	// Validate NodePackages
	if settings.NodePackages != nil {
		for i, pm := range settings.NodePackages.PackageManagers {
			if !pm.IsValid() {
				return fmt.Errorf("invalid default PackageManagerType at index %d: %d", i, pm)
			}
		}
	}

	// Validate GoPackages
	if settings.GoPackages != nil {
		if !settings.GoPackages.CleanCache.IsValid() {
			return fmt.Errorf("invalid default CacheCleanupMode for CleanCache: %d", settings.GoPackages.CleanCache)
		}
		if !settings.GoPackages.CleanTestCache.IsValid() {
			return fmt.Errorf("invalid default CacheCleanupMode for CleanTestCache: %d", settings.GoPackages.CleanTestCache)
		}
		if !settings.GoPackages.CleanModCache.IsValid() {
			return fmt.Errorf("invalid default CacheCleanupMode for CleanModCache: %d", settings.GoPackages.CleanModCache)
		}
		if !settings.GoPackages.CleanBuildCache.IsValid() {
			return fmt.Errorf("invalid default CacheCleanupMode for CleanBuildCache: %d", settings.GoPackages.CleanBuildCache)
		}
		if !settings.GoPackages.CleanLintCache.IsValid() {
			return fmt.Errorf("invalid default CacheCleanupMode for CleanLintCache: %d", settings.GoPackages.CleanLintCache)
		}
	}

	// Validate CargoPackages
	if settings.CargoPackages != nil {
		if !settings.CargoPackages.Autoclean.IsValid() {
			return fmt.Errorf("invalid default CacheCleanupMode in CargoPackages: %d", settings.CargoPackages.Autoclean)
		}
	}

	// Validate BuildCache
	if settings.BuildCache != nil {
		for i, tt := range settings.BuildCache.ToolTypes {
			if !tt.IsValid() {
				return fmt.Errorf("invalid default BuildToolType at index %d: %d", i, tt)
			}
		}
	}

	// Validate Docker
	if settings.Docker != nil {
		if !settings.Docker.PruneMode.IsValid() {
			return fmt.Errorf("invalid default DockerPruneMode: %d", settings.Docker.PruneMode)
		}
	}

	// Validate SystemCache
	if settings.SystemCache != nil {
		for i, ct := range settings.SystemCache.CacheTypes {
			if !ct.IsValid() {
				return fmt.Errorf("invalid default CacheType at index %d: %d", i, ct)
			}
		}
	}

	// Validate LangVersionManager
	if settings.LangVersionManager != nil {
		for i, vmt := range settings.LangVersionManager.ManagerTypes {
			if !vmt.IsValid() {
				return fmt.Errorf("invalid default VersionManagerType at index %d: %d", i, vmt)
			}
		}
	}

	// Validate ProjectsManagementAutomation
	if settings.ProjectsManagementAutomation != nil {
		if !settings.ProjectsManagementAutomation.ClearCache.IsValid() {
			return fmt.Errorf("invalid default CacheCleanupMode in ProjectsManagementAutomation: %d", settings.ProjectsManagementAutomation.ClearCache)
		}
	}

	return nil
}