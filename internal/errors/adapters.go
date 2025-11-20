package errors

import (
	"os"
	"strings"
	"syscall"
)

// isConfigurationError checks if error is configuration-related
func isConfigurationError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Common indicators of configuration errors
	configIndicators := []string{
		"config", "configuration", "yaml", "json", "toml",
		"parse", "invalid", "missing", "required",
	}

	lowerErrStr := strings.ToLower(errStr)
	for _, indicator := range configIndicators {
		if strings.Contains(lowerErrStr, indicator) {
			return true
		}
	}

	return false
}

// FileSystemErrorAdapter wraps file system errors with proper type safety
type FileSystemErrorAdapter struct{}

// Adapt converts OS-level errors to structured errors
func (fsa *FileSystemErrorAdapter) Adapt(err error) *CleanWizardError {
	if err == nil {
		return nil
	}

	// Handle specific OS errors
	if pathErr, ok := err.(*os.PathError); ok {
		switch {
		case pathErr.Err == syscall.ENOENT || pathErr.Err == os.ErrNotExist:
			return NewError(ErrCodeFileNotFound,
				"File not found: "+pathErr.Path).WithCause(err).WithCaller()
		case pathErr.Err == syscall.EACCES || pathErr.Err == syscall.EPERM || pathErr.Err == os.ErrPermission:
			return NewError(ErrCodePermissionError,
				"Permission denied: "+pathErr.Path).WithCause(err).WithCaller()
		case pathErr.Err == syscall.ENOSPC:
			return NewError(ErrCodeDiskFull,
				"Disk full: "+pathErr.Path).WithCause(err).WithCaller()
		default:
			return NewError(ErrCodeFilesystem,
				"File system error: "+pathErr.Err.Error()).WithCause(err).WithCaller()
		}
	}

	// Handle other OS errors
	if os.IsNotExist(err) {
		return NewError(ErrCodeFileNotFound,
			"File or directory not found").WithCause(err).WithCaller()
	}

	if os.IsPermission(err) {
		return NewError(ErrCodePermissionError,
			"Permission denied").WithCause(err).WithCaller()
	}

	// Not a file system error, return nil to let other adapters handle it
	return nil
}

// ConfigErrorAdapter wraps configuration errors with proper type safety
type ConfigErrorAdapter struct{}

// Adapt converts configuration errors to structured errors
func (cea *ConfigErrorAdapter) Adapt(err error, context string) *CleanWizardError {
	if err == nil {
		return nil
	}

	// Only handle actual configuration-related errors, not generic errors
	// This allows the default handler to catch non-config errors
	if isConfigurationError(err) {
		return WrapError(err, ErrCodeInvalidConfig,
			"Configuration error: "+context).WithCaller()
	}

	// Not a configuration error
	return nil
}

// ValidationErrorAdapter wraps validation errors with proper type safety
type ValidationErrorAdapter struct{}

// Adapt converts validation errors to structured errors
func (vea *ValidationErrorAdapter) Adapt(err error, field string) *CleanWizardError {
	if err == nil {
		return nil
	}

	// Only handle actual validation errors
	if isValidationError(err) {
		return WrapErrorf(err, ErrCodeValidationFailed,
			"Validation failed for field '%s': %v", field, err).WithCaller()
	}

	// Not a validation error
	return nil
}

// isValidationError checks if error is validation-related
func isValidationError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Strong indicators of validation errors
	validationIndicators := []string{
		"validation failed", "validation error", "invalid format",
		"invalid range", "required field", "constraint violation",
		"unacceptable value", "validation rule", "invalid value",
	}

	lowerErrStr := strings.ToLower(errStr)
	for _, indicator := range validationIndicators {
		if strings.Contains(lowerErrStr, indicator) {
			return true
		}
	}

	return false
}

// NetworkErrorAdapter wraps network errors with proper type safety
type NetworkErrorAdapter struct{}

// Adapt converts network errors to structured errors
func (nea *NetworkErrorAdapter) Adapt(err error, operation string) *CleanWizardError {
	if err == nil {
		return nil
	}

	// Only handle actual network errors
	if !isNetworkError(err) {
		return nil
	}

	return WrapError(err, ErrCodeConnectionFailed,
		"Network operation failed: "+operation).WithCaller()
}

// isNetworkError checks if error is network-related
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Common indicators of network errors
	networkIndicators := []string{
		"connection", "timeout", "network", "dns", "host", "dial",
		"refused", "unreachable", "network unreachable", "connection refused",
	}

	lowerErrStr := strings.ToLower(errStr)
	for _, indicator := range networkIndicators {
		if strings.Contains(lowerErrStr, indicator) {
			return true
		}
	}

	return false
}

// SystemErrorAdapter wraps system-level errors with proper type safety
type SystemErrorAdapter struct{}

// Adapt converts system errors to structured errors
func (sea *SystemErrorAdapter) Adapt(err error, component string) *CleanWizardError {
	if err == nil {
		return nil
	}

	// Only handle actual system errors
	if !isSystemError(err) {
		return nil
	}

	return WrapError(err, ErrCodeProcessFailed,
		"System component failed: "+component).WithCaller()
}

// isSystemError checks if error is system-related
func isSystemError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Common indicators of system errors
	systemIndicators := []string{
		"permission denied", "access denied", "operation not permitted",
		"no such file", "file not found", "device", "resource", "memory",
		"disk", "space", "quota", "limit", "system", "kernel",
	}

	lowerErrStr := strings.ToLower(errStr)
	for _, indicator := range systemIndicators {
		if strings.Contains(lowerErrStr, indicator) {
			return true
		}
	}

	return false
}

// ExternalToolErrorAdapter wraps external tool errors with proper type safety
type ExternalToolErrorAdapter struct{}

// Adapt converts external tool errors to structured errors
func (eta *ExternalToolErrorAdapter) Adapt(err error, tool string) *CleanWizardError {
	if err == nil {
		return nil
	}

	// Only handle actual external tool errors
	if !isExternalToolError(err) {
		return nil
	}

	return WrapError(err, ErrCodeCleanupFailed,
		"External tool '"+tool+"' failed").WithCaller()
}

// isExternalToolError checks if error is external tool-related
func isExternalToolError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Common indicators of external tool errors
	externalToolIndicators := []string{
		"nix", "brew", "docker", "kubectl", "git", "curl", "wget",
		"command not found", "executable", "binary", "tool", "utility",
		"external command", "subprocess", "child process", "fork",
	}

	lowerErrStr := strings.ToLower(errStr)
	for _, indicator := range externalToolIndicators {
		if strings.Contains(lowerErrStr, indicator) {
			return true
		}
	}

	return false
}
