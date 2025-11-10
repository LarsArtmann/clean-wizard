package errors

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ErrorCode represents standardized error codes
type ErrorCode int

const (
	// General errors
	ErrUnknown ErrorCode = iota
	ErrInvalidInput
	ErrNotFound
	ErrPermissionDenied
	ErrTimeout

	// Configuration errors
	ErrConfigLoad
	ErrConfigSave
	ErrConfigValidation

	// Nix-specific errors
	ErrNixNotAvailable
	ErrNixCommandFailed
	ErrNixStoreCorrupted

	// Cleaning errors
	ErrCleaningFailed
	ErrCleaningTimeout
	ErrCleanupRollback
)

// String returns string representation of error code
func (e ErrorCode) String() string {
	switch e {
	case ErrUnknown:
		return "UNKNOWN"
	case ErrInvalidInput:
		return "INVALID_INPUT"
	case ErrNotFound:
		return "NOT_FOUND"
	case ErrPermissionDenied:
		return "PERMISSION_DENIED"
	case ErrTimeout:
		return "TIMEOUT"
	case ErrConfigLoad:
		return "CONFIG_LOAD"
	case ErrConfigSave:
		return "CONFIG_SAVE"
	case ErrConfigValidation:
		return "CONFIG_VALIDATION"
	case ErrNixNotAvailable:
		return "NIX_NOT_AVAILABLE"
	case ErrNixCommandFailed:
		return "NIX_COMMAND_FAILED"
	case ErrNixStoreCorrupted:
		return "NIX_STORE_CORRUPTED"
	case ErrCleaningFailed:
		return "CLEANING_FAILED"
	case ErrCleaningTimeout:
		return "CLEANING_TIMEOUT"
	case ErrCleanupRollback:
		return "CLEANUP_ROLLBACK"
	default:
		return "UNDEFINED"
	}
}

// ErrorLevel represents severity of error
type ErrorLevel int

const (
	LevelInfo ErrorLevel = iota
	LevelWarn
	LevelError
	LevelFatal
)

// String returns string representation of error level
func (e ErrorLevel) String() string {
	switch e {
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNDEFINED"
	}
}

// CleanWizardError represents structured error with context
type CleanWizardError struct {
	Code      ErrorCode
	Level     ErrorLevel
	Message   string
	Operation string
	Details   map[string]any
	Timestamp time.Time
	Stack     string
}

// Error implements error interface
func (e *CleanWizardError) Error() string {
	if e.Details == nil {
		return fmt.Sprintf("[%s] %s: %s", e.Level.String(), e.Code.String(), e.Message)
	}

	details := make([]string, 0, len(e.Details))
	for key, value := range e.Details {
		details = append(details, fmt.Sprintf("%s=%v", key, value))
	}

	return fmt.Sprintf("[%s] %s: %s (details: %s)",
		e.Level.String(), e.Code.String(), e.Message, strings.Join(details, ", "))
}

// Unwrap returns underlying error if any
func (e *CleanWizardError) Unwrap() error {
	return nil
}

// NewError creates new CleanWizardError
func NewError(code ErrorCode, message string) *CleanWizardError {
	return &CleanWizardError{
		Code:      code,
		Level:     LevelError,
		Message:   message,
		Details:   make(map[string]any),
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
		Details:   make(map[string]any),
		Timestamp: time.Now(),
		Stack:     captureStack(),
	}
}

// NewErrorWithDetails creates new CleanWizardError with context details
func NewErrorWithDetails(code ErrorCode, message string, details map[string]any) *CleanWizardError {
	err := &CleanWizardError{
		Code:      code,
		Level:     LevelError,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Stack:     captureStack(),
	}

	// Add automatic details
	if err.Details == nil {
		err.Details = make(map[string]any)
	}

	return err
}

// WithOperation adds operation context to error
func (e *CleanWizardError) WithOperation(operation string) *CleanWizardError {
	e.Operation = operation
	return e
}

// WithDetail adds single detail to error
func (e *CleanWizardError) WithDetail(key string, value any) *CleanWizardError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// WithLevel updates error level
func (e *CleanWizardError) WithLevel(level ErrorLevel) *CleanWizardError {
	e.Level = level
	return e
}

// IsLevel checks if error is at or above specified level
func (e *CleanWizardError) IsLevel(level ErrorLevel) bool {
	return e.Level >= level
}

// IsErrorCode checks if error has specific code
func (e *CleanWizardError) IsErrorCode(code ErrorCode) bool {
	return e.Code == code
}

// captureStack captures current stack trace
func captureStack() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	if n > 0 && n < len(buf) {
		return string(buf[:n])
	}
	return ""
}

// IsRetryable determines if error operation can be retried
func (e *CleanWizardError) IsRetryable() bool {
	switch e.Code {
	case ErrTimeout, ErrNixCommandFailed:
		return true
	default:
		return false
	}
}

// IsUserFriendly determines if error is suitable for end-user display
func (e *CleanWizardError) IsUserFriendly() bool {
	switch e.Level {
	case LevelInfo, LevelWarn:
		return true
	default:
		return e.Code == ErrNixNotAvailable || e.Code == ErrConfigValidation
	}
}

// Log logs error with appropriate level
func (e *CleanWizardError) Log() {
	fields := logrus.Fields{
		"code":      e.Code.String(),
		"level":     e.Level.String(),
		"operation": e.Operation,
		"timestamp": e.Timestamp,
	}

	if e.Details != nil {
		for key, value := range e.Details {
			fields[key] = value
		}
	}

	entry := logrus.WithFields(fields)

	switch e.Level {
	case LevelInfo:
		entry.Info(e.Message)
	case LevelWarn:
		entry.Warn(e.Message)
	case LevelError:
		entry.Error(e.Message)
	case LevelFatal:
		entry.Fatal(e.Message)
	}
}
