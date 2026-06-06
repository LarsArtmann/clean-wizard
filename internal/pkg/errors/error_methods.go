package errors

import (
	"runtime/debug"
)

// WithOperation adds operation context to error.
func (e *CleanWizardError) WithOperation(operation string) *CleanWizardError {
	e.Operation = operation

	return e
}

// detailFieldHandler defines a handler for a specific detail field.
type detailFieldHandler func(details **ErrorDetails, value any)

// detailFieldHandlers maps detail keys to their handlers.
var detailFieldHandlers = map[string]detailFieldHandler{ //nolint:gochecknoglobals
	"field":       setStringFieldStrictHandler,
	"value":       setStringFieldHandler, //nolint:goconst
	"expected":    setStringFieldHandler,
	"actual":      setStringFieldHandler,
	"operation":   setStringFieldStrictHandler,
	"file_path":   setStringFieldStrictHandler,
	"line_number": setIntFieldHandler,
	"retry_count": setIntFieldHandler,
	"duration":    setStringFieldStrictHandler,
}

func setStringFieldStrictHandler(details **ErrorDetails, value any) {
	setStringFieldStrict(&(*details).Field, value)
}

func setStringFieldHandler(details **ErrorDetails, value any) {
	setStringField(&(*details).Value, value)
}

func setIntFieldHandler(details **ErrorDetails, value any) {
	setIntField(&(*details).LineNumber, value)
}

// WithDetail adds single detail to error.
func (e *CleanWizardError) WithDetail(key string, value any) *CleanWizardError {
	ensureDetails(&e.Details)

	if handler, ok := detailFieldHandlers[key]; ok {
		handler(&e.Details, value)
	} else {
		// Add to metadata for custom keys
		e.Details.Metadata = addToMetadata(e.Details.Metadata, key, value)
	}

	return e
}

// WithLevel sets error level.
func (e *CleanWizardError) WithLevel(level ErrorLevel) *CleanWizardError {
	e.Level = level

	return e
}

// IsLevel checks if error matches specific level.
func (e *CleanWizardError) IsLevel(level ErrorLevel) bool {
	return e.Level == level
}

// IsErrorCode checks if error matches specific error code.
func (e *CleanWizardError) IsErrorCode(code ErrorCode) bool {
	return e.Code == code
}

// captureStack captures the current call stack.
func captureStack() string {
	// Use debug.Stack() for simpler stack capture
	return string(debug.Stack())
}

// IsRetryable checks if error is retryable.
func (e *CleanWizardError) IsRetryable() bool {
	switch e.Code {
	case ErrTimeout, ErrNixCommandFailed, ErrCleaningFailed:
		return true
	case ErrConfigLoad, ErrConfigSave:
		return true
	case ErrUnknown, ErrInvalidInput, ErrNotFound, ErrPermissionDenied:
		return false
	case ErrConfigValidation, ErrNixNotAvailable, ErrNixStoreCorrupted:
		return false
	case ErrCleaningTimeout, ErrCleanupRollback:
		return false
	default:
		return false
	}
}

// IsUserFriendly checks if error is safe to show to users.
func (e *CleanWizardError) IsUserFriendly() bool {
	switch e.Code {
	case ErrPermissionDenied, ErrTimeout, ErrConfigValidation:
		return true
	case ErrInvalidInput, ErrNotFound:
		return true
	case ErrUnknown, ErrConfigLoad, ErrConfigSave:
		return false
	case ErrNixNotAvailable, ErrNixCommandFailed, ErrNixStoreCorrupted:
		return false
	case ErrCleaningFailed, ErrCleaningTimeout, ErrCleanupRollback:
		return false
	default:
		return false
	}
}

// Log logs the error with appropriate level.
func (e *CleanWizardError) Log() {
	level := e.Level.LogLevel()
	switch level {
	case "debug":
		e.debugLog()
	case "info":
		e.infoLog()
	case "warning":
		e.warnLog()
	case "error": //nolint:goconst
		e.errorLog()
	case "fatal":
		e.fatalLog()
	case "panic":
		e.panicLog()
	default:
		e.errorLog()
	}
}

// Helper logging methods.
func (e *CleanWizardError) debugLog() {
	// Implementation depends on logging library choice
	// For now, using simple approach
	_ = "[DEBUG] " + e.Error()
}

func (e *CleanWizardError) infoLog() {
	// Implementation depends on logging library choice
	_ = "[INFO] " + e.Error()
}

func (e *CleanWizardError) warnLog() {
	// Implementation depends on logging library choice
	_ = "[WARN] " + e.Error()
}

func (e *CleanWizardError) errorLog() {
	// Implementation depends on logging library choice
	_ = "[ERROR] " + e.Error()
}

func (e *CleanWizardError) fatalLog() {
	// Implementation depends on logging library choice
	_ = "[FATAL] " + e.Error()
}

func (e *CleanWizardError) panicLog() {
	// Implementation depends on logging library choice
	_ = "[PANIC] " + e.Error()
}
