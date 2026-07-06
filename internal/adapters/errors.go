package adapters

import (
	errorfamily "github.com/larsartmann/go-error-family"
)

// ErrInvalidConfig creates a configuration validation error classified as
// Rejection — the user provided invalid configuration values.
func ErrInvalidConfig(message string) error {
	return errorfamily.NewRejection("config.invalid", message)
}
