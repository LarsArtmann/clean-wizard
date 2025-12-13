package errors

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// ErrorCode represents error code with type safety
type ErrorCode int

const (
	// System errors
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeTimeout
	ErrorCodeCancelled
	ErrorCodeRetryable

	// Validation errors
	ErrorCodeValidation
	ErrorCodeInvalidInput
	ErrorCodeInvalidState

	// Domain errors
	ErrorCodeNotFound
	ErrorCodeAlreadyExists
	ErrorCodeConflict

	// Permission errors
	ErrorCodeUnauthorized
	ErrorCodeForbidden
	ErrorCodeDenied

	// Configuration errors
	ErrorCodeConfig
	ErrorCodeProfile
	ErrorCodeSettings

	// Operation errors
	ErrorCodeOperation
	ErrorCodeExecution
	ErrorCodeRecovery

	// Security errors
	ErrorCodeSecurity
	ErrorCodeThreat
	ErrorCodeBlock
)

// String returns string representation of error code
func (ec ErrorCode) String() string {
	switch ec {
	case ErrorCodeUnknown:
		return "UNKNOWN"
	case ErrorCodeTimeout:
		return "TIMEOUT"
	case ErrorCodeCancelled:
		return "CANCELLED"
	case ErrorCodeRetryable:
		return "RETRYABLE"
	case ErrorCodeValidation:
		return "VALIDATION"
	case ErrorCodeInvalidInput:
		return "INVALID_INPUT"
	case ErrorCodeInvalidState:
		return "INVALID_STATE"
	case ErrorCodeNotFound:
		return "NOT_FOUND"
	case ErrorCodeAlreadyExists:
		return "ALREADY_EXISTS"
	case ErrorCodeConflict:
		return "CONFLICT"
	case ErrorCodeUnauthorized:
		return "UNAUTHORIZED"
	case ErrorCodeForbidden:
		return "FORBIDDEN"
	case ErrorCodeDenied:
		return "DENIED"
	case ErrorCodeConfig:
		return "CONFIG"
	case ErrorCodeProfile:
		return "PROFILE"
	case ErrorCodeSettings:
		return "SETTINGS"
	case ErrorCodeOperation:
		return "OPERATION"
	case ErrorCodeExecution:
		return "EXECUTION"
	case ErrorCodeRecovery:
		return "RECOVERY"
	case ErrorCodeSecurity:
		return "SECURITY"
	case ErrorCodeThreat:
		return "THREAT"
	case ErrorCodeBlock:
		return "BLOCK"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(ec))
	}
}

// IsValid checks if error code is valid
func (ec ErrorCode) IsValid() bool {
	return ec >= ErrorCodeUnknown && ec <= ErrorCodeBlock
}

// ErrorSeverity represents error severity
type ErrorSeverity int

const (
	ErrorSeverityInfo ErrorSeverity = iota
	ErrorSeverityWarning
	ErrorSeverityError
	ErrorSeverityCritical
)

// String returns string representation of error severity
func (es ErrorSeverity) String() string {
	switch es {
	case ErrorSeverityInfo:
		return "INFO"
	case ErrorSeverityWarning:
		return "WARNING"
	case ErrorSeverityError:
		return "ERROR"
	case ErrorSeverityCritical:
		return "CRITICAL"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(es))
	}
}

// IsValid checks if error severity is valid
func (es ErrorSeverity) IsValid() bool {
	return es >= ErrorSeverityInfo && es <= ErrorSeverityCritical
}

// DomainError represents domain-specific error
type DomainError struct {
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Severity  ErrorSeverity          `json:"severity"`
	Context   string                 `json:"context"`
	Details   map[string]interface{} `json:"details"`
	Cause     error                  `json:"cause"`
	Timestamp time.Time              `json:"timestamp"`
	Stack     string                 `json:"stack"`
	Retryable bool                   `json:"retryable"`
}

// Error implements error interface
func (de *DomainError) Error() string {
	if de.Context != "" {
		return fmt.Sprintf("[%s:%s] %s (%s)", de.Code, de.Severity, de.Message, de.Context)
	}
	return fmt.Sprintf("[%s:%s] %s", de.Code, de.Severity, de.Message)
}

// Unwrap returns underlying cause
func (de *DomainError) Unwrap() error {
	return de.Cause
}

// IsRetryable returns true if error is retryable
func (de *DomainError) IsRetryable() bool {
	return de.Retryable
}

// WithDetail adds detail to error
func (de *DomainError) WithDetail(key string, value interface{}) *DomainError {
	if de.Details == nil {
		de.Details = make(map[string]interface{})
	}
	de.Details[key] = value
	return de
}

// WithContext adds context to error
func (de *DomainError) WithContext(context string) *DomainError {
	de.Context = context
	return de
}

// WithCause adds cause to error
func (de *DomainError) WithCause(cause error) *DomainError {
	de.Cause = cause
	return de
}

