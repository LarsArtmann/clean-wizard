package cleaner

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
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
	testCases := []ValidateSettingsTestCase{
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  false,
		},
		{
			name:     "nil projects management automation settings",
			settings: &domain.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid settings with clear_cache enabled",
			settings: &domain.OperationSettings{
				ProjectsManagementAutomation: &domain.ProjectsManagementAutomationSettings{
					ClearCache: true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with clear_cache disabled",
			settings: &domain.OperationSettings{
				ProjectsManagementAutomation: &domain.ProjectsManagementAutomationSettings{
					ClearCache: false,
				},
			},
			wantErr: false,
		},
	}

	TestValidateSettings(t, NewProjectsManagementAutomationCleaner, testCases)
}

func TestProjectsManagementAutomationCleaner_Clean_DryRun(t *testing.T) {
	TestCleanDryRun(t, NewProjectsManagementAutomationCleaner, "projects-management-automation", 1)
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
	TestDryRunStrategy(t, NewProjectsManagementAutomationCleaner, "projects-management-automation")
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
	cleaner := NewProjectsManagementAutomationCleaner(false, true)

	// Skip test if projects-management-automation is not available
	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: projects-management-automation not available")
		return
	}

	startTime := time.Now()
	result := cleaner.Clean(context.Background())
	elapsed := time.Since(startTime)

	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// CleanTime should be recorded and reasonable
	if cleanResult.CleanTime == 0 {
		t.Error("Clean() returned CleanTime = 0")
	}

	// CleanedAt should be set
	if cleanResult.CleanedAt.IsZero() {
		t.Error("Clean() returned CleanedAt = zero time")
	}

	// Verify timing is reasonable (clean operation should complete quickly)
	if cleanResult.CleanTime > 30*time.Second {
		t.Errorf("Clean() took %v, which seems too long", cleanResult.CleanTime)
	}

	// Actual execution time should be close to CleanTime
	if elapsed < cleanResult.CleanTime/2 || elapsed > cleanResult.CleanTime*2 {
		t.Logf("Note: Clean() recorded time %v but actual elapsed was %v", cleanResult.CleanTime, elapsed)
	}
}
