package errors

import (
	"fmt"
	"os/exec"
)

// HandleCommandError standardizes command execution errors
func HandleCommandError(cmd *exec.Cmd, err error) *CleanWizardError {
	if err == nil {
		return nil
	}

	// Create base error
	baseErr := NewError(ErrNixCommandFailed, fmt.Sprintf("Command failed: %s", err.Error()))

	// Add command context
	baseErr = baseErr.
		WithOperation(fmt.Sprintf("exec: %s", cmd.String())).
		WithDetail("command", cmd.Args).
		WithDetail("path", cmd.Path)

	// Add specific error details based on error type
	if exitErr, ok := err.(*exec.ExitError); ok {
		baseErr = baseErr.
			WithDetail("exit_code", exitErr.ExitCode).
			WithDetail("signal", exitErr.ProcessState.String())
	}

	return baseErr
}

// HandleNixNotAvailable standardizes Nix availability errors
func HandleNixNotAvailable(operation string) *CleanWizardError {
	return NewErrorWithLevel(ErrNixNotAvailable, LevelWarn,
		"Nix package manager is not available on this system").
		WithOperation(operation).
		WithDetail("suggestion", "Please install Nix or use mock mode for testing").
		WithDetail("documentation", "https://nixos.org/download.html")
}

// HandleConfigError standardizes configuration errors
func HandleConfigError(operation string, err error) *CleanWizardError {
	baseErr := NewError(ErrConfigLoad, fmt.Sprintf("Configuration error: %s", err.Error()))
	baseErr.Operation = operation
	return baseErr
}

// HandleValidationError standardizes validation errors
func HandleValidationError(operation string, err error) *CleanWizardError {
	baseErr := NewError(ErrConfigValidation, fmt.Sprintf("Validation error: %s", err.Error()))
	baseErr.Operation = operation
	baseErr = baseErr.WithDetail("validation_type", "comprehensive")
	return baseErr
}

// HandleValidationErrorWithDetails standardizes validation errors with detailed context
func HandleValidationErrorWithDetails(operation, field string, value any, reason string) *CleanWizardError {
	cleanErr := NewErrorWithDetails(ErrConfigValidation,
		fmt.Sprintf("Validation failed for %s: %s", field, reason),
		&ErrorDetails{
			Operation: operation,
			Field:     field,
			Value:     fmt.Sprintf("%v", value),
			Metadata: map[string]string{
				"reason": reason,
			},
		})
	// Set root-level Operation field for grouping/logging
	cleanErr.Operation = operation
	return cleanErr
}

// WrapError wraps existing error with CleanWizardError context
func WrapError(err error, code ErrorCode, operation string) *CleanWizardError {
	if err == nil {
		return nil
	}

	cleanErr := NewError(code, err.Error())
	cleanErr.Operation = operation
	cleanErr.WithDetail("wrapped_error", err.Error())

	// If it's already a CleanWizardError, preserve details
	if wizardErr, ok := err.(*CleanWizardError); ok {
		cleanErr.Details = wizardErr.Details
		cleanErr.Stack = wizardErr.Stack
		cleanErr.Timestamp = wizardErr.Timestamp
	}

	return cleanErr
}

// IsNixAvailable checks if Nix is available on the system
func IsNixAvailable() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}
