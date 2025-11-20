package errors

import (
	"fmt"
	"runtime"
)

// ErrorCode represents type-safe error codes with compile-time safety
type ErrorCode int

const (
	// Domain errors (1000-1999)
	ErrCodeInvalidGeneration  ErrorCode = 1000
	ErrCodeInvalidSettings    ErrorCode = 1001
	ErrCodeOptimizationFailed ErrorCode = 1002
	ErrCodeCleanupFailed      ErrorCode = 1003
	ErrCodeValidationFailed   ErrorCode = 1004

	// Configuration errors (2000-2999)
	ErrCodeInvalidConfig    ErrorCode = 2000
	ErrCodeMissingProfile   ErrorCode = 2001
	ErrCodeInvalidOperation ErrorCode = 2002
	ErrCodeSafetyViolation  ErrorCode = 2003
	ErrCodePermissionDenied ErrorCode = 2004

	// File system errors (3000-3999)
	ErrCodeFileNotFound    ErrorCode = 3000
	ErrCodePermissionError ErrorCode = 3001
	ErrCodeDiskFull        ErrorCode = 3002
	ErrCodeCorruption      ErrorCode = 3003
	ErrCodePathInvalid     ErrorCode = 3004

	// Network errors (4000-4999)
	ErrCodeConnectionFailed ErrorCode = 4000
	ErrCodeTimeout          ErrorCode = 4001
	ErrCodeRateLimited      ErrorCode = 4002

	// System errors (5000-5999)
	ErrCodeOutOfMemory       ErrorCode = 5000
	ErrCodeProcessFailed     ErrorCode = 5001
	ErrCodeResourceExhausted ErrorCode = 5002
)

// ErrorSeverity represents error severity levels with type safety
type ErrorSeverity int

const (
	SeverityDebug ErrorSeverity = iota
	SeverityInfo
	SeverityWarning
	SeverityError
	SeverityCritical
)

// String returns string representation of ErrorSeverity
func (es ErrorSeverity) String() string {
	switch es {
	case SeverityDebug:
		return "debug"
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ErrorType represents different categories of errors
type ErrorType int

const (
	ErrorTypeDomain ErrorType = iota
	ErrorTypeConfig
	ErrorTypeFileSystem
	ErrorTypeNetwork
	ErrorTypeSystem
	ErrorTypeValidation
	ErrorTypePermission
)

// String returns string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeDomain:
		return "domain"
	case ErrorTypeConfig:
		return "config"
	case ErrorTypeFileSystem:
		return "filesystem"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeSystem:
		return "system"
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypePermission:
		return "permission"
	default:
		return "unknown"
	}
}

// CleanWizardError represents structured error with type safety
type CleanWizardError struct {
	Code      ErrorCode     `json:"code"`
	Type      ErrorType     `json:"type"`
	Severity  ErrorSeverity `json:"severity"`
	Message   string        `json:"message"`
	Details   any           `json:"details,omitempty"`
	Cause     error         `json:"cause,omitempty"`
	Caller    string        `json:"caller,omitempty"`
	Timestamp int64         `json:"timestamp"`
}

// Error implements error interface
func (e *CleanWizardError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s (caused by: %v)", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *CleanWizardError) Unwrap() error {
	return e.Cause
}

// IsType checks if error is of specific type
func (e *CleanWizardError) IsType(errorType ErrorType) bool {
	return e.Type == errorType
}

// IsCode checks if error has specific code
func (e *CleanWizardError) IsCode(code ErrorCode) bool {
	return e.Code == code
}

// IsSeverity checks if error has specific severity level or higher (more severe)
func (e *CleanWizardError) IsSeverity(severity ErrorSeverity) bool {
	return e.Severity >= severity
}

// WithCaller adds caller information to error
func (e *CleanWizardError) WithCaller() *CleanWizardError {
	_, file, line, _ := runtime.Caller(2)
	e.Caller = fmt.Sprintf("%s:%d", file, line)
	return e
}

// WithDetails adds additional details to error
func (e *CleanWizardError) WithDetails(details any) *CleanWizardError {
	e.Details = details
	return e
}

// WithCause adds underlying cause to error
func (e *CleanWizardError) WithCause(cause error) *CleanWizardError {
	e.Cause = cause
	return e
}

// WithSeverity sets error severity
func (e *CleanWizardError) WithSeverity(severity ErrorSeverity) *CleanWizardError {
	e.Severity = severity
	return e
}

// NewError creates a new structured error
func NewError(code ErrorCode, message string) *CleanWizardError {
	return &CleanWizardError{
		Code:      code,
		Type:      inferErrorType(code),
		Severity:  inferErrorSeverity(code),
		Message:   message,
		Timestamp: getCurrentTimestamp(),
	}
}

// NewErrorf creates a new structured error with formatted message
func NewErrorf(code ErrorCode, format string, args ...any) *CleanWizardError {
	return &CleanWizardError{
		Code:      code,
		Type:      inferErrorType(code),
		Severity:  inferErrorSeverity(code),
		Message:   fmt.Sprintf(format, args...),
		Timestamp: getCurrentTimestamp(),
	}
}

// WrapError wraps an existing error with additional context
func WrapError(cause error, code ErrorCode, message string) *CleanWizardError {
	return &CleanWizardError{
		Code:      code,
		Type:      inferErrorType(code),
		Severity:  inferErrorSeverity(code),
		Message:   message,
		Cause:     cause,
		Timestamp: getCurrentTimestamp(),
	}
}

// WrapErrorf wraps an existing error with formatted message
func WrapErrorf(cause error, code ErrorCode, format string, args ...any) *CleanWizardError {
	return &CleanWizardError{
		Code:      code,
		Type:      inferErrorType(code),
		Severity:  inferErrorSeverity(code),
		Message:   fmt.Sprintf(format, args...),
		Cause:     cause,
		Timestamp: getCurrentTimestamp(),
	}
}

// inferErrorType maps error codes to error types
func inferErrorType(code ErrorCode) ErrorType {
	switch code {
	case ErrCodeValidationFailed:
		return ErrorTypeValidation
	case ErrCodePermissionDenied, ErrCodePermissionError:
		return ErrorTypePermission
	default:
		switch {
		case code >= 1000 && code < 2000:
			return ErrorTypeDomain
		case code >= 2000 && code < 3000:
			return ErrorTypeConfig
		case code >= 3000 && code < 4000:
			return ErrorTypeFileSystem
		case code >= 4000 && code < 5000:
			return ErrorTypeNetwork
		case code >= 5000 && code < 6000:
			return ErrorTypeSystem
		default:
			return ErrorTypeDomain
		}
	}
}

// inferErrorSeverity maps error codes to severity levels
func inferErrorSeverity(code ErrorCode) ErrorSeverity {
	switch code {
	case ErrCodeFileNotFound:
		return SeverityWarning // File not found is typically a warning
	case ErrCodeValidationFailed:
		return SeverityError // Validation errors are serious
	case ErrCodeInvalidConfig:
		return SeverityInfo // Config errors are typically info level
	default:
		switch {
		case code >= 5000:
			return SeverityCritical
		case code >= 4000:
			return SeverityError
		case code >= 3000:
			return SeverityError // File system errors are typically errors
		case code >= 2000:
			return SeverityInfo // Config errors are typically info level
		default:
			return SeverityError
		}
	}
}

// getCurrentTimestamp returns current Unix timestamp (placeholder)
func getCurrentTimestamp() int64 {
	// In real implementation, use time.Now().Unix()
	return 0
}
