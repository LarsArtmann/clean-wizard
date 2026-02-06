package cleaner

import (
	"reflect"
	"testing"
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
}, toString func(T) string, testName string) {
	for _, tt := range tests {
		t.Run(tt.Want, func(t *testing.T) {
			got := toString(tt.Item)
			if got != tt.Want {
				t.Errorf("%s(%v) = %v, want %v", testName, tt.Item, got, tt.Want)
			}
		})
	}
}

// testNewCleanerConstructor is a helper function that tests cleaner constructors
// with different combinations of verbose and dryRun parameters.
// This eliminates duplicate test code across multiple cleaner test files.
func testNewCleanerConstructor(
	t *testing.T,
	constructor func(bool, bool) interface {
		// We need to access the verbose and dryRun fields
		// Since they're not part of an interface, we use reflection
	},
	cleanerName string,
) {
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
			cleaner := constructor(tt.verbose, tt.dryRun)

			if cleaner == nil {
				t.Fatalf("%s(%v, %v) returned nil cleaner", cleanerName, tt.verbose, tt.dryRun)
			}

			// Use reflection to access the verbose and dryRun fields
			cleanerValue := reflect.ValueOf(cleaner).Elem()

			verboseField := cleanerValue.FieldByName("verbose")
			if !verboseField.IsValid() {
				t.Fatalf("%s cleaner does not have 'verbose' field", cleanerName)
			}
			if verboseField.Bool() != tt.verbose {
				t.Errorf("verbose = %v, want %v", verboseField.Bool(), tt.verbose)
			}

			dryRunField := cleanerValue.FieldByName("dryRun")
			if !dryRunField.IsValid() {
				t.Fatalf("%s cleaner does not have 'dryRun' field", cleanerName)
			}
			if dryRunField.Bool() != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", dryRunField.Bool(), tt.dryRun)
			}
		})
	}
}