package cleaner

import (
	"reflect"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// availableItemsTestHelper is a helper function for testing Available* functions.
// This is called by type-specific test wrappers.
func availableItemsTestHelper[T comparable](t *testing.T, expectedItems []T, availableFn func() []T, testName string) {
	items := availableFn()

	if len(items) != len(expectedItems) {
		t.Errorf("%s() returned %d items, want %d", testName, len(items), len(expectedItems))
	}

	for i, item := range items {
		if !reflect.DeepEqual(item, expectedItems[i]) {
			t.Errorf("%s()[%d] = %v, want %v", testName, i, item, expectedItems[i])
		}
	}
}

// stringTypesTestHelper is a helper function for testing string representations.
// This is called by type-specific test wrappers.
func stringTypesTestHelper[T comparable](t *testing.T, tests []struct {
	Item T
	Want string
}, toString func(T) string, testName string,
) {
	for _, tt := range tests {
		t.Run(tt.Want, func(t *testing.T) {
			got := toString(tt.Item)
			if got != tt.Want {
				t.Errorf("%s(%v) = %v, want %v", testName, tt.Item, got, tt.Want)
			}
		})
	}
}

// BooleanSettingsCleanerTestCase represents a test case for cleaners with boolean settings.
type BooleanSettingsCleanerTestCase struct {
	Name   string
	Config BooleanSettingsTestConfig
}

// TestBooleanSettingsCleaners runs boolean settings tests for all applicable cleaners.
// This consolidates duplicate test functions across cleaner test files.
func TestBooleanSettingsCleaners(t *testing.T) {
	testCases := []BooleanSettingsCleanerTestCase{
		{
			Name: "Cargo",
			Config: BooleanSettingsTestConfig{
				TestName:          "Cargo",
				ToolName:          "Cargo",
				SettingsFieldName: "cargo packages",
				ExpectedItems:     1,
				Constructor:       NewBooleanSettingsCleanerTestConstructor(NewCargoCleaner),
				CreateSettingsFunc: func(enabled bool) *domain.OperationSettings {
					cleanupMode := domain.CacheCleanupDisabled
					if enabled {
						cleanupMode = domain.CacheCleanupEnabled
					}
					return &domain.OperationSettings{
						CargoPackages: &domain.CargoPackagesSettings{
							Autoclean: cleanupMode,
						},
					}
				},
			},
		},
		{
			Name: "ProjectsManagementAutomation",
			Config: BooleanSettingsTestConfig{
				TestName:          "ProjectsManagementAutomation",
				ToolName:          "projects-management-automation",
				SettingsFieldName: "projects management automation",
				ExpectedItems:     1,
				Constructor:       NewBooleanSettingsCleanerTestConstructor(NewProjectsManagementAutomationCleaner),
				CreateSettingsFunc: func(enabled bool) *domain.OperationSettings {
					cleanupMode := domain.CacheCleanupDisabled
					if enabled {
						cleanupMode = domain.CacheCleanupEnabled
					}
					return &domain.OperationSettings{
						ProjectsManagementAutomation: &domain.ProjectsManagementAutomationSettings{
							ClearCache: cleanupMode,
						},
					}
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			CreateBooleanSettingsTest(t, tc.Config)
		})
	}
}
