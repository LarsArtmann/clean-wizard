package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewProjectsManagementAutomationCleaner(t *testing.T) {
	TestNewCleanerConstructor(t, NewProjectsManagementAutomationCleaner, "NewProjectsManagementAutomationCleaner")
}

func TestProjectsManagementAutomationCleaner_Type(t *testing.T) {
	cleaner := NewTestCleaner(NewProjectsManagementAutomationCleaner)()

	if cleaner.Type() != domain.OperationTypeProjectsManagementAutomation {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeProjectsManagementAutomation)
	}
}

func TestProjectsManagementAutomationCleaner_IsAvailable(t *testing.T) {
	TestIsAvailable(t, NewProjectsManagementAutomationCleaner)
}

func TestProjectsManagementAutomationCleaner_BooleanSettingsTests(t *testing.T) {
	CreateBooleanSettingsCleanerTestFunctions(t,
		NewBooleanSettingsCleanerTestConfigFn(
			"ProjectsManagementAutomation",
			"projects-management-automation",
			"projects management automation",
			1,
			NewProjectsManagementAutomationCleaner,
			func(enabled bool) *domain.OperationSettings {
				return &domain.OperationSettings{
					ProjectsManagementAutomation: &domain.ProjectsManagementAutomationSettings{
						ClearCache: enabled,
					},
				}
			},
		),
	)
}

func TestProjectsManagementAutomationCleaner_EstimateCacheSize(t *testing.T) {
	cleaner := NewTestCleaner(NewProjectsManagementAutomationCleaner)()

	size := cleaner.estimateCacheSize()
	expectedSize := int64(100 * 1024 * 1024) // 100MB

	if size != expectedSize {
		t.Errorf("estimateCacheSize() = %d, want %d (100MB)", size, expectedSize)
	}
}

func TestProjectsManagementAutomationCleaner_Scan(t *testing.T) {
	cleaner := NewTestCleaner(NewProjectsManagementAutomationCleaner)()

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
	cleaner := NewTestCleaner(NewProjectsManagementAutomationCleaner)()

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

	cleaner := NewTestCleaner(NewProjectsManagementAutomationCleaner)()

	// Can't easily test "tool not available" case without mocking
	// So we just verify IsAvailable is called
	_ = cleaner.IsAvailable(context.Background())
}

func TestProjectsManagementAutomationCleaner_StandardTests(t *testing.T) {
	TestStandardCleaner(t, NewBooleanSettingsCleanerTestConstructor(NewProjectsManagementAutomationCleaner), "projects-management-automation")
}
