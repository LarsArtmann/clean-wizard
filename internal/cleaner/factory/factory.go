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

// CleanerConstructorFunc builds a single cleaner from the run flags and the
// resolved profile operation settings. Settings may be nil, in which case the
// cleaner falls back to its factory defaults.
type CleanerConstructorFunc func(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error)

// cleanerRegistration pairs a cleaner registry name with its constructor.
type cleanerRegistration struct {
	name        string
	constructor CleanerConstructorFunc
}

// cleanerConstructors is the canonical list of cleaners that make up the
// default registry. Both DefaultRegistryWithConfig and the DI container's
// per-cleaner providers consume this table, keeping construction logic in a
// single place.
var cleanerConstructors = []cleanerRegistration{ //nolint:gochecknoglobals
	{cleaner.CleanerNix, Nix},
	{cleaner.CleanerHomebrew, Homebrew},
	{cleaner.CleanerDocker, Docker},
	{cleaner.CleanerCargo, Cargo},
	{cleaner.CleanerGo, Go},
	{cleaner.CleanerNode, Node},
	{cleaner.CleanerBuildCache, BuildCache},
	{cleaner.CleanerSystemCache, SystemCache},
	{cleaner.CleanerTempFiles, TempFiles},
	{cleaner.CleanerProjects, Projects},
	{cleaner.CleanerProjectExec, ProjectExec},
	{cleaner.CleanerCompiledBinaries, CompiledBinaries},
	{cleaner.CleanerGolangciLint, GolangciLint},
}

// DefaultRegistryWithConfig creates a registry with cleaners configured for the
// given verbosity, dry-run, and operation settings. Settings may carry per-cleaner
// configuration from the active user profile; a nil settings value (or an absent
// section) keeps each cleaner's factory default. This is the single factory used
// by the DI container to create the cleaner registry.
// Returns an error if any cleaner fails to initialize or has invalid settings.
func DefaultRegistryWithConfig(
	verbose, dryRun bool,
	settings *operations.OperationSettings,
) (*cleaner.Registry, error) {
	registry := cleaner.NewRegistry()

	if err := registerAllCleaners(registry, verbose, dryRun, settings); err != nil {
		return nil, errorfamily.WrapRejection(err, "cleaner.registry_create", "failed to create registry with config")
	}

	return registry, nil
}

// registerAllCleaners registers all available cleaners with the given configuration.
// Returns an error if any cleaner fails to initialize.
func registerAllCleaners(
	registry *cleaner.Registry,
	verbose, dryRun bool,
	settings *operations.OperationSettings,
) error {
	for _, rc := range cleanerConstructors {
		c, err := rc.constructor(verbose, dryRun, settings)
		if err != nil {
			return errorfamily.WrapRejectionf(
				err,
				"cleaner.create",
				"cleaner=%s construction failed",
				rc.name,
			)
		}

		registry.Register(rc.name, c)
	}

	return nil
}

// validateConstructor validates a freshly constructed cleaner's settings against
// the resolved operation settings (when the cleaner supports settings).
func validateConstructor(c cleaner.Cleaner, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	if withSettings, ok := c.(cleaner.CleanerWithSettings); ok {
		if err := withSettings.ValidateSettings(settings); err != nil {
			return nil, errorfamily.WrapRejectionf(
				err,
				"cleaner.settings_invalid",
				"cleaner=%s has invalid operation settings",
				c.Name(),
			)
		}
	}

	return c, nil
}

// Nix builds the Nix generations cleaner (keep count from settings).
func Nix(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	return validateConstructor(nix.NewNixCleaner(verbose, dryRun, resolveNixKeepCount(settings)...), settings)
}

// Homebrew builds the Homebrew cleaner (cleanup mode from settings).
func Homebrew(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	return validateConstructor(homebrew.NewHomebrewCleaner(verbose, dryRun, resolveHomebrewMode(settings)), settings)
}

// Docker builds the Docker cleaner (prune mode from settings).
func Docker(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	return validateConstructor(docker.NewDockerCleaner(verbose, dryRun, resolveDockerPruneMode(settings)), settings)
}

