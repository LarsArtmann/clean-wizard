package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewBuildCacheCleaner(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		dryRun    bool
		olderThan string
		excludes  []string
		basePaths []string
		wantErr   bool
	}{
		{
			name:      "valid configuration",
			verbose:   false,
			dryRun:    false,
			olderThan: "30d",
			excludes:  []string{},
			basePaths: []string{},
			wantErr:   false,
		},
		{
			name:      "verbose dry-run",
			verbose:   true,
			dryRun:    true,
			olderThan: "7d",
			excludes:  []string{"/keep"},
			basePaths: []string{"/custom/path"},
			wantErr:   false,
		},
		{
			name:      "invalid duration",
			verbose:   false,
			dryRun:    false,
			olderThan: "invalid",
			excludes:  []string{},
			basePaths: []string{},
			wantErr:   true,
		},
		{
			name:      "empty duration",
			verbose:   false,
			dryRun:    false,
			olderThan: "",
			excludes:  []string{},
			basePaths: []string{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner, err := NewBuildCacheCleaner(tt.verbose, tt.dryRun, tt.olderThan, tt.excludes, tt.basePaths)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewBuildCacheCleaner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cleaner == nil {
				t.Error("NewBuildCacheCleaner() returned nil cleaner")
			}

			if cleaner != nil {
				if cleaner.verbose != tt.verbose {
					t.Errorf("verbose = %v, want %v", cleaner.verbose, tt.verbose)
				}

				if cleaner.dryRun != tt.dryRun {
					t.Errorf("dryRun = %v, want %v", cleaner.dryRun, tt.dryRun)
				}
			}
		})
	}
}

func TestBuildCacheCleaner_Type(t *testing.T) {
	cleaner, err := NewBuildCacheCleaner(false, false, "30d", []string{}, []string{})
	if err != nil {
		t.Fatalf("NewBuildCacheCleaner() error = %v", err)
	}

	if cleaner.Type() != domain.OperationTypeBuildCache {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeBuildCache)
	}
}

func TestBuildCacheCleaner_IsAvailable(t *testing.T) {
	cleaner, err := NewBuildCacheCleaner(false, false, "30d", []string{}, []string{})
	if err != nil {
		t.Fatalf("NewBuildCacheCleaner() error = %v", err)
	}

	available := cleaner.IsAvailable(context.Background())

	// Build cache cleaner should always be available
	if !available {
		t.Error("IsAvailable() should always return true for BuildCacheCleaner")
	}
}

