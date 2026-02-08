package domain

import "gopkg.in/yaml.v3"

// OperationSettings provides type-safe configuration for different operation types
// This eliminates map[string]any violations while maintaining flexibility.
type OperationSettings struct {
	// Nix Generations Settings
	NixGenerations *NixGenerationsSettings `json:"nix_generations,omitempty" yaml:"nix_generations,omitempty"`

	// Temp Files Settings
	TempFiles *TempFilesSettings `json:"temp_files,omitempty" yaml:"temp_files,omitempty"`

	// Homebrew Settings
	Homebrew *HomebrewSettings `json:"homebrew,omitempty" yaml:"homebrew,omitempty"`

	// Node Packages Settings
	NodePackages *NodePackagesSettings `json:"node_packages,omitempty" yaml:"node_packages,omitempty"`

	// Go Packages Settings
	GoPackages *GoPackagesSettings `json:"go_packages,omitempty" yaml:"go_packages,omitempty"`

	// Cargo Packages Settings
	CargoPackages *CargoPackagesSettings `json:"cargo_packages,omitempty" yaml:"cargo_packages,omitempty"`

	// Build Cache Settings
	BuildCache *BuildCacheSettings `json:"build_cache,omitempty" yaml:"build_cache,omitempty"`

	// Docker Settings
	Docker *DockerSettings `json:"docker,omitempty" yaml:"docker,omitempty"`

	// System Cache Settings
	SystemCache *SystemCacheSettings `json:"system_cache,omitempty" yaml:"system_cache,omitempty"`

	// Language Version Manager Settings
	LangVersionManager *LangVersionManagerSettings `json:"lang_version_manager,omitempty" yaml:"lang_version_manager,omitempty"`

	// System Temp Settings
	SystemTemp *SystemTempSettings `json:"system_temp,omitempty" yaml:"system_temp,omitempty"`

	// Projects Management Automation Settings
	ProjectsManagementAutomation *ProjectsManagementAutomationSettings `json:"projects_management_automation,omitempty" yaml:"projects_management_automation,omitempty"`
}

// NixGenerationsSettings provides type-safe settings for Nix generations cleanup.
type NixGenerationsSettings struct {
	Generations int              `json:"generations"       yaml:"generations"`
	Optimize    OptimizationMode `json:"optimize"          yaml:"optimize"`
	DryRun      ExecutionMode    `json:"dry_run,omitempty" yaml:"dry_run,omitempty"`
}

// TempFilesSettings provides type-safe settings for temporary files cleanup.
type TempFilesSettings struct {
	OlderThan string   `json:"older_than"         yaml:"older_than"`
	Excludes  []string `json:"excludes,omitempty" yaml:"excludes,omitempty"`
}

// HomebrewSettings provides type-safe settings for Homebrew cleanup.
type HomebrewSettings struct {
	UnusedOnly HomebrewMode `json:"unused_only"     yaml:"unused_only"`
	Prune      string       `json:"prune,omitempty" yaml:"prune,omitempty"`
}

// NodePackagesSettings provides type-safe settings for Node.js package manager cleanup.
type NodePackagesSettings struct {
	PackageManagers []PackageManagerType `json:"package_managers" yaml:"package_managers"`
}

// GoPackagesSettings provides type-safe settings for Go language cleanup.
type GoPackagesSettings struct {
	CleanCache      CacheCleanupMode `json:"clean_cache,omitempty"       yaml:"clean_cache,omitempty"`
	CleanTestCache  CacheCleanupMode `json:"clean_test_cache,omitempty"  yaml:"clean_test_cache,omitempty"`
	CleanModCache   CacheCleanupMode `json:"clean_mod_cache,omitempty"   yaml:"clean_mod_cache,omitempty"`
	CleanBuildCache CacheCleanupMode `json:"clean_build_cache,omitempty" yaml:"clean_build_cache,omitempty"`
	CleanLintCache  CacheCleanupMode `json:"clean_lint_cache,omitempty"  yaml:"clean_lint_cache,omitempty"`
}

