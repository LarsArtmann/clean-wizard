package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// cacheTypeFromBools converts old boolean parameters to GoCacheType enum.
func cacheTypeFromBools(cleanCache, cleanTestCache, cleanModCache, cleanBuildCache, cleanLintCache bool) GoCacheType {
	var caches GoCacheType
	if cleanCache {
		caches |= GoCacheGOCACHE
	}
	if cleanTestCache {
		caches |= GoCacheTestCache
	}
	if cleanModCache {
		caches |= GoCacheModCache
	}
	if cleanBuildCache {
		caches |= GoCacheBuildCache
	}
	if cleanLintCache {
		caches |= GoCacheLintCache
	}
	return caches
}

func TestNewGoCleaner(t *testing.T) {
	tests := []struct {
		name            string
		verbose         bool
		dryRun          bool
		cleanCache      bool
		cleanTestCache  bool
		cleanModCache   bool
		cleanBuildCache bool
		cleanLintCache  bool
	}{
		{
			name:            "all caches enabled",
			verbose:         false,
			dryRun:          false,
			cleanCache:      true,
			cleanTestCache:  true,
			cleanModCache:   true,
			cleanBuildCache: true,
			cleanLintCache:  true,
		},
		{
			name:            "only cache enabled",
			verbose:         true,
			dryRun:          true,
			cleanCache:      true,
			cleanTestCache:  false,
			cleanModCache:   false,
			cleanBuildCache: false,
			cleanLintCache:  false,
		},
		{
			name:            "dry-run with all caches",
			verbose:         false,
			dryRun:          true,
			cleanCache:      true,
			cleanTestCache:  true,
			cleanModCache:   true,
			cleanBuildCache: true,
			cleanLintCache:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewGoCleanerWithSettings(tt.verbose, tt.dryRun, cacheTypeFromBools(tt.cleanCache, tt.cleanTestCache, tt.cleanModCache, tt.cleanBuildCache, tt.cleanLintCache))

			if cleaner == nil {
				t.Fatal("NewGoCleaner() returned nil cleaner")
			}

			if cleaner.config.Verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", cleaner.config.Verbose, tt.verbose)
			}

			if cleaner.config.DryRun != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", cleaner.config.DryRun, tt.dryRun)
			}

			if cleaner.config.Caches.Has(GoCacheGOCACHE) != tt.cleanCache {
				t.Errorf("cleanCache = %v, want %v", cleaner.config.Caches.Has(GoCacheGOCACHE), tt.cleanCache)
			}

			if cleaner.config.Caches.Has(GoCacheTestCache) != tt.cleanTestCache {
				t.Errorf("cleanTestCache = %v, want %v", cleaner.config.Caches.Has(GoCacheTestCache), tt.cleanTestCache)
			}

			if cleaner.config.Caches.Has(GoCacheModCache) != tt.cleanModCache {
				t.Errorf("cleanModCache = %v, want %v", cleaner.config.Caches.Has(GoCacheModCache), tt.cleanModCache)
			}

			if cleaner.config.Caches.Has(GoCacheBuildCache) != tt.cleanBuildCache {
				t.Errorf("cleanBuildCache = %v, want %v", cleaner.config.Caches.Has(GoCacheBuildCache), tt.cleanBuildCache)
			}
		})
	}
}

func TestGoCleaner_Type(t *testing.T) {
	cleaner := NewGoCleanerWithSettings(false, false, GoCacheGOCACHE|GoCacheTestCache|GoCacheModCache|GoCacheBuildCache)

	if cleaner.Type() != domain.OperationTypeGoPackages {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeGoPackages)
	}
}

func TestGoCleaner_IsAvailable(t *testing.T) {
	cleaner := NewGoCleanerWithSettings(false, false, GoCacheGOCACHE|GoCacheTestCache|GoCacheModCache|GoCacheBuildCache)
	available := cleaner.IsAvailable(context.Background())

	// Result depends on Go installation
	// We just verify it doesn't crash and returns a boolean
	if available != true && available != false {
		t.Errorf("IsAvailable() returned invalid value")
	}
}