func TestBuildCacheCleaner_ValidateSettings(t *testing.T) {
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
			name:     "nil build cache settings",
			settings: &domain.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid settings with all tools",
			settings: &domain.OperationSettings{
				BuildCache: &domain.BuildCacheSettings{
					ToolTypes: []string{"gradle", "maven", "sbt"},
					OlderThan: "30d",
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with single tool",
			settings: &domain.OperationSettings{
				BuildCache: &domain.BuildCacheSettings{
					ToolTypes: []string{"gradle"},
					OlderThan: "7d",
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with no tools",
			settings: &domain.OperationSettings{
				BuildCache: &domain.BuildCacheSettings{
					ToolTypes: []string{},
					OlderThan: "30d",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid tool type",
			settings: &domain.OperationSettings{
				BuildCache: &domain.BuildCacheSettings{
					ToolTypes: []string{"invalid-tool"},
					OlderThan: "30d",
				},
			},
			wantErr: true,
		},
		{
			name: "mixed valid and invalid tools",
			settings: &domain.OperationSettings{
				BuildCache: &domain.BuildCacheSettings{
					ToolTypes: []string{"gradle", "invalid-tool"},
					OlderThan: "30d",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner, err := NewBuildCacheCleaner(false, false, "30d", []string{}, []string{})
			if err != nil {
				t.Fatalf("NewBuildCacheCleaner() error = %v", err)
			}

			err = cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildCacheCleaner_Clean_DryRun(t *testing.T) {
	cleaner, err := NewBuildCacheCleaner(false, true, "30d", []string{}, []string{})
	if err != nil {
		t.Fatalf("NewBuildCacheCleaner() error = %v", err)
	}

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Dry-run should report items for all tool types (3 tools)
	if cleanResult.ItemsRemoved != 3 {
		t.Errorf("Clean() removed %d items, want 3", cleanResult.ItemsRemoved)
	}

	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}

func TestBuildCacheCleaner_Scan(t *testing.T) {
	cleaner, err := NewBuildCacheCleaner(false, false, "30d", []string{}, []string{})
	if err != nil {
		t.Fatalf("NewBuildCacheCleaner() error = %v", err)
	}

	result := cleaner.Scan(context.Background())

	// Scan may not find any items if build tools aren't installed
	// Just verify it doesn't crash
	if result.IsErr() {
		t.Fatalf("Scan() error = %v", result.Error())
	}

	items := result.Value()

	// Items count depends on whether build tools are installed
	if len(items) == 0 {
		t.Log("Scan() found 0 items (build tools may not be installed)")
	}
}

func TestBuildCacheCleaner_GetHomeDir(t *testing.T) {
	_, err := NewBuildCacheCleaner(false, false, "30d", []string{}, []string{})
	if err != nil {
		t.Fatalf("NewBuildCacheCleaner() error = %v", err)
	}

	// Test GetHomeDir doesn't crash
	home, err := GetHomeDir()

	// May return empty string if home cannot be determined
	if home == "" && err == nil {
		t.Error("GetHomeDir() returned empty string and no error")
	}

	if home != "" {
		t.Logf("GetHomeDir() = %s", home)
	}
}

func TestBuildCacheCleaner_GetDirSize(t *testing.T) {
	_, err := NewBuildCacheCleaner(false, false, "30d", []string{}, []string{})
	if err != nil {
		t.Fatalf("NewBuildCacheCleaner() error = %v", err)
	}

	// Test with non-existent path
	size := GetDirSize("/non/existent/path/12345")
	// Should return 0 for non-existent path
	if size != 0 {
		t.Errorf("GetDirSize() for non-existent path = %d, want 0", size)
	}

	// Test with temp directory
	tmpDir := t.TempDir()
	size = GetDirSize(tmpDir)
	// Should be 0 for empty directory
	if size != 0 {
		t.Errorf("GetDirSize() for empty dir = %d, want 0", size)
	}
}

func TestBuildCacheCleaner_GetDirModTime(t *testing.T) {
	_, err := NewBuildCacheCleaner(false, false, "30d", []string{}, []string{})
	if err != nil {
		t.Fatalf("NewBuildCacheCleaner() error = %v", err)
	}

	// Test with non-existent path
	modTime := GetDirModTime("/non/existent/path/12345")
	if !modTime.IsZero() {
		t.Errorf("GetDirModTime() for non-existent path = %v, want zero time", modTime)
	}

	// Test with temp directory
	tmpDir := t.TempDir()
	modTime = GetDirModTime(tmpDir)
	if modTime.IsZero() {
		t.Error("GetDirModTime() for temp dir returned zero time")
	}
}

func TestAvailableBuildTools(t *testing.T) {
	expectedTools := []BuildToolType{
		BuildToolGradle,
		BuildToolMaven,
		BuildToolSBT,
	}
	availableItemsTestHelper(t, expectedTools, AvailableBuildTools, "AvailableBuildTools")
}

func TestBuildToolType_String(t *testing.T) {
	TestTypeString(t, "BuildToolType", []BuildToolType{
		BuildToolGradle,
		BuildToolMaven,
		BuildToolSBT,
	})
}

func TestBuildCacheCleaner_DryRunStrategy(t *testing.T) {
	cleaner, err := NewBuildCacheCleaner(false, true, "30d", []string{}, []string{})
	if err != nil {
		t.Fatalf("NewBuildCacheCleaner() error = %v", err)
	}

	TestDryRunStrategy(t, SimpleCleanerConstructorFromInstance(cleaner), "build-cache")
}

func TestBuildCacheCleaner_ParseDuration(t *testing.T) {
	tests := []struct {
		duration  string
		wantValid bool
	}{
		{"1h", true},
		{"24h", true},
		{"7d", true},
		{"30d", true},
		{"1w", false}, // Not supported
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.duration, func(t *testing.T) {
			cleaner, err := NewBuildCacheCleaner(false, false, tt.duration, []string{}, []string{})

			if tt.wantValid && err != nil {
				t.Errorf("NewBuildCacheCleaner() with duration %s should succeed, got error: %v", tt.duration, err)
			}

			if !tt.wantValid && err == nil {
				t.Errorf("NewBuildCacheCleaner() with duration %s should fail", tt.duration)
			}

			if cleaner != nil {
				// Verify duration was parsed correctly
				if cleaner.olderThan <= 0 {
					t.Errorf("olderThan = %v, want > 0", cleaner.olderThan)
				}
			}
		})
	}
}
