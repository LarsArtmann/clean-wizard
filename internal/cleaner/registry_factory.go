package cleaner

import (
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	errorfamily "github.com/larsartmann/go-error-family"
)

// DefaultRegistryWithConfig creates a registry with cleaners configured for the
// given verbosity, dry-run, and operation settings. Settings may carry per-cleaner
// configuration from the active user profile; a nil settings value (or an absent
// section) keeps each cleaner's factory default. This is the single factory used
// by the DI container to create the cleaner registry.
// Returns an error if any cleaner fails to initialize or has invalid settings.
func DefaultRegistryWithConfig(verbose, dryRun bool, settings *domain.OperationSettings) (*Registry, error) {
	registry := NewRegistry()

	err := registerAllCleaners(registry, verbose, dryRun, settings)
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "cleaner.registry_create", "failed to create registry with config")
	}

	return registry, nil
}

// registerAllCleaners registers all available cleaners with the given configuration.
// This helper function eliminates duplication between DefaultRegistry and DefaultRegistryWithConfig.
// Returns an error if any cleaner fails to initialize.
func registerAllCleaners(registry *Registry, verbose, dryRun bool, settings *domain.OperationSettings) error {
	// Nix cleaner (keep count from settings; constructor default 5 when unset)
	nixKeepCount := resolveNixKeepCount(settings)
	if err := registerValidated(
		registry,
		CleanerNix,
		NewNixCleaner(verbose, dryRun, nixKeepCount...),
		settings,
	); err != nil {
		return err
	}

	// Homebrew cleaner
	homebrewMode := resolveHomebrewMode(settings)
	if err := registerValidated(
		registry,
		CleanerHomebrew,
		NewHomebrewCleaner(verbose, dryRun, homebrewMode),
		settings,
	); err != nil {
		return err
	}

	// Docker cleaner
	dockerPruneMode := resolveDockerPruneMode(settings)
	if err := registerValidated(
		registry,
		CleanerDocker,
		NewDockerCleaner(verbose, dryRun, dockerPruneMode),
		settings,
	); err != nil {
		return err
	}

	// Cargo cleaner
	if err := registerValidated(registry, CleanerCargo, NewCargoCleaner(verbose, dryRun), settings); err != nil {
		return err
	}

	// Go cleaner (cache selection from settings)
	goCleaner, err := NewGoCleaner(verbose, dryRun, resolveGoCaches(settings))
	if err != nil {
		return errorfamily.WrapRejection(err, "cleaner.go_create", "failed to create Go cleaner")
	}

	if err := registerValidated(registry, CleanerGo, goCleaner, settings); err != nil {
		return err
	}

	// Node packages cleaner (falls back to all available package managers)
	if err := registerValidated(
		registry,
		CleanerNode,
		NewNodePackageManagerCleaner(verbose, dryRun, resolveNodePackageManagers(settings)),
		settings,
	); err != nil {
		return err
	}

	// Build cache cleaner
	buildCacheCleaner, err := NewBuildCacheCleaner(verbose, dryRun, resolveBuildCacheOlderThan(settings), nil, nil)
	if err != nil {
		return errorfamily.WrapRejection(err, "cleaner.buildcache_create", "failed to create BuildCache cleaner")
	}

	if err := registerValidated(registry, CleanerBuildCache, buildCacheCleaner, settings); err != nil {
		return err
	}

	// System cache cleaner
	systemCacheOlderThan, systemCacheTypes := resolveSystemCache(settings)

	systemCacheCleaner, err := NewSystemCacheCleaner(verbose, dryRun, systemCacheOlderThan, systemCacheTypes)
	if err != nil {
		return errorfamily.WrapRejection(err, "cleaner.systemcache_create", "failed to create SystemCache cleaner")
	}

	if err := registerValidated(registry, CleanerSystemCache, systemCacheCleaner, settings); err != nil {
		return err
	}

	// Temp files cleaner (standard temp paths stay fixed; settings control age and excludes)
	tempFilesOlderThan, tempFilesExcludes := resolveTempFiles(settings)

	tempFilesCleaner, err := NewTempFilesCleaner(
		verbose,
		dryRun,
		tempFilesOlderThan,
		tempFilesExcludes,
		[]string{filepath.Join("/", "tmp")},
	)
	if err != nil {
		return errorfamily.WrapRejection(err, "cleaner.tempfiles_create", "failed to create TempFiles cleaner")
	}

	if err := registerValidated(registry, CleanerTempFiles, tempFilesCleaner, settings); err != nil {
		return err
	}

	// Projects management automation cleaner
	if err := registerValidated(
		registry,
		CleanerProjects,
		NewProjectsManagementAutomationCleaner(verbose, dryRun),
		settings,
	); err != nil {
		return err
	}

	// Project executables cleaner
	projectExecExcludeExtensions, projectExecExcludePatterns := resolveProjectExecutables(settings)
	if err := registerValidated(
		registry,
		CleanerProjectExec,
		NewProjectExecutablesCleaner(verbose, dryRun, projectExecExcludeExtensions, projectExecExcludePatterns),
		settings,
	); err != nil {
		return err
	}

	// Compiled binaries cleaner
	compiledMinSizeMB, compiledOlderThan, compiledBasePaths, compiledExcludePatterns := resolveCompiledBinaries(
		settings,
	)
	compiledBinariesCleaner := NewCompiledBinariesCleaner(
		verbose, dryRun, compiledMinSizeMB, compiledOlderThan, compiledBasePaths, compiledExcludePatterns,
	)

	if err := registerValidated(registry, CleanerCompiledBinaries, compiledBinariesCleaner, settings); err != nil {
		return err
	}

	// golangci-lint cache cleaner (uses `golangci-lint cache status` for accurate sizing)
	if err := registerValidated(
		registry,
		CleanerGolangciLint,
		NewGolangciLintCacheCleaner(verbose, dryRun),
		settings,
	); err != nil {
		return err
	}

	return nil
}

// registerValidated validates the cleaner's settings against the resolved
// operation settings (when the cleaner supports settings), then registers it.
func registerValidated(registry *Registry, name string, c Cleaner, settings *domain.OperationSettings) error {
	if withSettings, ok := c.(CleanerWithSettings); ok {
		if err := withSettings.ValidateSettings(settings); err != nil {
			return errorfamily.WrapRejectionf(
				err,
				"cleaner.settings_invalid",
				"cleaner=%s has invalid operation settings",
				name,
			)
		}
	}

	registry.Register(name, c)

	return nil
}