// Cargo builds the Cargo package cleaner.
func Cargo(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	return validateConstructor(cargo.NewCargoCleaner(verbose, dryRun), settings)
}

// Go builds the Go cleaner (cache selection from settings).
func Go(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	c, err := golang.NewGoCleaner(verbose, dryRun, resolveGoCaches(settings))
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "cleaner.go_create", "failed to create Go cleaner")
	}

	return validateConstructor(c, settings)
}

// Node builds the Node packages cleaner (package managers from settings, falling
// back to all available package managers).
func Node(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	return validateConstructor(
		nodepackages.NewNodePackageManagerCleaner(verbose, dryRun, resolveNodePackageManagers(settings)),
		settings,
	)
}

// BuildCache builds the build cache cleaner (age filter from settings).
func BuildCache(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	c, err := buildcache.NewBuildCacheCleaner(
		verbose, dryRun, resolveBuildCacheOlderThan(settings), nil, nil,
	)
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "cleaner.buildcache_create", "failed to create BuildCache cleaner")
	}

	return validateConstructor(c, settings)
}

// SystemCache builds the system cache cleaner (types and age from settings).
func SystemCache(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	olderThan, cacheTypes := resolveSystemCache(settings)

	c, err := systemcache.NewSystemCacheCleaner(verbose, dryRun, olderThan, cacheTypes)
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "cleaner.systemcache_create", "failed to create SystemCache cleaner")
	}

	return validateConstructor(c, settings)
}

// TempFiles builds the temp files cleaner (age and excludes from settings;
// standard temp paths stay fixed).
func TempFiles(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	olderThan, excludes := resolveTempFiles(settings)

	c, err := tempfiles.NewTempFilesCleaner(
		verbose,
		dryRun,
		olderThan,
		excludes,
		[]string{filepath.Join("/", "tmp")},
	)
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "cleaner.tempfiles_create", "failed to create TempFiles cleaner")
	}

	return validateConstructor(c, settings)
}

// Projects builds the projects management automation cleaner.
func Projects(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	return validateConstructor(
		projectsmanagementautomation.NewProjectsManagementAutomationCleaner(verbose, dryRun),
		settings,
	)
}

// ProjectExec builds the project executables cleaner (excludes from settings).
func ProjectExec(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	excludeExtensions, excludePatterns := resolveProjectExecutables(settings)

	return validateConstructor(
		projectexecutables.NewProjectExecutablesCleaner(verbose, dryRun, excludeExtensions, excludePatterns),
		settings,
	)
}

// CompiledBinaries builds the compiled binaries cleaner (size and age filters
// from settings).
func CompiledBinaries(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	minSizeMB, olderThan, basePaths, excludePatterns := resolveCompiledBinaries(settings)

	return validateConstructor(
		compiledbinaries.NewCompiledBinariesCleaner(verbose, dryRun, minSizeMB, olderThan, basePaths, excludePatterns),
		settings,
	)
}

// GolangciLint builds the golangci-lint cache cleaner (uses `golangci-lint cache
// status` for accurate sizing).
func GolangciLint(verbose, dryRun bool, settings *operations.OperationSettings) (cleaner.Cleaner, error) {
	return validateConstructor(golangcilint.NewGolangciLintCacheCleaner(verbose, dryRun), settings)
}

// CleanerRegistration describes a single cleaner for container wiring.
type CleanerRegistration struct {
	Name        string
	Constructor CleanerConstructorFunc
}

// CleanerRegistrations returns the canonical cleaner list in registration
// order. The DI container uses it to register one provider per cleaner and to
// assemble the registry from those providers.
func CleanerRegistrations() []CleanerRegistration {
	registrations := make([]CleanerRegistration, 0, len(cleanerConstructors))
	for _, rc := range cleanerConstructors {
		registrations = append(registrations, CleanerRegistration{Name: rc.name, Constructor: rc.constructor})
	}

	return registrations
}