// CargoPackagesSettings provides type-safe settings for Cargo package manager cleanup.
type CargoPackagesSettings struct {
	Autoclean CacheCleanupMode `json:"autoclean,omitempty" yaml:"autoclean,omitempty"`
}

// BuildCacheSettings provides type-safe settings for build cache cleanup.
type BuildCacheSettings struct {
	ToolTypes []BuildToolType `json:"tool_types,omitempty" yaml:"tool_types,omitempty"`
	OlderThan string            `json:"older_than,omitempty" yaml:"older_than,omitempty"`
}

// DockerSettings provides type-safe settings for Docker cleanup.
type DockerSettings struct {
	PruneMode DockerPruneMode `json:"prune_mode,omitempty" yaml:"prune_mode,omitempty"`
}

// SystemCacheSettings provides type-safe settings for macOS system cache cleanup.
type SystemCacheSettings struct {
	CacheTypes []CacheType `json:"cache_types,omitempty" yaml:"cache_types,omitempty"`
	OlderThan  string     `json:"older_than,omitempty"  yaml:"older_than,omitempty"`
}

// LangVersionManagerSettings provides type-safe settings for language version manager cleanup.
type LangVersionManagerSettings struct {
	ManagerTypes []VersionManagerType `json:"manager_types,omitempty" yaml:"manager_types,omitempty"`
}

// SystemTempSettings provides type-safe settings for system temp cleanup.
type SystemTempSettings struct {
	Paths     []string `json:"paths"      yaml:"paths"`
	OlderThan string   `json:"older_than" yaml:"older_than"`
}

// ProjectsManagementAutomationSettings provides type-safe settings for projects management automation cleanup.
type ProjectsManagementAutomationSettings struct {
	ClearCache CacheCleanupMode `json:"clear_cache" yaml:"clear_cache"`
}

// OperationType represents different types of cleanup operations.
type OperationType string

const (
	OperationTypeNixGenerations               OperationType = "nix-generations"
	OperationTypeTempFiles                    OperationType = "temp-files"
	OperationTypeHomebrew                     OperationType = "homebrew-cleanup"
	OperationTypeNodePackages                 OperationType = "node-packages"
	OperationTypeGoPackages                   OperationType = "go-packages"
	OperationTypeCargoPackages                OperationType = "cargo-packages"
	OperationTypeBuildCache                   OperationType = "build-cache"
	OperationTypeDocker                       OperationType = "docker"
	OperationTypeSystemCache                  OperationType = "system-cache"
	OperationTypeLangVersionManager           OperationType = "lang-version-manager"
	OperationTypeSystemTemp                   OperationType = "system-temp"
	OperationTypeProjectsManagementAutomation OperationType = "projects-management-automation"
)

// GetOperationType returns the operation type from operation name.
func GetOperationType(name string) OperationType {
	switch name {
	case "nix-generations":
		return OperationTypeNixGenerations
	case "temp-files":
		return OperationTypeTempFiles
	case "homebrew-cleanup":
		return OperationTypeHomebrew
	case "node-packages":
		return OperationTypeNodePackages
	case "go-packages":
		return OperationTypeGoPackages
	case "cargo-packages":
		return OperationTypeCargoPackages
	case "build-cache":
		return OperationTypeBuildCache
	case "docker":
		return OperationTypeDocker
	case "system-cache":
		return OperationTypeSystemCache
	case "lang-version-manager":
		return OperationTypeLangVersionManager
	case "system-temp":
		return OperationTypeSystemTemp
	case "projects-management-automation":
		return OperationTypeProjectsManagementAutomation
	default:
		return OperationType(name) // Fallback for custom types
	}
}

