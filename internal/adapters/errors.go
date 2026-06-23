package adapters

import (
	"fmt"
)

// ErrInvalidConfig creates a configuration validation error.
func ErrInvalidConfig(message string) error {
	return fmt.Errorf("configuration error: %s", message)
}