// WithRetryable marks error as retryable
func (de *DomainError) WithRetryable(retryable bool) *DomainError {
	de.Retryable = retryable
	return de
}

// ErrorFactory creates domain errors
type ErrorFactory struct {
	context string
	details map[string]interface{}
}

// NewErrorFactory creates new error factory
func NewErrorFactory(context string) *ErrorFactory {
	return &ErrorFactory{
		context: context,
		details: make(map[string]interface{}),
	}
}

// WithDetail adds detail to factory
func (ef *ErrorFactory) WithDetail(key string, value interface{}) *ErrorFactory {
	ef.details[key] = value
	return ef
}

// WithContext updates factory context
func (ef *ErrorFactory) WithContext(context string) *ErrorFactory {
	ef.context = context
	return ef
}

// Unknown creates unknown error
func (ef *ErrorFactory) Unknown(message string) *DomainError {
	return ef.createError(ErrorCodeUnknown, message, ErrorSeverityError)
}

// Timeout creates timeout error
func (ef *ErrorFactory) Timeout(message string) *DomainError {
	return ef.createError(ErrorCodeTimeout, message, ErrorSeverityError)
}

// Cancelled creates cancelled error
func (ef *ErrorFactory) Cancelled(message string) *DomainError {
	return ef.createError(ErrorCodeCancelled, message, ErrorSeverityWarning)
}

// Retryable creates retryable error
func (ef *ErrorFactory) Retryable(message string) *DomainError {
	return ef.createError(ErrorCodeRetryable, message, ErrorSeverityWarning)
}

// Validation creates validation error
func (ef *ErrorFactory) Validation(message string) *DomainError {
	return ef.createError(ErrorCodeValidation, message, ErrorSeverityError)
}

// InvalidInput creates invalid input error
func (ef *ErrorFactory) InvalidInput(message string) *DomainError {
	return ef.createError(ErrorCodeInvalidInput, message, ErrorSeverityError)
}

// InvalidState creates invalid state error
func (ef *ErrorFactory) InvalidState(message string) *DomainError {
	return ef.createError(ErrorCodeInvalidState, message, ErrorSeverityCritical)
}

// NotFound creates not found error
func (ef *ErrorFactory) NotFound(message string) *DomainError {
	return ef.createError(ErrorCodeNotFound, message, ErrorSeverityError)
}

// AlreadyExists creates already exists error
func (ef *ErrorFactory) AlreadyExists(message string) *DomainError {
	return ef.createError(ErrorCodeAlreadyExists, message, ErrorSeverityWarning)
}

// Conflict creates conflict error
func (ef *ErrorFactory) Conflict(message string) *DomainError {
	return ef.createError(ErrorCodeConflict, message, ErrorSeverityError)
}

// Unauthorized creates unauthorized error
func (ef *ErrorFactory) Unauthorized(message string) *DomainError {
	return ef.createError(ErrorCodeUnauthorized, message, ErrorSeverityError)
}

// Forbidden creates forbidden error
func (ef *ErrorFactory) Forbidden(message string) *DomainError {
	return ef.createError(ErrorCodeForbidden, message, ErrorSeverityError)
}

// Denied creates denied error
func (ef *ErrorFactory) Denied(message string) *DomainError {
	return ef.createError(ErrorCodeDenied, message, ErrorSeverityError)
}

// Config creates config error
func (ef *ErrorFactory) Config(message string) *DomainError {
	return ef.createError(ErrorCodeConfig, message, ErrorSeverityError)
}

// Profile creates profile error
func (ef *ErrorFactory) Profile(message string) *DomainError {
	return ef.createError(ErrorCodeProfile, message, ErrorSeverityError)
}

// Settings creates settings error
func (ef *ErrorFactory) Settings(message string) *DomainError {
	return ef.createError(ErrorCodeSettings, message, ErrorSeverityError)
}

// Operation creates operation error
func (ef *ErrorFactory) Operation(message string) *DomainError {
	return ef.createError(ErrorCodeOperation, message, ErrorSeverityError)
}

// Execution creates execution error
func (ef *ErrorFactory) Execution(message string) *DomainError {
	return ef.createError(ErrorCodeExecution, message, ErrorSeverityCritical)
}

// Recovery creates recovery error
func (ef *ErrorFactory) Recovery(message string) *DomainError {
	return ef.createError(ErrorCodeRecovery, message, ErrorSeverityError)
}

// Security creates security error
func (ef *ErrorFactory) Security(message string) *DomainError {
	return ef.createError(ErrorCodeSecurity, message, ErrorSeverityCritical)
}

// Threat creates threat error
func (ef *ErrorFactory) Threat(message string) *DomainError {
	return ef.createError(ErrorCodeThreat, message, ErrorSeverityCritical)
}

// Block creates block error
func (ef *ErrorFactory) Block(message string) *DomainError {
	return ef.createError(ErrorCodeBlock, message, ErrorSeverityError)
}

