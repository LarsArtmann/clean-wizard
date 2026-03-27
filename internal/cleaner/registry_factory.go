package cleaner

import (
	"context"
	"path/filepath"

	"github.com/cockroachdb/errors"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// DefaultRegistry creates a new registry with all cleaners registered.
// Cleaners are created with default settings (verbose=false, dryRun=false).
// This is useful for availability checks and discovery.
// Returns an error if any cleaner fails to initialize.
func DefaultRegistry() (*Registry, error) {
	registry := NewRegistry()
	if err := registerAllCleaners(registry, false, false); err != nil {
		return nil, errors.Wrap(err, "failed to create default registry")
	}

	return registry, nil
}

// DefaultRegistryWithConfig creates a registry with cleaners configured for actual cleaning.
// This should be used when performing clean operations.
// Returns an error if any cleaner fails to initialize.
func DefaultRegistryWithConfig(verbose, dryRun bool) (*Registry, error) {
	registry := NewRegistry()
	if err := registerAllCleaners(registry, verbose, dryRun); err != nil {
		return nil, errors.Wrap(err, "failed to create registry with config")
	}

	return registry, nil
}

// registerAllCleaners registers all available cleaners with the given configuration.
// This helper function eliminates duplication between DefaultRegistry and DefaultRegistryWithConfig.
// Returns an error if any cleaner fails to initialize.
func registerAllCleaners(registry *Registry, verbose, dryRun bool) error {
	// Nix cleaner
	registry.Register(CleanerNix, NewNixCleaner(verbose, dryRun))

	// Homebrew cleaner (default mode: all)
	registry.Register(CleanerHomebrew, NewHomebrewCleaner(verbose, dryRun, domain.HomebrewModeAll))

	// Docker cleaner (default: prune all)
	registry.Register(CleanerDocker, NewDockerCleaner(verbose, dryRun, domain.DockerPruneAll))

	// Cargo cleaner
	registry.Register(CleanerCargo, NewCargoCleaner(verbose, dryRun))

	// Go cleaner (default: all cache types)
	goCleaner, err := NewGoCleaner(
		verbose,
		dryRun,
		GoCacheGOCACHE|GoCacheTestCache|GoCacheModCache|GoCacheBuildCache,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create Go cleaner")
	}

	registry.Register(CleanerGo, goCleaner)

	// Node packages cleaner (default: all available package managers)
	registry.Register(
		CleanerNode,
		NewNodePackageManagerCleaner(verbose, dryRun, AvailableNodePackageManagers()),
	)

	// Build cache cleaner (default: 30d, all tools)
	buildCacheCleaner, err := NewBuildCacheCleaner(verbose, dryRun, "30d", []string{}, []string{})
	if err != nil {
		return errors.Wrap(err, "failed to create BuildCache cleaner")
	}

	registry.Register(CleanerBuildCache, buildCacheCleaner)

	// System cache cleaner (default: 30d, all cache types)
	systemCacheCleaner, err := NewSystemCacheCleaner(verbose, dryRun, "30d", nil)
	if err != nil {
		return errors.Wrap(err, "failed to create SystemCache cleaner")
	}

	registry.Register(CleanerSystemCache, systemCacheCleaner)

	// Temp files cleaner (default: 7d, standard temp paths)
	tempFilesCleaner, err := NewTempFilesCleaner(
		verbose,
		dryRun,
		"7d",
		[]string{},
		[]string{filepath.Join("/", "tmp")},
	)
	if err != nil {
		return errors.Wrap(err, "failed to create TempFiles cleaner")
	}

	registry.Register(CleanerTempFiles, tempFilesCleaner)

	// Projects management automation cleaner
	registry.Register(CleanerProjects, NewProjectsManagementAutomationCleaner(verbose, dryRun))

	// Project executables cleaner
	registry.Register(
		CleanerProjectExec,
		NewProjectExecutablesCleaner(verbose, dryRun, []string{".sh"}, []string{}),
	)

	// Compiled binaries cleaner (default: 10MB, any age, ~/projects)
	compiledBinariesCleaner := NewCompiledBinariesCleaner(
		verbose, dryRun, DefaultMinSizeMB, DefaultOlderThan, nil, []string{})
	registry.Register(CleanerCompiledBinaries, compiledBinariesCleaner)

	return nil
}

// AvailableCleaners returns a list of available cleaner names from the default registry.
// This is a convenience function for quick availability checks.
// Returns an error if the registry cannot be created.
func AvailableCleaners(ctx context.Context) ([]string, error) {
	registry, err := DefaultRegistry()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get available cleaners")
	}
	available := registry.Available(ctx)
	names := make([]string, 0, len(available))

	// Map cleaners back to names
	allNames := registry.Names()
	for _, name := range allNames {
		if cleaner, ok := registry.Get(name); ok && cleaner.IsAvailable(ctx) {
			names = append(names, name)
		}
	}

	return names, nil
}