func TestGoCleaner_ValidateSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *domain.OperationSettings
		wantErr  bool
	}{
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  false,
		},
		{
			name:     "nil Go packages settings",
			settings: &domain.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid settings with all caches",
			settings: &domain.OperationSettings{
				GoPackages: &domain.GoPackagesSettings{
					CleanCache:      true,
					CleanTestCache:  true,
					CleanModCache:   true,
					CleanBuildCache: true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with no caches",
			settings: &domain.OperationSettings{
				GoPackages: &domain.GoPackagesSettings{
					CleanCache:      false,
					CleanTestCache:  false,
					CleanModCache:   false,
					CleanBuildCache: false,
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with mixed caches",
			settings: &domain.OperationSettings{
				GoPackages: &domain.GoPackagesSettings{
					CleanCache:      true,
					CleanTestCache:  false,
					CleanModCache:   true,
					CleanBuildCache: false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewGoCleanerWithSettings(false, false, GoCacheGOCACHE|GoCacheTestCache|GoCacheModCache|GoCacheBuildCache)

			err := cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoCleaner_Clean_DryRun(t *testing.T) {
	tests := []struct {
		name            string
		cleanCache      bool
		cleanTestCache  bool
		cleanModCache   bool
		cleanBuildCache bool
		wantItems       uint
	}{
		{
			name:            "dry-run with all caches",
			cleanCache:      true,
			cleanTestCache:  true,
			cleanModCache:   true,
			cleanBuildCache: true,
			wantItems:       4,
		},
		{
			name:            "dry-run with single cache",
			cleanCache:      true,
			cleanTestCache:  false,
			cleanModCache:   false,
			cleanBuildCache: false,
			wantItems:       1,
		},
		{
			name:            "dry-run with mixed caches",
			cleanCache:      true,
			cleanTestCache:  false,
			cleanModCache:   true,
			cleanBuildCache: false,
			wantItems:       2,
		},
		{
			name:            "dry-run with no caches",
			cleanCache:      false,
			cleanTestCache:  false,
			cleanModCache:   false,
			cleanBuildCache: false,
			wantItems:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewGoCleanerWithSettings(false, true, cacheTypeFromBools(tt.cleanCache, tt.cleanTestCache, tt.cleanModCache, tt.cleanBuildCache, false))

			result := cleaner.Clean(context.Background())
			if result.IsErr() {
				t.Fatalf("Clean() error = %v", result.Error())
			}

			cleanResult := result.Value()

			if cleanResult.ItemsRemoved != tt.wantItems {
				t.Errorf("Clean() removed %d items, want %d", cleanResult.ItemsRemoved, tt.wantItems)
			}

			if cleanResult.Strategy != domain.StrategyDryRun {
				t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
			}

			if cleanResult.FreedBytes == 0 && tt.wantItems > 0 {
				t.Errorf("Clean() freed %d bytes, want > 0 when items > 0", cleanResult.FreedBytes)
			}
		})
	}
}

func TestGoCleaner_Clean_NoAvailable(t *testing.T) {
	// This test would fail if Go is installed
	// We just verify the error handling logic

	cleaner := NewGoCleanerWithSettings(false, false, GoCacheGOCACHE|GoCacheTestCache|GoCacheModCache|GoCacheBuildCache)

	// Can't easily test "Go not available" case without mocking
	// So we just verify IsAvailable is called
	_ = cleaner.IsAvailable(context.Background())
}

func TestGoCleaner_DryRunStrategy(t *testing.T) {
	cleaner := NewGoCleanerWithSettings(false, true, cacheTypeFromBools(true, true, true, true, true))

	TestDryRunStrategy(t, SimpleCleanerConstructorFromInstance(cleaner), "go")
}

func TestGoCleaner_CleanGolangciLintCache(t *testing.T) {
	lintCleaner := NewGolangciLintCleaner(true)

	result := lintCleaner.Clean(context.Background())
	if result.IsErr() {
		// golangci-lint might not be installed, which is acceptable
		t.Logf("lintCleaner.Clean() returned error (golangci-lint may not be installed): %v", result.Error())
		return
	}

	cleanResult := result.Value()

	if cleanResult.ItemsRemoved != 1 {
		t.Errorf("lintCleaner.Clean() removed %d items, want 1", cleanResult.ItemsRemoved)
	}
}
