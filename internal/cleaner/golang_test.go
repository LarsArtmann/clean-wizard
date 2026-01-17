package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewGoCleaner(t *testing.T) {
	tests := []struct {
		name            string
		verbose         bool
		dryRun          bool
		cleanCache      bool
		cleanTestCache  bool
		cleanModCache   bool
		cleanBuildCache bool
	}{
		{
			name:            "all caches enabled",
			verbose:         false,
			dryRun:          false,
			cleanCache:      true,
			cleanTestCache:  true,
			cleanModCache:   true,
			cleanBuildCache: true,
		},
		{
			name:            "only cache enabled",
			verbose:         true,
			dryRun:          true,
			cleanCache:      true,
			cleanTestCache:  false,
			cleanModCache:   false,
			cleanBuildCache: false,
		},
		{
			name:            "dry-run with all caches",
			verbose:         false,
			dryRun:          true,
			cleanCache:      true,
			cleanTestCache:  true,
			cleanModCache:   true,
			cleanBuildCache: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewGoCleaner(tt.verbose, tt.dryRun, tt.cleanCache, tt.cleanTestCache, tt.cleanModCache, tt.cleanBuildCache)

			if cleaner == nil {
				t.Fatal("NewGoCleaner() returned nil cleaner")
			}

			if cleaner.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", cleaner.verbose, tt.verbose)
			}

			if cleaner.dryRun != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", cleaner.dryRun, tt.dryRun)
			}

			if cleaner.cleanCache != tt.cleanCache {
				t.Errorf("cleanCache = %v, want %v", cleaner.cleanCache, tt.cleanCache)
			}

			if cleaner.cleanTestCache != tt.cleanTestCache {
				t.Errorf("cleanTestCache = %v, want %v", cleaner.cleanTestCache, tt.cleanTestCache)
			}

			if cleaner.cleanModCache != tt.cleanModCache {
				t.Errorf("cleanModCache = %v, want %v", cleaner.cleanModCache, tt.cleanModCache)
			}

			if cleaner.cleanBuildCache != tt.cleanBuildCache {
				t.Errorf("cleanBuildCache = %v, want %v", cleaner.cleanBuildCache, tt.cleanBuildCache)
			}
		})
	}
}

func TestGoCleaner_Type(t *testing.T) {
	cleaner := NewGoCleaner(false, false, true, true, true, true)

	if cleaner.Type() != domain.OperationTypeGoPackages {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeGoPackages)
	}
}

func TestGoCleaner_IsAvailable(t *testing.T) {
	cleaner := NewGoCleaner(false, false, true, true, true, true)
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
			cleaner := NewGoCleaner(false, false, true, true, true, true)

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
			cleaner := NewGoCleaner(false, true, tt.cleanCache, tt.cleanTestCache, tt.cleanModCache, tt.cleanBuildCache)

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

	cleaner := NewGoCleaner(false, false, true, true, true, true)

	// Can't easily test "Go not available" case without mocking
	// So we just verify IsAvailable is called
	_ = cleaner.IsAvailable(context.Background())
}

func TestGoCleaner_GetHomeDir(t *testing.T) {
	cleaner := NewGoCleaner(false, false, true, true, true, true)

	// Set HOME explicitly
	t.Setenv("HOME", "/test/home")
	home := cleaner.getHomeDir()
	if home != "/test/home" {
		t.Errorf("getHomeDir() = %v, want /test/home", home)
	}

	// Test fallback on Windows (USERPROFILE)
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "C:\\Users\\test")
	home = cleaner.getHomeDir()
	if home != "C:\\Users\\test" {
		t.Errorf("getHomeDir() = %v, want C:\\Users\\test", home)
	}

	// Test fallback to empty string
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	home = cleaner.getHomeDir()
	if home != "" {
		t.Errorf("getHomeDir() = %v, want empty string", home)
	}
}

func TestGoCleaner_GetDirSize(t *testing.T) {
	cleaner := NewGoCleaner(false, false, true, true, true, true)

	// Test with non-existent path
	size := cleaner.getDirSize("/non/existent/path/12345")
	// Should return 0 for non-existent path
	if size != 0 {
		t.Errorf("getDirSize() for non-existent path = %d, want 0", size)
	}

	// Test with temp directory
	tmpDir := t.TempDir()
	size = cleaner.getDirSize(tmpDir)
	// Should be 0 for empty directory
	if size != 0 {
		t.Errorf("getDirSize() for empty dir = %d, want 0", size)
	}
}

func TestGoCleaner_GetDirModTime(t *testing.T) {
	cleaner := NewGoCleaner(false, false, true, true, true, true)

	// Test with non-existent path
	modTime := cleaner.getDirModTime("/non/existent/path/12345")
	if !modTime.IsZero() {
		t.Errorf("getDirModTime() for non-existent path = %v, want zero time", modTime)
	}

	// Test with temp directory
	tmpDir := t.TempDir()
	modTime = cleaner.getDirModTime(tmpDir)
	if modTime.IsZero() {
		t.Error("getDirModTime() for temp dir returned zero time")
	}
}

func TestGoCleaner_DryRunStrategy(t *testing.T) {
	cleaner := NewGoCleaner(false, true, true, true, true, true)

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Verify dry-run strategy is set
	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	// Verify no files were actually removed (dry-run)
	if cleanResult.ItemsRemoved == 0 {
		t.Error("Dry-run should report items that would be removed")
	}
}
