package errors

import (
	"context"
	"fmt"
)

// ErrorHandler interface for centralized error handling
type ErrorHandler interface {
	Handle(err error) *CleanWizardError
	Recover() *CleanWizardError
	WithContext(ctx context.Context) ErrorHandler
}

// DefaultErrorHandler provides default error handling implementation
type DefaultErrorHandler struct {
	context context.Context
}

// NewErrorHandler creates a new error handler
func NewErrorHandler() ErrorHandler {
	return &DefaultErrorHandler{
		context: context.Background(),
	}
}

// Handle processes an error and returns structured error
func (deh *DefaultErrorHandler) Handle(err error) *CleanWizardError {
	if err == nil {
		return nil
	}

	// If it's already a CleanWizardError, return as-is
	if cwErr, ok := err.(*CleanWizardError); ok {
		return cwErr.WithCaller()
	}

	// Handle by type with appropriate adapters
	return deh.handleByType(err)
}

// Recover handles panics and converts to structured error
func (deh *DefaultErrorHandler) Recover() *CleanWizardError {
	if r := recover(); r != nil {
		var cause error
		if err, ok := r.(error); ok {
			cause = err
		} else {
			cause = NewErrorf(ErrCodeProcessFailed, "panic recovered: %v", r)
		}
		return NewError(ErrCodeProcessFailed, "System panic occurred").
			WithCause(cause).WithCaller()
	}
	return nil
}

// WithContext adds context to error handler
func (deh *DefaultErrorHandler) WithContext(ctx context.Context) ErrorHandler {
	return &DefaultErrorHandler{
		context: ctx,
	}
}

// handleByType processes errors by their type
func (deh *DefaultErrorHandler) handleByType(err error) *CleanWizardError {
	// Try different adapters based on error characteristics
	fsAdapter := &FileSystemErrorAdapter{}
	if fsErr := fsAdapter.Adapt(err); fsErr != nil {
		return fsErr
	}

	configAdapter := &ConfigErrorAdapter{}
	if configErr := configAdapter.Adapt(err, "configuration"); configErr != nil {
		return configErr
	}

	validationAdapter := &ValidationErrorAdapter{}
	if validationErr := validationAdapter.Adapt(err, "field"); validationErr != nil {
		return validationErr
	}

	// Default fallback
	return WrapError(err, ErrCodeValidationFailed, "Unhandled error type").WithCaller()
}

// Common convenience functions

// DomainError creates domain-level errors
func DomainError(code ErrorCode, message string) *CleanWizardError {
	return NewError(code, message).WithCaller()
}

// DomainErrorf creates domain-level errors with formatting
func DomainErrorf(code ErrorCode, format string, args ...any) *CleanWizardError {
	return NewErrorf(code, format, args...).WithCaller()
}

// ConfigError creates configuration errors
func ConfigError(message string) *CleanWizardError {
	return NewError(ErrCodeInvalidConfig, message).WithCaller()
}

// ConfigErrorf creates configuration errors with formatting
func ConfigErrorf(format string, args ...any) *CleanWizardError {
	return NewErrorf(ErrCodeInvalidConfig, format, args...).WithCaller()
}

// ValidationError creates validation errors
func ValidationError(field, message string) *CleanWizardError {
	return NewErrorf(ErrCodeValidationFailed, "Field '%s': %s", field, message).WithCaller()
}

// PermissionError creates permission errors
func PermissionError(resource string) *CleanWizardError {
	return NewErrorf(ErrCodePermissionDenied, "Permission denied for: %s", resource).WithCaller()
}

// FileNotFoundError creates file not found errors
func FileNotFoundError(path string) *CleanWizardError {
	return NewErrorf(ErrCodeFileNotFound, "File not found: %s", path).WithCaller()
}

// CleanupError creates cleanup operation errors
func CleanupError(operation string, cause error) *CleanWizardError {
	return WrapErrorf(cause, ErrCodeCleanupFailed, "Cleanup operation '%s' failed", operation).WithCaller()
}

// OptimizationError creates optimization errors
func OptimizationError(component string, cause error) *CleanWizardError {
	return WrapErrorf(cause, ErrCodeOptimizationFailed, "Optimization failed for %s", component).WithCaller()
}

// SystemError creates system-level errors
func SystemError(component, message string) *CleanWizardError {
	return NewErrorf(ErrCodeProcessFailed, "System component '%s': %s", component, message).WithCaller()
}

// SystemErrorf creates system-level errors with formatting
func SystemErrorf(component, format string, args ...any) *CleanWizardError {
	return NewErrorf(ErrCodeProcessFailed, "System component '%s': %s", component,
		fmt.Sprintf(format, args...)).WithCaller()
}

// IsCleanWizardError checks if error is a CleanWizardError
func IsCleanWizardError(err error) bool {
	_, ok := err.(*CleanWizardError)
	return ok
}

// GetCleanWizardError converts any error to CleanWizardError
func GetCleanWizardError(err error) *CleanWizardError {
	if err == nil {
		return nil
	}

	if cwErr, ok := err.(*CleanWizardError); ok {
		return cwErr
	}

	handler := NewErrorHandler()
	return handler.Handle(err)
}

// ErrorBuilder provides fluent error building
type ErrorBuilder struct {
	error *CleanWizardError
}

// NewErrorBuilder creates a new error builder
func NewErrorBuilder(code ErrorCode) *ErrorBuilder {
	return &ErrorBuilder{
		error: NewError(code, ""),
	}
}

// WithMessage sets the error message
func (eb *ErrorBuilder) WithMessage(message string) *ErrorBuilder {
	eb.error.Message = message
	return eb
}

// WithMessagef sets the error message with formatting
func (eb *ErrorBuilder) WithMessagef(format string, args ...any) *ErrorBuilder {
	eb.error.Message = fmt.Sprintf(format, args...)
	return eb
}

// WithDetails adds additional details to the error
func (eb *ErrorBuilder) WithDetails(details any) *ErrorBuilder {
	eb.error.Details = details
	return eb
}

// WithCause adds underlying cause to the error
func (eb *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	eb.error.Cause = cause
	return eb
}

// WithSeverity sets error severity
func (eb *ErrorBuilder) WithSeverity(severity ErrorSeverity) *ErrorBuilder {
	eb.error.Severity = severity
	return eb
}

// Build constructs the final error
func (eb *ErrorBuilder) Build() *CleanWizardError {
	return eb.error.WithCaller()
}
