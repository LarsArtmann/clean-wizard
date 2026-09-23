package cleaner

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// Factory defaults. They preserve the behavior of cleaners when a user profile
// does not configure the corresponding settings section, so they intentionally
// mirror the literals previously hardcoded in registerAllCleaners.
const (
	defaultBuildCacheOlderThan  = "30d"
	defaultSystemCacheOlderThan = "30d"
	defaultTempFilesOlderThan   = "7d"
)

// resolveNixKeepCount returns the variadic keepCount argument for NewNixCleaner.
// An empty result keeps the constructor default (5).
func resolveNixKeepCount(settings *domain.OperationSettings) []int {
	if settings == nil {
		return nil
	}

	if settings.NixGenerations != nil && settings.NixGenerations.Generations > 0 {
		return []int{settings.NixGenerations.Generations}
	}

	return nil
}

// resolveHomebrewMode maps the homebrew settings section to the cleaner's mode.
func resolveHomebrewMode(settings *domain.OperationSettings) domain.HomebrewMode {
	if settings == nil {
		return domain.HomebrewModeAll
	}

	if settings.Homebrew != nil {
		return settings.Homebrew.UnusedOnly
	}

	return domain.HomebrewModeAll
}

// resolveDockerPruneMode maps the docker settings section to the cleaner's prune mode.
func resolveDockerPruneMode(settings *domain.OperationSettings) domain.DockerPruneMode {
	if settings == nil {
		return domain.DockerPruneAll
	}

	if settings.Docker != nil {
		return settings.Docker.PruneMode
	}

	return domain.DockerPruneAll
}

// resolveGoCaches maps the go_packages settings section to Go cache flags.
// A section with every cache disabled cannot produce a valid cache set, so it
// falls back to the factory default rather than constructing an invalid cleaner.
func resolveGoCaches(settings *domain.OperationSettings) GoCacheType {
	const defaultGoCaches = GoCacheGOCACHE | GoCacheTestCache | GoCacheModCache | GoCacheBuildCache

	if settings == nil {
		return defaultGoCaches
	}

	goSettings := settings.GoPackages
	if goSettings == nil {
		return defaultGoCaches
	}

	var caches GoCacheType

	if goSettings.CleanCache.IsEnabled() {
		caches |= GoCacheGOCACHE
	}

	if goSettings.CleanTestCache.IsEnabled() {
		caches |= GoCacheTestCache
	}

	if goSettings.CleanModCache.IsEnabled() {
		caches |= GoCacheModCache
	}

	if goSettings.CleanBuildCache.IsEnabled() {
		caches |= GoCacheBuildCache
	}

	if goSettings.CleanLintCache.IsEnabled() {
		caches |= GoCacheLintCache
	}

	if !caches.IsValid() {
		return defaultGoCaches
	}

	return caches
}

// resolveNodePackageManagers maps the node_packages settings section to the
// package managers to clean, falling back to whatever is available on this system.
func resolveNodePackageManagers(settings *domain.OperationSettings) []domain.PackageManagerType {
	if settings == nil {
		return AvailableNodePackageManagers()
	}

	if settings.NodePackages != nil && len(settings.NodePackages.PackageManagers) > 0 {
		return settings.NodePackages.PackageManagers
	}

	return AvailableNodePackageManagers()
}

// resolveBuildCacheOlderThan maps the build_cache settings section to the age filter.
func resolveBuildCacheOlderThan(settings *domain.OperationSettings) string {
	if settings == nil {
		return defaultBuildCacheOlderThan
	}

	if settings.BuildCache != nil && settings.BuildCache.OlderThan != "" {
		return settings.BuildCache.OlderThan
	}

	return defaultBuildCacheOlderThan
}

// resolveSystemCache maps the system_cache settings section to age filter and cache types.
func resolveSystemCache(settings *domain.OperationSettings) (string, []domain.CacheType) {
	olderThan := defaultSystemCacheOlderThan

	var cacheTypes []domain.CacheType

	if settings == nil {
		return olderThan, cacheTypes
	}

	if systemSettings := settings.SystemCache; systemSettings != nil {
		if systemSettings.OlderThan != "" {
			olderThan = systemSettings.OlderThan
		}

		cacheTypes = systemSettings.CacheTypes
	}

	return olderThan, cacheTypes
}

// resolveTempFiles maps the temp_files settings section to age filter and excludes.
func resolveTempFiles(settings *domain.OperationSettings) (string, []string) {
	olderThan := defaultTempFilesOlderThan

	var excludes []string

	if settings == nil {
		return olderThan, excludes
	}

	if tempSettings := settings.TempFiles; tempSettings != nil {
		if tempSettings.OlderThan != "" {
			olderThan = tempSettings.OlderThan
		}

		excludes = tempSettings.Excludes
	}

	return olderThan, excludes
}

// resolveProjectExecutables maps the project_executables settings section to
// the exclusion filters. A nil extension list keeps the constructor default (.sh).
func resolveProjectExecutables(settings *domain.OperationSettings) ([]string, []string) {
	if settings == nil {
		return nil, nil
	}

	if executableSettings := settings.ProjectExecutables; executableSettings != nil {
		return executableSettings.ExcludeExtensions, executableSettings.ExcludePatterns
	}

	return nil, nil
}

// resolveCompiledBinaries maps the compiled_binaries settings section to size,
// age, and path filters.
func resolveCompiledBinaries(settings *domain.OperationSettings) (int, string, []string, []string) {
	minSizeMB := DefaultMinSizeMB
	olderThan := DefaultOlderThan

	var basePaths, excludePatterns []string

	if settings == nil {
		return minSizeMB, olderThan, basePaths, excludePatterns
	}

	if binarySettings := settings.CompiledBinaries; binarySettings != nil {
		if binarySettings.MinSizeMB > 0 {
			minSizeMB = binarySettings.MinSizeMB
		}

		if binarySettings.OlderThan != "" {
			olderThan = binarySettings.OlderThan
		}

		basePaths = binarySettings.BasePaths
		excludePatterns = binarySettings.ExcludePatterns
	}

	return minSizeMB, olderThan, basePaths, excludePatterns
}
