package factory

import (
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/compiledbinaries"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/docker"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/golang"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/homebrew"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/nix"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/tempfiles"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	errorfamily "github.com/larsartmann/go-error-family"
)

func TestResolveOperationSettings_Defaults(t *testing.T) {
	t.Parallel()

	if got := resolveNixKeepCount(nil); got != nil {
		t.Errorf("resolveNixKeepCount() = %v, want nil (constructor default)", got)
	}

	if got := resolveHomebrewMode(nil); got != enums.HomebrewModeAll {
		t.Errorf("resolveHomebrewMode() = %v, want ALL", got)
	}

	if got := resolveDockerPruneMode(nil); got != enums.DockerPruneAll {
		t.Errorf("resolveDockerPruneMode() = %v, want ALL", got)
	}

	if got := resolveGoCaches(nil); got != golang.GoCacheGOCACHE|golang.GoCacheTestCache|golang.GoCacheModCache|golang.GoCacheBuildCache {
		t.Errorf("resolveGoCaches() = %v, want factory default flags", got)
	}

	if got := resolveBuildCacheOlderThan(nil); got != "30d" {
		t.Errorf("resolveBuildCacheOlderThan() = %q, want 30d", got)
	}

	olderThan, cacheTypes := resolveSystemCache(nil)
	if olderThan != "30d" || cacheTypes != nil {
		t.Errorf("resolveSystemCache() = (%q, %v), want (30d, nil)", olderThan, cacheTypes)
	}

	olderThan, excludes := resolveTempFiles(nil)
	if olderThan != "7d" || excludes != nil {
		t.Errorf("resolveTempFiles() = (%q, %v), want (7d, nil)", olderThan, excludes)
	}

	extensions, patterns := resolveProjectExecutables(nil)
	if extensions != nil || patterns != nil {
		t.Errorf("resolveProjectExecutables() = (%v, %v), want (nil, nil)", extensions, patterns)
	}

	minSizeMB, olderThan, basePaths, excludePatterns := resolveCompiledBinaries(nil)
	if minSizeMB != compiledbinaries.DefaultMinSizeMB || olderThan != compiledbinaries.DefaultOlderThan ||
		basePaths != nil || excludePatterns != nil {
		t.Errorf(
			"resolveCompiledBinaries() = (%d, %q, %v, %v), want defaults",
			minSizeMB, olderThan, basePaths, excludePatterns,
		)
	}

	if got := resolveNodePackageManagers(nil); len(got) == 0 {
		t.Error("resolveNodePackageManagers() = empty, want available package managers fallback")
	}
}