// DefaultSettings returns default settings for the given operation type.
func DefaultSettings(opType OperationType) *OperationSettings {
	switch opType {
	case OperationTypeNixGenerations:
		return &OperationSettings{
			NixGenerations: &NixGenerationsSettings{
				Generations: 1,
				Optimize:    OptimizationModeDisabled,
				DryRun:      ExecutionModeNormal,
			},
		}
	case OperationTypeTempFiles:
		return &OperationSettings{
			TempFiles: &TempFilesSettings{
				OlderThan: "7d",
				Excludes:  []string{"/tmp/keep"},
			},
		}
	case OperationTypeHomebrew:
		return &OperationSettings{
			Homebrew: &HomebrewSettings{
				UnusedOnly: HomebrewModeUnusedOnly,
			},
		}
	case OperationTypeNodePackages:
		return &OperationSettings{
			NodePackages: &NodePackagesSettings{
				PackageManagers: []PackageManagerType{PackageManagerNpm, PackageManagerPnpm, PackageManagerYarn, PackageManagerBun},
			},
		}
	case OperationTypeGoPackages:
		return &OperationSettings{
			GoPackages: &GoPackagesSettings{
				CleanCache:      CacheCleanupEnabled,
				CleanTestCache:  CacheCleanupEnabled,
				CleanModCache:   CacheCleanupDisabled,
				CleanBuildCache: CacheCleanupEnabled,
				CleanLintCache:  CacheCleanupDisabled,
			},
		}
	case OperationTypeCargoPackages:
		return &OperationSettings{
			CargoPackages: &CargoPackagesSettings{
				Autoclean: CacheCleanupEnabled,
			},
		}
	case OperationTypeBuildCache:
		return &OperationSettings{
			BuildCache: &BuildCacheSettings{
				ToolTypes: []BuildToolType{BuildToolJava, BuildToolScala},
				OlderThan: "30d",
			},
		}
	case OperationTypeDocker:
		return &OperationSettings{
			Docker: &DockerSettings{
				PruneMode: DockerPruneAll,
			},
		}
	case OperationTypeSystemCache:
		return &OperationSettings{
			SystemCache: &SystemCacheSettings{
				CacheTypes: []CacheType{CacheTypeSpotlight, CacheTypeXcode, CacheTypeCocoapods, CacheTypeHomebrew},
				OlderThan:  "30d",
			},
		}
	case OperationTypeLangVersionManager:
		return &OperationSettings{
			LangVersionManager: &LangVersionManagerSettings{
				ManagerTypes: []VersionManagerType{VersionManagerNvm, VersionManagerPyenv, VersionManagerRbenv},
			},
		}
	case OperationTypeSystemTemp:
		return &OperationSettings{
			SystemTemp: &SystemTempSettings{
				Paths:     []string{"/tmp", "/var/tmp"},
				OlderThan: "30d",
			},
		}
	case OperationTypeProjectsManagementAutomation:
		return &OperationSettings{
			ProjectsManagementAutomation: &ProjectsManagementAutomationSettings{
				ClearCache: CacheCleanupEnabled,
			},
		}
	default:
		return &OperationSettings{} // Empty settings for custom types
	}
}

// ValidateSettings validates settings for the given operation type.
func (os *OperationSettings) ValidateSettings(opType OperationType) error {
	switch opType {
	case OperationTypeNixGenerations:
		if os.NixGenerations == nil {
			return nil // Optional settings
		}
		if os.NixGenerations.Generations < 0 || os.NixGenerations.Generations > 10 {
			return &ValidationError{
				Field:   "nix_generations.generations",
				Message: "generations must be between 0 and 10 (0 = keep only current)",
				Value:   os.NixGenerations.Generations,
			}
		}
	case OperationTypeTempFiles:
		if os.TempFiles == nil {
			return nil
		}
		if os.TempFiles.OlderThan == "" {
			return &ValidationError{
				Field:   "temp_files.older_than",
				Message: "older_than is required",
				Value:   os.TempFiles.OlderThan,
			}
		}
		if _, err := ParseCustomDuration(os.TempFiles.OlderThan); err != nil {
			return &ValidationError{
				Field:   "temp_files.older_than",
				Message: "older_than must be a valid duration (e.g., '7d', '24h')",
				Value:   os.TempFiles.OlderThan,
			}
		}
	case OperationTypeHomebrew:
		// Homebrew settings are always valid
	case OperationTypeSystemTemp:
		if os.SystemTemp == nil {
			return nil
		}
		if len(os.SystemTemp.Paths) == 0 {
			return &ValidationError{
				Field:   "system_temp.paths",
				Message: "paths are required",
				Value:   os.SystemTemp.Paths,
			}
		}
	}
	return nil
}

