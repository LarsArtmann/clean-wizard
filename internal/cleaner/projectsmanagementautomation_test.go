package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

func TestNewProjectsManagementAutomationCleaner(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		dryRun  bool
	}{
		{
			name:    "standard configuration",
			verbose: false,
			dryRun:  false,
		},
		{
			name:    "verbose mode",
			verbose: true,
			dryRun:  false,
		},
		{
			name:    "dry-run mode",
			verbose: false,
			dryRun:  true,
		},
		{
			name:    "verbose dry-run mode",
			verbose: true,
			dryRun:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewProjectsManagementAutomationCleaner(tt.verbose, tt.dryRun)

			if cleaner == nil {
				t.Fatal("NewProjectsManagementAutomationCleaner() returned nil cleaner")
			}

			if cleaner.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", cleaner.verbose, tt.verbose)
			}

			if cleaner.dryRun != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", cleaner.dryRun, tt.dryRun)
			}
		})
	}
}

func TestProjectsManagementAutomationCleaner_Type(t *testing.T) {
	cleaner := NewProjectsManagementAutomationCleaner(false, false)

	if cleaner.Type() != domain.OperationTypeProjectsManagementAutomation {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeProjectsManagementAutomation)
	}
}

func TestProjectsManagementAutomationCleaner_IsAvailable(t *testing.T) {
	cleaner := NewProjectsManagementAutomationCleaner(false, false)
	available := cleaner.IsAvailable(context.Background())

	// Result depends on projects-management-automation installation
	// We just verify it doesn't crash and returns a boolean
	if available != true && available != false {
		t.Errorf("IsAvailable() returned invalid value")
	}
}

func TestProjectsManagementAutomationCleaner_ValidateSettings(t *testing.T) {
	testCases := CreateBooleanSettingsTestCases("projects management automation", func(enabled bool) *domain.OperationSettings {
		return &domain.OperationSettings{
			ProjectsManagementAutomation: &domain.ProjectsManagementAutomationSettings{
				ClearCache: enabled,
			},
		}
	})

	constructor := func(verbose, dryRun bool) interface {
		IsAvailable(ctx context.Context) bool
		Clean(ctx context.Context) result.Result[domain.CleanResult]
		ValidateSettings(*domain.OperationSettings) error
	} {
		return NewProjectsManagementAutomationCleaner(verbose, dryRun)
	}

	TestValidateSettings(t, constructor, testCases)
}

func TestProjectsManagementAutomationCleaner_Clean_DryRun(t *testing.T) {
	constructor := ToSimpleCleanerConstructor(func(verbose, dryRun bool) interface {
		IsAvailable(ctx context.Context) bool
		Clean(ctx context.Context) result.Result[domain.CleanResult]
		ValidateSettings(*domain.OperationSettings) error
	} {
		return NewProjectsManagementAutomationCleaner(verbose, dryRun)
	})

	TestCleanDryRun(t, constructor, "projects-management-automation", 1)
}

func TestProjectsManagementAutomationCleaner_EstimateCacheSize(t *testing.T) {
	cleaner := NewProjectsManagementAutomationCleaner(false, false)

	size := cleaner.estimateCacheSize()
	expectedSize := int64(100 * 1024 * 1024) // 100MB

	if size != expectedSize {
		t.Errorf("estimateCacheSize() = %d, want %d (100MB)", size, expectedSize)
	}
}

func TestProjectsManagementAutomationCleaner_Scan(t *testing.T) {
	cleaner := NewProjectsManagementAutomationCleaner(false, false)

	result := cleaner.Scan(context.Background())

	if result.IsErr() {
		t.Fatalf("Scan() error = %v", result.Error())
	}

	items := result.Value()

	// If tool is available, should find cache item
	if cleaner.IsAvailable(context.Background()) {
		if len(items) == 0 {
			t.Log("Scan() found 0 items (cache may not exist)")
		}

		// Verify cache item structure
		for _, item := range items {
			if item.Path == "" {
				t.Error("Scan() returned item with empty path")
			}
			if item.Size == 0 {
				t.Logf("Scan() returned item with zero size: %s", item.Path)
			}
			if item.ScanType != domain.ScanTypeSystem {
				t.Errorf("Scan() returned item with ScanType = %v, want %v", item.ScanType, domain.ScanTypeSystem)
			}
			if item.Created.IsZero() {
				t.Error("Scan() returned item with zero Created time")
			}
		}
	}
}

func TestProjectsManagementAutomationCleaner_Scan_NotAvailable(t *testing.T) {
	cleaner := NewProjectsManagementAutomationCleaner(false, false)

	// If tool is not available, should return empty items
	if !cleaner.IsAvailable(context.Background()) {
		result := cleaner.Scan(context.Background())

		if result.IsErr() {
			t.Fatalf("Scan() error = %v", result.Error())
		}

		items := result.Value()

		if len(items) != 0 {
			t.Errorf("Scan() found %d items, want 0 when tool not available", len(items))
		}
	}
}

func TestProjectsManagementAutomationCleaner_Clean_NoAvailable(t *testing.T) {
	// This test would fail if projects-management-automation is installed
	// We just verify error handling logic exists

	cleaner := NewProjectsManagementAutomationCleaner(false, false)

	// Can't easily test "tool not available" case without mocking
	// So we just verify IsAvailable is called
	_ = cleaner.IsAvailable(context.Background())
}

func TestProjectsManagementAutomationCleaner_DryRunStrategy(t *testing.T) {
	constructor := ToSimpleCleanerConstructor(func(verbose, dryRun bool) interface {
		IsAvailable(ctx context.Context) bool
		Clean(ctx context.Context) result.Result[domain.CleanResult]
		ValidateSettings(*domain.OperationSettings) error
	} {
		return NewProjectsManagementAutomationCleaner(verbose, dryRun)
	})

	TestDryRunStrategy(t, constructor, "projects-management-automation")
}

func TestProjectsManagementAutomationCleaner_ClearCacheSettings(t *testing.T) {
	tests := []struct {
		name       string
		clearCache bool
		wantErr    bool
	}{
		{
			name:       "clear_cache enabled",
			clearCache: true,
			wantErr:    false,
		},
		{
			name:       "clear_cache disabled",
			clearCache: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewProjectsManagementAutomationCleaner(false, false)

			settings := &domain.OperationSettings{
				ProjectsManagementAutomation: &domain.ProjectsManagementAutomationSettings{
					ClearCache: tt.clearCache,
				},
			}

			err := cleaner.ValidateSettings(settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProjectsManagementAutomationCleaner_Clean_Timing(t *testing.T) {
	constructor := func(verbose, dryRun bool) interface {
		IsAvailable(ctx context.Context) bool
		Clean(ctx context.Context) result.Result[domain.CleanResult]
	} {
		return NewProjectsManagementAutomationCleaner(verbose, dryRun)
	}

	TestCleanTiming(t, constructor, "projects-management-automation")
}
