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
	CleanerTypeLangVersionMgr               CleanerType = "langversion"
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

// registryNameToCleanerType maps registry cleaner names to CleanerType.
var registryNameToCleanerType = map[string]CleanerType{
	"nix":                 CleanerTypeNix,
	"homebrew":            CleanerTypeHomebrew,
	"tempfiles":           CleanerTypeTempFiles,
	"node":                CleanerTypeNodePackages,
	"go":                  CleanerTypeGoPackages,
	"cargo":               CleanerTypeCargoPackages,
	"buildcache":          CleanerTypeBuildCache,
	"docker":              CleanerTypeDocker,
	"systemcache":         CleanerTypeSystemCache,
	"langversion":         CleanerTypeLangVersionMgr,
	"projects":            CleanerTypeProjectsManagementAutomation,
	"compiled-binaries":   CleanerTypeCompiledBinaries,
	"project-executables": CleanerTypeProjectExecutables,
	"golangci-lint-cache": CleanerTypeGolangciLintCache,
}
