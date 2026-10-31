package errorfamily

import (
	"errors"
	"fmt"
	"time"
)

// New creates a new Error with the given family, code, and message.
func New(family Family, code, message string) *Error {
	return &Error{
		code:       code,
		message:    message,
		family:     family,
		cause:      nil,
		context:    make(map[string]string),
		timestamp:  time.Now().UTC(),
		exitCode:   0,
		httpStatus: 0,
	}
}

// Newf creates a new Error with a formatted message.
func Newf(family Family, code, format string, args ...any) *Error {
	return New(family, code, fmt.Sprintf(format, args...))
}

// Wrap wraps an existing error with family, code, and message.
// Returns nil if err is nil.
func Wrap(err error, family Family, code, message string) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		code:       code,
		message:    message,
		family:     family,
		cause:      err,
		context:    make(map[string]string),
		timestamp:  time.Now().UTC(),
		exitCode:   0,
		httpStatus: 0,
	}
}

// Wrapf wraps an error with a formatted message.
// Returns nil if err is nil.
func Wrapf(err error, family Family, code, format string, args ...any) *Error {
	return Wrap(err, family, code, fmt.Sprintf(format, args...))
}

// Family-specific constructors.
// These make the error's behavioral classification explicit at the call site.

// NewRejection creates a Rejection error (bad input, unauthorized, not found).
func NewRejection(code, message string) *Error {
	return New(Rejection, code, message)
}

// NewConflict creates a Conflict error (version mismatch, duplicate).
func NewConflict(code, message string) *Error {
	return New(Conflict, code, message)
}

// NewTransient creates a Transient error (temporary failure, retryable).
func NewTransient(code, message string) *Error {
	return New(Transient, code, message)
}

// NewCorruption creates a Corruption error (data damaged, not self-healable).
func NewCorruption(code, message string) *Error {
	return New(Corruption, code, message)
}

// NewInfrastructure creates an Infrastructure error (system cannot serve).
func NewInfrastructure(code, message string) *Error {
	return New(Infrastructure, code, message)
}

// NewOrchestration creates an Orchestration error (internal coordination failure).
// Use this for bugs, misconfigurations, and internal logic failures that are
// neither I/O failures nor user input errors — e.g. "render status tree",
// "build CLI command", "wire dependencies". Not retryable. No user-facing fix.
func NewOrchestration(code, message string) *Error {
	return New(Orchestration, code, message)
}

// Wrap variants for each family.

// WrapRejection wraps an error as Rejection.
func WrapRejection(err error, code, message string) *Error {
	return Wrap(err, Rejection, code, message)
}

// WrapConflict wraps an error as Conflict.
func WrapConflict(err error, code, message string) *Error {
	return Wrap(err, Conflict, code, message)
}

// WrapTransient wraps an error as Transient (retryable).
func WrapTransient(err error, code, message string) *Error {
	return Wrap(err, Transient, code, message)
}

// WrapCorruption wraps an error as Corruption.
func WrapCorruption(err error, code, message string) *Error {
	return Wrap(err, Corruption, code, message)
}

// WrapInfrastructure wraps an error as Infrastructure.
func WrapInfrastructure(err error, code, message string) *Error {
	return Wrap(err, Infrastructure, code, message)
}

// WrapOrchestration wraps an error as Orchestration (internal coordination failure).
func WrapOrchestration(err error, code, message string) *Error {
	return Wrap(err, Orchestration, code, message)
}

// Formatted Wrap variants for each family.
// These mirror the Newf/Wrapf family: they accept a printf-style format string.

// WrapRejectionf wraps an error as Rejection with a formatted message.
func WrapRejectionf(err error, code, format string, args ...any) *Error {
	return Wrap(err, Rejection, code, fmt.Sprintf(format, args...))
}

// WrapConflictf wraps an error as Conflict with a formatted message.
func WrapConflictf(err error, code, format string, args ...any) *Error {
	return Wrap(err, Conflict, code, fmt.Sprintf(format, args...))
}

// WrapTransientf wraps an error as Transient (retryable) with a formatted message.
func WrapTransientf(err error, code, format string, args ...any) *Error {
	return Wrap(err, Transient, code, fmt.Sprintf(format, args...))
}

// WrapCorruptionf wraps an error as Corruption with a formatted message.
func WrapCorruptionf(err error, code, format string, args ...any) *Error {
	return Wrap(err, Corruption, code, fmt.Sprintf(format, args...))
}

// WrapInfrastructuref wraps an error as Infrastructure with a formatted message.
func WrapInfrastructuref(err error, code, format string, args ...any) *Error {
	return Wrap(err, Infrastructure, code, fmt.Sprintf(format, args...))
}

// WrapOrchestrationf wraps an error as Orchestration with a formatted message.
func WrapOrchestrationf(err error, code, format string, args ...any) *Error {
	return Wrap(err, Orchestration, code, fmt.Sprintf(format, args...))
}

// WrapOnce wraps an error only if it is not already a *Error.
// Use this at API boundaries where errors may have been classified by
// an inner layer. Prevents double-wrapping chains like:
//
//	[transient:db.timeout] outer: [transient:db.timeout] inner: cause
//
// If err is already a *Error, it is returned unchanged.
// If err is nil, returns nil (same nil-safety as [Wrap]).
func WrapOnce(err error, family Family, code, message string) *Error {
	if err == nil {
		return nil
	}

	if existing, ok := errors.AsType[*Error](err); ok {
		return existing
	}

	return Wrap(err, family, code, message)
}

// WrapOncef is the formatted-message variant of [WrapOnce].
// Wraps an error only if it is not already a *Error, using a printf-style
// format string for the message.
// Returns nil if err is nil. Returns the existing *Error unchanged if already classified.
func WrapOncef(err error, family Family, code, format string, args ...any) *Error {
	return WrapOnce(err, family, code, fmt.Sprintf(format, args...))
}
