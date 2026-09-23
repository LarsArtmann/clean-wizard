package factory

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/cargo"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/projectsmanagementautomation"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
)

// TestBooleanSettingsCleaners runs boolean settings tests for all applicable cleaners.
// This consolidates duplicate test functions across cleaner test files.
func TestBooleanSettingsCleaners(t *testing.T) {
	t.Parallel()

	testCases := []cleaner.BooleanSettingsCleanerTestCase{
		{
			Name: "Cargo",
			Config: cleaner.BooleanSettingsTestConfig{
				TestName:          "Cargo",
				ToolName:          "Cargo",
				SettingsFieldName: "cargo packages",
				ExpectedItems:     2,
				Constructor:       cleaner.NewBooleanSettingsCleanerTestConstructor(cargo.NewCargoCleaner),
				CreateSettingsFunc: func(enabled bool) *operations.OperationSettings {
					cleanupMode := enums.CacheCleanupDisabled
					if enabled {
						cleanupMode = enums.CacheCleanupEnabled
					}

					return &operations.OperationSettings{
						CargoPackages: &operations.CargoPackagesSettings{
							Autoclean: cleanupMode,
						},
					}
				},
			},
		},
		{
			Name: "ProjectsManagementAutomation",
			Config: cleaner.BooleanSettingsTestConfig{
				TestName:          "ProjectsManagementAutomation",
				ToolName:          "projects-management-automation",
				SettingsFieldName: "projects management automation",
				ExpectedItems:     1,
				Constructor: cleaner.NewBooleanSettingsCleanerTestConstructor(
					projectsmanagementautomation.NewProjectsManagementAutomationCleaner,
				),
				CreateSettingsFunc: func(enabled bool) *operations.OperationSettings {
					cleanupMode := enums.CacheCleanupDisabled
					if enabled {
						cleanupMode = enums.CacheCleanupEnabled
					}

					return &operations.OperationSettings{
						ProjectsManagementAutomation: &operations.ProjectsManagementAutomationSettings{
							ClearCache: cleanupMode,
						},
					}
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			cleaner.CreateBooleanSettingsTest(t, tc.Config)
		})
	}
}
