package commands

// CleanerType represents available cleaner types for TUI selection.
type CleanerType string

const (
	CleanerTypeNix                          CleanerType = "nix"
	CleanerTypeHomebrew                     CleanerType = "homebrew"
	CleanerTypeTempFiles                    CleanerType = "tempfiles"
	CleanerTypeNodePackages                 CleanerType = "node"
	CleanerTypeGoPackages                   CleanerType = "go"
	CleanerTypeCargoPackages                CleanerType = "cargo"
	CleanerTypeBuildCache                   CleanerType = "buildcache"
	CleanerTypeDocker                       CleanerType = "docker"
	CleanerTypeSystemCache                  CleanerType = "systemcache"
	CleanerTypeProjectsManagementAutomation CleanerType = "projects"
	CleanerTypeCompiledBinaries             CleanerType = "compiled-binaries"
	CleanerTypeProjectExecutables           CleanerType = "project-executables"
	CleanerTypeGolangciLintCache            CleanerType = "golangci-lint-cache"
)

// CleanerAvailability represents the availability status of a cleaner.
type CleanerAvailability string

const (
	CleanerAvailabilityAvailable   CleanerAvailability = "available"
	CleanerAvailabilityUnavailable CleanerAvailability = "unavailable"
)

type cleanerMetadataEntry struct {
	RegistryName string
	DisplayName  string
	Description  string
	Icon         string
}

var cleanerMetadata = map[CleanerType]cleanerMetadataEntry{
	CleanerTypeNix: {
		RegistryName: "nix",
		DisplayName:  "Nix",
		Description:  "Clean old Nix store generations and optimize store",
		Icon:         "❄️",
	},
	CleanerTypeHomebrew: {
		RegistryName: "homebrew",
		DisplayName:  "Homebrew",
		Description:  "Clean Homebrew cache and unused packages",
		Icon:         "🍺",
	},
	CleanerTypeTempFiles: {
		RegistryName: "tempfiles",
		DisplayName:  "Temp Files",
		Description:  "Clean /tmp files (not dirs) older than 7 days",
		Icon:         "🗂️",
	},
	CleanerTypeNodePackages: {
		RegistryName: "node",
		DisplayName:  "Node.js Packages",
		Description:  "Clean npm, pnpm, yarn, bun caches",
		Icon:         "📦",
	},
	CleanerTypeGoPackages: {
		RegistryName: "go",
		DisplayName:  "Go Packages",
		Description:  "Clean Go module, test, and build caches",
		Icon:         "🐹",
	},
	CleanerTypeCargoPackages: {
		RegistryName: "cargo",
		DisplayName:  "Cargo Packages",
		Description:  "Clean Rust/Cargo registry and source caches",
		Icon:         "🦀",
	},
	CleanerTypeBuildCache: {
		RegistryName: "buildcache",
		DisplayName:  "Build Cache",
		Description:  "Clean Gradle, Maven, and SBT caches",
		Icon:         "🔨",
	},
	CleanerTypeDocker: {
		RegistryName: "docker",
		DisplayName:  "Docker",
		Description:  "Clean Docker images, containers, and volumes",
		Icon:         "🐳",
	},
	CleanerTypeSystemCache: {
		RegistryName: "systemcache",
		DisplayName:  "System Cache",
		Description:  "Clean macOS Spotlight, Xcode, CocoaPods caches",
		Icon:         "⚙️",
	},
	CleanerTypeProjectsManagementAutomation: {
		RegistryName: "projects",
		DisplayName:  "Projects Management Automation",
		Description:  "Clear projects-management-automation cache",
		Icon:         "⚙️",
	},
	CleanerTypeCompiledBinaries: {
		RegistryName: "compiled-binaries",
		DisplayName:  "Compiled Binaries",
		Description:  "Clean compiled binary files in project directories",
		Icon:         "🔧",
	},
	CleanerTypeProjectExecutables: {
		RegistryName: "project-executables",
		DisplayName:  "Project Executables",
		Description:  "Remove executable files (not scripts) from project directories",
		Icon:         "📁",
	},
	CleanerTypeGolangciLintCache: {
		RegistryName: "golangci-lint-cache",
		DisplayName:  "golangci-lint Cache",
		Description:  "Clean golangci-lint cache (uses cache status for accurate sizing)",
		Icon:         "🐹",
	},
}

var registryNameToCleanerType map[string]CleanerType

func init() {
	registryNameToCleanerType = make(map[string]CleanerType, len(cleanerMetadata))
	for ct, meta := range cleanerMetadata {
		registryNameToCleanerType[meta.RegistryName] = ct
	}
}