func TestResolveOperationSettings_ConfiguredSections(t *testing.T) {
	t.Parallel()

	settings := &operations.OperationSettings{
		NixGenerations: &operations.NixGenerationsSettings{
			Generations: 3,
		},
		Homebrew: &operations.HomebrewSettings{
			UnusedOnly: enums.HomebrewModeUnusedOnly,
		},
		Docker: &operations.DockerSettings{
			PruneMode: enums.DockerPruneVolumes,
		},
		GoPackages: &operations.GoPackagesSettings{
			CleanCache: enums.CacheCleanupEnabled,
		},
		NodePackages: &operations.NodePackagesSettings{
			PackageManagers: []enums.PackageManagerType{enums.PackageManagerBun},
		},
		BuildCache: &operations.BuildCacheSettings{
			OlderThan: "14d",
		},
		SystemCache: &operations.SystemCacheSettings{
			OlderThan:  "21d",
			CacheTypes: []enums.CacheType{enums.CacheTypePip},
		},
		TempFiles: &operations.TempFilesSettings{
			OlderThan: "14d",
			Excludes:  []string{"/tmp/keep"},
		},
		ProjectExecutables: &operations.ProjectExecutablesSettings{
			ExcludeExtensions: []string{".bin"},
			ExcludePatterns:   []string{"vendor/**"},
		},
		CompiledBinaries: &operations.CompiledBinariesSettings{
			MinSizeMB:       50,
			OlderThan:       "30d",
			BasePaths:       []string{"~/src"},
			ExcludePatterns: []string{"node_modules/**"},
		},
	}

	if got := resolveNixKeepCount(settings); len(got) != 1 || got[0] != 3 {
		t.Errorf("resolveNixKeepCount() = %v, want [3]", got)
	}

	if got := resolveHomebrewMode(settings); got != enums.HomebrewModeUnusedOnly {
		t.Errorf("resolveHomebrewMode() = %v, want UNUSED_ONLY", got)
	}

	if got := resolveDockerPruneMode(settings); got != enums.DockerPruneVolumes {
		t.Errorf("resolveDockerPruneMode() = %v, want VOLUMES", got)
	}

	if got := resolveGoCaches(settings); got != golang.GoCacheGOCACHE {
		t.Errorf("resolveGoCaches() = %v, want GOCACHE only", got)
	}

	managers := resolveNodePackageManagers(settings)
	if len(managers) != 1 || managers[0] != enums.PackageManagerBun {
		t.Errorf("resolveNodePackageManagers() = %v, want [bun]", managers)
	}

	if got := resolveBuildCacheOlderThan(settings); got != "14d" {
		t.Errorf("resolveBuildCacheOlderThan() = %q, want 14d", got)
	}

	olderThan, cacheTypes := resolveSystemCache(settings)
	if olderThan != "21d" || len(cacheTypes) != 1 || cacheTypes[0] != enums.CacheTypePip {
		t.Errorf("resolveSystemCache() = (%q, %v), want (21d, [PIP])", olderThan, cacheTypes)
	}

	olderThan, excludes := resolveTempFiles(settings)
	if olderThan != "14d" || len(excludes) != 1 || excludes[0] != "/tmp/keep" {
		t.Errorf("resolveTempFiles() = (%q, %v), want (14d, [/tmp/keep])", olderThan, excludes)
	}

	extensions, patterns := resolveProjectExecutables(settings)
	if len(extensions) != 1 || extensions[0] != ".bin" || len(patterns) != 1 || patterns[0] != "vendor/**" {
		t.Errorf("resolveProjectExecutables() = (%v, %v), want ([.bin], [vendor/**])", extensions, patterns)
	}

	minSizeMB, olderThan, basePaths, excludePatterns := resolveCompiledBinaries(settings)
	if minSizeMB != 50 || olderThan != "30d" ||
		len(basePaths) != 1 || basePaths[0] != "~/src" ||
		len(excludePatterns) != 1 || excludePatterns[0] != "node_modules/**" {
		t.Errorf(
			"resolveCompiledBinaries() = (%d, %q, %v, %v), want configured values",
			minSizeMB, olderThan, basePaths, excludePatterns,
		)
	}
}