// ValidationError represents a settings validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// CacheCleanupMode represents cache cleanup behavior as a type-safe enum.
type CacheCleanupMode int

const (
	// CacheCleanupDisabled represents disabled cache cleanup.
	CacheCleanupDisabled CacheCleanupMode = iota
	// CacheCleanupEnabled represents enabled cache cleanup.
	CacheCleanupEnabled
)

// String returns string representation of cache cleanup mode.
func (cm CacheCleanupMode) String() string {
	switch cm {
	case CacheCleanupDisabled:
		return "DISABLED"
	case CacheCleanupEnabled:
		return "ENABLED"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if cache cleanup mode is valid.
func (cm CacheCleanupMode) IsValid() bool {
	return cm >= CacheCleanupDisabled && cm <= CacheCleanupEnabled
}

// Values returns all possible cache cleanup modes.
func (cm CacheCleanupMode) Values() []CacheCleanupMode {
	return []CacheCleanupMode{
		CacheCleanupDisabled,
		CacheCleanupEnabled,
	}
}

// IsEnabled checks if cache cleanup is enabled.
func (cm CacheCleanupMode) IsEnabled() bool {
	return cm == CacheCleanupEnabled
}

// MarshalYAML implements yaml.Marshaler interface for CacheCleanupMode.
func (cm CacheCleanupMode) MarshalYAML() (any, error) {
	return int(cm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for CacheCleanupMode.
// Accepts both string and integer representations.
func (cm *CacheCleanupMode) UnmarshalYAML(value *yaml.Node) error {
	return unmarshalBinaryEnum(
		value, cm,
		binaryEnumStringMap(CacheCleanupDisabled, CacheCleanupEnabled),
		binaryEnumIntMap(CacheCleanupDisabled, CacheCleanupEnabled),
		"cache cleanup mode",
	)
}

// DockerPruneMode represents Docker prune behavior as a type-safe enum.
type DockerPruneMode int

const (
	// DockerPruneAll represents pruning all resources.
	DockerPruneAll DockerPruneMode = iota
	// DockerPruneImages represents pruning only images.
	DockerPruneImages
	// DockerPruneContainers represents pruning only containers.
	DockerPruneContainers
	// DockerPruneVolumes represents pruning only volumes.
	DockerPruneVolumes
	// DockerPruneBuilds represents pruning only build cache.
	DockerPruneBuilds
)

// String returns string representation of Docker prune mode.
func (pm DockerPruneMode) String() string {
	switch pm {
	case DockerPruneAll:
		return "ALL"
	case DockerPruneImages:
		return "IMAGES"
	case DockerPruneContainers:
		return "CONTAINERS"
	case DockerPruneVolumes:
		return "VOLUMES"
	case DockerPruneBuilds:
		return "BUILDS"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if Docker prune mode is valid.
func (pm DockerPruneMode) IsValid() bool {
	return pm >= DockerPruneAll && pm <= DockerPruneBuilds
}

// Values returns all possible Docker prune modes.
func (pm DockerPruneMode) Values() []DockerPruneMode {
	return []DockerPruneMode{
		DockerPruneAll,
		DockerPruneImages,
		DockerPruneContainers,
		DockerPruneVolumes,
		DockerPruneBuilds,
	}
}

// MarshalYAML implements yaml.Marshaler interface for DockerPruneMode.
func (pm DockerPruneMode) MarshalYAML() (any, error) {
	return int(pm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for DockerPruneMode.
func (pm *DockerPruneMode) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, pm, map[string]DockerPruneMode{
		"ALL":        DockerPruneAll,
		"IMAGES":     DockerPruneImages,
		"CONTAINERS": DockerPruneContainers,
		"VOLUMES":    DockerPruneVolumes,
		"BUILDS":     DockerPruneBuilds,
	}, "invalid docker prune mode")
}

// BuildToolType represents build tool types as a type-safe enum.
type BuildToolType int

const (
	// BuildToolGo represents Go build tools.
	BuildToolGo BuildToolType = iota
	// BuildToolRust represents Rust/Cargo build tools.
	BuildToolRust
	// BuildToolNode represents Node.js build tools.
	BuildToolNode
	// BuildToolPython represents Python build tools.
	BuildToolPython
	// BuildToolJava represents Java build tools (Maven, Gradle).
	BuildToolJava
	// BuildToolScala represents Scala build tools (SBT).
	BuildToolScala
)

// String returns string representation of build tool type.
func (bt BuildToolType) String() string {
	switch bt {
	case BuildToolGo:
		return "GO"
	case BuildToolRust:
		return "RUST"
	case BuildToolNode:
		return "NODE"
	case BuildToolPython:
		return "PYTHON"
	case BuildToolJava:
		return "JAVA"
	case BuildToolScala:
		return "SCALA"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if build tool type is valid.
func (bt BuildToolType) IsValid() bool {
	return bt >= BuildToolGo && bt <= BuildToolScala
}

// Values returns all possible build tool types.
func (bt BuildToolType) Values() []BuildToolType {
	return []BuildToolType{
		BuildToolGo,
		BuildToolRust,
		BuildToolNode,
		BuildToolPython,
		BuildToolJava,
		BuildToolScala,
	}
}

// MarshalYAML implements yaml.Marshaler interface for BuildToolType.
func (bt BuildToolType) MarshalYAML() (any, error) {
	return int(bt), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for BuildToolType.
func (bt *BuildToolType) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, bt, map[string]BuildToolType{
		"GO":     BuildToolGo,
		"RUST":   BuildToolRust,
		"NODE":   BuildToolNode,
		"PYTHON": BuildToolPython,
		"JAVA":   BuildToolJava,
		"SCALA":  BuildToolScala,
	}, "invalid build tool type")
}

// CacheType represents system cache types as a type-safe enum.
type CacheType int

const (
	// CacheTypeSpotlight represents macOS Spotlight cache.
	CacheTypeSpotlight CacheType = iota
	// CacheTypeXcode represents Xcode derived data cache.
	CacheTypeXcode
	// CacheTypeCocoapods represents CocoaPods cache.
	CacheTypeCocoapods
	// CacheTypeHomebrew represents Homebrew cache.
	CacheTypeHomebrew
	// CacheTypePip represents Python pip cache.
	CacheTypePip
	// CacheTypeNpm represents Node.js npm cache.
	CacheTypeNpm
	// CacheTypeYarn represents Yarn cache.
	CacheTypeYarn
	// CacheTypeCcache represents ccache.
	CacheTypeCcache
)

// String returns string representation of cache type.
func (ct CacheType) String() string {
	switch ct {
	case CacheTypeSpotlight:
		return "SPOTLIGHT"
	case CacheTypeXcode:
		return "XCODE"
	case CacheTypeCocoapods:
		return "COCOAPODS"
	case CacheTypeHomebrew:
		return "HOMEBREW"
	case CacheTypePip:
		return "PIP"
	case CacheTypeNpm:
		return "NPM"
	case CacheTypeYarn:
		return "YARN"
	case CacheTypeCcache:
		return "CCACHE"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if cache type is valid.
func (ct CacheType) IsValid() bool {
	return ct >= CacheTypeSpotlight && ct <= CacheTypeCcache
}

// Values returns all possible cache types.
func (ct CacheType) Values() []CacheType {
	return []CacheType{
		CacheTypeSpotlight,
		CacheTypeXcode,
		CacheTypeCocoapods,
		CacheTypeHomebrew,
		CacheTypePip,
		CacheTypeNpm,
		CacheTypeYarn,
		CacheTypeCcache,
	}
}

// MarshalYAML implements yaml.Marshaler interface for CacheType.
func (ct CacheType) MarshalYAML() (any, error) {
	return int(ct), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for CacheType.
func (ct *CacheType) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, ct, map[string]CacheType{
		"SPOTLIGHT": CacheTypeSpotlight,
		"XCODE":      CacheTypeXcode,
		"COCOAPODS":  CacheTypeCocoapods,
		"HOMEBREW":   CacheTypeHomebrew,
		"PIP":        CacheTypePip,
		"NPM":        CacheTypeNpm,
		"YARN":       CacheTypeYarn,
		"CCACHE":     CacheTypeCcache,
	}, "invalid cache type")
}

// VersionManagerType represents language version manager types as a type-safe enum.
type VersionManagerType int

const (
	// VersionManagerNvm represents Node Version Manager.
	VersionManagerNvm VersionManagerType = iota
	// VersionManagerPyenv represents Python Version Manager.
	VersionManagerPyenv
	// VersionManagerGvm represents Go Version Manager.
	VersionManagerGvm
	// VersionManagerRbenv represents Ruby Version Manager.
	VersionManagerRbenv
	// VersionManagerSdkman represents SDKMAN for Java/Kotlin.
	VersionManagerSdkman
	// VersionManagerJenv represents Java Environment Manager.
	VersionManagerJenv
)

// String returns string representation of version manager type.
func (vm VersionManagerType) String() string {
	switch vm {
	case VersionManagerNvm:
		return "NVM"
	case VersionManagerPyenv:
		return "PYENV"
	case VersionManagerGvm:
		return "GVM"
	case VersionManagerRbenv:
		return "RBENV"
	case VersionManagerSdkman:
		return "SDKMAN"
	case VersionManagerJenv:
		return "JENV"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if version manager type is valid.
func (vm VersionManagerType) IsValid() bool {
	return vm >= VersionManagerNvm && vm <= VersionManagerJenv
}

// Values returns all possible version manager types.
func (vm VersionManagerType) Values() []VersionManagerType {
	return []VersionManagerType{
		VersionManagerNvm,
		VersionManagerPyenv,
		VersionManagerGvm,
		VersionManagerRbenv,
		VersionManagerSdkman,
		VersionManagerJenv,
	}
}

// MarshalYAML implements yaml.Marshaler interface for VersionManagerType.
func (vm VersionManagerType) MarshalYAML() (any, error) {
	return int(vm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for VersionManagerType.
func (vm *VersionManagerType) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, vm, map[string]VersionManagerType{
		"NVM":     VersionManagerNvm,
		"PYENV":   VersionManagerPyenv,
		"GVM":     VersionManagerGvm,
		"RBENV":   VersionManagerRbenv,
		"SDKMAN":  VersionManagerSdkman,
		"JENV":    VersionManagerJenv,
	}, "invalid version manager type")
}

// PackageManagerType represents Node.js package manager types as a type-safe enum.
type PackageManagerType int

const (
	// PackageManagerNpm represents npm.
	PackageManagerNpm PackageManagerType = iota
	// PackageManagerPnpm represents pnpm.
	PackageManagerPnpm
	// PackageManagerYarn represents Yarn.
	PackageManagerYarn
	// PackageManagerBun represents Bun.
	PackageManagerBun
)

// String returns string representation of package manager type.
func (pm PackageManagerType) String() string {
	switch pm {
	case PackageManagerNpm:
		return "NPM"
	case PackageManagerPnpm:
		return "PNPM"
	case PackageManagerYarn:
		return "YARN"
	case PackageManagerBun:
		return "BUN"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if package manager type is valid.
func (pm PackageManagerType) IsValid() bool {
	return pm >= PackageManagerNpm && pm <= PackageManagerBun
}

// Values returns all possible package manager types.
func (pm PackageManagerType) Values() []PackageManagerType {
	return []PackageManagerType{
		PackageManagerNpm,
		PackageManagerPnpm,
		PackageManagerYarn,
		PackageManagerBun,
	}
}

// MarshalYAML implements yaml.Marshaler interface for PackageManagerType.
func (pm PackageManagerType) MarshalYAML() (any, error) {
	return int(pm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for PackageManagerType.
func (pm *PackageManagerType) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, pm, map[string]PackageManagerType{
		"NPM":  PackageManagerNpm,
		"PNPM": PackageManagerPnpm,
		"YARN": PackageManagerYarn,
		"BUN":  PackageManagerBun,
	}, "invalid package manager type")
}
