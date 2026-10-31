package finding

import (
	"errors"
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
)

// Sentinel errors for use with errors.Is.
var (
	ErrValidation = errors.New("finding: validation error")
	ErrIO         = errors.New("finding: I/O error")
	ErrParse      = errors.New("finding: parse error")
	ErrConflict   = errors.New("finding: conflict error")
	ErrInternal   = errors.New("finding: internal error")
)

// ErrorCategory categorizes errors for programmatic handling.
type ErrorCategory string

const (
	// ErrCategoryValidation indicates validation errors.
	ErrCategoryValidation ErrorCategory = "validation"
	// ErrCategoryIO indicates file system or network errors.
	ErrCategoryIO ErrorCategory = "io"
	// ErrCategoryParse indicates parsing errors.
	ErrCategoryParse ErrorCategory = "parse"
	// ErrCategoryConflict indicates conflicting operations.
	ErrCategoryConflict ErrorCategory = "conflict"
	// ErrCategoryInternal indicates internal logic errors.
	ErrCategoryInternal ErrorCategory = "internal"
)

// IsValid returns true if the error category is a non-empty string matching
// the lowercase-hyphenated convention (e.g., "validation", "io").
func (c ErrorCategory) IsValid() bool {
	return isValidLowercaseHyphen(string(c))
}

// FindingError provides structured error information with context.
//
//nolint:revive // stuttering name is intentional for clarity
type FindingError struct {
	Category ErrorCategory // Category of error
	Finding  *Finding      // Associated finding (may be nil)
	Message  string        // Human-readable message
	Cause    error         // Underlying cause (may be nil)
	File     FilePath      // File path (if applicable)
	Position *Position     // Position in file (if applicable)
}

// Error implements the error interface.
func (e *FindingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Category, e.Message, e.Cause)
	}

	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

// Unwrap returns the underlying cause for error inspection.
func (e *FindingError) Unwrap() error {
	return e.Cause
}

// Is supports errors.Is by matching sentinel errors.
func (e *FindingError) Is(target error) bool {
	switch e.Category {
	case ErrCategoryValidation:
		return target == ErrValidation
	case ErrCategoryIO:
		return target == ErrIO
	case ErrCategoryParse:
		return target == ErrParse
	case ErrCategoryConflict:
		return target == ErrConflict
	case ErrCategoryInternal:
		return target == ErrInternal
	default:
		return false
	}
}

// ErrorCode implements errorfamily.Coded.
func (e *FindingError) ErrorCode() string {
	return "finding." + string(e.Category)
}

// ErrorFamily implements errorfamily.Classified, mapping ErrorCategory to
// go-error-family families for integration with Classify().
func (e *FindingError) ErrorFamily() errorfamily.Family {
	switch e.Category {
	case ErrCategoryValidation, ErrCategoryParse:
		return errorfamily.Rejection
	case ErrCategoryConflict:
		return errorfamily.Conflict
	case ErrCategoryIO:
		return errorfamily.Transient
	case ErrCategoryInternal:
		return errorfamily.Infrastructure
	default:
		return errorfamily.Transient
	}
}

// WithFinding sets the finding on a copy of the FindingError and returns it.
func (e *FindingError) WithFinding(f Finding) *FindingError {
	clone := *e
	clone.Finding = &f
	clone.File = f.Position.File
	clone.Position = &f.Position

	return &clone
}

// WithPosition sets the position on a copy of the FindingError and returns it.
func (e *FindingError) WithPosition(pos Position) *FindingError {
	clone := *e
	clone.Position = &pos
	clone.File = pos.File

	return &clone
}

// newCategorizedError creates a FindingError with the given category.
func newCategorizedError(cat ErrorCategory, message string, cause error) *FindingError {
	return &FindingError{
		Category: cat,
		Message:  message,
		Cause:    cause,
	}
}

// NewValidationError creates a validation error.
func NewValidationError(message string, cause error) *FindingError {
	return newCategorizedError(ErrCategoryValidation, message, cause)
}

// NewIOError creates an IO error.
func NewIOError(message string, cause error) *FindingError {
	return newCategorizedError(ErrCategoryIO, message, cause)
}

// NewParseError creates a parse error.
func NewParseError(message string, cause error) *FindingError {
	return newCategorizedError(ErrCategoryParse, message, cause)
}

// NewConflictError creates a conflict error.
func NewConflictError(message string, cause error) *FindingError {
	return newCategorizedError(ErrCategoryConflict, message, cause)
}

// NewInternalError creates an internal error.
func NewInternalError(message string, cause error) *FindingError {
	return newCategorizedError(ErrCategoryInternal, message, cause)
}

// IsFindingError returns true if err is a *FindingError.
func IsFindingError(err error) bool {
	_, ok := errors.AsType[*FindingError](
		err,
	) //nolint:erraudit // ok-pattern type assertion; first return is a value, not an ignored error

	return ok
}

// CategoryOf returns the category of the error, or empty string if not a FindingError.
func CategoryOf(err error) ErrorCategory {
	if findingErr, ok := errors.AsType[*FindingError](err); ok {
		return findingErr.Category
	}

	return ""
}

// IsCategory returns true if err is a FindingError with the given category.
func IsCategory(err error, cat ErrorCategory) bool {
	return CategoryOf(err) == cat
}

// must returns v or panics if err is non-nil. Used by Must-style constructors
// (MustParseCategory, MustParseSeverity, Builder.MustBuild) to eliminate the
// repetitive `if err != nil { panic(err) }` boilerplate.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
