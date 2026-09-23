package operations

import (
	"fmt"
	"runtime"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
)

// settingsFactory maps operation types to their default settings factories.
var settingsFactory = map[OperationType]func() *OperationSettings{ //nolint:gochecknoglobals
	OperationTypeNixGenerations:               defaultNixGenerationsSettings,
	OperationTypeTempFiles:                    defaultTempFilesSettings,
	OperationTypeHomebrew:                     defaultHomebrewSettings,
	OperationTypeNodePackages:                 defaultNodePackagesSettings,
	OperationTypeGoPackages:                   defaultGoPackagesSettings,
	OperationTypeCargoPackages:                defaultCargoPackagesSettings,
	OperationTypeBuildCache:                   func() *OperationSettings { return &OperationSettings{BuildCache: defaultBuildCacheSettings()} }, //nolint:exhaustruct
	OperationTypeDocker:                       defaultDockerSettings,
	OperationTypeSystemCache:                  defaultSystemCacheSettings,
	OperationTypeSystemTemp:                   defaultSystemTempSettings,
	OperationTypeProjectsManagementAutomation: defaultProjectsManagementAutomationSettings,
	OperationTypeProjectExecutables:           defaultProjectExecutablesSettings,
	OperationTypeCompiledBinaries:             defaultCompiledBinariesSettings,
	OperationTypeGitHistory:                   func() *OperationSettings { return &OperationSettings{GitHistory: &GitHistorySettings{}} }, //nolint:exhaustruct
	OperationTypeGolangciLintCache:            func() *OperationSettings { return &OperationSettings{} },                                  //nolint:exhaustruct
}

// DefaultSettings returns default settings for given operation type.
func DefaultSettings(opType OperationType) *OperationSettings {
	factory, ok := settingsFactory[opType]
	if !ok {
		return &OperationSettings{} //nolint:exhaustruct
	}

	settings := factory()

	err := validateEnumDefaults(settings, opType)
	if err != nil {
		panic(fmt.Sprintf("DefaultSettings validation failed for %s: %v", opType, err))
	}

	return settings
}

func defaultNixGenerationsSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		NixGenerations: &NixGenerationsSettings{
			Generations: 1,
			Optimize:    enums.OptimizationModeDisabled,
			DryRun:      enums.ExecutionModeNormal,
		},
	}
}

func defaultTempFilesSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		TempFiles: &TempFilesSettings{
			OlderThan: "7d",
			Excludes:  []string{"/tmp/keep"},
		},
	}
}

func defaultHomebrewSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		Homebrew: &HomebrewSettings{ //nolint:exhaustruct
			UnusedOnly: enums.HomebrewModeUnusedOnly,
		},
	}
}

func defaultNodePackagesSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		NodePackages: &NodePackagesSettings{
			PackageManagers: []enums.PackageManagerType{
				enums.PackageManagerNpm,
				enums.PackageManagerPnpm,
				enums.PackageManagerYarn,
				enums.PackageManagerBun,
			},
		},
	}
}

func defaultGoPackagesSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		GoPackages: &GoPackagesSettings{
			CleanCache:      enums.CacheCleanupEnabled,
			CleanTestCache:  enums.CacheCleanupEnabled,
			CleanModCache:   enums.CacheCleanupDisabled,
			CleanBuildCache: enums.CacheCleanupEnabled,
			CleanLintCache:  enums.CacheCleanupDisabled,
		},
	}
}

func defaultCargoPackagesSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		CargoPackages: &CargoPackagesSettings{
			Autoclean: enums.CacheCleanupEnabled,
		},
	}
}

func defaultDockerSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		Docker: &DockerSettings{
			PruneMode: enums.DockerPruneAll,
		},
	}
}

func defaultSystemCacheSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		SystemCache: &SystemCacheSettings{
			CacheTypes: getDefaultSystemCacheTypes(),
			OlderThan:  "30d",
		},
	}
}

func defaultSystemTempSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		SystemTemp: &SystemTempSettings{
			Paths:     []string{"/tmp", "/var/tmp"},
			OlderThan: "30d",
		},
	}
}

func defaultProjectsManagementAutomationSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		ProjectsManagementAutomation: &ProjectsManagementAutomationSettings{
			ClearCache: enums.CacheCleanupEnabled,
		},
	}
}

func defaultProjectExecutablesSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		ProjectExecutables: &ProjectExecutablesSettings{ //nolint:exhaustruct
			ExcludeExtensions: []string{".sh"},
		},
	}
}

func defaultCompiledBinariesSettings() *OperationSettings {
	return &OperationSettings{ //nolint:exhaustruct
		CompiledBinaries: &CompiledBinariesSettings{ //nolint:exhaustruct
			MinSizeMB: 10,
			OlderThan: "0",
		},
	}
}

// validateEnumDefaults validates all enum values in default settings.
func validateEnumDefaults(settings *OperationSettings, opType OperationType) error {
	if settings == nil {
		return fmt.Errorf("nil settings for operation type: %s", opType)
	}

	if err := validateNixGenerationsDefaults(settings.NixGenerations); err != nil {
		return err
	}

	if err := validateHomebrewDefaults(settings.Homebrew); err != nil {
		return err
	}

	if err := validateNodePackagesDefaults(settings.NodePackages); err != nil {
		return err
	}

	if err := validateGoPackagesDefaults(settings.GoPackages); err != nil {
		return err
	}

	if err := validateCargoPackagesDefaults(settings.CargoPackages); err != nil {
		return err
	}

	if err := validateBuildCacheDefaults(settings.BuildCache); err != nil {
		return err
	}

	if err := validateDockerDefaults(settings.Docker); err != nil {
		return err
	}

	if err := validateSystemCacheDefaults(settings.SystemCache); err != nil {
		return err
	}

	return validateProjectsAutomationDefaults(settings.ProjectsManagementAutomation)
}

