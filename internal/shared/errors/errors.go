package errors

import (
	"fmt"
)

// ConfigLoadError creates a configuration loading error
func ConfigLoadError(err error) error {
	return fmt.Errorf("config load error: %w", err)
}

// ConfigSaveError creates a configuration saving error
func ConfigSaveError(err error) error {
	return fmt.Errorf("config save error: %w", err)
}

// ConfigValidateError creates a configuration validation error
func ConfigValidateError(message string) error {
	return fmt.Errorf("config validation error: %s", message)
}

// CleanOperationError creates a cleaning operation error
func CleanOperationError(operation string, err error) error {
	return fmt.Errorf("clean operation '%s' failed: %w", operation, err)
}

// ScanOperationError creates a scanning operation error
func ScanOperationError(scanner string, err error) error {
	return fmt.Errorf("scan operation '%s' failed: %w", scanner, err)
}

// SafetyError creates a safety check error
func SafetyError(message string) error {
	return fmt.Errorf("safety check failed: %s", message)
}

// InputValidationError creates an input validation error
func InputValidationError(field, value string) error {
	return fmt.Errorf("invalid input for field '%s': %s", field, value)
}
