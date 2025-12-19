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
	ensureDetails(&e.Details)

	switch key {
	case "field":
		setStringFieldStrict(&e.Details.Field, value)
	case "value":
		setStringField(&e.Details.Value, value)
	case "expected":
		setStringField(&e.Details.Expected, value)
	case "actual":
		setStringField(&e.Details.Actual, value)
	case "operation":
		setStringFieldStrict(&e.Details.Operation, value)
	case "file_path":
		setStringFieldStrict(&e.Details.FilePath, value)
	case "line_number":
		setIntField(&e.Details.LineNumber, value)
	case "retry_count":
		setIntField(&e.Details.RetryCount, value)
	case "duration":
		setStringFieldStrict(&e.Details.Duration, value)
	default:
		// Add to metadata for custom keys
		e.Details.Metadata = addToMetadata(e.Details.Metadata, key, value)
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
