package cleaner

import (
	"fmt"
)

// validateSettings validates that all string items in the slice are valid types.
// It returns an error if any item is not in the validItems map.
func validateSettings(items []string, validItems map[string]bool, itemName, validValuesDescription string) error {
	for _, item := range items {
		if !validItems[item] {
			return fmt.Errorf("invalid %s: %s (must be %s)", itemName, item, validValuesDescription)
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