// createError creates domain error with factory settings
func (ef *ErrorFactory) createError(code ErrorCode, message string, severity ErrorSeverity) *DomainError {
	error := &DomainError{
		Code:      code,
		Message:   message,
		Severity:  severity,
		Context:   ef.context,
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
		Retryable: code == ErrorCodeTimeout || code == ErrorCodeRetryable,
	}

	// Copy details from factory
	for k, v := range ef.details {
		error.Details[k] = v
	}

	// Add stack trace for critical errors
	if severity >= ErrorSeverityCritical {
		error.Stack = ef.captureStack()
	}

	return error
}

// captureStack captures current stack trace
func (ef *ErrorFactory) captureStack() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stack := strings.Split(string(buf[:n]), "\n")

	// Clean up stack trace
	var cleanStack []string
	for i, line := range stack {
		if i > 0 && i%2 == 0 {
			cleanStack = append(cleanStack, line)
		}
	}

	return strings.Join(cleanStack, "\n")
}

// Wrap wraps existing error with domain error
func (ef *ErrorFactory) Wrap(err error, code ErrorCode, message string) *DomainError {
	if err == nil {
		return nil
	}

	domainErr := ef.createError(code, message, ErrorSeverityError)
	domainErr.Cause = err

	// If wrapped error is already domain error, preserve its details
	if de, ok := err.(*DomainError); ok {
		for k, v := range de.Details {
			domainErr.Details[k] = v
		}
	}

	return domainErr
}

// ToResult converts error to result
func (de *DomainError) ToResult() result.Result[any] {
	return result.Err[any](de)
}

// ToResultWithValue converts error to result with type
func ToResultWithValue[T any](de *DomainError) result.Result[T] {
	return result.Err[T](de)
}

// ErrorHandler handles errors according to severity and type
type ErrorHandler struct {
	logger Logger
}

// Logger interface for error handling
type Logger interface {
	Info(message string, fields ...interface{})
	Warn(message string, fields ...interface{})
	Error(message string, fields ...interface{})
	Critical(message string, fields ...interface{})
}

// NewErrorHandler creates new error handler
func NewErrorHandler(logger Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// Handle handles error according to severity and type
func (eh *ErrorHandler) Handle(err error) error {
	if err == nil {
		return nil
	}

	de, ok := err.(*DomainError)
	if !ok {
		// Wrap non-domain errors
		factory := NewErrorFactory("unknown")
		de = factory.Unknown(err.Error())
	}

	// Log according to severity
	switch de.Severity {
	case ErrorSeverityInfo:
		eh.logger.Info(de.Error(), "code", de.Code, "context", de.Context)
	case ErrorSeverityWarning:
		eh.logger.Warn(de.Error(), "code", de.Code, "context", de.Context)
	case ErrorSeverityError:
		eh.logger.Error(de.Error(), "code", de.Code, "context", de.Context, "details", de.Details)
	case ErrorSeverityCritical:
		eh.logger.Critical(de.Error(), "code", de.Code, "context", de.Context, "details", de.Details, "stack", de.Stack)
	}

	return de
}

// IsRetryable checks if error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	if de, ok := err.(*DomainError); ok {
		return de.IsRetryable()
	}

	return false
}

// GetErrorCode returns error code from error
func GetErrorCode(err error) ErrorCode {
	if err == nil {
		return ErrorCodeUnknown
	}

	if de, ok := err.(*DomainError); ok {
		return de.Code
	}

	return ErrorCodeUnknown
}

// GetErrorSeverity returns error severity from error
func GetErrorSeverity(err error) ErrorSeverity {
	if err == nil {
		return ErrorSeverityInfo
	}

	if de, ok := err.(*DomainError); ok {
		return de.Severity
	}

	return ErrorSeverityError
}

// AddContext adds context to error
func AddContext(err error, context string) *DomainError {
	if err == nil {
		return nil
	}

	if de, ok := err.(*DomainError); ok {
		return de.WithContext(context)
	}

	factory := NewErrorFactory(context)
	return factory.Unknown(err.Error())
}

// AddDetail adds detail to error
func AddDetail(err error, key string, value interface{}) *DomainError {
	if err == nil {
		return nil
	}

	if de, ok := err.(*DomainError); ok {
		return de.WithDetail(key, value)
	}

	factory := NewErrorFactory("unknown")
	return factory.Unknown(err.Error()).WithDetail(key, value)
}

// CombineErrors combines multiple errors
func CombineErrors(errors ...error) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return errors[0]
	}

	factory := NewErrorFactory("combined")
	messages := make([]string, 0, len(errors))

	for _, err := range errors {
		if err != nil {
			messages = append(messages, err.Error())
		}
	}

	if len(messages) == 0 {
		return nil
	}

	return factory.Unknown(fmt.Sprintf("multiple errors: %s", strings.Join(messages, "; ")))
}
