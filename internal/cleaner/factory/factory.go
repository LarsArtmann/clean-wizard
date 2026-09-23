package factory

import (
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/buildcache"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/cargo"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/compiledbinaries"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/docker"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/golang"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/golangcilint"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/homebrew"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/nix"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/nodepackages"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/projectexecutables"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/projectsmanagementautomation"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/systemcache"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/tempfiles"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	errorfamily "github.com/larsartmann/go-error-family"
)

// DefaultRegistryWithConfig creates a registry with cleaners configured for the
// given verbosity, dry-run, and operation settings. Settings may carry per-cleaner
// configuration from the active user profile; a nil settings value (or an absent
// section) keeps each cleaner's factory default. This is the single factory used
// by the DI container to create the cleaner registry.
// Returns an error if any cleaner fails to initialize or has invalid settings.
func DefaultRegistryWithConfig(verbose, dryRun bool, settings *operations.OperationSettings) (*cleaner.Registry, error) {
	registry := cleaner.NewRegistry()

	err := registerAllCleaners(registry, verbose, dryRun, settings)
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "cleaner.registry_create", "failed to create registry with config")
	}

	return registry, nil
}

// registerAllCleaners registers all available cleaners with the given configuration.
// This helper function eliminates duplication between DefaultRegistry and DefaultRegistryWithConfig.
// Returns an error if any cleaner fails to initialize.
func registerAllCleaners(registry *cleaner.Registry, verbose, dryRun bool, settings *operations.OperationSettings) error {
	// Nix cleaner (keep count from settings; constructor default 5 when unset)
	nixKeepCount := resolveNixKeepCount(settings)
	if err := registerValidated(
		registry,
		cleaner.CleanerNix,
		nix.NewNixCleaner(verbose, dryRun, nixKeepCount...),
		settings,
	); err != nil {
		return err
	}

	// Homebrew cleaner
	homebrewMode := resolveHomebrewMode(settings)
	if err := registerValidated(
		registry,
		cleaner.CleanerHomebrew,
		homebrew.NewHomebrewCleaner(verbose, dryRun, homebrewMode),
		settings,
	); err != nil {
		return err
	}

	// Docker cleaner
	dockerPruneMode := resolveDockerPruneMode(settings)
	if err := registerValidated(
		registry,
		cleaner.CleanerDocker,
		docker.NewDockerCleaner(verbose, dryRun, dockerPruneMode),
		settings,
	); err != nil {
		return err
	}

	// Cargo cleaner
	if err := registerValidated(registry, cleaner.CleanerCargo, cargo.NewCargoCleaner(verbose, dryRun), settings); err != nil {
		return err
	}

	// Go cleaner (cache selection from settings)
	goCleaner, err := golang.NewGoCleaner(verbose, dryRun, resolveGoCaches(settings))
	if err != nil {
		return errorfamily.WrapRejection(err, "cleaner.go_create", "failed to create Go cleaner")
	}

	if err := registerValidated(registry, cleaner.CleanerGo, goCleaner, settings); err != nil {
		return err
	}

	// Node packages cleaner (falls back to all available package managers)
	if err := registerValidated(
		registry,
		cleaner.CleanerNode,
		nodepackages.NewNodePackageManagerCleaner(verbose, dryRun, resolveNodePackageManagers(settings)),
		settings,
	); err != nil {
		return err
	}

	// Build cache cleaner
	buildCacheCleaner, err := buildcache.NewBuildCacheCleaner(verbose, dryRun, resolveBuildCacheOlderThan(settings), nil, nil)
	if err != nil {
		return errorfamily.WrapRejection(err, "cleaner.buildcache_create", "failed to create BuildCache cleaner")
	}

	if err := registerValidated(registry, cleaner.CleanerBuildCache, buildCacheCleaner, settings); err != nil {
		return err
	}

	// System cache cleaner
	systemCacheOlderThan, systemCacheTypes := resolveSystemCache(settings)

	systemCacheCleaner, err := systemcache.NewSystemCacheCleaner(verbose, dryRun, systemCacheOlderThan, systemCacheTypes)
	if err != nil {
		return errorfamily.WrapRejection(err, "cleaner.systemcache_create", "failed to create SystemCache cleaner")
	}

	if err := registerValidated(registry, cleaner.CleanerSystemCache, systemCacheCleaner, settings); err != nil {
		return err
	}

	// Temp files cleaner (standard temp paths stay fixed; settings control age and excludes)
	tempFilesOlderThan, tempFilesExcludes := resolveTempFiles(settings)

	tempFilesCleaner, err := tempfiles.NewTempFilesCleaner(
		verbose,
		dryRun,
		tempFilesOlderThan,
		tempFilesExcludes,
		[]string{filepath.Join("/", "tmp")},
	)
	if err != nil {
		return errorfamily.WrapRejection(err, "cleaner.tempfiles_create", "failed to create TempFiles cleaner")
	}

	if err := registerValidated(registry, cleaner.CleanerTempFiles, tempFilesCleaner, settings); err != nil {
		return err
	}

	// Projects management automation cleaner
	if err := registerValidated(
		registry,
		cleaner.CleanerProjects,
		projectsmanagementautomation.NewProjectsManagementAutomationCleaner(verbose, dryRun),
		settings,
	); err != nil {
		return err
	}

	// Project executables cleaner
	projectExecExcludeExtensions, projectExecExcludePatterns := resolveProjectExecutables(settings)
	if err := registerValidated(
		registry,
		cleaner.CleanerProjectExec,
		projectexecutables.NewProjectExecutablesCleaner(verbose, dryRun, projectExecExcludeExtensions, projectExecExcludePatterns),
		settings,
	); err != nil {
		return err
	}

	// Compiled binaries cleaner
	compiledMinSizeMB, compiledOlderThan, compiledBasePaths, compiledExcludePatterns := resolveCompiledBinaries(
		settings,
	)
	compiledBinariesCleaner := compiledbinaries.NewCompiledBinariesCleaner(
		verbose, dryRun, compiledMinSizeMB, compiledOlderThan, compiledBasePaths, compiledExcludePatterns,
	)

	if err := registerValidated(registry, cleaner.CleanerCompiledBinaries, compiledBinariesCleaner, settings); err != nil {
		return err
	}

	// golangci-lint cache cleaner (uses `golangci-lint cache status` for accurate sizing)
	if err := registerValidated(
		registry,
		cleaner.CleanerGolangciLint,
		golangcilint.NewGolangciLintCacheCleaner(verbose, dryRun),
		settings,
	); err != nil {
		return err
	}

	return nil
}

// registerValidated validates the cleaner's settings against the resolved
// operation settings (when the cleaner supports settings), then registers it.
func registerValidated(registry *cleaner.Registry, name string, c cleaner.Cleaner, settings *operations.OperationSettings) error {
	if withSettings, ok := c.(cleaner.CleanerWithSettings); ok {
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
