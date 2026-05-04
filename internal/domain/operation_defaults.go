package domain

import (
	"fmt"
	"runtime"
)

// settingsFactory maps operation types to their default settings factories.
var settingsFactory = map[OperationType]func() *OperationSettings{
	OperationTypeNixGenerations:               defaultNixGenerationsSettings,
	OperationTypeTempFiles:                    defaultTempFilesSettings,
	OperationTypeHomebrew:                     defaultHomebrewSettings,
	OperationTypeNodePackages:                 defaultNodePackagesSettings,
	OperationTypeGoPackages:                   defaultGoPackagesSettings,
	OperationTypeCargoPackages:                defaultCargoPackagesSettings,
	OperationTypeBuildCache:                   func() *OperationSettings { return &OperationSettings{BuildCache: defaultBuildCacheSettings()} },
	OperationTypeDocker:                       defaultDockerSettings,
	OperationTypeSystemCache:                  defaultSystemCacheSettings,
	OperationTypeSystemTemp:                   defaultSystemTempSettings,
	OperationTypeProjectsManagementAutomation: defaultProjectsManagementAutomationSettings,
	OperationTypeProjectExecutables:           defaultProjectExecutablesSettings,
	OperationTypeCompiledBinaries:             defaultCompiledBinariesSettings,
	OperationTypeGitHistory:                   func() *OperationSettings { return &OperationSettings{GitHistory: &GitHistorySettings{}} },
	OperationTypeGolangciLintCache:            func() *OperationSettings { return &OperationSettings{} },
}

// DefaultSettings returns default settings for given operation type.
func DefaultSettings(opType OperationType) *OperationSettings {
	factory, ok := settingsFactory[opType]
	if !ok {
		return &OperationSettings{}
	}

	settings := factory()

	err := validateEnumDefaults(settings, opType)
	if err != nil {
		panic(fmt.Sprintf("DefaultSettings validation failed for %s: %v", opType, err))
	}

	return settings
}

func defaultNixGenerationsSettings() *OperationSettings {
	return &OperationSettings{
		NixGenerations: &NixGenerationsSettings{
			Generations: 1,
			Optimize:    OptimizationModeDisabled,
			DryRun:      ExecutionModeNormal,
		},
	}
}

func defaultTempFilesSettings() *OperationSettings {
	return &OperationSettings{
		TempFiles: &TempFilesSettings{
			OlderThan: "7d",
			Excludes:  []string{"/tmp/keep"},
		},
	}
}

func defaultHomebrewSettings() *OperationSettings {
	return &OperationSettings{
		Homebrew: &HomebrewSettings{
			UnusedOnly: HomebrewModeUnusedOnly,
		},
	}
}

func defaultNodePackagesSettings() *OperationSettings {
	return &OperationSettings{
		NodePackages: &NodePackagesSettings{
			PackageManagers: []PackageManagerType{
				PackageManagerNpm,
				PackageManagerPnpm,
				PackageManagerYarn,
				PackageManagerBun,
			},
		},
	}
}

func defaultGoPackagesSettings() *OperationSettings {
	return &OperationSettings{
		GoPackages: &GoPackagesSettings{
			CleanCache:      CacheCleanupEnabled,
			CleanTestCache:  CacheCleanupEnabled,
			CleanModCache:   CacheCleanupDisabled,
			CleanBuildCache: CacheCleanupEnabled,
			CleanLintCache:  CacheCleanupDisabled,
		},
	}
}

func defaultCargoPackagesSettings() *OperationSettings {
	return &OperationSettings{
		CargoPackages: &CargoPackagesSettings{
			Autoclean: CacheCleanupEnabled,
		},
	}
}

func defaultDockerSettings() *OperationSettings {
	return &OperationSettings{
		Docker: &DockerSettings{
			PruneMode: DockerPruneAll,
		},
	}
}

func defaultSystemCacheSettings() *OperationSettings {
	return &OperationSettings{
		SystemCache: &SystemCacheSettings{
			CacheTypes: getDefaultSystemCacheTypes(),
			OlderThan:  "30d",
		},
	}
}

func defaultSystemTempSettings() *OperationSettings {
	return &OperationSettings{
		SystemTemp: &SystemTempSettings{
			Paths:     []string{"/tmp", "/var/tmp"},
			OlderThan: "30d",
		},
	}
}

func defaultProjectsManagementAutomationSettings() *OperationSettings {
	return &OperationSettings{
		ProjectsManagementAutomation: &ProjectsManagementAutomationSettings{
			ClearCache: CacheCleanupEnabled,
		},
	}
}

func defaultProjectExecutablesSettings() *OperationSettings {
	return &OperationSettings{
		ProjectExecutables: &ProjectExecutablesSettings{
			ExcludeExtensions: []string{".sh"},
		},
	}
}

func defaultCompiledBinariesSettings() *OperationSettings {
	return &OperationSettings{
		CompiledBinaries: &CompiledBinariesSettings{
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

func validateNodePackagesDefaults(s *NodePackagesSettings) error {
	if s == nil {
		return nil
	}
	for i, pm := range s.PackageManagers {
		if !pm.IsValid() {
			return fmt.Errorf("invalid default PackageManagerType at index %d: %d", i, pm)
		}
	}
	return nil
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
	for i, tt := range s.ToolTypes {
		if !tt.IsValid() {
			return fmt.Errorf("invalid default BuildToolType at index %d: %d", i, tt)
		}
	}
	return nil
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
	for i, ct := range s.CacheTypes {
		if !ct.IsValid() {
			return fmt.Errorf("invalid default CacheType at index %d: %d", i, ct)
		}
	}
	return nil
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
		ToolTypes: []BuildToolType{
			BuildToolGo,
			BuildToolRust,
			BuildToolNode,
			BuildToolPython,
			BuildToolJava,
			BuildToolScala,
		},
		OlderThan: "30d",
	}
}

// getDefaultSystemCacheTypes returns platform-appropriate default cache types.
func getDefaultSystemCacheTypes() []CacheType {
	switch runtime.GOOS {
	case "darwin":
		return []CacheType{
			CacheTypeSpotlight,
			CacheTypeXcode,
			CacheTypeCocoapods,
			CacheTypeHomebrew,
		}
	case "linux":
		return []CacheType{
			CacheTypeXdgCache,
			CacheTypeThumbnails,
			CacheTypeHomebrew,
		}
	default:
		return []CacheType{
			CacheTypeHomebrew,
		}
	}
}
