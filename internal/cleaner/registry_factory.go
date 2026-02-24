package cleaner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// DefaultRegistry creates a new registry with all cleaners registered.
// Cleaners are created with default settings (verbose=false, dryRun=false).
// This is useful for availability checks and discovery.
func DefaultRegistry() *Registry {
	registry := NewRegistry()
	registerAllCleaners(registry, false, false)
	return registry
}

// DefaultRegistryWithConfig creates a registry with cleaners configured for actual cleaning.
// This should be used when performing clean operations.
func DefaultRegistryWithConfig(verbose, dryRun bool) *Registry {
	registry := NewRegistry()
	registerAllCleaners(registry, verbose, dryRun)
	return registry
}

// registerAllCleaners registers all available cleaners with the given configuration.
// This helper function eliminates duplication between DefaultRegistry and DefaultRegistryWithConfig.
// Panics if any cleaner fails to initialize - this indicates a programming error in defaults.
func registerAllCleaners(registry *Registry, verbose, dryRun bool) {
	// Nix cleaner
	registry.Register(CleanerNix, NewNixCleaner(verbose, dryRun))

	// Homebrew cleaner (default mode: all)
	registry.Register(CleanerHomebrew, NewHomebrewCleaner(verbose, dryRun, domain.HomebrewModeAll))

	// Docker cleaner (default: prune all)
	registry.Register(CleanerDocker, NewDockerCleaner(verbose, dryRun, domain.DockerPruneAll))

	// Cargo cleaner
	registry.Register(CleanerCargo, NewCargoCleaner(verbose, dryRun))

	// Go cleaner (default: all cache types)
	goCleaner, err := NewGoCleaner(verbose, dryRun, GoCacheGOCACHE|GoCacheTestCache|GoCacheModCache|GoCacheBuildCache)
	if err != nil {
		panic(fmt.Sprintf("failed to create Go cleaner: %v", err))
	}
	registry.Register(CleanerGo, goCleaner)

	// Node packages cleaner (default: all available package managers)
	registry.Register(CleanerNode, NewNodePackageManagerCleaner(verbose, dryRun, AvailableNodePackageManagers()))

	// Build cache cleaner (default: 30d, all tools)
	buildCacheCleaner, err := NewBuildCacheCleaner(verbose, dryRun, "30d", []string{}, []string{})
	if err != nil {
		panic(fmt.Sprintf("failed to create BuildCache cleaner: %v", err))
	}
	registry.Register(CleanerBuildCache, buildCacheCleaner)

	// System cache cleaner (default: 30d, all cache types)
	systemCacheCleaner, err := NewSystemCacheCleaner(verbose, dryRun, "30d", nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create SystemCache cleaner: %v", err))
	}
	registry.Register(CleanerSystemCache, systemCacheCleaner)

	// Temp files cleaner (default: 7d, standard temp paths)
	tempFilesCleaner, err := NewTempFilesCleaner(verbose, dryRun, "7d", []string{}, []string{filepath.Join("/", "tmp")})
	if err != nil {
		panic(fmt.Sprintf("failed to create TempFiles cleaner: %v", err))
	}
	registry.Register(CleanerTempFiles, tempFilesCleaner)

	// Projects management automation cleaner
	registry.Register(CleanerProjects, NewProjectsManagementAutomationCleaner(verbose, dryRun))

	// Project executables cleaner
	registry.Register(CleanerProjectExec, NewProjectExecutablesCleaner(verbose, dryRun, []string{".sh"}, []string{}))

	// Compiled binaries cleaner (default: 10MB, any age, ~/projects)
	compiledBinariesCleaner := NewCompiledBinariesCleaner(
		verbose, dryRun, DefaultMinSizeMB, DefaultOlderThan, nil, []string{})
	registry.Register(CleanerCompiledBinaries, compiledBinariesCleaner)
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
