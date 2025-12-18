package errors

import (
	"fmt"
	"runtime/debug"
)

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
		}
	default:
		// Add to metadata for custom keys
		if e.Details.Metadata == nil {
			e.Details.Metadata = make(map[string]string)
		}
		e.Details.Metadata[key] = fmt.Sprintf("%v", value)
	}
	return e
}

// WithLevel sets error level
func (e *CleanWizardError) WithLevel(level ErrorLevel) *CleanWizardError {
	e.Level = level
	return e
}

// IsLevel checks if error matches specific level
func (e *CleanWizardError) IsLevel(level ErrorLevel) bool {
	return e.Level == level
}

// IsErrorCode checks if error matches specific error code
func (e *CleanWizardError) IsErrorCode(code ErrorCode) bool {
	return e.Code == code
}

// captureStack captures the current call stack
func captureStack() string {
	// Use debug.Stack() for simpler stack capture
	return string(debug.Stack())
}

// IsRetryable checks if error is retryable
func (e *CleanWizardError) IsRetryable() bool {
	switch e.Code {
	case ErrTimeout, ErrNixCommandFailed, ErrCleaningFailed:
		return true
	case ErrConfigLoad, ErrConfigSave:
		return true
	default:
		return false
	}
}

// IsUserFriendly checks if error is safe to show to users
func (e *CleanWizardError) IsUserFriendly() bool {
	switch e.Code {
	case ErrPermissionDenied, ErrTimeout, ErrConfigValidation:
		return true
	case ErrInvalidInput, ErrNotFound:
		return true
	default:
		return false
	}
}

// Log logs the error with appropriate level
func (e *CleanWizardError) Log() {
	level := e.Level.LogLevel()
	switch level {
	case "debug":
		e.debugLog()
	case "info":
		e.infoLog()
	case "warning":
		e.warnLog()
	case "error":
		e.errorLog()
	case "fatal":
		e.fatalLog()
	case "panic":
		e.panicLog()
	default:
		e.errorLog()
	}
}

// Helper logging methods
func (e *CleanWizardError) debugLog() {
	// Implementation depends on logging library choice
	// For now, using simple approach
	_ = fmt.Sprintf("[DEBUG] %s", e.Error())
}

func (e *CleanWizardError) infoLog() {
	// Implementation depends on logging library choice
	_ = fmt.Sprintf("[INFO] %s", e.Error())
}

func (e *CleanWizardError) warnLog() {
	// Implementation depends on logging library choice
	_ = fmt.Sprintf("[WARN] %s", e.Error())
}

func (e *CleanWizardError) errorLog() {
	// Implementation depends on logging library choice
	_ = fmt.Sprintf("[ERROR] %s", e.Error())
}

func (e *CleanWizardError) fatalLog() {
	// Implementation depends on logging library choice
	_ = fmt.Sprintf("[FATAL] %s", e.Error())
}

func (e *CleanWizardError) panicLog() {
	// Implementation depends on logging library choice
	_ = fmt.Sprintf("[PANIC] %s", e.Error())
}
