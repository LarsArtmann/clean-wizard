package errors

import "fmt"

// CleanWizardError represents a custom error type for Clean Wizard
type CleanWizardError struct {
	Code    string
	Message string
	Cause   error
}

// Error implements the error interface
func (e *CleanWizardError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *CleanWizardError) Unwrap() error {
	return e.Cause
}

// NewCleanWizardError creates a new CleanWizardError
func NewCleanWizardError(code, message string, cause error) *CleanWizardError {
	return &CleanWizardError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Predefined error codes
const (
	ErrCodeConfigLoad     = "CONFIG_LOAD"
	ErrCodeConfigSave     = "CONFIG_SAVE"
	ErrCodeConfigValidate = "CONFIG_VALIDATE"
	ErrCodeScanFailed     = "SCAN_FAILED"
	ErrCodeCleanFailed    = "CLEAN_FAILED"
	ErrCodeBackupFailed   = "BACKUP_FAILED"
	ErrCodeProtectedPath  = "PROTECTED_PATH"
	ErrCodeCommandFailed  = "COMMAND_FAILED"
	ErrCodePermission     = "PERMISSION"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeInvalidInput   = "INVALID_INPUT"
	ErrCodeSystemError    = "SYSTEM_ERROR"
)

// ConfigLoadError creates a new config load error
func ConfigLoadError(cause error) error {
	return NewCleanWizardError(ErrCodeConfigLoad, "failed to load configuration", cause)
}

// ConfigSaveError creates a new config save error
func ConfigSaveError(cause error) error {
	return NewCleanWizardError(ErrCodeConfigSave, "failed to save configuration", cause)
}

// ConfigValidateError creates a new config validation error
func ConfigValidateError(message string) error {
	return NewCleanWizardError(ErrCodeConfigValidate, message, nil)
}

// ScanFailedError creates a new scan failed error
func ScanFailedError(component string, cause error) error {
	return NewCleanWizardError(ErrCodeScanFailed, fmt.Sprintf("failed to scan %s", component), cause)
}

// CleanFailedError creates a new clean failed error
func CleanFailedError(operation string, cause error) error {
	return NewCleanWizardError(ErrCodeCleanFailed, fmt.Sprintf("failed to clean %s", operation), cause)
}

// BackupFailedError creates a new backup failed error
func BackupFailedError(cause error) error {
	return NewCleanWizardError(ErrCodeBackupFailed, "failed to create backup", cause)
}

// ProtectedPathError creates a new protected path error
func ProtectedPathError(path string) error {
	return NewCleanWizardError(ErrCodeProtectedPath, fmt.Sprintf("path is protected: %s", path), nil)
}

// CommandFailedError creates a new command failed error
func CommandFailedError(command string, cause error) error {
	return NewCleanWizardError(ErrCodeCommandFailed, fmt.Sprintf("command failed: %s", command), cause)
}

// PermissionError creates a new permission error
func PermissionError(operation string, cause error) error {
	return NewCleanWizardError(ErrCodePermission, fmt.Sprintf("permission denied for %s", operation), cause)
}

// NotFoundError creates a new not found error
func NotFoundError(what string) error {
	return NewCleanWizardError(ErrCodeNotFound, fmt.Sprintf("%s not found", what), nil)
}

// InvalidInputError creates a new invalid input error
func InvalidInputError(message string) error {
	return NewCleanWizardError(ErrCodeInvalidInput, message, nil)
}

// SystemError creates a new system error
func SystemError(message string, cause error) error {
	return NewCleanWizardError(ErrCodeSystemError, message, cause)
}

// IsErrorCode checks if an error is a CleanWizardError with the given code
func IsErrorCode(err error, code string) bool {
	if cwErr, ok := err.(*CleanWizardError); ok {
		return cwErr.Code == code
	}
	return false
}

// GetErrorCode returns the error code from a CleanWizardError, or empty string if not a CleanWizardError
func GetErrorCode(err error) string {
	if cwErr, ok := err.(*CleanWizardError); ok {
		return cwErr.Code
	}
	return ""
}

// GetErrorMessage returns the error message from a CleanWizardError, or the error string if not a CleanWizardError
func GetErrorMessage(err error) string {
	if cwErr, ok := err.(*CleanWizardError); ok {
		return cwErr.Message
	}
	return err.Error()
}