func validateNixGenerationsDefaults(s *NixGenerationsSettings) error {
	if s == nil {
		return nil
	}

	if !s.Optimize.IsValid() {
		return fmt.Errorf("invalid default OptimizationMode in NixGenerations: %d", s.Optimize)
	}

	if !s.DryRun.IsValid() {
		return fmt.Errorf("invalid default ExecutionMode in NixGenerations: %d", s.DryRun)
	}

	return nil
}

func validateHomebrewDefaults(s *HomebrewSettings) error {
	if s == nil {
		return nil
	}

	if !s.UnusedOnly.IsValid() {
		return fmt.Errorf("invalid default HomebrewMode: %d", s.UnusedOnly)
	}

	return nil
}

// validateEnumSliceDefaults reports the first invalid enum element in items.
// typeName identifies the enum in the error message (e.g. "PackageManagerType").
func validateEnumSliceDefaults[T interface{ IsValid() bool }](items []T, typeName string) error {
	for i, item := range items {
		if !item.IsValid() {
			return fmt.Errorf("invalid default %s at index %d: %d", typeName, i, item)
		}
	}

	return nil
}

func validateNodePackagesDefaults(s *NodePackagesSettings) error {
	if s == nil {
		return nil
	}

	return validateEnumSliceDefaults(s.PackageManagers, "PackageManagerType")
}

func validateGoPackagesDefaults(s *GoPackagesSettings) error {
	if s == nil {
		return nil
	}

	if !s.CleanCache.IsValid() {
		return fmt.Errorf("invalid default CacheCleanupMode for CleanCache: %d", s.CleanCache)
	}

	if !s.CleanTestCache.IsValid() {
		return fmt.Errorf(
			"invalid default CacheCleanupMode for CleanTestCache: %d",
			s.CleanTestCache,
		)
	}

	if !s.CleanModCache.IsValid() {
		return fmt.Errorf("invalid default CacheCleanupMode for CleanModCache: %d", s.CleanModCache)
	}

	if !s.CleanBuildCache.IsValid() {
		return fmt.Errorf(
			"invalid default CacheCleanupMode for CleanBuildCache: %d",
			s.CleanBuildCache,
		)
	}

	if !s.CleanLintCache.IsValid() {
		return fmt.Errorf(
			"invalid default CacheCleanupMode for CleanLintCache: %d",
			s.CleanLintCache,
		)
	}

	return nil
}

func validateCargoPackagesDefaults(s *CargoPackagesSettings) error {
	if s == nil {
		return nil
	}

	if !s.Autoclean.IsValid() {
		return fmt.Errorf("invalid default CacheCleanupMode in CargoPackages: %d", s.Autoclean)
	}

	return nil
}

func validateBuildCacheDefaults(s *BuildCacheSettings) error {
	if s == nil {
		return nil
	}

	return validateEnumSliceDefaults(s.ToolTypes, "BuildToolType")
}

func validateDockerDefaults(s *DockerSettings) error {
	if s == nil {
		return nil
	}

	if !s.PruneMode.IsValid() {
		return fmt.Errorf("invalid default DockerPruneMode: %d", s.PruneMode)
	}

	return nil
}

func validateSystemCacheDefaults(s *SystemCacheSettings) error {
	if s == nil {
		return nil
	}

	return validateEnumSliceDefaults(s.CacheTypes, "CacheType")
}

func validateProjectsAutomationDefaults(s *ProjectsManagementAutomationSettings) error {
	if s == nil {
		return nil
	}

	if !s.ClearCache.IsValid() {
		return fmt.Errorf(
			"invalid default CacheCleanupMode in ProjectsManagementAutomation: %d",
			s.ClearCache,
		)
	}

	return nil
}

// defaultBuildCacheSettings returns default settings for build cache cleanup.
func defaultBuildCacheSettings() *BuildCacheSettings {
	return &BuildCacheSettings{
		ToolTypes: []enums.BuildToolType{
			enums.BuildToolGo,
			enums.BuildToolRust,
			enums.BuildToolNode,
			enums.BuildToolPython,
			enums.BuildToolJava,
			enums.BuildToolScala,
		},
		OlderThan: "30d",
	}
}

// getDefaultSystemCacheTypes returns platform-appropriate default cache types.
func getDefaultSystemCacheTypes() []enums.CacheType {
	switch runtime.GOOS {
	case "darwin":
		return []enums.CacheType{
			enums.CacheTypeSpotlight,
			enums.CacheTypeXcode,
			enums.CacheTypeCocoapods,
			enums.CacheTypeHomebrew,
		}
	case "linux":
		return []enums.CacheType{
			enums.CacheTypeXdgCache,
			enums.CacheTypeThumbnails,
			enums.CacheTypePip,
			enums.CacheTypeNpm,
			enums.CacheTypeYarn,
			enums.CacheTypeCcache,
		}
	default:
		return []enums.CacheType{
			enums.CacheTypeHomebrew,
		}
	}
}
