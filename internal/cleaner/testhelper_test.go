package cleaner

import (
	"context"
	"reflect"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/onsi/gomega"
)

// assertValidationError is a helper for testing that ValidateSettings
// returns expected errors. Consolidates duplicate validation test patterns
// across ginkgo test files.
func assertValidationError(
	cleaner CleanerWithSettings, settings *domain.OperationSettings, expectedErrSubstring string,
) {
	err := cleaner.ValidateSettings(settings)
	gomega.Expect(err).To(gomega.HaveOccurred())
	gomega.Expect(err.Error()).To(gomega.ContainSubstring(expectedErrSubstring))
}

// availableItemsTestHelper is a helper function for testing Available* functions.
// This is called by type-specific test wrappers.
func availableItemsTestHelper[T comparable](
	t *testing.T,
	expectedItems []T,
	availableFn func() []T,
	testName string,
) {
	t.Helper()
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

// AssertNoItemsToCleanResult verifies that a cleaner returns the expected conservative
// result when there are no items to clean. This consolidates duplicate test patterns
// across Ginkgo test files.
func AssertNoItemsToCleanResult(ctx context.Context, cleaner Cleaner, setupEmptyState func()) {
	GinkgoNoItemsToCleanTest(ctx, cleaner, setupEmptyState)
}

// VerboseDryRunCleaner is an interface for cleaners that have verbose and dryRun fields.
// Used for testing common cleaner initialization patterns.
type VerboseDryRunCleaner interface {
	GetVerbose() bool
	GetDryRun() bool
}

// assertCleanerBooleanFields validates that a cleaner's verbose and dryRun fields
// match the expected values. This eliminates duplicate assertion code across cleaner test files.
//
// Usage:
//
//	if cleaner != nil {
//		assertCleanerBooleanFields(t, cleaner, tt.verbose, tt.dryRun)
//	}
func assertCleanerBooleanFields(
	t *testing.T,
	cleaner VerboseDryRunCleaner,
	wantVerbose, wantDryRun bool,
) {
	t.Helper()

	if got := cleaner.GetVerbose(); got != wantVerbose {
		t.Errorf("verbose = %v, want %v", got, wantVerbose)
	}

	if got := cleaner.GetDryRun(); got != wantDryRun {
		t.Errorf("dryRun = %v, want %v", got, wantDryRun)
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
				Constructor: NewBooleanSettingsCleanerTestConstructor(
					NewProjectsManagementAutomationCleaner,
				),
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
