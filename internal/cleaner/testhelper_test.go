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