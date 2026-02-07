package domain

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
	PackageManagers []string `json:"package_managers" yaml:"package_managers"`
}

// GoPackagesSettings provides type-safe settings for Go language cleanup.
type GoPackagesSettings struct {
	CleanCache      bool `json:"clean_cache,omitempty"       yaml:"clean_cache,omitempty"`
	CleanTestCache  bool `json:"clean_test_cache,omitempty"  yaml:"clean_test_cache,omitempty"`
	CleanModCache   bool `json:"clean_mod_cache,omitempty"   yaml:"clean_mod_cache,omitempty"`
	CleanBuildCache bool `json:"clean_build_cache,omitempty" yaml:"clean_build_cache,omitempty"`
	CleanLintCache  bool `json:"clean_lint_cache,omitempty"  yaml:"clean_lint_cache,omitempty"`
}

// CargoPackagesSettings provides type-safe settings for Cargo package manager cleanup.
type CargoPackagesSettings struct {
	Autoclean bool `json:"autoclean,omitempty" yaml:"autoclean,omitempty"`
}

// BuildCacheSettings provides type-safe settings for build cache cleanup.
type BuildCacheSettings struct {
	ToolTypes []string `json:"tool_types,omitempty" yaml:"tool_types,omitempty"`
	OlderThan string   `json:"older_than,omitempty" yaml:"older_than,omitempty"`
}

// DockerSettings provides type-safe settings for Docker cleanup.
type DockerSettings struct {
	PruneMode string `json:"prune_mode,omitempty" yaml:"prune_mode,omitempty"`
}

// SystemCacheSettings provides type-safe settings for macOS system cache cleanup.
type SystemCacheSettings struct {
	CacheTypes []string `json:"cache_types,omitempty" yaml:"cache_types,omitempty"`
	OlderThan  string   `json:"older_than,omitempty"  yaml:"older_than,omitempty"`
}

// LangVersionManagerSettings provides type-safe settings for language version manager cleanup.
type LangVersionManagerSettings struct {
	ManagerTypes []string `json:"manager_types,omitempty" yaml:"manager_types,omitempty"`
}

// SystemTempSettings provides type-safe settings for system temp cleanup.
type SystemTempSettings struct {
	Paths     []string `json:"paths"      yaml:"paths"`
	OlderThan string   `json:"older_than" yaml:"older_than"`
}

// ProjectsManagementAutomationSettings provides type-safe settings for projects management automation cleanup.
type ProjectsManagementAutomationSettings struct {
	ClearCache bool `json:"clear_cache" yaml:"clear_cache"`
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
				PackageManagers: []string{"npm", "pnpm", "yarn", "bun"},
			},
		}
	case OperationTypeGoPackages:
		return &OperationSettings{
			GoPackages: &GoPackagesSettings{
				CleanCache:      true,
				CleanTestCache:  true,
				CleanModCache:   false,
				CleanBuildCache: true,
				CleanLintCache:  false,
			},
		}
	case OperationTypeCargoPackages:
		return &OperationSettings{
			CargoPackages: &CargoPackagesSettings{
				Autoclean: true,
			},
		}
	case OperationTypeBuildCache:
		return &OperationSettings{
			BuildCache: &BuildCacheSettings{
				ToolTypes: []string{"gradle", "maven", "sbt"},
				OlderThan: "30d",
			},
		}
	case OperationTypeDocker:
		return &OperationSettings{
			Docker: &DockerSettings{
				PruneMode: "standard",
			},
		}
	case OperationTypeSystemCache:
		return &OperationSettings{
			SystemCache: &SystemCacheSettings{
				CacheTypes: []string{"spotlight", "xcode", "cocoapods", "homebrew"},
				OlderThan:  "30d",
			},
		}
	case OperationTypeLangVersionManager:
		return &OperationSettings{
			LangVersionManager: &LangVersionManagerSettings{
				ManagerTypes: []string{"nvm", "pyenv", "rbenv"},
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
				ClearCache: true,
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
