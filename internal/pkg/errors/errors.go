package errors

import (
	"fmt"

	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
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

// ErrorDetails represents strongly-typed error context information
type ErrorDetails struct {
	Field       string `json:"field,omitempty"`
	Value       string `json:"value,omitempty"`
	Expected    string `json:"expected,omitempty"`
	Actual      string `json:"actual,omitempty"`
	Operation   string `json:"operation,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	LineNumber  int    `json:"line_number,omitempty"`
	RetryCount  int    `json:"retry_count,omitempty"`
	Duration    string `json:"duration,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CleanWizardError represents structured error with context
type CleanWizardError struct {
	Code      ErrorCode
	Level     ErrorLevel
	Message   string
	Operation string
	Details   *ErrorDetails `json:"details,omitempty"`
	Timestamp time.Time
	Stack     string
}

// Error implements error interface
func (e *CleanWizardError) Error() string {
	if e.Details == nil {
		return fmt.Sprintf("[%s] %s: %s", e.Level.String(), e.Code.String(), e.Message)
	}

	var details []string
	if e.Details.Field != "" {
		details = append(details, fmt.Sprintf("field=%s", e.Details.Field))
	}
	if e.Details.Value != "" {
		details = append(details, fmt.Sprintf("value=%s", e.Details.Value))
	}
	if e.Details.Expected != "" {
		details = append(details, fmt.Sprintf("expected=%s", e.Details.Expected))
	}
	if e.Details.Actual != "" {
		details = append(details, fmt.Sprintf("actual=%s", e.Details.Actual))
	}
	if e.Details.Operation != "" {
		details = append(details, fmt.Sprintf("operation=%s", e.Details.Operation))
	}
	if e.Details.FilePath != "" {
		details = append(details, fmt.Sprintf("file=%s", e.Details.FilePath))
	}
	if e.Details.LineNumber > 0 {
		details = append(details, fmt.Sprintf("line=%d", e.Details.LineNumber))
	}
	if e.Details.RetryCount > 0 {
		details = append(details, fmt.Sprintf("retries=%d", e.Details.RetryCount))
	}
	if e.Details.Duration != "" {
		details = append(details, fmt.Sprintf("duration=%s", e.Details.Duration))
	}

	// Add metadata details
	for key, value := range e.Details.Metadata {
		details = append(details, fmt.Sprintf("%s=%s", key, value))
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

// WithOperation adds operation context to error
func (e *CleanWizardError) WithOperation(operation string) *CleanWizardError {
	e.Operation = operation
	return e
}

// WithDetail adds single detail to error
func (e *CleanWizardError) WithDetail(key string, value any) *CleanWizardError {
	if e.Details == nil {
		e.Details = &ErrorDetails{
			Metadata: make(map[string]string),
		}
	}

	switch key {
	case "field":
		if v, ok := value.(string); ok {
			e.Details.Field = v
		}
	case "value":
		if v, ok := value.(string); ok {
			e.Details.Value = v
		} else {
			e.Details.Value = fmt.Sprintf("%v", value)
		}
	case "expected":
		if v, ok := value.(string); ok {
			e.Details.Expected = v
		} else {
			e.Details.Expected = fmt.Sprintf("%v", value)
		}
	case "actual":
		if v, ok := value.(string); ok {
			e.Details.Actual = v
		} else {
			e.Details.Actual = fmt.Sprintf("%v", value)
		}
	case "operation":
		if v, ok := value.(string); ok {
			e.Details.Operation = v
		}
	case "file_path":
		if v, ok := value.(string); ok {
			e.Details.FilePath = v
		}
	case "line_number":
		if v, ok := value.(int); ok {
			e.Details.LineNumber = v
		}
	case "retry_count":
		if v, ok := value.(int); ok {
			e.Details.RetryCount = v
		}
	case "duration":
		if v, ok := value.(string); ok {
			e.Details.Duration = v
		} else if d, ok := value.(time.Duration); ok {
			e.Details.Duration = d.String()
		} else {
			e.Details.Duration = fmt.Sprintf("%v", value)
		}
	default:
		// Store unknown keys in metadata
		if e.Details.Metadata == nil {
			e.Details.Metadata = make(map[string]string)
		}
		e.Details.Metadata[key] = fmt.Sprintf("%v", value)
	}

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
	event := log.Info().
		Str("code", e.Code.String()).
		Str("level", e.Level.String()).
		Str("operation", e.Operation).
		Time("timestamp", e.Timestamp)

	if e.Details != nil {
		if e.Details.Field != "" {
			event = event.Str("detail_field", e.Details.Field)
		}
		if e.Details.Value != "" {
			event = event.Str("detail_value", e.Details.Value)
		}
		if e.Details.Expected != "" {
			event = event.Str("detail_expected", e.Details.Expected)
		}
		if e.Details.Actual != "" {
			event = event.Str("detail_actual", e.Details.Actual)
		}
		if e.Details.Operation != "" {
			event = event.Str("detail_operation", e.Details.Operation)
		}
		if e.Details.FilePath != "" {
			event = event.Str("detail_file_path", e.Details.FilePath)
		}
		if e.Details.LineNumber > 0 {
			event = event.Int("detail_line_number", e.Details.LineNumber)
		}
		if e.Details.RetryCount > 0 {
			event = event.Int("detail_retry_count", e.Details.RetryCount)
		}
		if e.Details.Duration != "" {
			event = event.Str("detail_duration", e.Details.Duration)
		}

		// Add metadata
		for key, value := range e.Details.Metadata {
			event = event.Str("meta_"+key, value)
		}
	}

	switch e.Level {
	case LevelInfo:
		event.Msg(e.Message)
	case LevelWarn:
		event.Msg(e.Message)
	case LevelError:
		event.Msg(e.Message)
	case LevelFatal:
		event.Msg(e.Message)
	}
}
