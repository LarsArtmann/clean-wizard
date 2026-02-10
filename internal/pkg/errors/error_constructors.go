package errors

import (
	"time"
)

// NewError creates new CleanWizardError.
func NewError(code ErrorCode, message string) *CleanWizardError {
	return &CleanWizardError{
		Code:      code,
		Level:     LevelError,
		Message:   message,
		Details:   nil,
		Timestamp: time.Now(),
		Stack:     captureStack(),
	}
}

// NewErrorWithLevel creates new CleanWizardError with custom level.
func NewErrorWithLevel(code ErrorCode, level ErrorLevel, message string) *CleanWizardError {
	return &CleanWizardError{
		Code:      code,
		Level:     level,
		Message:   message,
		Details:   nil,
		Timestamp: time.Now(),
		Stack:     captureStack(),
	}
}

// NewErrorWithDetails creates new CleanWizardError with context details.
func NewErrorWithDetails(code ErrorCode, message string, details *ErrorDetails) *CleanWizardError {
	err := &CleanWizardError{
		Code:      code,
		Level:     LevelError,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Stack:     captureStack(),
	}

	return err
}

// ConfigValidateError creates config validation error.
func ConfigValidateError(message string) error {
	return NewErrorWithDetails(ErrConfigValidation, message, NewErrorDetails().
		WithOperation("config_validation").
		Build())
}

// ConfigLoadError creates config loading error.
func ConfigLoadError(message string) error {
	return NewErrorWithDetails(ErrConfigLoad, message, NewErrorDetails().
		WithOperation("config_load").
		Build())
}

// ConfigSaveError creates config saving error.
func ConfigSaveError(message string) error {
	return NewErrorWithDetails(ErrConfigSave, message, NewErrorDetails().
		WithOperation("config_save").
		Build())
}

// NixCommandError creates nix command error.
func NixCommandError(message string) error {
	return NewErrorWithDetails(ErrNixCommandFailed, message, NewErrorDetails().
		WithOperation("nix_command").
		Build())
}

// CleaningError creates cleaning operation error.
func CleaningError(message string) error {
	return NewErrorWithDetails(ErrCleaningFailed, message, NewErrorDetails().
		WithOperation("cleaning").
		Build())
}

// ValidationError creates validation error.
func ValidationError(field, value, expected string) error {
	return NewErrorWithDetails(ErrInvalidInput, "Validation failed", NewErrorDetails().
		WithField(field).
		WithValue(value).
		WithExpected(expected).
		WithOperation("validation").
		Build())
}
