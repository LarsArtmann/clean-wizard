package cleaner

import (
	"fmt"
)

// formatValidItems formats the valid items map into a readable string list.
func formatValidItems(validItems map[string]bool) string {
	items := make([]string, 0, len(validItems))
	for item := range validItems {
		items = append(items, item)
	}
	return fmt.Sprintf("%v", items)
}

// validateSettings validates that all string items in the slice are valid types.
// It returns an error if any item is not in the validItems map.
// The error message includes full context of invalid items and valid options.
func validateSettings(items []string, validItems map[string]bool, itemName, validValuesDescription string) error {
	for i, item := range items {
		if !validItems[item] {
			return fmt.Errorf(
				"invalid %s '%s' at index %d. Valid options are: %s. Input items: %v",
				itemName,
				item,
				i,
				formatValidItems(validItems),
				items,
			)
		}
	}
	return nil
}

// toStringMap converts a slice of types to a map with string keys.
func toStringMap[T ~string](items []T) map[string]bool {
	result := make(map[string]bool, len(items))
	for _, item := range items {
		result[string(item)] = true
	}
	return result
}