func TestResolveOperationSettings_ZeroValueSemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *operations.OperationSettings
		assert   func(t *testing.T, settings *operations.OperationSettings)
	}{
		{
			name: "empty older_than falls back to default",
			settings: &operations.OperationSettings{
				TempFiles: &operations.TempFilesSettings{Excludes: []string{"/tmp/keep"}},
			},
			assert: func(t *testing.T, settings *operations.OperationSettings) {
				t.Helper()

				olderThan, excludes := resolveTempFiles(settings)
				if olderThan != "7d" {
					t.Errorf("olderThan = %q, want default 7d", olderThan)
				}

				if len(excludes) != 1 {
					t.Errorf("excludes = %v, want configured excludes kept", excludes)
				}
			},
		},
		{
			name: "fully disabled go packages fall back to default caches",
			settings: &operations.OperationSettings{
				GoPackages: &operations.GoPackagesSettings{},
			},
			assert: func(t *testing.T, settings *operations.OperationSettings) {
				t.Helper()

				got := resolveGoCaches(settings)
				if !got.IsValid() {
					t.Errorf("resolveGoCaches() = %v, want a valid default cache set", got)
				}
			},
		},
		{
			name: "nix generations of zero keep the constructor default",
			settings: &operations.OperationSettings{
				NixGenerations: &operations.NixGenerationsSettings{},
			},
			assert: func(t *testing.T, settings *operations.OperationSettings) {
				t.Helper()

				if got := resolveNixKeepCount(settings); got != nil {
					t.Errorf("resolveNixKeepCount() = %v, want nil", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.assert(t, tt.settings)
		})
	}
}

func TestDefaultRegistryWithConfig_Settings(t *testing.T) {
	t.Parallel()

	t.Run("nil settings build the full default registry", func(t *testing.T) {
		t.Parallel()

		registry, err := DefaultRegistryWithConfig(false, true, nil)
		if err != nil {
			t.Fatalf("DefaultRegistryWithConfig() error = %v", err)
		}

		if registry.Count() == 0 {
			t.Fatal("registry is empty")
		}

		c, ok := registry.Get(cleaner.CleanerHomebrew)
		if !ok {
			t.Fatal("homebrew cleaner not registered")
		}

		if got := c.(*homebrew.HomebrewCleaner).GetUnusedOnly(); got != enums.HomebrewModeAll {
			t.Errorf("homebrew mode = %v, want ALL", got)
		}
	})

	t.Run("profile settings reach the cleaner constructors", func(t *testing.T) {
		t.Parallel()

		settings := &operations.OperationSettings{
			NixGenerations: &operations.NixGenerationsSettings{Generations: 3},
			Homebrew:       &operations.HomebrewSettings{UnusedOnly: enums.HomebrewModeUnusedOnly},
			Docker:         &operations.DockerSettings{PruneMode: enums.DockerPruneVolumes},
			TempFiles:      &operations.TempFilesSettings{OlderThan: "14d"},
			GoPackages:     &operations.GoPackagesSettings{CleanCache: enums.CacheCleanupEnabled},
		}

		registry, err := DefaultRegistryWithConfig(false, true, settings)
		if err != nil {
			t.Fatalf("DefaultRegistryWithConfig() error = %v", err)
		}

		if got := mustGetCleaner[*nix.NixCleaner](t, registry, cleaner.CleanerNix).GetKeepCount(); got != 3 {
			t.Errorf("nix keepCount = %d, want 3", got)
		}

		if got := mustGetCleaner[*homebrew.HomebrewCleaner](
			t,
			registry,
			cleaner.CleanerHomebrew,
		).GetUnusedOnly(); got != enums.HomebrewModeUnusedOnly {
			t.Errorf("homebrew mode = %v, want UNUSED_ONLY", got)
		}

		if got := mustGetCleaner[*docker.DockerCleaner](
			t,
			registry,
			cleaner.CleanerDocker,
		).GetPruneMode(); got != enums.DockerPruneVolumes {
			t.Errorf("docker prune mode = %v, want VOLUMES", got)
		}

		if got := mustGetCleaner[*tempfiles.TempFilesCleaner](t, registry, cleaner.CleanerTempFiles).GetOlderThan(); got != 14*24*time.Hour {
			t.Errorf("temp files olderThan = %v, want 14d", got)
		}

		if got := mustGetCleaner[*golang.GoCleaner](t, registry, cleaner.CleanerGo).GetCaches(); got != golang.GoCacheGOCACHE {
			t.Errorf("go caches = %v, want GOCACHE only", got)
		}
	})

	t.Run("invalid settings fail registry creation", func(t *testing.T) {
		t.Parallel()

		settings := &operations.OperationSettings{
			TempFiles: &operations.TempFilesSettings{OlderThan: "not-a-duration"},
		}

		_, err := DefaultRegistryWithConfig(false, true, settings)
		if err == nil {
			t.Fatal("DefaultRegistryWithConfig() error = nil, want invalid settings rejection")
		}

		if family := errorfamily.Classify(err); family != errorfamily.Rejection {
			t.Errorf("error family = %v, want Rejection", family)
		}
	})
}

// mustGetCleaner fetches a cleaner from the registry and type-asserts it.
func mustGetCleaner[T cleaner.Cleaner](t *testing.T, registry *cleaner.Registry, name string) T {
	t.Helper()

	c, ok := registry.Get(name)
	if !ok {
		t.Fatalf("cleaner %q not registered", name)
	}

	typed, ok := c.(T)
	if !ok {
		t.Fatalf("cleaner %q has type %T, want %T", name, c, typed)
	}

	return typed
}
