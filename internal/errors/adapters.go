package errors

import (
	"os"
	"syscall"
)

// FileSystemErrorAdapter wraps file system errors with proper type safety
type FileSystemErrorAdapter struct{}

// Adapt converts OS-level errors to structured errors
func (fsa *FileSystemErrorAdapter) Adapt(err error) *CleanWizardError {
	if err == nil {
		return nil
	}

	// Handle specific OS errors
	var pathErr *os.PathError
	if err, ok := err.(*os.PathError); ok {
		switch err.Err {
		case syscall.ENOENT:
			return NewError(ErrCodeFileNotFound, 
				"File not found: "+pathErr.Path).WithCause(err).WithCaller()
		case syscall.EACCES, syscall.EPERM:
			return NewError(ErrCodePermissionDenied,
				"Permission denied: "+pathErr.Path).WithCause(err).WithCaller()
		case syscall.ENOSPC:
			return NewError(ErrCodeDiskFull,
				"Disk full: "+pathErr.Path).WithCause(err).WithCaller()
		default:
			return NewError(ErrCodeFileNotFound,
				"File system error: "+pathErr.Err.Error()).WithCause(err).WithCaller()
		}
	}

	// Handle other OS errors
	if os.IsNotExist(err) {
		return NewError(ErrCodeFileNotFound,
			"File or directory not found").WithCause(err).WithCaller()
	}

	if os.IsPermission(err) {
		return NewError(ErrCodePermissionDenied,
			"Permission denied").WithCause(err).WithCaller()
	}

	return NewError(ErrCodeValidationFailed,
		"File system validation failed").WithCause(err).WithCaller()
}

// ConfigErrorAdapter wraps configuration errors with proper type safety
type ConfigErrorAdapter struct{}

// Adapt converts configuration errors to structured errors
func (cea *ConfigErrorAdapter) Adapt(err error, context string) *CleanWizardError {
	if err == nil {
		return nil
	}

	return WrapError(err, ErrCodeInvalidConfig,
		"Configuration error: "+context).WithCaller()
}

// ValidationErrorAdapter wraps validation errors with proper type safety
type ValidationErrorAdapter struct{}

// Adapt converts validation errors to structured errors
func (vea *ValidationErrorAdapter) Adapt(err error, field string) *CleanWizardError {
	if err == nil {
		return nil
	}

	return WrapErrorf(err, ErrCodeValidationFailed,
		"Validation failed for field '%s': %v", field, err).WithCaller()
}

// NetworkErrorAdapter wraps network errors with proper type safety
type NetworkErrorAdapter struct{}

// Adapt converts network errors to structured errors
func (nea *NetworkErrorAdapter) Adapt(err error, operation string) *CleanWizardError {
	if err == nil {
		return nil
	}

	return WrapError(err, ErrCodeConnectionFailed,
		"Network operation failed: "+operation).WithCaller()
}

// SystemErrorAdapter wraps system-level errors with proper type safety
type SystemErrorAdapter struct{}

// Adapt converts system errors to structured errors
func (sea *SystemErrorAdapter) Adapt(err error, component string) *CleanWizardError {
	if err == nil {
		return nil
	}

	return WrapError(err, ErrCodeProcessFailed,
		"System component failed: "+component).WithCaller()
}

// ExternalToolErrorAdapter wraps external tool errors with proper type safety
type ExternalToolErrorAdapter struct{}

// Adapt converts external tool errors to structured errors
func (eta *ExternalToolErrorAdapter) Adapt(err error, tool string) *CleanWizardError {
	if err == nil {
		return nil
	}

	return WrapError(err, ErrCodeCleanupFailed,
		"External tool '"+tool+"' failed").WithCaller()
}