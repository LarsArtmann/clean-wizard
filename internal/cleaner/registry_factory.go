package cleaner

import (
	"context"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// DefaultRegistry creates a new registry with all cleaners registered.
// Cleaners are created with default settings (verbose=false, dryRun=false).
// This is useful for availability checks and discovery.
func DefaultRegistry() *Registry {
	registry := NewRegistry()

	// Nix cleaner
	registry.Register("nix", NewNixCleaner(false, false))

	// Homebrew cleaner (default mode: all)
	registry.Register("homebrew", NewHomebrewCleaner(false, false, domain.HomebrewModeAll))

	// Docker cleaner (default: prune all)
	registry.Register("docker", NewDockerCleaner(false, false, domain.DockerPruneAll))

	// Cargo cleaner
	registry.Register("cargo", NewCargoCleaner(false, false))

	// Go cleaner (default: all cache types)
	goCleaner, _ := NewGoCleaner(false, false, GoCacheGOCACHE|GoCacheTestCache|GoCacheModCache|GoCacheBuildCache)
	registry.Register("go", goCleaner)

	// Node packages cleaner (default: all available package managers)
	registry.Register("node", NewNodePackageManagerCleaner(false, false, AvailableNodePackageManagers()))

	// Build cache cleaner (default: 30d, all tools)
	buildCacheCleaner, _ := NewBuildCacheCleaner(false, false, "30d", []string{}, []string{})
	registry.Register("buildcache", buildCacheCleaner)

	// System cache cleaner (default: 30d)
	systemCacheCleaner, _ := NewSystemCacheCleaner(false, false, "30d")
	registry.Register("systemcache", systemCacheCleaner)

	// Temp files cleaner (default: 7d, standard temp paths)
	tempFilesCleaner, _ := NewTempFilesCleaner(false, false, "7d", []string{}, []string{filepath.Join("/", "tmp")})
	registry.Register("tempfiles", tempFilesCleaner)

	// Language version manager cleaner (default: all available managers)
	registry.Register("langversion", NewLanguageVersionManagerCleaner(false, false, AvailableLangVersionManagers()))

	// Projects management automation cleaner
	registry.Register("projects", NewProjectsManagementAutomationCleaner(false, false))

	return registry
}

// DefaultRegistryWithConfig creates a registry with cleaners configured for actual cleaning.
// This should be used when performing clean operations.
func DefaultRegistryWithConfig(verbose, dryRun bool) *Registry {
	registry := NewRegistry()

	// Nix cleaner
	registry.Register("nix", NewNixCleaner(verbose, dryRun))

	// Homebrew cleaner (default mode: all)
	registry.Register("homebrew", NewHomebrewCleaner(verbose, dryRun, domain.HomebrewModeAll))

	// Docker cleaner (default: prune all)
	registry.Register("docker", NewDockerCleaner(verbose, dryRun, domain.DockerPruneAll))

	// Cargo cleaner
	registry.Register("cargo", NewCargoCleaner(verbose, dryRun))

	// Go cleaner (default: all cache types)
	goCleaner, _ := NewGoCleaner(verbose, dryRun, GoCacheGOCACHE|GoCacheTestCache|GoCacheModCache|GoCacheBuildCache)
	registry.Register("go", goCleaner)

	// Node packages cleaner (default: all available package managers)
	registry.Register("node", NewNodePackageManagerCleaner(verbose, dryRun, AvailableNodePackageManagers()))

	// Build cache cleaner (default: 30d, all tools)
	buildCacheCleaner, _ := NewBuildCacheCleaner(verbose, dryRun, "30d", []string{}, []string{})
	registry.Register("buildcache", buildCacheCleaner)

	// System cache cleaner (default: 30d)
	systemCacheCleaner, _ := NewSystemCacheCleaner(verbose, dryRun, "30d")
	registry.Register("systemcache", systemCacheCleaner)

	// Temp files cleaner (default: 7d, standard temp paths)
	tempFilesCleaner, _ := NewTempFilesCleaner(verbose, dryRun, "7d", []string{}, []string{filepath.Join("/", "tmp")})
	registry.Register("tempfiles", tempFilesCleaner)

	// Language version manager cleaner (default: all available managers)
	registry.Register("langversion", NewLanguageVersionManagerCleaner(verbose, dryRun, AvailableLangVersionManagers()))

	// Projects management automation cleaner
	registry.Register("projects", NewProjectsManagementAutomationCleaner(verbose, dryRun))

	return registry
}

// AvailableCleaners returns a list of available cleaner names from the default registry.
// This is a convenience function for quick availability checks.
func AvailableCleaners(ctx context.Context) []string {
	registry := DefaultRegistry()
	available := registry.Available(ctx)
	names := make([]string, 0, len(available))

	// Map cleaners back to names
	allNames := registry.Names()
	for _, name := range allNames {
		if cleaner, ok := registry.Get(name); ok && cleaner.IsAvailable(ctx) {
			names = append(names, name)
		}
	}

	return names
}
