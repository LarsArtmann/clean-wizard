package errors

import (
	"time"
)

// NewError creates new CleanWizardError
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

// NewErrorWithLevel creates new CleanWizardError with custom level
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

// NewErrorWithDetails creates new CleanWizardError with context details
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

// ConfigValidateError creates config validation error
func ConfigValidateError(message string) error {
	return NewErrorWithDetails(ErrConfigValidation, message, &ErrorDetails{
		Operation: "config_validation",
	})
}

// ConfigLoadError creates config loading error
func ConfigLoadError(message string) error {
	return NewErrorWithDetails(ErrConfigLoad, message, &ErrorDetails{
		Operation: "config_load",
	})
}

// ConfigSaveError creates config saving error
func ConfigSaveError(message string) error {
	return NewErrorWithDetails(ErrConfigSave, message, &ErrorDetails{
		Operation: "config_save",
	})
}

// NixCommandError creates nix command error
func NixCommandError(message string) error {
	return NewErrorWithDetails(ErrNixCommandFailed, message, &ErrorDetails{
		Operation: "nix_command",
	})
}

// CleaningError creates cleaning operation error
func CleaningError(message string) error {
	return NewErrorWithDetails(ErrCleaningFailed, message, &ErrorDetails{
		Operation: "cleaning",
	})
}

// ValidationError creates validation error
func ValidationError(field, value, expected string) error {
	return NewErrorWithDetails(ErrInvalidInput, "Validation failed", &ErrorDetails{
		Field:     field,
		Value:     value,
		Expected:  expected,
		Operation: "validation",
	})
}
